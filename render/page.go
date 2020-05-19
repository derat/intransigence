// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package render

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"

	// Importing this as gopkg.in/russross/blackfriday.v2 doesn't work since the
	// package's go.mod currently contains the github.com URL instead:
	// https://github.com/russross/blackfriday/issues/500
	md "github.com/russross/blackfriday/v2"

	yaml "gopkg.in/yaml.v3"
)

const (
	mobileMaxWidth  = 640
	desktopMinWidth = mobileMaxWidth + 1

	// WebPExt is the extension for generated WebP image files.
	WebPExt = ".webp"
)

// Page renders and returns the page described by the supplied Markdown data.
// The amp parameter specifies whether the AMP or non-AMP version of the page should be rendered.
func Page(si SiteInfo, markdown []byte, amp bool) ([]byte, error) {
	r := newRenderer(si, amp)
	b := md.Run(markdown, md.WithRenderer(r), md.WithExtensions(md.CommonExtensions&^md.Autolink))
	if r.err != nil {
		return nil, r.err
	}
	return b, nil
}

// pageInfo holds information about the current page that is used by page.tmpl.
type pageInfo struct {
	Title           string `yaml:"title"`             // used in <title> element
	ID              string `yaml:"id"`                // ID to highlight in navbox
	Desc            string `yaml:"desc"`              // meta description
	ImgURL          string `yaml:"img_url"`           // structured data image URL (must be 696 pixels wide)
	ImgWidth        int    `yaml:"img_width"`         // structured data image width
	ImgHeight       int    `yaml:"img_height"`        // structured data image width
	Created         string `yaml:"created"`           // creation date as 'YYYY-MM-DD'
	Modified        string `yaml:"modified"`          // last-modified date as 'YYYY-MM-DD'
	HideTitleSuffix bool   `yaml:"hide_title_suffix"` // don't append SiteInfo.TitleSuffix
	HideBackToTop   bool   `yaml:"hide_back_to_top"`  // hide footer link to jump to top
	HideDates       bool   `yaml:"hide_dates"`        // hide footer created andmodified dates
	HasMap          bool   `yaml:"has_map"`           // page contains a map
	HasGraph        bool   `yaml:"has_graph"`         // page contains one or more maps

	NavText    string  `yaml:"-"` // text to display next to logo in nav area
	LogoHTML   imgInfo `yaml:"-"` // header logo for non-AMP
	LogoAMP    imgInfo `yaml:"-"` // header logo for AMP
	NavToggle  imgInfo `yaml:"-"` // nav toggle icon for non-AMP
	MenuButton imgInfo `yaml:"-"` // menu button for AMP

	NavItem *NavItem `yaml:"-"` // nav item corresponding to current page

	GoogleAnalyticsCode string `yaml:"-"`

	LinkRel  string `yaml:"-"` // rel attribute for <link>, e.g. "canonical"
	LinkHref string `yaml:"-"` // href attribute for <link>

	HTMLStyle        template.CSS  `yaml:"-"`
	HTMLScripts      []template.JS `yaml:"-"`
	AMPStyle         template.CSS  `yaml:"-"`
	AMPNoscriptStyle template.CSS  `yaml:"-"`
	AMPCustomStyle   template.CSS  `yaml:"-"`
	CSPMeta          template.HTML `yaml:"-"`
	StructData       structData    `yaml:"-"`
}

// imgInfo holds information used by img.tmpl.
type imgInfo struct {
	Path   string `html:"path" yaml:"path"`     // path, e.g. "files/img.png" or "files/img-*.png"
	Width  int    `html:"width" yaml:"width"`   // 100% width in pixels; inferred if empty
	Height int    `html:"height" yaml:"height"` // 100% height in pixels; inferred if empty
	Alt    string `html:"alt" yaml:"alt"`       // alt text
	Lazy   bool   `html:"lazy" yaml:"lazy"`     // whether image should be lazy-loaded

	// These fields are set programatically, mostly by finishImgInfo.
	Attr               []template.HTMLAttr // additional attributes to include (can be modified before/after finishImgInfo)
	Src, WebPSrc       string              // 'src' attr values for original and WebP images (set by finishImgInfo)
	Srcset, WebPSrcset string              // 'srcset' attr values for original and WebP images (set by finishImgInfo)
	Sizes              string              // 'sizes' attr value (set by finishImgInfo but can be modified after)
	biggestSrc         string              // highest-res version of Src (set by finishImgInfo)
	widths             []int               // ascending widths in pixels of images if multi-res (set by finishImgInfo)
	layout             string              // AMP layout (consumed by finishImgInfo; "responsive" used if empty)
}

