// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package render

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
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
	// TODO: All of the uses of these dirs need to prepend the site dir.
	inlineDir   = "inline"
	staticDir   = "static"
	templateDir = "templates"
	navFile     = "nav.yaml"

	indexID = "index" // index page

	baseURL     = "https://www.erat.org/"
	titleSuffix = " - erat.org"          // appended to page titles
	defaultDesc = "Dan Erat's home page" // default meta description

	// Structured data.
	authorName          = "Daniel Erat"
	authorEmail         = "dan-website@erat.org"
	publisherName       = "erat.org"
	publisherLogoURL    = baseURL + "resources/erat-logo-20161109.png"
	publisherLogoWidth  = 167
	publisherLogoHeight = 60

	googleAnalyticsCode = "UA-187852-1"

	mobileMaxWidth  = 640
	desktopMinWidth = 641

	// Dates in resource filenames.
	// TODO: This is terrible; delete it.
	logoImgDate     = "20160914"
	collapseImgDate = "20160915"
)

func Page(markdown []byte, dir string, amp bool) ([]byte, error) {
	var ni []*navItem
	if b, err := ioutil.ReadFile(filepath.Join(dir, navFile)); err != nil {
		return nil, err
	} else if err := yaml.Unmarshal(b, &ni); err != nil {
		return nil, fmt.Errorf("failed parsing nav items: %v", err)
	}

	r := newRenderer(ni, amp)
	b := md.Run(markdown, md.WithRenderer(r))
	if r.err != nil {
		return nil, r.err
	}
	return b, nil
}

func readInline(fn string) string {
	b, err := ioutil.ReadFile(filepath.Join(inlineDir, fn))
	if err != nil {
		log.Fatal("Failed reading inline file: ", err)
	}
	return string(b)
}

type pageInfo struct {
	Title           string `yaml:"title"`             // used in <title> element
	ID              string `yaml:"id"`                // ID to highlight in navbox
	Desc            string `yaml:"desc"`              // meta description
	ImgURL          string `yaml:"img_url"`           // structured data image URL (must be 696 pixels wide)
	ImgWidth        int    `yaml:"img_width"`         // structured data image width
	ImgHeight       int    `yaml:"img_height"`        // structured data image width
	Created         string `yaml:"created"`           // creation date as 'YYYY-MM-DD'
	Modified        string `yaml:"modified"`          // last-modified date as 'YYYY-MM-DD'
	HideTitleSuffix bool   `yaml:"hide_title_suffix"` // don't append titleSuffix
	HideBackToTop   bool   `yaml:"hide_back_to_top"`  // hide footer link to jump to top
	HasMap          bool   `yaml:"has_map"`           // page contains a map
	HasGraph        bool   `yaml:"has_graph"`         // page contains one or more maps

	NavItems []*navItem `yaml:"-"` // hierarchy of nav items
	NavItem  *navItem   `yaml:"-"` // nav item corresponding to current page

	LogoImgDate         string `yaml:"-"`
	CollapseImgDate     string `yaml:"-"`
	GoogleAnalyticsCode string `yaml:"-"`

	LinkRel  string `yaml:"-"`
	LinkHref string `yaml:"-"`

	HTMLStyle        template.CSS  `yaml:"-"`
	HTMLScripts      []template.JS `yaml:"-"`
	AMPStyle         template.CSS  `yaml:"-"`
	AMPNoscriptStyle template.CSS  `yaml:"-"`
	AMPCustomStyle   template.CSS  `yaml:"-"`
	CSPMeta          template.HTML `yaml:"-"`

	StructData          structData    `yaml:"-"`
	BeginContentComment template.HTML `yaml:"-"`
	EndContentComment   template.HTML `yaml:"-"`
}

type imageTagInfo struct {
	Prefix string `html:"prefix",yaml:"prefix"` // relative path prefix (before width)
	Suffix string `html:"suffix",yaml:"suffix"` // path suffix (empty if prefix contains full path)
	Width  int    `html:"width",yaml:"width"`   // 100% width
	Height int    `html:"height",yaml:"height"` // 100% height
	Alt    string `html:"alt",yaml:"alt"`       // alt text

	Layout string   // AMP layout ("responsive" used if empty)
	Inline bool     // add "inline" class
	Attr   []string // additional attributes to include
}

