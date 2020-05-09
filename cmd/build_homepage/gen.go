// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"golang.org/x/net/html"

	"github.com/derat/homepage/render"
	"github.com/derat/htmlpretty"
)

const (
	iframeSubdir = "iframes" // subdir under output dir for generated iframe pages
	minSuffix    = ".min"    // suffix for minified CSS and JS files

	fileMode = 0644 // mode for creating files
	dirMode  = 0755 // mode for creating dirs

	indent    = "  "
	wrapWidth = 120
)

// generatePages renders non-AMP and AMP versions of all normal pages and writes them
// to the appropriate subdirectory under out.
func generatePages(si *render.SiteInfo, out string, pretty bool) error {
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
			if pretty {
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

// generateIframes renders all iframe pages and writes them to the appropriate subdirectory under out.
func generateIframes(si *render.SiteInfo, out string, pretty bool) error {
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
		if pretty {
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
	if err := htmlpretty.Print(&b, node, indent, wrapWidth); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// minifyInline updates the minified versions of CSS or JS files within dir if needed.
// It seems like it'd be cleaner to write the minified files into a temporary dir instead
// of alongside the source files, but yui-compressor is slow (1 second for a 9 KB JS file!)
// so it seems better to skip unnecessary calls.
func minifyInline(dir string) error {
	// filepath.Glob doesn't appear to support globs like "*.{css,js}".
	ps, err := filepath.Glob(filepath.Join(dir, "*.css"))
	if err != nil {
		return err
	} else if js, err := filepath.Glob(filepath.Join(dir, "*.js")); err != nil {
		return err
	} else {
		ps = append(ps, js...)
	}

	for _, p := range ps {
		pi, err := os.Stat(p)
		if err != nil {
			return err
		}
		mp := p + minSuffix
		mi, err := os.Stat(mp)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		// Minify if the minimized version doesn't exist or has an mtime before the original file's.
		if err != nil || mi.ModTime().Before(pi.ModTime()) {
			if err := exec.Command("yui-compressor", "-o", mp, p).Run(); err != nil {
				return err
			}
		}
	}

	return nil
}