// figureInfo holds information used by figure.tmpl.
type figureInfo struct {
	Align   string        `yaml:"align"`   // left, right, center, desktop_left, desktop_right, desktop_alt
	Caption template.HTML `yaml:"caption"` // <figcaption> text; template.HTML to permit links and not escape quotes
	Class   string        `yaml:"class"`   // CSS class
}

type renderer struct {
	si   SiteInfo
	pi   pageInfo
	tmpl *templater
	hr   *md.HTMLRenderer
	err  error // error encountered during rendering
	amp  bool  // rendering an AMP page

	startingBox bool         // in the middle of a level-1 header
	boxTitle    bytes.Buffer // text seen while startingBox is true
	inBox       bool         // rendering a box

	lastFigureAlign string // last "align" value used for a figure
	numMapMarkers   int    // number of boxes with "map_marker"
}

func newRenderer(si SiteInfo, amp bool) *renderer {
	r := renderer{
		si: si,
		pi: pageInfo{
			NavText:             si.NavText,
			Desc:                si.DefaultDesc,
			GoogleAnalyticsCode: si.GoogleAnalyticsCode,
		},
		hr:  md.NewHTMLRenderer(md.HTMLRendererParameters{}),
		amp: amp,
	}
	r.tmpl = newTemplater(filepath.Join(si.TemplateDir()), template.FuncMap{
		"amp": func() bool {
			return r.amp
		},
		"currentID": func() string {
			return r.pi.NavItem.ID
		},
		"formatDate": func(date, layout string) string {
			t, err := time.Parse("2006-01-02", date)
			if err != nil {
				r.setErrorf("failed to parse date %q: %v", date, err)
				return ""
			}
			return t.Format(layout)
		},
		"topNavItems": func() []*NavItem {
			return r.si.NavItems
		},
	})

	return &r
}

// setError saves err to r.err if it isn't already set. err is returned.
// Use this instead of setting r.err directly to avoid overwriting an earlier error.
func (r *renderer) setError(err error) error {
	if r.err == nil {
		r.err = err
	}
	return err
}

// setErrorf constructs a new error and saves it to r.err if it isn't already set.
// The new error is returned.
// Use this instead of setting r.err directly to avoid overwriting an earlier error.
func (r *renderer) setErrorf(format string, args ...interface{}) error {
	err := fmt.Errorf(format, args...)
	if r.err == nil {
		r.err = err
	}
	return err
}

func (r *renderer) RenderNode(w io.Writer, node *md.Node, entering bool) md.WalkStatus {
	// Bail if we already set an error in RenderHeader.
	if r.err != nil {
		return md.Terminate
	}

	// Capture box title text.
	if node.Type != md.Heading && r.startingBox {
		return r.hr.RenderNode(&r.boxTitle, node, entering)
	}

	switch node.Type {
	case md.Code:
		return r.mangleOutput(w, node, entering, unescapeQuotes)
	case md.CodeBlock:
		return r.renderCodeBlock(w, node, entering)
	case md.Heading:
		return r.renderHeading(w, node, entering)
	case md.HTMLSpan:
		return r.renderHTMLSpan(w, node, entering)
	case md.Link:
		// Make sure that we don't rewrite the link a second time when exiting.
		if entering {
			link, err := r.rewriteLink(string(node.LinkData.Destination))
			if err != nil {
				r.setError(err)
				return md.Terminate
			}
			node.LinkData.Destination = []byte(link)
		}
		// Fall through and let Blackfriday render the possibly-updated URL as normal.
	case md.Text:
		return r.mangleOutput(w, node, entering, unescapeQuotes|escapeEmdashes)
	}

	// Fall back to Blackfriday's default HTML rendering.
	return r.hr.RenderNode(w, node, entering)
}