type imageboxInfo struct {
	Align   string        `yaml:"align"`   // left, right, center, desktop_left, desktop_right, desktop_alt
	Caption template.HTML `yaml:"caption"` // <figcaption> text; template.HTML to permit links and not escape quotes
	Class   string        `yaml:"class"`   // CSS class
}

type renderer struct {
	tmpls map[string]*template.Template
	hr    *md.HTMLRenderer
	pi    pageInfo
	err   error // error encountered during rendering
	amp   bool  // rendering an AMP page

	startingBox bool         // in the middle of a level-1 header
	boxTitle    bytes.Buffer // text seen while startingBox is true
	inBox       bool         // rendering a box

	lastImageboxAlign string // last "align" value used for an imagebox
	numMapMarkers     int    // number of boxes with "map_marker"
}

func newRenderer(navItems []*navItem, amp bool) *renderer {
	return &renderer{
		tmpls: make(map[string]*template.Template),
		hr:    md.NewHTMLRenderer(md.HTMLRendererParameters{}),
		pi: pageInfo{
			Desc:                defaultDesc,
			NavItems:            navItems,
			LogoImgDate:         logoImgDate,
			CollapseImgDate:     collapseImgDate,
			GoogleAnalyticsCode: googleAnalyticsCode,
			BeginContentComment: template.HTML("<!-- begin content -->"),
			EndContentComment:   template.HTML("<!-- end content -->"),
		},
		amp: amp,
	}
}

// tagRegexp matches the tag name at the beginning of an HTML tag, e.g. "span" or "/span".
var tagRegexp = regexp.MustCompile(`^<([^\s>]+)`)

// template runs a template consisting of the named files using the supplied
// data and functions. The template is cached after it's loaded for the first
// time, and some standard functions are also included.
func (r *renderer) template(w io.Writer, names []string, data interface{}, funcs template.FuncMap) error {
	tname := names[0]
	tmpl, ok := r.tmpls[tname]
	if !ok {
		var paths []string
		for _, n := range names {
			paths = append(paths, filepath.Join(templateDir, n))
		}

		fm := template.FuncMap{
			"amp": func() bool {
				return r.amp
			},
			"formatDate": func(date, layout string) string {
				t, err := time.Parse("2006-01-02", date)
				if err != nil {
					r.err = fmt.Errorf("failed to parse date %q: %v", date, err)
					return ""
				}
				return t.Format(layout)
			},
			"srcset": func(pre, suf string) string {
				glob := filepath.Join(staticDir, pre+"*"+suf)
				ps, err := filepath.Glob(glob)
				if err != nil {
					r.err = fmt.Errorf("failed to list image files %q: %v", glob, err)
					return ""
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
				// TODO: Remove this; it's just here to match Ruby's Dir.glob.
				if ps, err = dirOrder(ps); err != nil {
					r.err = fmt.Errorf("srcset failed to sort images: %v", err)
					return ""
				}
				var srcs []string
				for _, p := range ps {
					width := p[len(filepath.Join(staticDir, pre)) : len(p)-len(suf)]
					srcs = append(srcs, fmt.Sprintf("%s%s%s %sw", pre, width, suf, width))
				}
				return strings.Join(srcs, ", ")
			},
		}
		for n, f := range funcs {
			fm[n] = f
		}

		tmpl = template.Must(template.New(tname).Funcs(fm).ParseFiles(paths...))
		r.tmpls[tname] = tmpl
	}
	return tmpl.Execute(w, data)
}

func (r *renderer) RenderNode(w io.Writer, node *md.Node, entering bool) md.WalkStatus {
	//log.Printf("RenderNode: %s %v %q", node.Type, entering, node.Literal)

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
	case md.Del:
		// Blackfriday emits <del>. Emit <s> for now to make comparisons easier.
		// TODO: Remove this later?
		if entering {
			io.WriteString(w, "<s>")
		} else {
			io.WriteString(w, "</s>")
		}
		return md.GoToNext
	case md.Heading:
		return r.renderHeading(w, node, entering)
	case md.HTMLSpan:
		return r.renderHTMLSpan(w, node, entering)
	case md.Link:
		var link string
		if link, r.err = r.rewriteLink(string(node.LinkData.Destination)); r.err != nil {
			return md.Terminate
		}
		node.LinkData.Destination = []byte(link)
		// Fall through and let Blackfriday render the possibly-updated URL as normal.
	case md.Strong:
		// Blackfriday emits <strong>. Emit <b> for now to make comparisons easier.
		// TODO: Remove this later?
		if entering {
			io.WriteString(w, "<b>")
		} else {
			io.WriteString(w, "</b>")
		}
		return md.GoToNext
	case md.Text:
		return r.mangleOutput(w, node, entering, unescapeQuotes|escapeEmdashes)
	}

	// Fall back to Blackfriday's default HTML rendering.
	return r.hr.RenderNode(w, node, entering)
}

