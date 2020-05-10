// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/derat/homepage/render"
	"github.com/otiai10/copy"
)

const (
	siteFile     = "site.yaml" // YAML file containing render.SiteInfo data
	outSubdir    = "out"       // subdir under site dir for generated site
	tmpOutPrefix = ".out."     // prefix passed to ioutil.TempDir for temp output dir
	oldOutSubdir = ".out.old"  // subdir under site dir containing previous generated site
)

func die(code int, args ...interface{}) {
	fmt.Fprintln(os.Stderr, args...)
	os.Exit(code)
}

func dief(code int, format string, args ...interface{}) {
	die(code, fmt.Sprintf(format, args...))
}

func main() {
	dir, err := os.Getwd()
	if err != nil {
		die(1, "Failed to get working dir:", err)
	}
	flag.StringVar(&dir, "dir", dir, "Site directory (defaults to working dir)")
	out := flag.String("out", "", "Destination directory (site is built under -dir if empty)")
	pretty := flag.Bool("pretty", true, "Pretty-print HTML")
	validate := flag.Bool("validate", true, "Validate generated files")
	flag.Parse()

	// TODO: Permit this once everything works.
	if *out == "" {
		die(2, "-out must be explicitly specified")
	}

	var flags buildFlags
	if *pretty {
		flags |= prettyPrint
	}
	if *validate {
		flags |= validatePages
	}
	if err := buildSite(context.Background(), dir, *out, flags); err != nil {
		die(1, "Failed to build site:", err)
	}
}

type buildFlags int

const (
	prettyPrint buildFlags = 1 << iota
	validatePages
)

// buildSite builds the site rooted at dir into the directory named by out.
// If out is empty, the site is built into outSubdir under the site directory.
func buildSite(ctx context.Context, dir, out string, flags buildFlags) error {
	si, err := render.NewSiteInfo(filepath.Join(dir, siteFile))
	if err != nil {
		return fmt.Errorf("failed to load site info: %v", err)
	}

	// If an output directory wasn't specified, create a temp dir within the site dir to build into.
	buildToSiteDir := false
	if out == "" {
		var err error
		if out, err = ioutil.TempDir(dir, tmpOutPrefix); err != nil {
			return fmt.Errorf("failed to create temp output dir: %v", err)
		}
		buildToSiteDir = true
		defer os.RemoveAll(out) // clean up temp dir on failure
	}

	// Minify inline files before they get included in pages and iframes.
	if err := minifyInline(si.InlineDir()); err != nil {
		return err
	}
	if err := generatePages(si, out, flags&prettyPrint != 0); err != nil {
		return err
	}
	if err := generateIframes(si, out, flags&prettyPrint != 0); err != nil {
		return err
	}
	if flags&validatePages != 0 {
		if err := validateDir(ctx, out); err != nil {
			return err
		}
	}

	// Copy over static content after finishing with validation.
	if err := copy.Copy(si.StaticDir(), out); err != nil {
		return err
	}
	for src, dst := range si.ExtraStaticDirs {
		dp := filepath.Clean(filepath.Join(out, dst))
		if !strings.HasPrefix(dp, out+"/") {
			return fmt.Errorf("dest path %v escapes out dir", dp)
		}
		if err := copy.Copy(filepath.Join(dir, src), dp); err != nil {
			return err
		}
	}

	// TODO: Generate sitemap.

	// Create gzipped versions of text-based files for the HTTP server to use.
	if err := compressDir(out); err != nil {
		return err
	}

	// If we built into the site dir, rename the temp dir that we used.
	if buildToSiteDir {
		dest := filepath.Join(dir, outSubdir)
		if _, err := os.Stat(dest); err == nil {
			// Copy the existing output dir's atimes and mtimes for files that didn't change.
			if err := copyFileTimes(dest, out); err != nil {
				return err
			}
			// Delete the old backup if present and back up the existing output dir.
			old := filepath.Join(dir, oldOutSubdir)
			if err := os.RemoveAll(old); err != nil {
				return err
			}
			if err := os.Rename(dest, old); err != nil {
				return err
			}
		} else if !os.IsNotExist(err) {
			return err
		}
		if err := os.Rename(out, dest); err != nil {
			return err
		}
	}
	return nil
}
