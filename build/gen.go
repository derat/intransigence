// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package build

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

	fileMode = 0644 // mode for creating files
	dirMode  = 0755 // mode for creating dirs

	indent    = "  "
	wrapWidth = 120
)

// generatePages renders non-AMP and AMP versions of all normal pages and writes them
// to the appropriate subdirectory under out. The generated files' paths are returned.
func generatePages(si *render.SiteInfo, out string, pretty bool) ([]string, error) {
	ps, err := filepath.Glob(filepath.Join(si.PageDir(), "*.md"))
	if err != nil {
		return nil, fmt.Errorf("failed to enumerate pages: %v", err)
	}

	defer clearStatus()
	var outPaths []string
	for i, p := range ps {
		statusf("Generating pages: [%d/%d]", i, len(ps))
		md, err := ioutil.ReadFile(p)
		if err != nil {
			return nil, err
		}

		base := filepath.Base(p)
		base = base[:len(base)-len(".md")]

		build := func(dest string, amp bool) error {
			outPaths = append(outPaths, dest)
			b, err := render.Page(*si, md, amp)
			if err != nil {
				return fmt.Errorf("failed to render %s: %v", filepath.Base(dest), err)
			}
			if pretty {
				if b, err = prettyPrintDoc(bytes.NewReader(b)); err != nil {
					return fmt.Errorf("failed to pretty-print %s: %v", filepath.Base(dest), err)
				}
			}
			return ioutil.WriteFile(dest, b, fileMode)
		}

		if err := build(filepath.Join(out, base+render.HTMLExt), false /* amp */); err != nil {
			return nil, err
		}
		if err := build(filepath.Join(out, base+render.AMPExt), true /* amp */); err != nil {
			return nil, err
		}
	}
	return outPaths, nil
}

// generateIframes renders all iframe pages and writes them to the appropriate subdirectory under out.
// The generated files' paths are returned.
func generateIframes(si *render.SiteInfo, out string, pretty bool) ([]string, error) {
	ps, err := filepath.Glob(filepath.Join(si.IframeDir(), "*.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to enumerate iframe data: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(out, iframeSubdir), dirMode); err != nil {
		return nil, err
	}
	var outPaths []string
	for _, p := range ps {
		data, err := ioutil.ReadFile(p)
		if err != nil {
			return nil, err
		}

		base := filepath.Base(p)
		base = base[:len(base)-len(".json")]
		dest := filepath.Join(out, iframeSubdir, base+render.HTMLExt)
		outPaths = append(outPaths, dest)

		b, err := render.Iframe(*si, data)
		if err != nil {
			return nil, fmt.Errorf("failed to render iframe %q: %v", base, err)
		}
		if pretty {
			if b, err = prettyPrintDoc(bytes.NewReader(b)); err != nil {
				return nil, fmt.Errorf("failed to pretty-print iframe %s: %v", filepath.Base(dest), err)
			}
		}
		if err := ioutil.WriteFile(dest, b, fileMode); err != nil {
			return nil, err
		}
	}
	return outPaths, nil
}

// prettyPrintDoc reads an HTML5 document from r and returns a buffer containing a pretty-printed
// copy of it.
func prettyPrintDoc(r io.Reader) ([]byte, error) {
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

// generateCSS runs sassc to generate minified CSS files from all .scss files within dir.
// The .scss extension is replaced with .css.min.
func generateCSS(dir string) error {
	ps, err := filepath.Glob(filepath.Join(dir, "*.scss"))
	if err != nil {
		return err
	}

	defer clearStatus()
	for i, p := range ps {
		statusf("Generating CSS: [%d/%d]", i, len(ps))
		dp := p[:len(p)-4] + "css" + minSuffix
		if err := exec.Command("sassc", "--style", "compressed", p, dp).Run(); err != nil {
			return err
		}
	}
	return nil
}