func (r *renderer) RenderHeader(w io.Writer, ast *md.Node) {
	if r.err != nil {
		return
	}

	fc := ast.FirstChild
	if fc == nil || fc.Type != md.CodeBlock || string(fc.CodeBlockData.Info) != "page" {
		r.setErrorf(`page doesn't start with "page" code block`)
		return
	}
	if err := yaml.Unmarshal(fc.Literal, &r.pi); err != nil {
		r.setErrorf("failed to parse page info from %q: %v", fc.Literal, err)
		return
	}

	// Add a fake nav item for the index page if it's current.
	// The Name and URL fields don't need to be set since this item is never rendered.
	if r.pi.ID == indexID {
		r.pi.NavItem = &NavItem{ID: indexID, Children: r.si.NavItems}
	} else {
		for _, n := range r.si.NavItems {
			if r.pi.NavItem = n.FindID(r.pi.ID); r.pi.NavItem != nil {
				break
			}
		}
		if r.pi.NavItem == nil {
			r.setErrorf("no page with ID %q", r.pi.ID)
			return
		}
	}

	r.pi.LogoHTML = imgInfo{
		Path: r.si.LogoPathHTML,
		Alt:  r.si.LogoAlt,
		Attr: []template.HTMLAttr{template.HTMLAttr(`id="nav-logo"`)},
	}
	if err := r.finishImgInfo(&r.pi.LogoHTML); err != nil {
		r.setErrorf("logo failed: %v", err)
		return
	}
	// If the second-smallest size isn't a 2x version of the smallest size, use it on mobile.
	// This is a hack to avoid needing to hardcode image sizes in code.
	if widths := r.pi.LogoHTML.widths; len(widths) >= 2 {
		if widths[1] != 2*widths[0] {
			r.pi.LogoHTML.Sizes = fmt.Sprintf("(max-width: %dpx) %dpx, %s",
				mobileMaxWidth, widths[1], r.pi.LogoHTML.Sizes)
		}
	}

	r.pi.LogoAMP = imgInfo{
		Path: r.si.LogoPathAMP,
		Alt:  r.si.LogoAlt,
	}
	if err := r.finishImgInfo(&r.pi.LogoAMP); err != nil {
		r.setErrorf("AMP logo failed: %v", err)
		return
	}

	// On mobile, collapse the navbox if the page doesn't have subpages.
	r.pi.NavToggle = imgInfo{
		Path: r.si.NavTogglePath,
		Alt:  "[toggle navigation]",
		Attr: []template.HTMLAttr{template.HTMLAttr(`id="nav-toggle-img"`)},
	}
	if len(r.pi.NavItem.Children) == 0 {
		r.pi.NavToggle.Attr = append(r.pi.NavToggle.Attr, template.HTMLAttr(`class="expand"`))
	}
	if err := r.finishImgInfo(&r.pi.NavToggle); err != nil {
		r.setErrorf("nav toggle failed: %v", err)
		return
	}

	r.pi.MenuButton = imgInfo{
		Path: r.si.MenuButtonPath,
		Alt:  "[toggle menu]",
		Attr: []template.HTMLAttr{
			template.HTMLAttr(`id="menu-button"`),
			template.HTMLAttr(`tabindex="0"`),
			template.HTMLAttr(`role="button"`),
			template.HTMLAttr(`on="tap:sidebar.open"`),
		},
	}
	if err := r.finishImgInfo(&r.pi.MenuButton); err != nil {
		r.setErrorf("menu button failed: %v", err)
		return
	}

	// It would be much simpler to just use a map[string]interface{} for this,
	// but the properties are marshaled in an arbitrary order then, making it
	// hard to compare the output against template_lib.rb.
	r.pi.StructData = structData{
		Context:       "http://schema.org",
		Type:          "Article",
		Headline:      r.pi.Title,
		DatePublished: r.pi.Created,
		Author: structDataAuthor{
			Type:  "Person",
			Name:  r.si.AuthorName,
			Email: r.si.AuthorEmail,
		},
		Publisher: structDataPublisher{
			Type: "Organization",
			Name: r.si.PublisherName,
			URL:  r.si.BaseURL,
			Logo: structDataImage{
				Type:   "ImageObject",
				URL:    r.si.PublisherLogoURL,
				Width:  r.si.PublisherLogoWidth,
				Height: r.si.PublisherLogoHeight,
			},
		},
	}
	var err error
	if r.pi.StructData.MainEntityOfPage, err = r.si.AbsURL(r.pi.NavItem.URL); err != nil {
		r.setError(err)
		return
	}
	if r.pi.Desc != "" && r.pi.Desc != r.si.DefaultDesc {
		r.pi.StructData.Description = r.pi.Desc
	}
	if r.pi.Modified != "" {
		r.pi.StructData.DateModified = r.pi.Modified
	}
	if r.pi.ImgURL != "" && r.pi.ImgWidth > 0 && r.pi.ImgHeight > 0 {
		r.pi.StructData.Image = &structDataImage{
			Type:   "ImageObject",
			Width:  r.pi.ImgWidth,
			Height: r.pi.ImgHeight,
		}
		if err := r.si.CheckStatic(r.pi.ImgURL); err != nil {
			r.setError(err)
			return
		}
		if r.pi.StructData.Image.URL, err = r.si.AbsURL(r.pi.ImgURL); err != nil {
			r.setError(err)
			return
		}
	}

	// Do this here so it's not included in structured data.
	if !r.pi.HideTitleSuffix {
		r.pi.Title += r.si.TitleSuffix
	}

	if r.amp {
		r.pi.LinkRel = "canonical"
		if r.pi.LinkHref, err = r.si.AbsURL(r.pi.NavItem.URL); err != nil {
			r.setError(err)
			return
		}

		r.pi.AMPStyle = template.CSS(r.si.ReadInline("amp-boilerplate.css"))
		r.pi.AMPNoscriptStyle = template.CSS(r.si.ReadInline("amp-boilerplate-noscript.css"))
		r.pi.AMPCustomStyle = template.CSS(
			r.si.ReadInline("base.css.min") +
				r.si.ReadInline("mobile.css.min") +
				r.si.ReadInline("mobile-amp.css.min"))

		// TODO: It looks like AMP runs
		// https://raw.githubusercontent.com/ampproject/amphtml/1476486609642/src/style-installer.js,
		// which tries to apply inline styles. How is this supposed to work?
		// https://github.com/ampproject/amphtml/blob/master/spec/amp-html-format.md
		// says "AMP HTML documents must not trigger errors when served with a Content
		// Security Policy that does not include the keywords unsafe-inline and
		// unsafe-eval. The AMP HTML format is designed so that is always the case." On
		// top of that, the validator appears to complain "The attribute 'http-equiv'
		// may not appear in tag 'meta name= and content='." in response to CSP <meta>
		// tags. Just omit CSP in AMP pages for now.
		// Revisit this, maybe:
		// https://amp.dev/documentation/guides-and-tutorials/optimize-and-measure/secure-pages/
	} else {
		r.pi.LinkRel = "amphtml"
		if r.pi.LinkHref, err = r.si.AbsURL(r.pi.NavItem.AMPURL()); err != nil {
			r.setError(err)
			return
		}

		r.pi.HTMLStyle = template.CSS(r.si.ReadInline("base.css.min") + r.si.ReadInline("base-nonamp.css.min") +
			fmt.Sprintf("@media(min-width:%dpx){%s}", desktopMinWidth, r.si.ReadInline("desktop.css.min")) +
			fmt.Sprintf("@media(max-width:%dpx){%s}", mobileMaxWidth, r.si.ReadInline("mobile.css.min")))
		r.pi.HTMLScripts = []template.JS{template.JS(r.si.ReadInline("base.js.min"))}
		if r.pi.HasMap {
			r.pi.HTMLScripts = append(r.pi.HTMLScripts, template.JS(r.si.ReadInline("map.js.min")))
		}

		csp := cspBuilder{}
		csp.add(cspDefault, cspNone)
		csp.add(cspChild, cspSelf)
		csp.add(cspImg, cspSelf)

		csp.hash(cspStyle, string(r.pi.HTMLStyle))
		for _, s := range r.pi.HTMLScripts {
			csp.hash(cspScript, string(s))
		}
		r.pi.CSPMeta = template.HTML(csp.tag())
	}

	r.setError(r.tmpl.runNamed(w, []string{"page.tmpl", "img.tmpl"}, "start", &r.pi, nil))
}

