// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package render

import (
	"errors"
	"fmt"
	"html/template"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	_ "golang.org/x/image/webp"
)

const (
	thumbnailSize = 4      // width/height in pixels for image thumbnails
	svgExt        = ".svg" // extension for SVG images
)

// imgInfo holds information used by img.tmpl.
type imgInfo struct {
	Path   string `html:"path" yaml:"path"`     // path, e.g. "files/img.png" or "files/img-*.png"
	URL    string `html:"url" yaml:"url"`       // URL, e.g. "data:image/gif;base64,..."
	Width  int    `html:"width" yaml:"width"`   // 100% width in pixels; inferred if empty
	Height int    `html:"height" yaml:"height"` // 100% height in pixels; inferred if empty
	Alt    string `html:"alt" yaml:"alt"`       // alt text
	Lazy   bool   `html:"lazy" yaml:"lazy"`     // whether image should be lazy-loaded

	// These fields are set programatically, mostly by finishImgInfo.
	ID      string              // DOM ID for image
	Classes []string            // CSS classes (can be modified before/after finishImgInfo)
	Attr    []template.HTMLAttr // additional attrs to include (can be modified before/after finishImgInfo)

	Src, Srcset                 string // attr values for preferred image
	FallbackSrc, FallbackSrcset string // attr values for fallback image (if any)

	ThumbSrc          template.URL // attr value for thumbnail placeholder image (if any)
	DefineThumbFilter bool         // true if #thumb-filter SVG filter should be defined

	Sizes      string // 'sizes' attr value (set by finishImgInfo but can be modified after)
	biggestSrc string // highest-res version of image (set by finishImgInfo)
	widths     []int  // ascending widths in pixels of images if multi-res (set by finishImgInfo)
	layout     string // AMP layout (consumed by finishImgInfo; "responsive" used if empty)
	noThumb    bool   // avoid generating a thumbnail (consumed by finishImgInfo)
}

