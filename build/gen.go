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
	"time"

	"golang.org/x/net/html"

	"github.com/derat/htmlpretty"
	"github.com/derat/plaingen/render"
)

const (
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
	if err := os.MkdirAll(filepath.Join(out, render.IframeOutDir), dirMode); err != nil {
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
		dest := filepath.Join(out, render.IframeOutDir, base+render.HTMLExt)
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
	// TODO: Also minify *.css?
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

// generateWebP runs cwebp to generate WebP versions of all GIF, JPEG, and PNG images under dir.
// Existing files are regenerated if needed.
func generateWebP(dir string) error {
	// Get mtimes for orig images and WebP files.
	imgTimes := make(map[string]time.Time)
	webPTimes := make(map[string]time.Time)
	if err := filepath.Walk(dir, func(p string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.Mode()&os.ModeType != 0 {
			return nil
		}
		switch filepath.Ext(p) {
		case ".gif", ".jpg", ".jpeg", ".png":
			imgTimes[p] = fi.ModTime()
		case render.WebPExt:
			webPTimes[p] = fi.ModTime()
		}
		return nil
	}); err != nil {
		return err
	}

	todo := make(map[string]string) // orig image path -> WebP path
	for ip, it := range imgTimes {
		wp := ip[:len(ip)-len(filepath.Ext(ip))] + render.WebPExt
		if wt, ok := webPTimes[wp]; !ok || wt.Before(it) {
			todo[ip] = wp
		}
	}

	num := 0
	defer clearStatus()
	for ip, wp := range todo {
		statusf("Generating WebP images: [%d/%d]", num, len(todo))
		switch filepath.Ext(ip) {
		case ".gif":
			if err := exec.Command("gif2webp", ip, "-o", wp).Run(); err != nil {
				return err
			}
		default:
			// https://chromium.googlesource.com/webm/libwebp/+/refs/heads/0.4.1/src/enc/config.c#52
			preset := "text"
			if ext := filepath.Ext(ip); ext == ".jpg" || ext == ".jpeg" {
				preset = "photo"
			}
			if err := exec.Command("cwebp", "-preset", preset, ip, "-o", wp).Run(); err != nil {
				return err
			}
		}
		num++
	}
	return nil
}