func (r *renderer) RenderHeader(w io.Writer, ast *md.Node) {
	fc := ast.FirstChild
	if fc == nil || fc.Type != md.CodeBlock || string(fc.CodeBlockData.Info) != "page_info" {
		r.err = errors.New("page doesn't start with page info code block")
		return
	}
	if err := yaml.Unmarshal(fc.Literal, &r.pi); err != nil {
		r.err = fmt.Errorf("failed to parse page info from %q: %v", fc.Literal, err)
		return
	}

	m := make(map[string]*navItem)
	for _, p := range r.pi.NavItems {
		p.update(r.pi.ID, m)
	}
	// Add a fake nav item for the index page if it's current.
	// URL fields don't need to be set since this item is never rendered.
	if r.pi.ID == indexID {
		r.pi.NavItem = &navItem{
			ID:       indexID,
			Current:  true,
			Expanded: true,
			Children: r.pi.NavItems,
		}
	} else if r.pi.NavItem = m[r.pi.ID]; r.pi.NavItem == nil {
		r.err = fmt.Errorf("no page with ID %q", r.pi.ID)
		return
	}

	// It would be much simpler to just use a map[string]interface{} for this,
	// but the properties are marshaled in an arbitrary order then, making it
	// hard to compare the output against template_lib.rb.
	r.pi.StructData = structData{
		Context:          "http://schema.org",
		Type:             "Article",
		MainEntityOfPage: r.pi.NavItem.AbsURL,
		Headline:         r.pi.Title,
		DatePublished:    r.pi.Created,
		Author: structDataAuthor{
			Type:  "Person",
			Name:  authorName,
			Email: authorEmail,
		},
		Publisher: structDataPublisher{
			Type: "Organization",
			Name: publisherName,
			URL:  baseURL,
			Logo: structDataImage{
				Type:   "ImageObject",
				URL:    publisherLogoURL,
				Width:  publisherLogoWidth,
				Height: publisherLogoHeight,
			},
		},
	}
	if r.pi.Desc != "" && r.pi.Desc != defaultDesc {
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
		if r.pi.StructData.Image.URL, r.err = absURL(r.pi.ImgURL); r.err != nil {
			return
		}
	}

	// Do this here so it's not included in structured data.
	if !r.pi.HideTitleSuffix {
		r.pi.Title += titleSuffix
	}

	if r.amp {
		r.pi.LinkRel = "canonical"
		r.pi.LinkHref = r.pi.NavItem.AbsURL

		r.pi.AMPStyle = template.CSS(readInline("amp-boilerplate.css.min"))
		r.pi.AMPNoscriptStyle = template.CSS(readInline("amp-boilerplate-noscript.css.min"))
		r.pi.AMPCustomStyle = template.CSS(
			readInline("base.css.min") + readInline("mobile.css.min") + readInline("mobile-amp.css.min"))

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
		r.pi.LinkHref = r.pi.NavItem.AbsAMPURL

		r.pi.HTMLStyle = template.CSS(readInline("base.css.min") + readInline("base-nonamp.css.min") +
			fmt.Sprintf("@media(min-width:%dpx){%s}", desktopMinWidth, readInline("desktop.css.min")) +
			fmt.Sprintf("@media(max-width:%dpx){%s}", mobileMaxWidth, readInline("mobile.css.min")))
		r.pi.HTMLScripts = []template.JS{template.JS(readInline("base.js.min"))}
		if r.pi.HasMap {
			r.pi.HTMLScripts = append(r.pi.HTMLScripts, template.JS(readInline("map.js.min")))
		}

		csp := cspHasher{}
		csp.add(cspDefault, cspNone)
		csp.add(cspChild, cspSelf)
		csp.add(cspImg, cspSelf)

		csp.hash(cspStyle, string(r.pi.HTMLStyle))
		for _, s := range r.pi.HTMLScripts {
			csp.hash(cspScript, string(s))
		}
		r.pi.CSPMeta = template.HTML(csp.tag())
	}

	r.err = r.template(w, []string{"page_header.tmpl"}, &r.pi, nil)
}