func (r *renderer) RenderFooter(w io.Writer, ast *md.Node) {
	// Bail if we set an error in RenderHeader or RenderNode.
	if r.err != nil {
		return
	}
	if r.inBox {
		if r.setError(r.tmpl.runNamed(w, []string{"box.tmpl"}, "end", nil, nil)) != nil {
			return
		}
	}
	r.setError(r.tmpl.runNamed(w, []string{"page.tmpl"}, "end", &r.pi, nil))
}

// Renders a node of type md.CodeBlock and returns the appropriate walk status.
// Sets r.err and returns md.Terminate if an error is encountered.
func (r *renderer) renderCodeBlock(w io.Writer, node *md.Node, entering bool) md.WalkStatus {
	// Rewrites the supplied iframe URL (e.g. iframes/my_map.html) to be either absolute or site-rooted.
	// The framed page isn't AMP-compliant, so it won't be served by the AMP cache: https://www.erat.org/amp.html#iframes
	// Absolute URLs could presumably also be used in the non-AMP case, but it makes development harder.
	iframeHref := func(s string) string {
		if r.amp {
			return r.si.BaseURL + s
		}
		return "/" + s
	}

	// Rewrites the supplied figure "align" value to handle "desktop_alt". Also updates lastFigureAlign.
	figureAlign := func(s string) string {
		v := func() string {
			if s != "desktop_alt" {
				return s
			}
			if r.lastFigureAlign != "desktop_left" {
				return "desktop_left"
			}
			return "desktop_right"
		}()
		r.lastFigureAlign = v
		return v
	}

	switch string(node.CodeBlockData.Info) {
	case "clear":
		if r.setError(r.tmpl.run(w, []string{"clear.tmpl"}, nil, nil)) != nil {
			return md.Terminate
		}
		return md.SkipChildren
	case "graph":
		var info struct {
			figureInfo `yaml:",inline"`
			Href       string `yaml:"href"`   // relative path to graph iframe page
			Name       string `yaml:"name"`   // graph data name
			Width      int    `yaml:"width"`  // graph width (without border)
			Height     int    `yaml:"height"` // graph height (without border)
		}
		if err := yaml.Unmarshal(node.Literal, &info); err != nil {
			r.setErrorf("failed to parse graph info from %q: %v", node.Literal, err)
			return md.Terminate
		}
		info.figureInfo.Align = figureAlign(info.figureInfo.Align)
		info.Href = iframeHref(info.Href)
		if r.setError(r.tmpl.run(w, []string{"graph.tmpl", "figure.tmpl"}, info, nil)) != nil {
			return md.Terminate
		}
		return md.SkipChildren
	case "image":
		var info struct {
			figureInfo `yaml:",inline"`
			imgInfo    `yaml:",inline"`
			Href       string `yaml:"href"`
		}
		if err := yaml.Unmarshal(node.Literal, &info); err != nil {
			r.setErrorf("failed to parse image info from %q: %v", node.Literal, err)
			return md.Terminate
		}
		if err := r.finishImgInfo(&info.imgInfo); err != nil {
			r.setErrorf("bad data in %q: %v", node.Literal, err)
			return md.Terminate
		}
		info.figureInfo.Align = figureAlign(info.figureInfo.Align)
		if len(info.Href) == 0 && info.imgInfo.biggestSrc != info.imgInfo.Src {
			info.Href = info.imgInfo.biggestSrc
		}
		if len(info.Href) > 0 {
			var err error
			if info.Href, err = r.rewriteLink(info.Href); err != nil {
				r.setError(err)
				return md.Terminate
			}
		}
		if r.setError(r.tmpl.run(w, []string{"image_block.tmpl", "figure.tmpl", "img.tmpl"}, info, nil)) != nil {
			return md.Terminate
		}
		return md.SkipChildren
	case "map":
		var info struct {
			imgInfo `yaml:",inline"` // placeholder image (also used for dimensions)
			Href    string           `yaml:"href"` // relative path to map iframe page
		}
		if err := yaml.Unmarshal(node.Literal, &info); err != nil {
			r.setErrorf("failed to parse map info from %q: %v", node.Literal, err)
			return md.Terminate
		}
		info.imgInfo.layout = "fill"
		info.imgInfo.Attr = append(info.imgInfo.Attr, template.HTMLAttr("placeholder"))
		info.imgInfo.Alt = "[map placeholder]"
		if err := r.finishImgInfo(&info.imgInfo); err != nil {
			r.setErrorf("bad data in %q: %v", node.Literal, err)
			return md.Terminate
		}
		info.Href = iframeHref(info.Href)
		if r.setError(r.tmpl.run(w, []string{"map.tmpl", "img.tmpl"}, info, nil)) != nil {
			return md.Terminate
		}
		return md.SkipChildren
	case "page":
		return md.SkipChildren // handled in RenderHeader
	default:
		node.CodeBlockData.Info = nil // prevent Blackfriday from adding e.g. "language-html" CSS class
		return r.mangleOutput(w, node, entering, unescapeQuotes|removeCodeNewline)
	}
}

