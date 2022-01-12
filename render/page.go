// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package render

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/html"

	// Importing this as gopkg.in/russross/blackfriday.v2 doesn't work since the
	// package's go.mod currently contains the github.com URL instead:
	// https://github.com/russross/blackfriday/issues/500
	bf "github.com/russross/blackfriday/v2"

	yaml "gopkg.in/yaml.v3"
)

const (
	mobileMaxWidth  = 640
	desktopMinWidth = mobileMaxWidth + 1

	dateLayout = "2006-01-02"

	// WebPExt is the extension for generated WebP image files.
	WebPExt = ".webp"
)

// Page renders and returns the page described by the supplied Markdown data.
// The id parameter specifies the page's default ID, although this can be overriden in the page block.
// The amp parameter specifies whether the AMP or non-AMP version of the page should be rendered.
// The returned feed info is nil if the page should not be included in the Atom feed.
func Page(si SiteInfo, id string, markdown []byte, amp bool) ([]byte, *PageFeedInfo, error) {
	r := newRenderer(si, id, amp)
	b := bf.Run(markdown, bf.WithRenderer(r),
		bf.WithExtensions((bf.CommonExtensions&^bf.Autolink)|bf.Footnotes))
	if r.err != nil {
		return nil, nil, r.err
	}

	var fi *PageFeedInfo
	if !r.pi.OmitFromFeed && r.pi.Created != "" {
		fi = &PageFeedInfo{
			Title: r.pi.Title,
			Desc:  r.pi.Desc,
		}
		var err error
		if fi.AbsURL, err = si.AbsURL(r.pi.NavItem.URL); err != nil {
			return nil, nil, err
		}
		if fi.Created, err = time.Parse(dateLayout, r.pi.Created); err != nil {
			return nil, nil, err
		}
	}

	return b, fi, nil
}

// PageFeedInfo contains metadata about a page that is needed to generate an Atom feed.
type PageFeedInfo struct {
	Title   string
	Desc    string
	AbsURL  string
	Created time.Time
}

// pageInfo holds information about the current page that is used by page.tmpl.
type pageInfo struct {
	Title           string `yaml:"title"`             // original title
	ID              string `yaml:"id"`                // NavItem.ID to highlight in navbox (inferred from filename if empty)
	Desc            string `yaml:"desc"`              // meta description
	ImagePath       string `yaml:"image_path"`        // path to structured data image in static dir
	Created         string `yaml:"created"`           // creation date as 'YYYY-MM-DD'
	Modified        string `yaml:"modified"`          // last-modified date as 'YYYY-MM-DD'
	HideTitleSuffix bool   `yaml:"hide_title_suffix"` // don't append SiteInfo.TitleSuffix
	HideBackToTop   bool   `yaml:"hide_back_to_top"`  // hide footer link to jump to top
	HideDates       bool   `yaml:"hide_dates"`        // hide footer created and modified dates
	OmitFromFeed    bool   `yaml:"omit_from_feed"`    // omit page from RSS feed
	HasMap          bool   `yaml:"has_map"`           // page contains a map
	HasGraph        bool   `yaml:"has_graph"`         // page contains one or more graphs
	PageStyle       string `yaml:"page_style"`        // optional custom page-specific CSS

	SiteInfo *SiteInfo `yaml:"-"` // site-level information
	NavItem  *NavItem  `yaml:"-"` // nav item corresponding to current page

	FullTitle string `yaml:"-"` // used in <title> element (possibly has suffix)

	LogoHTML   imgInfo `yaml:"-"` // header logo for non-AMP
	LogoAMP    imgInfo `yaml:"-"` // header logo for AMP
	NavToggle  imgInfo `yaml:"-"` // nav toggle icon for non-AMP
	MenuButton imgInfo `yaml:"-"` // menu button for AMP

	LinkRel  string `yaml:"-"` // rel attribute for AMP/non-AMP <link>, e.g. "canonical"
	LinkHref string `yaml:"-"` // href attribute for AMP/non-AMP <link>
	FeedHref string `yaml:"-"` // href attribute for Atom feed <link>

	HTMLStyle        template.CSS  `yaml:"-"` // inline CSS for non-AMP page
	HTMLScripts      []template.JS `yaml:"-"` // inline JS for non-AMP page
	AMPStyle         template.CSS  `yaml:"-"` // inline boilerplate CSS for AMP page
	AMPNoscriptStyle template.CSS  `yaml:"-"` // inline boilerplate <noscript> CSS for AMP page
	AMPCustomStyle   template.CSS  `yaml:"-"` // inline custom CSS for AMP page
	CSPMeta          template.HTML `yaml:"-"` // <meta> tag for Content Security policy
	StructData       structData    `yaml:"-"` // structured data JSON info
}