func (r *renderer) RenderFooter(w io.Writer, ast *md.Node) {
	// Bail if we set an error in RenderHeader or RenderNode.
	if r.err != nil {
		return
	}

	if r.inBox {
		if r.err = r.template(w, []string{"box_footer.tmpl"}, nil, nil); r.err != nil {
			return
		}
	}
	r.err = r.template(w, []string{"page_footer.tmpl"}, &r.pi, nil)
}

// Renders a node of type md.CodeBlock and returns the appropriate walk status.
// Sets r.err and returns md.Terminate if an error is encountered.
func (r *renderer) renderCodeBlock(w io.Writer, node *md.Node, entering bool) md.WalkStatus {
	// Rewrites the supplied iframe URL (e.g. iframes/my_map.html) to be either absolute or site-rooted.
	// The framed page isn't AMP-compliant, so it won't be served by the AMP cache: https://www.erat.org/amp.html#iframes
	// Absolute URLs could presumably also be used in the non-AMP case, but it makes development harder.
	iframeHref := func(s string) string {
		if r.amp {
			return baseURL + s
		}
		return "/" + s
	}

	// Rewrites the supplied imagebox "align" value to handle "desktop_alt". Also updates lastImageboxAlign.
	imageboxAlign := func(s string) string {
		v := func() string {
			if s != "desktop_alt" {
				return s
			}
			if r.lastImageboxAlign != "desktop_left" {
				return "desktop_left"
			}
			return "desktop_right"
		}()
		r.lastImageboxAlign = v
		return v
	}

	// TODO: Prefix these by '!', or maybe just use custom HTML elements instead.
	switch string(node.CodeBlockData.Info) {
	case "graph":
		var info struct {
			imageboxInfo `yaml:",inline"`
			Href         string `yaml:"href"`   // relative path to graph iframe page
			Name         string `yaml:"name"`   // graph data name
			Width        int    `yaml:"width"`  // graph width (without border)
			Height       int    `yaml:"height"` // graph height (without border)
		}
		if err := yaml.Unmarshal(node.Literal, &info); err != nil {
			r.err = fmt.Errorf("failed to parse graph info from %q: %v", node.Literal, err)
			return md.Terminate
		}
		info.imageboxInfo.Align = imageboxAlign(info.imageboxInfo.Align)
		info.Href = iframeHref(info.Href)
		if r.err = r.template(w, []string{"graph.tmpl", "imagebox.tmpl"}, info, nil); r.err != nil {
			return md.Terminate
		}
		return md.SkipChildren
	case "image":
		var info struct {
			imageboxInfo `yaml:",inline"`
			imageTagInfo `yaml:",inline"`
			Href         string `yaml:"href"`
		}
		if err := yaml.Unmarshal(node.Literal, &info); err != nil {
			r.err = fmt.Errorf("failed to parse image info from %q: %v", node.Literal, err)
			return md.Terminate
		}
		info.imageboxInfo.Align = imageboxAlign(info.imageboxInfo.Align)
		if len(info.Href) == 0 && len(info.Suffix) > 0 {
			info.Href = fmt.Sprintf("%s%d%s", info.Prefix, 2*info.Width, info.Suffix)
			// TODO: check_static_file
		}
		if len(info.Href) > 0 {
			if info.Href, r.err = r.rewriteLink(info.Href); r.err != nil {
				return md.Terminate
			}
		}
		if r.err = r.template(w, []string{"block_image.tmpl", "imagebox.tmpl", "image_tag.tmpl"}, info, nil); r.err != nil {
			return md.Terminate
		}
		return md.SkipChildren
	case "map":
		var info struct {
			imageTagInfo `yaml:",inline"` // placeholder image
			Href         string           `yaml:"href"` // relative path to map iframe page
		}
		if err := yaml.Unmarshal(node.Literal, &info); err != nil {
			r.err = fmt.Errorf("failed to parse map info from %q: %v", node.Literal, err)
			return md.Terminate
		}
		info.imageTagInfo.Layout = "fill"
		info.imageTagInfo.Attr = []string{"placeholder"}
		info.imageTagInfo.Alt = "[map placeholder]"
		info.Href = iframeHref(info.Href)
		if r.err = r.template(w, []string{"map.tmpl", "image_tag.tmpl"}, info, nil); r.err != nil {
			return md.Terminate
		}
		return md.SkipChildren
	case "page_info":
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
			if r.err = r.template(w, []string{"box_footer.tmpl"}, nil, nil); r.err != nil {
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
	if r.err = r.template(w, []string{"box_header.tmpl"}, &info, nil); r.err != nil {
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
		case "clear-floats":
			if token.Type == html.StartTagToken {
				io.WriteString(w, `<span class="clear">`)
			} else if token.Type == html.EndTagToken {
				io.WriteString(w, "</span>")
			}
			return md.SkipChildren, nil
		case "code-url":
			if token.Type == html.StartTagToken {
				io.WriteString(w, `<code class="url">`)
			} else if token.Type == html.EndTagToken {
				io.WriteString(w, "</code>")
			}
			return md.SkipChildren, nil
		case "img-inline":
			if token.Type == html.StartTagToken {
				var info = imageTagInfo{Inline: true, Layout: "fixed"}
				if err := unmarshalAttrs(token.Attr, &info); err != nil {
					return 0, err
				}
				if err := r.template(w, []string{"inline_image.tmpl", "image_tag.tmpl"}, info, nil); err != nil {
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
					return 0, fmt.Errorf("missing attribute in %q", node.Literal)
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
		r.err = fmt.Errorf("Rendering HTML span %q failed: %v", node.Literal, err)
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

	// Non-AMP links don't need to be rewritten.
	if (!r.amp && !forceAMP) || forceNonAMP {
		return link, nil
	}

	// TODO: Revisit this decision.
	if link[0] == '/' {
		return "", fmt.Errorf("link %q shouldn't have leading slash", link)
	}
	// If this isn't a regular page, it may not be served by the AMP CDN. Use an absolute URL.
	if !isPage(link) {
		return baseURL + link, nil
	}

	// TODO: Support forcing non-AMP.
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
// TODO: Consider removing this after the port from Ruby to Go is done -- this
// just helps cut down on differences in the new output.
func (r *renderer) mangleOutput(w io.Writer, node *md.Node, entering bool, ops mangleOps) md.WalkStatus {
	// TODO: Consider using bufio to stream the output instead of storing it in a buffer.
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
		r.err = fmt.Errorf("failed writing mangled %v node: %v", node.Type, err)
		return md.Terminate
	}
	return ret
}

type structData struct {
	Context          string `json:"@context"`
	Type             string `json:"@type"`
	MainEntityOfPage string `json:"mainEntityOfPage"`
	Headline         string `json:"headline"`
	DatePublished    string `json:"datePublished"`

	Author    structDataAuthor    `json:"author"`
	Publisher structDataPublisher `json:"publisher"`

	// TODO: Move these up. They're down here to match the order in template_lib.rb.
	Description  string `json:"description,omitempty"`
	DateModified string `json:"dateModified,omitempty"`

	Image *structDataImage `json:"image,omitempty"`
}

type structDataAuthor struct {
	Type  string `json:"@type"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type structDataPublisher struct {
	Type string          `json:"@type"`
	Name string          `json:"name"`
	URL  string          `json:"url"`
	Logo structDataImage `json:"logo"`
}

type structDataImage struct {
	Type   string `json:"@type"`
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

func dirOrder(ps []string) ([]string, error) {
	var dp string
	var fm = map[string]struct{}{}
	for _, p := range ps {
		if fd := filepath.Dir(p); dp == "" {
			dp = fd
		} else if fd != dp {
			return nil, fmt.Errorf("file %v not in dir %v", p, dp)
		}
		fm[filepath.Base(p)] = struct{}{}
	}
	if len(ps) == 0 {
		return ps, nil
	}

	d, err := os.Open(dp)
	if err != nil {
		return nil, err
	}
	defer d.Close()

	fns, err := d.Readdirnames(0)
	if err != nil {
		return nil, err
	}

	var ordered []string
	for _, fn := range fns {
		if _, ok := fm[fn]; ok {
			ordered = append(ordered, filepath.Join(dp, fn))
		}
	}
	return ordered, nil
}