// Renders a node of type md.Heading and returns the appropriate walk status.
// Sets r.err and returns md.Terminate if an error is encountered.
func (r *renderer) renderHeading(w io.Writer, node *md.Node, entering bool) md.WalkStatus {
	// Let Blackfriday render everything apart from level-1 headings, which we hijack toplevel headings to render boxes.
	if node.HeadingData.Level != 1 {
		return r.hr.RenderNode(w, node, entering)
	}

	// The contents of the heading appear as text between when we enter and
	// leaving the heading node, and we grab it literally to pass to the
	// template. The box is rendered when we leave the heading node.
	if entering {
		if r.inBox {
			if r.setError(r.tmpl.runNamed(w, []string{"box.tmpl"}, "end", nil, nil)) != nil {
				return md.Terminate
			}
			r.inBox = false
		}
		r.startingBox = true
		r.boxTitle.Reset()
		return md.GoToNext
	}

	// Additional attributes can be passed as slash-separated values in the
	// ID string, e.g. "# Heading #{myid/desktop_only/narrow}".
	var info = struct {
		ID          string // value for id attribute
		Title       template.HTML
		DesktopOnly bool   // only display on desktop
		MobileOnly  bool   // only display on mobile
		Narrow      bool   // make the box narrow
		MapLabel    string // letter label for map marker
	}{Title: template.HTML(r.boxTitle.String())}
	for i, v := range strings.Split(node.HeadingData.HeadingID, "/") {
		switch {
		case i == 0:
			info.ID = v
		case v == "desktop_only":
			info.DesktopOnly = true
		case v == "mobile_only":
			info.MobileOnly = true
		case v == "narrow":
			info.Narrow = true
		case v == "map_marker":
			info.MapLabel = string('A' + r.numMapMarkers)
			r.numMapMarkers++
		}
	}

	r.startingBox = false
	r.inBox = true
	if r.setError(r.tmpl.runNamed(w, []string{"box.tmpl"}, "start", &info, nil)) != nil {
		return md.Terminate
	}
	return md.GoToNext
}

