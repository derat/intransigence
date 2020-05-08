// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"golang.org/x/net/html"

	"github.com/derat/homepage/pretty"
	"github.com/derat/homepage/render"
)

const (
	siteFile     = "site.yaml" // YAML file containing render.SiteInfo data
	outSubdir    = "out"       // subdir under site dir for generated site
	tmpOutPrefix = ".out."     // prefix passed to ioutil.TempDir for temp output dir
	oldOutSubdir = ".out.old"  // subdir under site dir containing previous generated site
	iframeSubdir = "iframes"   // subdir under output dir for generated iframe pages
	fileMode     = 0644        // mode for creating files
	dirMode      = 0755        // mode for creating dirs
	indent       = "  "
	wrapWidth    = 120
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
		die(1, "Failed to get working dir: ", err)
	}
	flag.StringVar(&dir, "dir", dir, "Site directory (defaults to working dir)")
	out := flag.String("out", "", "Destination directory (site is built under -dir if empty)")
	pprint := flag.Bool("pretty", true, "Pretty-print HTML")
	flag.Parse()

	// TODO: Permit this once everything works.
	if *out == "" {
		die(2, "-out must be explicitly specified")
	}

	if err := buildSite(dir, *out, *pprint); err != nil {
		die(1, "Failed to build site: ", err)
	}
}

// buildSite builds the site rooted at dir into the directory named by out.
// If out is empty, the site is built into outSubdir under the site directory.
func buildSite(dir, out string, pprint bool) error {
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

	if err := buildPages(si, out, pprint); err != nil {
		return err
	}
	if err := buildIframes(si, out, pprint); err != nil {
		return err
	}

	// If we built into the site dir, rename the temp dir that we used.
	if buildToSiteDir {
		dest := filepath.Join(dir, outSubdir)
		// Rename the existing out dir after deleting the old backup if present.
		if _, err := os.Stat(dest); err == nil {
			old := filepath.Join(dir, oldOutSubdir)
			if err := os.RemoveAll(old); err != nil {
				return err
			}
			if err := os.Rename(dest, old); err != nil {
				return err
			}
		}
		if err := os.Rename(out, dest); err != nil {
			return err
		}
	}
	return nil
}

// buildPages renders non-AMP and AMP versions of all normal pages and writes them
// to the appropriate subdirectory under out.
func buildPages(si *render.SiteInfo, out string, pprint bool) error {
	ps, err := filepath.Glob(filepath.Join(si.PageDir(), "*.md"))
	if err != nil {
		return fmt.Errorf("failed to enumerate pages: %v", err)
	}
	for _, p := range ps {
		md, err := ioutil.ReadFile(p)
		if err != nil {
			return err
		}

		base := filepath.Base(p)
		base = base[:len(base)-len(".md")]

		build := func(dest string, amp bool) error {
			b, err := render.Page(*si, md, amp)
			if err != nil {
				return fmt.Errorf("failed to render %s: %v", filepath.Base(dest), err)
			}
			if pprint {
				if b, err = prettyPrint(bytes.NewReader(b)); err != nil {
					return fmt.Errorf("failed to pretty-print %s: %v", filepath.Base(dest), err)
				}
			}
			return ioutil.WriteFile(dest, b, fileMode)
		}

		if err := build(filepath.Join(out, base+render.HTMLExt), false /* amp */); err != nil {
			return err
		}
		if err := build(filepath.Join(out, base+render.AMPExt), true /* amp */); err != nil {
			return err
		}
	}
	return nil
}

// buildIframe renders all iframe pages and writes them to the appropriate subdirectory under out.
func buildIframes(si *render.SiteInfo, out string, pprint bool) error {
	ps, err := filepath.Glob(filepath.Join(si.IframeDir(), "*.json"))
	if err != nil {
		return fmt.Errorf("failed to enumerate iframe data: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(out, iframeSubdir), dirMode); err != nil {
		return err
	}
	for _, p := range ps {
		data, err := ioutil.ReadFile(p)
		if err != nil {
			return err
		}

		base := filepath.Base(p)
		base = base[:len(base)-len(".json")]
		dest := filepath.Join(out, iframeSubdir, base+render.HTMLExt)

		b, err := render.Iframe(*si, data)
		if err != nil {
			return fmt.Errorf("failed to render iframe %q: %v", base, err)
		}
		if pprint {
			if b, err = prettyPrint(bytes.NewReader(b)); err != nil {
				return fmt.Errorf("failed to pretty-print iframe %s: %v", filepath.Base(dest), err)
			}
		}
		if err := ioutil.WriteFile(dest, b, fileMode); err != nil {
			return err
		}
	}
	return nil
}

// prettyPrint reads an HTML5 document from r and returns a buffer containing a pretty-printed
// copy of it.
func prettyPrint(r io.Reader) ([]byte, error) {
	node, err := html.Parse(r)
	if err != nil {
		return nil, err
	}
	var b bytes.Buffer
	if err := pretty.Print(&b, node, indent, wrapWidth); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}