// figureInfo holds information used by figure.tmpl.
type figureInfo struct {
	Align       string        `yaml:"align"`        // left, right, center, desktop_left, desktop_right, desktop_alt
	Caption     template.HTML `yaml:"caption"`      // <figcaption> text; template.HTML to permit links and not escape quotes
	Class       string        `yaml:"class"`        // CSS class
	DesktopOnly bool          `yaml:"desktop_only"` // only display on desktop
	MobileOnly  bool          `yaml:"mobile_only"`  // only display on mobile
}

// renderer converts a single Markdown page into HTML.
// It implements the bf.Renderer interface.
type renderer struct {
	si   *SiteInfo        // site-level info
	pi   pageInfo         // page-level info
	tmpl *templater       // loads and executes HTML templates
	hr   *bf.HTMLRenderer // standard BlackFriday HTML renderer
	err  error            // error encountered during rendering
	amp  bool             // rendering an AMP page

	startingBox bool         // currently in the middle of a level-1 header
	boxTitle    bytes.Buffer // text seen while startingBox is true
	inBox       bool         // currently rendering a box

	lastFigureAlign string // last "align" value used for a figure
	numMapMarkers   int    // number of boxes with "map_marker"
	didThumb        bool   // already rendered an image with a thumbnail placeholder
}