// Renders a node of type md.HTMLSpan and returns the appropriate walk status.
// Sets r.err and returns md.Terminate if an error is encountered.
func (r *renderer) renderHTMLSpan(w io.Writer, node *md.Node, entering bool) md.WalkStatus {
	// Wrap all the logic in a function to make it easy to consolidate error-handling.
	// md.Terminate is always returned if an error occurs.
	ws, err := func() (md.WalkStatus, error) {
		token, err := parseTag(node.Literal)
		if err != nil {
			return 0, err
		}

		switch token.Data {
		case "code-url":
			if token.Type == html.StartTagToken {
				io.WriteString(w, `<code class="url">`)
			} else if token.Type == html.EndTagToken {
				io.WriteString(w, "</code>")
			}
			return md.SkipChildren, nil
		case "image":
			if token.Type == html.StartTagToken {
				var info = imgInfo{
					Attr:   []template.HTMLAttr{template.HTMLAttr(`class="inline"`)},
					layout: "fixed",
				}
				if err := unmarshalAttrs(token.Attr, &info); err != nil {
					return 0, err
				}
				if err := r.finishImgInfo(&info); err != nil {
					return 0, err
				}
				if err := r.tmpl.runNamed(w, []string{"img.tmpl"}, "img", info, nil); err != nil {
					return 0, err
				}
			}
			return md.SkipChildren, nil
		case "only-amp":
			if !r.amp {
				if token.Type == html.StartTagToken {
					io.WriteString(w, "<!-- AMP-only content \n")
				} else if token.Type == html.EndTagToken {
					io.WriteString(w, "\n-->")
				}
			}
			return md.GoToNext, nil // process nested nodes
		case "only-nonamp":
			if r.amp {
				if token.Type == html.StartTagToken {
					io.WriteString(w, "<!-- non-AMP-only content \n")
				} else if token.Type == html.EndTagToken {
					io.WriteString(w, "\n-->")
				}
			}
			return md.GoToNext, nil // process nested nodes
		case "text-size":
			if token.Type == html.StartTagToken {
				var info struct {
					Small bool `html:"small"`
					Tiny  bool `html:"tiny"`
				}
				if err := unmarshalAttrs(token.Attr, &info); err != nil {
					return 0, err
				}
				var class string
				if info.Small {
					class = "small"
				} else if info.Tiny {
					class = "real-small"
				} else {
					return 0, errors.New("missing attribute")
				}
				fmt.Fprintf(w, `<span class="%s">`, class)
			} else if token.Type == html.EndTagToken {
				io.WriteString(w, "</span>")
			}
			return md.GoToNext, nil // process nested content
		default:
			return 0, errors.New("unsupported tag")
		}
	}()
	if err != nil {
		r.setErrorf("rendering HTML span %q failed: %v", node.Literal, err)
		return md.Terminate
	}
	return ws
}