// finish validates info and fills additional fields.
// amp should be true if the image will be used for an AMP page.
// didThumb is used to track whether an image with ThumbSrc has been finished.
func (info *imgInfo) finish(si *SiteInfo, amp bool, didThumb *bool) error {
	if (info.Path == "" && info.URL == "") || (info.Path != "" && info.URL != "") {
		return errors.New("exactly one of path or url must be set")
	}
	if info.Alt == "" {
		return errors.New("alt must be set")
	}
	if amp {
		if info.layout == "" {
			info.layout = "responsive"
		}
		info.Attr = append(info.Attr, template.HTMLAttr(fmt.Sprintf(`layout="%s"`, info.layout)))
	}
	if info.Lazy && !amp { // <amp-img> already lazy-loads
		info.Attr = append(info.Attr, template.HTMLAttr(`loading="lazy"`))
	}

	if info.URL != "" {
		if info.Width == 0 || info.Height == 0 {
			return errors.New("width and height must be set for URLs")
		}
		info.Src = info.URL
		return nil
	}

	if wc := strings.IndexByte(info.Path, '*'); wc == -1 {
		// There's no wildcard, so there's just one size.
		if strings.HasSuffix(info.Path, WebPExt) || strings.HasSuffix(info.Path, svgExt) {
			info.Src = info.Path
		} else {
			info.Src = removeExt(info.Path) + WebPExt
			info.FallbackSrc = info.Path
		}
		info.biggestSrc = info.Path

		// If the image's display dimensions weren't supplied, get them from the file.
		if info.Width <= 0 || info.Height <= 0 {
			var err error
			if info.Width, info.Height, err = imageSize(filepath.Join(si.StaticDir(), info.Path)); err != nil {
				return fmt.Errorf("failed getting %v size: %v", info.Path, err)
			}
		}
		info.Srcset = fmt.Sprintf("%s %dw", info.Src, info.Width)
		if info.FallbackSrc != "" {
			info.FallbackSrcset = fmt.Sprintf("%s %dw", info.FallbackSrc, info.Width)
		}
	} else {
		// There's a wildcard, so we have multiple sizes.
		pre := info.Path[:wc]
		suf := info.Path[wc+1:]
		var err error
		var srcset string
		if srcset, info.widths, err = makeSrcset(si.StaticDir(), pre, suf); err != nil {
			return err
		} else if srcset == "" {
			if srcset, info.widths, err = makeSrcset(si.StaticGenDir(), pre, suf); err != nil {
				return err
			}
		}
		if srcset == "" {
			return fmt.Errorf("no images matched by prefix %q and suffix %q", pre, suf)
		}

		// If the image's display dimensions weren't supplied, get them from the files.
		if info.Width <= 0 || info.Height <= 0 {
			var p string
			if info.Width <= 0 && len(info.widths) >= 2 {
				// If there are 1x and 2x images, use the dimensions of the 1x image.
				for _, w := range info.widths[1:] {
					if w == 2*info.widths[0] {
						p = fmt.Sprintf("%s%d%s", pre, info.widths[0], suf)
					}
				}
			} else if info.Width > 0 {
				// If the width was supplied, use that file's height.
				p = fmt.Sprintf("%s%d%s", pre, info.Width, suf)
			}
			if p == "" {
				return errors.New("dimensions could not be determined")
			}
			if info.Width, info.Height, err = imageSize(filepath.Join(si.StaticDir(), p)); err != nil {
				return fmt.Errorf("failed getting %v dimensions: %v", p, err)
			}
		}

		src := fmt.Sprintf("%s%d%s", pre, info.Width, suf)
		info.biggestSrc = fmt.Sprintf("%s%d%s", pre, info.widths[len(info.widths)-1], suf)

		if strings.HasSuffix(info.Path, WebPExt) || strings.HasSuffix(info.Path, svgExt) {
			// If this was already a set of WebP images or a vector SVG image, use it directly.
			info.Src, info.Srcset = src, srcset
		} else {
			// Otherwise, make a WebP srcset and use the original images as a fallback.
			info.Src = removeExt(src) + WebPExt
			wsuf := removeExt(suf) + WebPExt
			if info.Srcset, _, err = makeSrcset(si.StaticDir(), pre, wsuf); err != nil {
				return err
			} else if info.Srcset == "" {
				if info.Srcset, _, err = makeSrcset(si.StaticGenDir(), pre, wsuf); err != nil {
					return err
				}
			}
			if info.Srcset == "" {
				return fmt.Errorf("no images matched by prefix %q and suffix %q", pre, wsuf)
			}
			info.FallbackSrc = src
			info.FallbackSrcset = srcset
		}
	}

	if info.Sizes == "" {
		info.Sizes = fmt.Sprintf("%dpx", info.Width)
	}

	if err := si.CheckStatic(info.Src); err != nil {
		return err
	}
	if info.FallbackSrc != "" {
		if err := si.CheckStatic(info.FallbackSrc); err != nil {
			return err
		}
	}
	if err := si.CheckStatic(info.biggestSrc); err != nil {
		return err
	}

	// Generate inline thumbnail.
	if !info.noThumb && !strings.HasSuffix(info.Src, svgExt) {
		origSrc := info.Src
		if info.FallbackSrc != "" {
			origSrc = info.FallbackSrc
		}
		// Ignore "webp: invalid format" errors that the webp package seems to return when passed
		// animated images.
		if thumb, err := genThumb(filepath.Join(si.StaticDir(), origSrc),
			thumbnailSize, thumbnailSize); err == nil {
			info.ThumbSrc = template.URL("data:image/gif;base64," + thumb)
		} else if err.Error() != "webp: invalid format" {
			return fmt.Errorf("failed generating thumbnail for %v: %v", origSrc, err)
		}
		// Define the SVG filter the first time we create a thumbnail.
		if !*didThumb && !amp {
			info.DefineThumbFilter = true
		}
		*didThumb = true
	}

	return nil
}

// imageSize returns the dimensions of the image at p.
func imageSize(p string) (w, h int, err error) {
	f, err := os.Open(p)
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()

	cfg, _, err := image.DecodeConfig(f)
	return cfg.Width, cfg.Height, err
}

// makeSrcset returns a srcset attribute value corresponding to the
// images matched by pre and suf under the supplied static dir.
// The returned slice contains image widths in ascending order.
// If no files are matched, an empty string is returned.
func makeSrcset(dir, pre, suf string) (string, []int, error) {
	glob := filepath.Join(dir, pre+"*"+suf)
	ps, err := filepath.Glob(glob)
	if err != nil {
		return "", nil, err
	}

	// Ascending order by embedded image width.
	sort.Slice(ps, func(i, j int) bool {
		if len(ps[i]) < len(ps[j]) {
			return true
		} else if len(ps[i]) > len(ps[j]) {
			return false
		}
		return ps[i] < ps[j]
	})

	var srcs []string
	var widths []int
	preLen := len(filepath.Join(dir, pre))
	for _, p := range ps {
		width, err := strconv.Atoi(p[preLen : len(p)-len(suf)])
		if err != nil {
			return "", nil, err
		}
		widths = append(widths, width)
		srcs = append(srcs, fmt.Sprintf("%s%d%s %dw", pre, width, suf, width))
	}
	return strings.Join(srcs, ", "), widths, nil
}