func newRenderer(si SiteInfo, id string, amp bool) *renderer {
	r := renderer{
		si: &si,
		pi: pageInfo{ID: id, SiteInfo: &si, Desc: si.DefaultDesc},
		hr: bf.NewHTMLRenderer(bf.HTMLRendererParameters{
			Flags: bf.FootnoteReturnLinks,
		}),
		amp: amp,
	}

	r.tmpl = newTemplater(filepath.Join(si.TemplateDir()), template.FuncMap{
		"amp": func() bool {
			return r.amp
		},
		"current": func() *NavItem {
			return r.pi.NavItem
		},
		"formatDate": func(date, layout string) string {
			t, err := time.Parse(dateLayout, date)
			if err != nil {
				r.setErrorf("failed to parse date %q: %v", date, err)
				return ""
			}
			return t.Format(layout)
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

func (r *renderer) RenderNode(w io.Writer, node *bf.Node, entering bool) bf.WalkStatus {
	// Bail if we already set an error in RenderHeader.
	if r.err != nil {
		return bf.Terminate
	}

	// Capture box title text.
	if node.Type != bf.Heading && r.startingBox {
		return r.hr.RenderNode(&r.boxTitle, node, entering)
	}

	switch node.Type {
	case bf.Code:
		return r.mangleOutput(w, node, entering, unescapeQuotes)
	case bf.CodeBlock:
		return r.renderCodeBlock(w, node, entering)
	case bf.Heading:
		return r.renderHeading(w, node, entering)
	case bf.HTMLSpan:
		return r.renderHTMLSpan(w, node, entering)
	case bf.Link:
		// Make sure that we don't rewrite the link a second time when exiting.
		if entering && node.LinkData.Footnote == nil {
			link, err := r.rewriteLink(string(node.LinkData.Destination))
			if err != nil {
				r.setError(err)
				return bf.Terminate
			}
			node.LinkData.Destination = []byte(link)
		}
		// Fall through and let Blackfriday render the possibly-updated URL as normal.
	case bf.Text:
		return r.mangleOutput(w, node, entering, unescapeQuotes|escapeEmdashes)
	}

	// Fall back to Blackfriday's default HTML rendering.
	return r.hr.RenderNode(w, node, entering)
}

func (r *renderer) RenderHeader(w io.Writer, ast *bf.Node) {
	if r.err != nil {
		return
	}

	fc := ast.FirstChild
	if fc == nil || fc.Type != bf.CodeBlock || string(fc.CodeBlockData.Info) != "page" {
		r.setErrorf(`page doesn't start with "page" code block`)
		return
	}
	if err := unmarshalYAML(fc.Literal, &r.pi); err != nil {
		r.setErrorf("failed to parse page info from %q: %v", fc.Literal, err)
		return
	}

	for _, n := range r.si.NavItems {
		if r.pi.NavItem = n.FindID(r.pi.ID); r.pi.NavItem != nil {
			break
		}
	}
	if r.pi.NavItem == nil {
		// Add a fake nav item for the index page if it's current and isn't listed.
		// The Name and URL fields don't need to be set since this item is never rendered.
		if r.pi.ID == indexID {
			r.pi.NavItem = &NavItem{ID: indexID}
		} else {
			r.setErrorf("no page with ID %q", r.pi.ID)
			return
		}
	}

	// Append the title suffix if needed.
	r.pi.FullTitle = r.pi.Title
	if !r.pi.HideTitleSuffix {
		r.pi.FullTitle += r.si.TitleSuffix
	}

	r.pi.LogoHTML = imgInfo{
		Path:    r.si.LogoPathHTML,
		Alt:     r.si.LogoAlt,
		ID:      "nav-logo",
		noThumb: true, // looks weird, and absolute positioning conflicts with placeholder CSS
	}
	if err := r.pi.LogoHTML.finish(r.si, r.amp, &r.didThumb); err != nil {
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
		Path:    r.si.LogoPathAMP,
		Alt:     r.si.LogoAlt,
		ID:      "nav-logo",
		noThumb: true, // looks weird
	}
	if err := r.pi.LogoAMP.finish(r.si, r.amp, &r.didThumb); err != nil {
		r.setErrorf("AMP logo failed: %v", err)
		return
	}

	// On mobile, collapse the navbox if the page doesn't have subpages.
	r.pi.NavToggle = imgInfo{
		Path:    r.si.NavTogglePath,
		Alt:     "[toggle navigation]",
		ID:      "nav-toggle-img",
		noThumb: true, // tiny
	}
	if len(r.pi.NavItem.Children) == 0 {
		r.pi.NavToggle.Classes = append(r.pi.NavToggle.Classes, "expand")
	}
	if err := r.pi.NavToggle.finish(r.si, r.amp, &r.didThumb); err != nil {
		r.setErrorf("nav toggle failed: %v", err)
		return
	}

	r.pi.MenuButton = imgInfo{
		Path: r.si.MenuButtonPath,
		Alt:  "[toggle menu]",
		ID:   "menu-button",
		Attr: []template.HTMLAttr{
			template.HTMLAttr(`tabindex="0"`),
			template.HTMLAttr(`role="button"`),
			template.HTMLAttr(`on="tap:sidebar.open"`),
		},
		noThumb: true, // tiny
	}
	if err := r.pi.MenuButton.finish(r.si, r.amp, &r.didThumb); err != nil {
		r.setErrorf("menu button failed: %v", err)
		return
	}

	var err error
	if r.pi.FeedHref, err = r.si.AbsURL(FeedFile); err != nil {
		r.setError(err)
		return
	}

	// It would be much simpler to just use a map[string]interface{} for this,
	// but the properties are marshaled in an arbitrary order then, making it
	// hard to compare the output against the previous version of the site.
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
		},
	}
	if r.si.PublisherLogoPath != "" {
		if r.pi.StructData.Publisher.Logo, err = r.newStructDataImage(r.si.PublisherLogoPath); err != nil {
			r.setError(err)
			return
		}
	}
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
	if r.pi.ImagePath != "" {
		if r.pi.StructData.Image, err = r.newStructDataImage(r.pi.ImagePath); err != nil {
			r.setError(err)
			return
		}
	}

	if r.amp {
		r.pi.LinkRel = "canonical"
		if r.pi.LinkHref, err = r.si.AbsURL(r.pi.NavItem.URL); err != nil {
			r.setError(err)
			return
		}

		r.pi.AMPStyle = template.CSS(getStdInline("amp-boilerplate.css"))
		r.pi.AMPNoscriptStyle = template.CSS(getStdInline("amp-boilerplate-noscript.css"))
		r.pi.AMPCustomStyle = template.CSS(
			getStdInline("base.css") + getStdInline("mobile.css") + getStdInline("amp.css") +
				r.si.ReadInline("base.css") + r.si.ReadInline("mobile.css") +
				r.si.ReadInline("amp.css") + r.si.ReadInline("page_"+r.pi.ID+".css") +
				r.pi.PageStyle)

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

		r.pi.HTMLStyle = template.CSS(getStdInline("base.css") + getStdInline("nonamp.css") +
			r.si.ReadInline("base.css") + r.si.ReadInline("nonamp.css") +
			fmt.Sprintf("@media(min-width:%dpx){%s%s}",
				desktopMinWidth, getStdInline("desktop.css"), r.si.ReadInline("desktop.css")) +
			fmt.Sprintf("@media(max-width:%dpx){%s%s}",
				mobileMaxWidth, getStdInline("mobile.css"), r.si.ReadInline("mobile.css")) +
			r.si.ReadInline("page_"+r.pi.ID+".css") + r.pi.PageStyle)
		r.pi.HTMLScripts = []template.JS{template.JS(getStdInline("base.js"))}
		if r.pi.HasMap {
			r.pi.HTMLScripts = append(r.pi.HTMLScripts, template.JS(getStdInline("map.js")))
		}
		if js := r.si.ReadInline("page_" + r.pi.ID + ".js"); js != "" {
			r.pi.HTMLScripts = append(r.pi.HTMLScripts, template.JS(js))
		}

		csp := cspBuilder{}
		csp.add(cspDefault, cspNone)
		csp.add(cspChild, cspSelf)
		csp.add(cspImg, cspSelf)
		csp.add(cspImg, "data:") // needed for inline image thumbnails

		if r.si.CloudflareAnalyticsToken != "" {
			csp.add(cspScript, cspSource(r.si.CloudflareAnalyticsScriptURL))
			csp.add(cspConnect, cspSource(r.si.CloudflareAnalyticsConnectPattern))
		}

		csp.hash(cspStyle, string(r.pi.HTMLStyle))
		for _, s := range r.pi.HTMLScripts {
			csp.hash(cspScript, string(s))
		}
		r.pi.CSPMeta = template.HTML(csp.tag())
	}

	r.setError(r.tmpl.runNamed(w, []string{"page.tmpl", "img.tmpl"}, "start", &r.pi, nil))
}

func (r *renderer) RenderFooter(w io.Writer, ast *bf.Node) {
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

// Renders a node of type bf.CodeBlock and returns the appropriate walk status.
// Sets r.err and returns bf.Terminate if an error is encountered.
func (r *renderer) renderCodeBlock(w io.Writer, node *bf.Node, entering bool) bf.WalkStatus {
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
			return bf.Terminate
		}
		return bf.SkipChildren
	case "graph":
		var info struct {
			figureInfo `yaml:",inline"`
			Href       string `yaml:"href"`   // relative path to graph iframe page
			Name       string `yaml:"name"`   // graph data name
			Width      int    `yaml:"width"`  // graph width (without border)
			Height     int    `yaml:"height"` // graph height (without border)
		}
		if err := unmarshalYAML(node.Literal, &info); err != nil {
			r.setErrorf("failed to parse graph info from %q: %v", node.Literal, err)
			return bf.Terminate
		}
		info.figureInfo.Align = figureAlign(info.figureInfo.Align)
		info.Href = iframeHref(info.Href)
		if r.setError(r.tmpl.run(w, []string{"graph.tmpl", "figure.tmpl"}, info, nil)) != nil {
			return bf.Terminate
		}
		return bf.SkipChildren
	case "image":
		var info struct {
			figureInfo `yaml:",inline"`
			imgInfo    `yaml:",inline"`
			Href       string `yaml:"href"`
			NoLink     bool   `yaml:"no_link"`
		}
		if err := unmarshalYAML(node.Literal, &info); err != nil {
			r.setErrorf("failed to parse image info from %q: %v", node.Literal, err)
			return bf.Terminate
		}
		if err := info.imgInfo.finish(r.si, r.amp, &r.didThumb); err != nil {
			r.setErrorf("bad data in %q: %v", node.Literal, err)
			return bf.Terminate
		}
		info.figureInfo.Align = figureAlign(info.figureInfo.Align)
		if info.Href == "" && !info.NoLink && info.imgInfo.biggestSrc != info.imgInfo.Src {
			info.Href = info.imgInfo.biggestSrc
		}
		if info.Href != "" {
			var err error
			if info.Href, err = r.rewriteLink(info.Href); err != nil {
				r.setError(err)
				return bf.Terminate
			}
		}
		if r.setError(r.tmpl.run(w, []string{"image_block.tmpl", "figure.tmpl", "img.tmpl"}, info, nil)) != nil {
			return bf.Terminate
		}
		return bf.SkipChildren
	case "map":
		var info struct {
			imgInfo `yaml:",inline"` // placeholder image (also used for dimensions)
			Href    string           `yaml:"href"` // relative path to map iframe page
		}
		if err := unmarshalYAML(node.Literal, &info); err != nil {
			r.setErrorf("failed to parse map info from %q: %v", node.Literal, err)
			return bf.Terminate
		}
		info.imgInfo.layout = "fill"
		info.imgInfo.Attr = append(info.imgInfo.Attr, template.HTMLAttr("placeholder"))
		info.imgInfo.Alt = "[map placeholder]"
		info.imgInfo.noThumb = true // already a placeholder
		if err := info.imgInfo.finish(r.si, r.amp, &r.didThumb); err != nil {
			r.setErrorf("bad data in %q: %v", node.Literal, err)
			return bf.Terminate
		}
		info.Href = iframeHref(info.Href)
		if r.setError(r.tmpl.run(w, []string{"map.tmpl", "img.tmpl"}, info, nil)) != nil {
			return bf.Terminate
		}
		return bf.SkipChildren
	case "page":
		return bf.SkipChildren // handled in RenderHeader
	default:
		node.CodeBlockData.Info = nil // prevent Blackfriday from adding e.g. "language-html" CSS class
		return r.mangleOutput(w, node, entering, unescapeQuotes|removeCodeNewline|wrapNoSelect)
	}
}

// Renders a node of type bf.Heading and returns the appropriate walk status.
// Sets r.err and returns bf.Terminate if an error is encountered.
func (r *renderer) renderHeading(w io.Writer, node *bf.Node, entering bool) bf.WalkStatus {
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
				return bf.Terminate
			}
			r.inBox = false
		}
		r.startingBox = true
		r.boxTitle.Reset()
		return bf.GoToNext
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
			info.MapLabel = string(rune('A' + r.numMapMarkers))
			r.numMapMarkers++
		}
	}

	r.startingBox = false
	r.inBox = true
	if r.setError(r.tmpl.runNamed(w, []string{"box.tmpl"}, "start", &info, nil)) != nil {
		return bf.Terminate
	}
	return bf.GoToNext
}