func (r *renderer) rewriteLink(link string) (string, error) {
	// Absolute links don't need to be rewritten.
	if u, err := url.Parse(link); err != nil {
		return "", fmt.Errorf("parsing link %q failed: %v", link, err)
	} else if u.IsAbs() {
		return link, nil
	}
	// Also use bare fragments as-is.
	if link[0] == '#' {
		return link, nil
	}

	// Check if we were explicitly asked to return an AMP or non-AMP page.
	var forceAMP, forceNonAMP bool
	if tl := strings.TrimSuffix(link, "!force_amp"); tl != link {
		forceAMP = true
		link = tl
	} else if tl := strings.TrimSuffix(link, "!force_nonamp"); tl != link {
		forceNonAMP = true
		link = tl
	}

	// TODO: Revisit this decision.
	if link[0] == '/' {
		return "", fmt.Errorf("link %q shouldn't have leading slash", link)
	}

	// Make sure that links to static resources work.
	if !isPage(link) {
		if err := r.si.CheckStatic(link); err != nil {
			return "", err
		}
	}

	// Non-AMP links don't need to be rewritten.
	if (!r.amp && !forceAMP) || forceNonAMP {
		return link, nil
	}

	// If this isn't a regular page, it may not be served by the AMP CDN. Use an absolute URL.
	if !isPage(link) {
		return r.si.BaseURL + link, nil
	}

	// Use the AMP version of the page instead.
	return ampPage(link), nil
}

// mangleOps describes operations that mangleOutput should perform.
type mangleOps int

const (
	unescapeQuotes    mangleOps = 1 << iota // replaces '&quot;' with '"'
	escapeEmdashes                          // replaces '—' with '&mdash;'
	removeCodeNewline                       // removes '\n' before trailing '</code></pre>'
)

