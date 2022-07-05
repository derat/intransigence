// Code generated by gen_filemap.go from 20fc399cdc2358bd8b5cca1ebc2004070f33f1b49aa5492a7022b829b9fa037b. DO NOT EDIT.

package render

var stdTemplates = map[string]string{
	"box.tmpl":         "{{/* Writes <section> for \"box\" code block. */}}\n{{define \"start\" -}}\n{{if eq .Level 1}}<div{{else}}<section{{end}} class=\"box\n{{- if .Narrow}} desktop-narrow{{end -}}\n\"\n{{- if .ID}} id=\"{{.ID}}\"{{end}}>\n  <h{{.Level}} class=\"title\">\n    {{- if .MapLabel}}<span class=\"location-label\">{{.MapLabel}}</span> {{end}}\n    {{- .Title -}}\n    {{- /* For non-AMP, a click handler is added on page load instead. */ -}}\n    {{- if .MapLabel}} (<a class=\"map-link\"{{if amp}} href=\"#map\"{{end}}>map</a>){{end}}\n  </h{{.Level}}>\n  <div class=\"body\">\n{{end}}\n\n{{/* Writes </section> for end of box created by \"box\" code block. */}}\n{{define \"end\" -}}\n  </div>\n{{if eq .Level 1}}</div>{{else}}</section>{{end}}\n{{end}}\n",
	"clear.tmpl":       "{{/* Writes empty <div> for \"clear\" code block. */}}\n<div class=\"clear\"></div>\n",
	"figure.tmpl":      "{{/* Writes <figure> for \"image\" and \"graph\" code blocks. */}}\n{{define \"figure_start\"}}\n<figure\n{{- if or .Align .Class .DesktopOnly .MobileOnly}} class=\"\n  {{- if eq .Align \"left\"}}left\n  {{- else if eq .Align \"right\"}}right\n  {{- else if eq .Align \"center\"}}center\n  {{- else if eq .Align \"desktop_left\"}}desktop-left mobile-center\n  {{- else if eq .Align \"desktop_right\"}}desktop-right mobile-center\n  {{- end -}}\n  {{- if .Class}} {{.Class}}{{end -}}\n  {{- if .DesktopOnly}} desktop-only{{end -}}\n  {{- if .MobileOnly}} mobile-only{{end -}}\n\"{{end}}>{{/**/ -}}\n{{end}}\n\n{{- /* Writes <figcaption></figcaption> and </figure> for \"image\" and \"graph\" code blocks. */}}\n{{define \"figure_end\" -}}\n{{if .Caption}}<figcaption>{{.Caption}}</figcaption>\n{{end -}}\n</figure>\n{{end}}\n",
	"graph.tmpl":       "{{/* Writes <figure> and <iframe> for \"graph\" code block. */ -}}\n{{template \"figure_start\" .}}\n{{- if amp}}<amp-iframe {{else}}<iframe {{end -}}\nclass=\"graph\" width={{.Width}} height={{.Height}}\n{{- if amp}} layout=\"responsive\" frameborder=\"0\" {{else}} {{end -}}\nsandbox=\"{{if not amp}}allow-same-origin {{end}}allow-scripts\" src=\"{{.Href}}?{{.Name}}\">\n{{- if amp}}</amp-iframe>{{else}}</iframe>{{end}}\n{{template \"figure_end\" .}}\n",
	"graph_page.tmpl":  "{{/* Writes graph iframe page. */ -}}\n<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n  <meta charset=\"utf-8\">\n  {{.CSPMeta}}\n  <meta name=\"robots\" content=\"noindex, nofollow\">\n  <title>graph</title>\n{{- range .ScriptURLs}}\n  <script src=\"{{.}}\"></script>\n{{end}}\n{{- range .InlineScripts}}\n  <script>{{.}}</script>\n{{end}}\n  <style>{{.InlineStyle}}</style>\n</head>\n<body>\n  <a id=\"graph-node\"></a>\n</body>\n</html>\n",
	"image_block.tmpl": "{{/* Writes <figure> and <img> for \"image\" code block. */ -}}\n{{template \"figure_start\" .}}\n{{if .Href}}<a href=\"{{.Href}}\">{{end -}}\n{{template \"img\" .}}\n{{- if .Href}}</a>{{end}}\n{{template \"figure_end\" .}}\n",
	"img.tmpl":         "{{/* Writes an image using either the amp-img or nonamp-img template. */}}\n{{define \"img\" -}}\n{{if amp}}{{template \"amp-img\" .}}{{else}}{{template \"nonamp-img\" .}}{{end -}}\n{{end}}\n\n{{/* Writes a <picture> containing the regular and fallback images, possibly wrapped\n     in a <span> with a thumbnail placeholder. Setting the background-image property\n     on the real <img> would far simpler, but we'd need to use inline 'style'\n     attributes to do that, which is forbidden by CSP. Using an <svg> lets us\n     just set its image's href attribute and also gives us more control over the blur\n     effect than a separate placeholder <img> with the CSS filter property. */}}\n{{define \"nonamp-img\" -}}\n{{if .ThumbSrc -}}\n<span class=\"img-wrapper\">{{/**/ -}}\n<svg width=\"100%\" height=\"100%\" viewBox=\"0 0 {{.Width}} {{.Height}}\">{{/**/ -}}\n  {{/* The ID namespace is unfortunately shared across all SVG images on the page,\n       so only define it in the first image that uses it. */ -}}\n  {{if .DefineThumbFilter -}}\n  <filter id=\"thumb-filter\">\n    <feGaussianBlur stdDeviation=\"12\"/>\n    {{/* Keep edges at full opacity: https://stackoverflow.com/a/24420004/6882947 */ -}}\n    <feComponentTransfer><feFuncA type=\"discrete\" tableValues=\"1 1\"/></feComponentTransfer>\n  </filter>{{/**/ -}}\n  {{end -}}\n  <image href=\"{{.ThumbSrc}}\" width=\"{{.Width}}\" height=\"{{.Height}}\" {{/**/ -}}\n      filter=\"url(#thumb-filter)\" preserveAspectRatio=\"none\"/>{{/**/ -}}\n</svg>\n{{- end -}}\n<picture>{{/**/ -}}\n  {{if .FallbackSrc -}}\n  <source type=\"image/webp\" sizes=\"{{.Sizes}}\" srcset=\"{{.Srcset}}\">{{/**/ -}}\n  {{end -}}\n  <img {{if .ID}}id=\"{{.ID}}\" {{end}}{{range .Attr}}{{.}} {{end -}}\n      {{if .Classes}}class=\"{{range .Classes}}{{.}} {{end}}\" {{end -}}\n      src=\"{{or .FallbackSrc .Src}}\" {{/**/ -}}\n      {{if .Sizes}}sizes=\"{{.Sizes}}\" {{end -}}\n      {{if .Srcset}}srcset=\"{{or .FallbackSrcset .Srcset}}\" {{end -}}\n      width=\"{{.Width}}\" height=\"{{.Height}}\" alt=\"{{.Alt}}\">{{/**/ -}}\n</picture>{{/**/ -}}\n{{if .ThumbSrc}}</span>{{end -}}\n{{end}}\n\n{{/* Writes <amp-img></amp-img> and a fallback (and maybe a thumbnail placeholder). */}}\n{{define \"amp-img\" -}}\n<amp-img {{if .ID}}id=\"{{.ID}}\" {{end}}{{range .Attr}}{{.}} {{end -}}\n    {{if .Classes}}class=\"{{range .Classes}}{{.}} {{end}}\" {{end -}}\n    src=\"{{.Src}}\" {{/**/ -}}\n    {{if .Sizes}}sizes=\"{{.Sizes}}\" {{end -}}\n    {{if .Srcset}}srcset=\"{{.Srcset}}\" {{end -}}\n    width=\"{{.Width}}\" height=\"{{.Height}}\" alt=\"{{.Alt}}\">{{/**/ -}}\n{{if .FallbackSrc -}}\n<amp-img fallback {{range .Attr}}{{.}} {{end -}}\n    {{if .Classes}}class=\"{{range .Classes}}{{.}} {{end}}\" {{end -}}\n    src=\"{{.FallbackSrc}}\" sizes=\"{{.Sizes}}\" srcset=\"{{.FallbackSrcset}}\" {{/**/ -}}\n    width=\"{{.Width}}\" height=\"{{.Height}}\" alt=\"{{.Alt}}\"></amp-img>{{/**/ -}}\n{{end -}}\n{{if .ThumbSrc -}}\n<amp-img placeholder {{range .Attr}}{{.}} {{end -}}\n    class=\"thumb{{range .Classes}} {{.}}{{end}}\" {{/**/ -}}\n    src=\"{{.ThumbSrc}}\" width=\"{{.Width}}\" height=\"{{.Height}}\" {{/**/ -}}\n    alt=\"{{.Alt}}\"></amp-img>{{/**/ -}}\n{{end -}}\n</amp-img>{{/**/ -}}\n{{end}}\n",
	"map.tmpl":         "{{/* Writes <iframe></iframe> for \"map\" code block. */ -}}\n<div class=\"mapbox\">\n  {{if amp}}<amp-iframe {{else}}<iframe {{end -}}\n  id=\"map\" width=\"{{.Width}}\" height=\"{{.Height}}\" {{/**/ -}}\n  {{if amp}}layout=\"responsive\" frameborder=\"0\" {{end -}}\n  referrerpolicy=\"unsafe-url\" {{/* referrer used by iframe to construct links */ -}}\n  sandbox=\"{{if not amp}}allow-same-origin {{end}}allow-scripts allow-top-navigation\" {{/**/ -}}\n  src=\"{{.Href}}\">{{/**/ -}}\n  {{if amp}}\n  {{template \"img\" .}}\n  {{end}}\n  {{if amp}}</amp-iframe>{{else}}</iframe>{{end}}\n</div>\n",
	"map_page.tmpl":    "{{/* Writes map iframe page. */ -}}\n<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n  <meta charset=\"utf-8\">\n  <meta name=\"robots\" content=\"noindex, nofollow\">\n  <title>map</title>\n{{- range .ScriptURLs}}\n  <script src=\"{{.}}\"></script>\n{{end}}\n{{- range .InlineScripts}}\n  <script>{{.}}</script>\n{{end}}\n  <style>{{.InlineStyle}}</style>\n</head>\n<body>\n  <div id=\"map-div\"></div>\n</body>\n</html>\n",
	"page.tmpl":        "{{/* Writes the top of a normal (AMP or non-AMP) page. */}}\n{{define \"start\" -}}\n<!DOCTYPE html>\n{{/* TODO: Make language configurable. */ -}}\n<html {{if amp}}amp {{end}}lang=\"en\">\n  <head>\n    <meta charset=\"utf-8\">\n    <link rel=\"{{.LinkRel}}\" href=\"{{.LinkHref}}\">\n    <link rel=\"alternate\" type=\"application/atom+xml\" href=\"{{.FeedHref}}\">\n    {{.CSPMeta}}\n    <meta name=\"viewport\" content=\"width=device-width, initial-scale=1, minimum-scale=1\">\n    <meta name=\"description\" content=\"{{.Desc}}\">\n    <meta name=\"robots\" content=\"NOODP\">\n\n    <title>{{.FullTitle}}</title>\n\n    {{if .SiteInfo.FaviconPath -}}\n    <link rel=\"apple-touch-icon-precomposed\" href=\"{{.SiteInfo.FaviconPath}}\">\n    <link rel=\"icon\" sizes=\"{{.SiteInfo.FaviconWidth}}x{{.SiteInfo.FaviconHeight}}\" href=\"{{.SiteInfo.FaviconPath}}\">\n    {{end -}}\n\n    <script type=\"application/ld+json\">{{.StructData}}</script>\n    {{if amp}}\n      <style amp-boilerplate>{{.AMPStyle}}</style>\n      <noscript><style amp-boilerplate>{{.AMPNoscriptStyle}}</style></noscript>\n      <style amp-custom>{{.AMPCustomStyle}}</style>\n      <script async custom-element=\"amp-sidebar\" src=\"https://cdn.ampproject.org/v0/amp-sidebar-0.1.js\"></script>\n      {{if or .HasGraph .HasMap -}}\n      <script async custom-element=\"amp-iframe\" src=\"https://cdn.ampproject.org/v0/amp-iframe-0.1.js\"></script>\n      {{end -}}\n      {{if .SiteInfo.GoogleAnalyticsCode -}}\n      <script async custom-element=\"amp-analytics\" src=\"https://cdn.ampproject.org/v0/amp-analytics-0.1.js\"></script>\n      {{end -}}\n      <script async src=\"https://cdn.ampproject.org/v0.js\"></script>\n    {{else}}{{/* non-AMP */}}\n      <style>{{.HTMLStyle}}</style>\n      {{range .HTMLScripts}}<script>{{.}}</script>\n      {{end -}}\n    {{end}}\n  </head>\n\n  <body{{if amp}} data-amp-auto-lightbox-disable data-prefers-dark-mode-class=\"dark\"{{end}}>\n    {{if amp}}{{template \"header_amp\" .}}{{else}}{{template \"header_html\" .}}{{end}}\n    <main>\n{{end}}\n\n{{/* Writes start-of-<body> data for non-AMP pages. */}}\n{{/* For desktop and responsive mobile, the logo and navbox are at the top of the page. */}}\n{{define \"header_html\"}}\n<script>{{.HTMLBodyScript}}</script>\n<header>\n  <nav>\n    {{template \"img\" .LogoHTML}}\n    {{/* This mirrors the box_header and box_footer templates. */ -}}\n    <div class=\"box\">\n      <div class=\"title\">\n        {{.SiteInfo.NavText}}\n        {{template \"img\" .NavToggle}}\n      </div>\n      <div class=\"body\">\n        {{/* On mobile, collapse the navbox if the page isn't the index and doesn't have subpages. */ -}}\n        <ul{{- if and (not .NavItem.IsIndex) (not .NavItem.VisibleChildren)}} class=\"collapsed-mobile\"{{end}}>\n          {{range .SiteInfo.NavItems}}{{template \"nav_item\" .}}{{end}}\n        </ul>\n      </div>\n    </div>\n  </nav>\n  {{/* Outside <nav> so it can have its own positioning. */ -}}\n  {{template \"img\" .DarkButton}}\n</header>\n{{end}}\n\n{{/* Writes start-of-<body> data for AMP pages. */}}\n{{/* For AMP, just the logo and a menu button go at the top. The navbox ends up in a sidebar. */}}\n{{define \"header_amp\"}}\n{{/* The validator barfs if the <amp-analytics> <script> tag doesn't have the \"type\" attribute. */ -}}\n{{if .SiteInfo.GoogleAnalyticsCode -}}\n<amp-analytics type=\"googleanalytics\">\n  <script type=\"application/json\">\n    {\n      \"vars\": {\n        \"account\": \"{{.SiteInfo.GoogleAnalyticsCode}}\"\n      },\n      \"triggers\": {\n        \"trackPageview\": {\n          \"on\": \"visible\",\n          \"request\": \"pageview\"\n        }\n      }\n    }\n  </script>\n</amp-analytics>\n{{end -}}\n\n<amp-sidebar id=\"sidebar\" layout=\"nodisplay\" side=\"right\">\n  {{/* This mirrors the box_header and box_footer templates. */ -}}\n  <nav>\n    <div class=\"box\">\n      <div class=\"title\">\n        {{.SiteInfo.NavText}}\n      </div>\n      <div class=\"body\">\n        <ul>\n          {{range .SiteInfo.NavItems}}{{template \"nav_item\" .}}{{end}}\n        </ul>\n      </div>\n    </div>\n  </nav>\n</amp-sidebar>\n\n<header>\n  {{template \"img\" .LogoAMP}}\n  <div class=\"spacer\"></div>\n  {{template \"img\" .DarkButton}}\n  {{template \"img\" .MenuButton}}\n</header>\n{{end}}\n\n{{/* Writes the bottom of a normal page. */}}\n{{define \"end\" -}}\n    </main>\n    {{if or (not .HideBackToTop) (and (not .HideDates) (or .Created .Modified)) -}}\n    <footer>\n      {{/* TODO: Make text configurable. */ -}}\n      {{if not .HideBackToTop}}<div class=\"back-to-top\"><a href=\"#top\">Back to top</a></div>{{end}}\n      {{if not .HideDates}}<div class=\"dates\">\n        {{if .Created}}<div class=\"created\">Page created in {{/**/ -}}\n          <time datetime=\"{{formatDate .Created \"2006\"}}\">{{formatDate .Created \"2006\"}}</time>.</div>{{end}}\n        {{if .Modified}}<div class=\"modified\">Last modified {{/**/ -}}\n          <time datetime=\"{{formatDate .Modified \"2006-01-02\"}}\">{{formatDate .Modified \"Jan 2, 2006\"}}</time>.</div>{{end}}\n      </div>{{end}}\n    </footer>{{/**/ -}}\n    {{end}}\n    {{if and .SiteInfo.CloudflareAnalyticsToken (not amp)}}<!-- Cloudflare Web Analytics --><script defer src=\"{{.SiteInfo.CloudflareAnalyticsScriptURL}}\" data-cf-beacon=\"{&quot;token&quot;:&quot;{{.SiteInfo.CloudflareAnalyticsToken}}&quot;}\"></script><!-- End Cloudflare Web Analytics -->\n    {{end}}\n  </body>\n</html>\n{{end}}\n\n{{/* Writes an <li> for a navigation item and its children. */}}\n{{define \"nav_item\" -}}\n<li>\n{{- if .HasID current.ID}}<span class=\"selected\">{{.Name}}</span>\n{{- else}}<a href=\"{{if amp}}{{.AMPURL}}{{else}}{{.URL}}{{end}}\">{{.Name}}</a>\n{{- end}}\n{{- if and .VisibleChildren (.FindID current.ID) (not current.OmitFromMenu)}}\n<ul>\n{{range .VisibleChildren}}{{template \"nav_item\" .}}{{end}}\n</ul>\n{{end -}}\n</li>\n{{end}}\n"}