// Renders a node of type bf.HTMLSpan and returns the appropriate walk status.
// Sets r.err and returns bf.Terminate if an error is encountered.
func (r *renderer) renderHTMLSpan(w io.Writer, node *bf.Node, entering bool) bf.WalkStatus {
	// Wrap all the logic in a function to make it easy to consolidate error-handling.
	// bf.Terminate is always returned if an error occurs.
	ws, err := func() (bf.WalkStatus, error) {
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
			return bf.SkipChildren, nil
		case "image":
			if token.Type == html.StartTagToken {
				var info = imgInfo{
					Classes: []string{"inline"},
					layout:  "fixed",
					noThumb: true, // probably small
				}
				if err := unmarshalAttrs(token.Attr, &info); err != nil {
					return 0, err
				}
				if err := info.finish(r.si, r.amp, &r.didThumb); err != nil {
					return 0, err
				}
				if err := r.tmpl.runNamed(w, []string{"img.tmpl"}, "img", info, nil); err != nil {
					return 0, err
				}
			}
			return bf.SkipChildren, nil
		case "only-amp":
			if !r.amp {
				if token.Type == html.StartTagToken {
					io.WriteString(w, "<!-- AMP-only content \n")
				} else if token.Type == html.EndTagToken {
					io.WriteString(w, "\n-->")
				}
			}
			return bf.GoToNext, nil // process nested nodes
		case "only-nonamp":
			if r.amp {
				if token.Type == html.StartTagToken {
					io.WriteString(w, "<!-- non-AMP-only content \n")
				} else if token.Type == html.EndTagToken {
					io.WriteString(w, "\n-->")
				}
			}
			return bf.GoToNext, nil // process nested nodes
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
			return bf.GoToNext, nil // process nested content
		default:
			return 0, errors.New("unsupported tag")
		}
	}()
	if err != nil {
		r.setErrorf("rendering HTML span %q failed: %v", node.Literal, err)
		return bf.Terminate
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

	// TODO: Revisit this decision.
	if link[0] == '/' {
		return "", fmt.Errorf("link %q shouldn't have leading slash", link)
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

	// Make sure that links to static resources work.
	if !isPage(link) {
		if err := r.si.CheckStatic(link); err != nil {
			return "", err
		}
	}

	if r.amp {
		// Resources that aren't AMP pages (or images?) probably won't be served by AMP caches. I
		// haven't been able to find anything in the AMP documentation guaranteeing that all caches
		// will rewrite relative links to these resources, though.
		//
		// https://github.com/ampproject/amphtml/blob/main/docs/spec/amp-cache-guidelines.md only
		// says that image/* should be accepted by caches.
		//
		// https://developers.google.com/amp/cache/overview says that at least in the case of the Google
		// AMP Cache, "outbound links are made absolute so that they continue to work when the document
		// is served from the Google AMP Cache origin instead of the publisher origin." Empirically,
		// Google's cache seems to rewrite relative links that were pointing to non-AMP pages.
		//
		// Using absolute URLs is annoying for development, but seems safer to do.
		if !isPage(link) || forceNonAMP {
			if abs, err := r.si.AbsURL(link); err != nil {
				return "", err
			} else {
				return abs, nil
			}
		}
		return ampPage(link), nil
	}

	if forceAMP {
		return ampPage(link), nil
	}
	return link, nil
}

// mangleOps describes operations that mangleOutput should perform.
type mangleOps int

const (
	unescapeQuotes    mangleOps = 1 << iota // replaces '&quot;' with '"'
	escapeEmdashes                          // replaces '—' with '&mdash;'
	removeCodeNewline                       // removes '\n' before trailing '</code></pre>'
	wrapNoSelect                            // wraps ‹›-enclosed text in '<span class="no-select">'
)

var noSelectRegexp = regexp.MustCompile(`‹([^›]*)›`)

// mangleOutput is a helper function that uses Blackfriday's standard
// HTMLRenderer to render the supplied node, and then performs the requested
// operations on the output before writing it to w.
func (r *renderer) mangleOutput(w io.Writer, node *bf.Node, entering bool, ops mangleOps) bf.WalkStatus {
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
	if ops&wrapNoSelect != 0 {
		s = noSelectRegexp.ReplaceAllString(s, `<span class="no-select">$1</span>`)
	}
	if _, err := io.WriteString(w, s); err != nil {
		r.setErrorf("failed writing mangled %v node: %v", node.Type, err)
		return bf.Terminate
	}
	return ret
}

// newStructDataImage returns a structDataImage for the image
// at the supplied path under the static dir.
func (r *renderer) newStructDataImage(path string) (*structDataImage, error) {
	img := &structDataImage{Type: "ImageObject"}
	if err := r.si.CheckStatic(path); err != nil {
		return nil, err
	}
	var err error
	if img.URL, err = r.si.AbsURL(path); err != nil {
		return nil, err
	}
	img.Width, img.Height, err = imageSize(filepath.Join(r.si.StaticDir(), path))
	return img, err
}

// removeExt removes the extension (e.g. ".txt") from p.
func removeExt(p string) string {
	return p[:len(p)-len(filepath.Ext(p))]
}

// getStdInline returns the contents of the named standard inline file from std_inline.go.
// It panics if the file does not exist.
func getStdInline(fn string) string {
	data, ok := stdInline[fn]
	if !ok {
		panic(fmt.Sprintf("No standard inline file %q", fn))
	}
	min, err := minifyData(data, filepath.Ext(fn))
	if err != nil {
		panic(fmt.Sprintf("Failed minifying %v: %v", fn, err))
	}
	return min
}

// unmarshalYAML decodes a YAML document within b into out.
// An error is returned if any unknown fields are found
// (differing from yaml.Unmarshal's behavior).
func unmarshalYAML(b []byte, out interface{}) error {
	dec := yaml.NewDecoder(bytes.NewReader(b))
	dec.KnownFields(true)
	return dec.Decode(out)
}