// mangleOutput is a helper function that uses Blackfriday's standard
// HTMLRenderer to render the supplied node, and then performs the requested
// operations on the output before writing it to w.
func (r *renderer) mangleOutput(w io.Writer, node *md.Node, entering bool, ops mangleOps) md.WalkStatus {
	var b bytes.Buffer
	ret := r.hr.RenderNode(&b, node, entering)
	s := b.String()
	if ops&unescapeQuotes != 0 {
		s = strings.ReplaceAll(s, "&quot;", "\"")
	}
	if ops&escapeEmdashes != 0 {
		s = strings.ReplaceAll(s, "—", "&mdash;")
	}
	if ops&removeCodeNewline != 0 {
		const suf = "\n</code></pre>\n"
		if strings.HasSuffix(s, suf) {
			s = s[:len(s)-len(suf)] + suf[1:]
		}
	}
	if _, err := io.WriteString(w, s); err != nil {
		r.setErrorf("failed writing mangled %v node: %v", node.Type, err)
		return md.Terminate
	}
	return ret
}

// finishImgInfo validates an imgInfo struct and fills additional fields.
func (r *renderer) finishImgInfo(info *imgInfo) error {
	if info.Path == "" {
		return errors.New("path must be set")
	}
	if info.Alt == "" {
		return errors.New("alt must be set")
	}
	if r.amp {
		if info.layout == "" {
			info.layout = "responsive"
		}
		info.Attr = append(info.Attr, template.HTMLAttr(fmt.Sprintf(`layout="%s"`, info.layout)))
	}
	if info.Lazy && !r.amp { // <amp-img> already lazy-loads
		info.Attr = append(info.Attr, template.HTMLAttr(`loading="lazy"`))
	}

	wc := strings.IndexByte(info.Path, '*')
	if wc == -1 {
		info.Src = info.Path
		info.WebPSrc = removeExt(info.Src) + WebPExt
		info.biggestSrc = info.Src

		// If the image's display dimensions weren't supplied, get them from the file.
		if info.Width <= 0 || info.Height <= 0 {
			var err error
			if info.Width, info.Height, err = imageSize(filepath.Join(r.si.StaticDir(), info.Src)); err != nil {
				return fmt.Errorf("failed getting %v size: %v", info.Src, err)
			}
		}
		info.Srcset = fmt.Sprintf("%s %dw", info.Src, info.Width)
		info.WebPSrcset = fmt.Sprintf("%s %dw", info.WebPSrc, info.Width)
	} else {
		pre := info.Path[:wc]
		suf := info.Path[wc+1:]
		var err error
		if info.Srcset, info.widths, err = r.makeSrcset(pre, suf); err != nil {
			return err
		} else if info.Srcset == "" {
			return fmt.Errorf("no images matched by prefix %q and suffix %q", pre, suf)
		}
		wsuf := removeExt(suf) + WebPExt
		if info.WebPSrcset, _, err = r.makeSrcset(pre, wsuf); err != nil {
			return err
		} else if info.WebPSrcset == "" {
			return fmt.Errorf("no images matched by prefix %q and suffix %q", pre, wsuf)
		}

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
			if info.Width, info.Height, err = imageSize(filepath.Join(r.si.StaticDir(), p)); err != nil {
				return fmt.Errorf("failed getting %v dimensions: %v", p, err)
			}
		}
		info.Src = fmt.Sprintf("%s%d%s", pre, info.Width, suf)
		info.WebPSrc = removeExt(info.Src) + WebPExt
		info.biggestSrc = fmt.Sprintf("%s%d%s", pre, info.widths[len(info.widths)-1], suf)
	}

	if info.Sizes == "" {
		info.Sizes = fmt.Sprintf("%dpx", info.Width)
	}

	if err := r.si.CheckStatic(info.Src); err != nil {
		return err
	}
	if err := r.si.CheckStatic(info.WebPSrc); err != nil {
		return err
	}
	if err := r.si.CheckStatic(info.biggestSrc); err != nil {
		return err
	}
	return nil
}

// makeSrcset returns a srcset attribute value corresponding to the images matched by
// pre and suf. The returned slice contains image widths in ascending order.
func (r *renderer) makeSrcset(pre, suf string) (string, []int, error) {
	glob := filepath.Join(r.si.StaticDir(), pre+"*"+suf)
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
	preLen := len(filepath.Join(r.si.StaticDir(), pre))
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

// removeExt removes the extension (e.g. ".txt") from p.
func removeExt(p string) string {
	return p[:len(p)-len(filepath.Ext(p))]
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
