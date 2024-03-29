// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package build

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/derat/intransigence/render"
	"github.com/otiai10/copy"
)

const (
	siteFile     = "site.yaml"     // YAML file containing render.SiteInfo data
	outSubdir    = "out"           // subdir under site dir for generated site
	tmpOutPrefix = ".out."         // prefix passed to ioutil.TempDir for temp output dir
	oldOutSubdir = ".out.old"      // subdir under site dir containing previous generated site
	sitemapFile  = "sitemap.xml"   // generated XML file containing sitemap data
	exeName      = "intransigence" // executable filename
)

// Flags specifies details of how the site should be built.
type Flags int

const (
	// PrettyPrint indicates that HTML output should be pretty-printed.
	PrettyPrint Flags = 1 << iota
	// Display a diff of changes and prompt before replacing the existing output dir.
	// Only has an effect when Build's out argument is empty.
	Prompt
	// Serve indicates that the new output dir should be served over HTTP while the diff is displayed.
	// Only has an effect when Prompt is true and Build's out argument is empty.
	Serve
	// Validate indicates that HTML and CSS output should be validated.
	Validate
)

// Build builds the site rooted at dir into the directory named by out.
// If out is empty, the site is built into outSubdir under the site directory.
func Build(ctx context.Context, dir, out string, flags Flags) error {
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
		// Make the temp directory world-readable in case amphtml-validator runs as another user.
		if err := os.Chmod(out, 0755); err != nil {
			return err
		}
		buildToSiteDir = true
		defer os.RemoveAll(out) // clean up temp dir on failure
	}

	// Generate inline CSS files before they get included in pages and iframes.
	// Also create WebP images so that their existence can be checked when generating pages.
	if err := generateCSS(si.InlineDir(), si.InlineGenDir()); err != nil {
		return err
	}
	if err := generateWebP(si.StaticDir(), si.StaticGenDir()); err != nil {
		return err
	}

	exeTime := getExeTime()
	var genPaths []string
	var feedInfos []render.PageFeedInfo
	if genPaths, feedInfos, err = generatePages(si, out, flags&PrettyPrint != 0, exeTime); err != nil {
		return err
	}
	if ps, err := generateIframes(si, out, flags&PrettyPrint != 0, exeTime); err != nil {
		return err
	} else {
		genPaths = append(genPaths, ps...)
	}

	// Only validate generated files -- we don't want to fail on issues in static files.
	if flags&Validate != 0 {
		if err := validateFiles(ctx, genPaths); err != nil {
			return err
		}
	}

	// Copy over static content.
	copyOpts := copy.Options{PreserveTimes: true}
	if err := copy.Copy(si.StaticDir(), out, copyOpts); err != nil {
		return err
	}
	for src, dst := range si.ExtraStaticDirs {
		dp := filepath.Clean(filepath.Join(out, dst))
		if !strings.HasPrefix(dp, out+"/") {
			return fmt.Errorf("dest path %v escapes out dir", dp)
		}
		if err := copy.Copy(filepath.Join(dir, src), dp, copyOpts); err != nil {
			return err
		}
	}
	// Also copy over generated WebP images.
	gd := si.StaticGenDir()
	if err := filepath.Walk(gd, func(p string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.Mode()&os.ModeType == 0 && filepath.Ext(p) == render.WebPExt {
			return copy.Copy(p, filepath.Join(out, p[len(gd)+1:]), copyOpts)
		}
		return nil
	}); err != nil {
		return err
	}

	// TODO: Try to preserve sitemap and feed timestamps somehow?
	if err := writeSitemap(filepath.Join(out, sitemapFile), si); err != nil {
		return fmt.Errorf("sitemap failed: %v", err)
	}
	if err := writeFeed(filepath.Join(out, render.FeedFile), si, feedInfos); err != nil {
		return fmt.Errorf("feed failed: %v", err)
	}

	// Create gzipped versions of text-based files for the HTTP server to use.
	if si.CompressPages {
		if err := compressDir(out); err != nil {
			return err
		}
	}

	// Update directory timestamps to improve rsync performance.
	if _, _, err := updateDirTimes(out); err != nil {
		return fmt.Errorf("updating dir times failed: %v", err)
	}

	// If we built into the site dir, rename the temp dir that we used.
	if buildToSiteDir {
		dest := filepath.Join(dir, outSubdir)
		if _, err := os.Stat(dest); err != nil && !os.IsNotExist(err) {
			return err
		} else if err == nil {
			// Show a diff and confirm that we should replace the existing output dir.
			if flags&Prompt != 0 {
				if ok, err := prompt(ctx, dest, out, flags&Serve != 0); err != nil {
					return fmt.Errorf("failed displaying diff: %v", err)
				} else if !ok {
					return errors.New("diff rejected")
				}
			}
			// Delete the old backup if present and back up the existing output dir.
			old := filepath.Join(dir, oldOutSubdir)
			if err := os.RemoveAll(old); err != nil {
				return err
			}
			if err := os.Rename(dest, old); err != nil {
				return err
			}
		}

		// Move the temp dir to the final location.
		if err := os.Rename(out, dest); err != nil {
			return err
		}
	}
	return nil
}

// getExeTime returns the intransigence executable's mtime.
// The zero time is returned during testing or if an error is encountered.
func getExeTime() time.Time {
	exe, err := os.Executable()
	if err != nil || filepath.Base(exe) != exeName {
		return time.Time{}
	}
	fi, err := os.Stat(exe)
	if err != nil {
		return time.Time{}
	}
	return fi.ModTime()
}
