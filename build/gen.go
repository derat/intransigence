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
	"sort"
	"time"

	"golang.org/x/net/html"

	"github.com/derat/htmlpretty"
	"github.com/derat/intransigence/render"
)

const (
	fileMode = 0644 // mode for creating files
	dirMode  = 0755 // mode for creating dirs

	indent    = "  "
	wrapWidth = 120
)

// generatePages renders non-AMP and AMP versions of all normal pages and writes them
// to the appropriate subdirectory under out. The generated files' paths are returned.
// The returned feed info structs are sorted newest-to-oldest.
func generatePages(si *render.SiteInfo, out string, pretty bool,
	exeTime time.Time) ([]string, []render.PageFeedInfo, error) {
	ps, err := filepath.Glob(filepath.Join(si.PageDir(), "*.md"))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to enumerate pages: %v", err)
	}

	defer clearStatus()
	var outPaths []string
	var feedInfos []render.PageFeedInfo
	for i, p := range ps {
		statusf("Generating pages: [%d/%d]", i, len(ps))
		md, err := ioutil.ReadFile(p)
		if err != nil {
			return nil, nil, err
		}
		pi, err := os.Stat(p)
		if err != nil {
			return nil, nil, err
		}

		base := filepath.Base(p)
		base = base[:len(base)-len(".md")]

		build := func(dest string, amp bool) error {
			outPaths = append(outPaths, dest)
			b, fi, err := render.Page(*si, base, md, amp)
			if err != nil {
				return fmt.Errorf("failed to render %s: %v", filepath.Base(dest), err)
			}
			if fi != nil && !amp {
				feedInfos = append(feedInfos, *fi)
			}
			if pretty {
				if b, err = prettyPrintDoc(bytes.NewReader(b)); err != nil {
					return fmt.Errorf("failed to pretty-print %s: %v", filepath.Base(dest), err)
				}
			}
			if err := ioutil.WriteFile(dest, b, fileMode); err != nil {
				return err
			}
			// Copy the Markdown file's mtime and atime.
			return os.Chtimes(dest, maxTime(getAtime(pi), exeTime), maxTime(pi.ModTime(), exeTime))
		}

		if err := build(filepath.Join(out, base+render.HTMLExt), false /* amp */); err != nil {
			return nil, nil, err
		}
		if err := build(filepath.Join(out, base+render.AMPExt), true /* amp */); err != nil {
			return nil, nil, err
		}
	}

	sort.Slice(feedInfos, func(i, j int) bool {
		return feedInfos[i].Created.After(feedInfos[j].Created)
	})

	return outPaths, feedInfos, nil
}

// generateIframes renders all iframe pages and writes them to the appropriate subdirectory under out.
// The generated files' paths are returned.
func generateIframes(si *render.SiteInfo, out string, pretty bool,
	exeTime time.Time) ([]string, error) {
	ps, err := filepath.Glob(filepath.Join(si.IframeDir(), "*.yaml"))
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
		fi, err := os.Stat(p)
		if err != nil {
			return nil, err
		}

		base := filepath.Base(p)
		base = base[:len(base)-len(".yaml")]
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
		// Copy the data file's mtime and atime.
		if err := os.Chtimes(dest, maxTime(getAtime(fi), exeTime), maxTime(fi.ModTime(), exeTime)); err != nil {
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

// generateCSS runs sassc to generate minified CSS of all .scss files in src.
// The files are written under dst with the .scss extension changed to .css.
func generateCSS(src, dst string) error {
	ps, err := filepath.Glob(filepath.Join(src, "*.scss"))
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	defer clearStatus()
	for i, p := range ps {
		statusf("Generating CSS: [%d/%d]", i, len(ps))
		base := filepath.Base(p)
		dp := filepath.Join(dst, base[:len(base)-4]+"css")
		if err := exec.Command("sassc", "--style", "compressed", p, dp).Run(); err != nil {
			return fmt.Errorf("failed running sassc on %v: %v", p, err)
		}
		if err := copyTimes(p, dp); err != nil {
			return err
		}
	}
	return nil
}

// generateWebP runs cwebp to generate WebP versions of all GIF, JPEG, and PNG images under src.
// The files are written under dst and are regenerated if stale.
func generateWebP(src, dst string) error {
	// Get mtimes for orig images and WebP files. Keys are paths relative to src and dst.
	imgTimes := make(map[string]time.Time)
	if err := filepath.Walk(src, func(p string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.Mode()&os.ModeType != 0 {
			return nil
		}
		switch filepath.Ext(p) {
		case ".gif", ".jpg", ".jpeg", ".png":
			imgTimes[p[len(src)+1:]] = fi.ModTime()
		}
		return nil
	}); err != nil {
		return err
	}

	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}
	webPTimes := make(map[string]time.Time)
	if err := filepath.Walk(dst, func(p string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.Mode()&os.ModeType == 0 && filepath.Ext(p) == render.WebPExt {
			webPTimes[p[len(dst)+1:]] = fi.ModTime()
		}
		return nil
	}); err != nil {
		return err
	}

	todo := make(map[string]string) // orig image path -> WebP path
	for ip, it := range imgTimes {
		wp := ip[:len(ip)-len(filepath.Ext(ip))] + render.WebPExt
		if wt, ok := webPTimes[wp]; !ok || wt.Before(it) {
			todo[filepath.Join(src, ip)] = filepath.Join(dst, wp)
		}
		delete(webPTimes, wp)
	}

	num := 0
	defer clearStatus()
	for ip, wp := range todo {
		statusf("Generating WebP images: [%d/%d]", num, len(todo))
		if err := os.MkdirAll(filepath.Dir(wp), 0755); err != nil {
			return err
		}
		switch filepath.Ext(ip) {
		case ".gif":
			if err := exec.Command("gif2webp", ip, "-o", wp).Run(); err != nil {
				return fmt.Errorf("failed running gif2webp on %v: %v", ip, err)
			}
		default:
			// https://chromium.googlesource.com/webm/libwebp/+/refs/heads/0.4.1/src/enc/config.c#52
			preset := "text"
			if ext := filepath.Ext(ip); ext == ".jpg" || ext == ".jpeg" {
				preset = "photo"
			}
			if err := exec.Command("cwebp", "-preset", preset, ip, "-o", wp).Run(); err != nil {
				return fmt.Errorf("failed running cwebp on %v: %v", ip, err)
			}
		}
		if err := copyTimes(ip, wp); err != nil {
			return err
		}
		num++
	}

	// Delete any WebP files corresponding to no-longer-present images.
	for p := range webPTimes {
		if err := os.Remove(filepath.Join(dst, p)); err != nil {
			return err
		}
	}

	return nil
}
