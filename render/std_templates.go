// Code generated by gen_filemap.go from a11f17027326f1870668640689dd302424bc56d79b9c5088e811ac68420327b0. DO NOT EDIT.

package render

var stdTemplates = map[string]string{
	"box.tmpl":         "{{/* Writes <div> for \"box\" code block. */}}\n{{define \"start\" -}}\n<div class=\"box content\n{{- if .DesktopOnly}} desktop-only{{end}}\n{{- if .MobileOnly}} mobile-only{{end}}\n{{- if .Narrow}} desktop-narrow{{end -}}\n\"\n{{- if .ID}} id=\"{{.ID}}\"{{end}}>\n  <div class=\"title\">\n    {{- if .MapLabel}}<span class=\"location-label\">{{.MapLabel}}</span> {{end}}\n    {{- .Title -}}\n    {{- /* For non-AMP, a click handler is added on page load instead. */ -}}\n    {{- if .MapLabel}} (<a class=\"map-link\"{{if amp}} href=\"#map\"{{end}}>map</a>){{end}}\n  </div>\n  <div class=\"body\">\n{{end}}\n\n{{/* Writes </div> for end of box created by \"box\" code block. */}}\n{{define \"end\" -}}\n  </div>\n</div>\n{{end}}\n",
	"clear.tmpl":       "{{/* Writes empty <div> for \"clear\" code block. */}}\n<div class=\"clear\"></div>\n",
	"figure.tmpl":      "{{/* Writes <figure> for \"image\" and \"graph\" code blocks. */}}\n{{define \"figure_start\"}}\n<figure\n{{- if or .Align .Class .DesktopOnly .MobileOnly}} class=\"\n  {{- if eq .Align \"left\"}}left\n  {{- else if eq .Align \"right\"}}right\n  {{- else if eq .Align \"center\"}}center\n  {{- else if eq .Align \"desktop_left\"}}desktop-left mobile-center\n  {{- else if eq .Align \"desktop_right\"}}desktop-right mobile-center\n  {{- end -}}\n  {{- if .Class}} {{.Class}}{{end -}}\n  {{- if .DesktopOnly}} desktop-only{{end -}}\n  {{- if .MobileOnly}} mobile-only{{end -}}\n\"{{end}}>\n{{end}}\n\n{{- /* Writes <figcaption></figcaption> and </figure> for \"image\" and \"graph\" code blocks. */}}\n{{define \"figure_end\" -}}\n{{if .Caption}}<figcaption>{{.Caption}}</figcaption>{{end}}\n</figure>\n{{end}}\n",
	"graph.tmpl":       "{{/* Writes <figure> and <iframe> for \"graph\" code block. */ -}}\n{{template \"figure_start\" .}}\n{{- if amp}}<amp-iframe {{else}}<iframe {{end -}}\nclass=\"graph\" width={{.Width}} height={{.Height}}\n{{- if amp}} layout=\"responsive\" frameborder=\"0\" {{else}} {{end -}}\nsandbox=\"allow-scripts\" src=\"{{.Href}}?{{.Name}}\">\n{{- if amp}}</amp-iframe>{{else}}</iframe>{{end}}\n{{template \"figure_end\" .}}\n",
	"graph_page.tmpl":  "{{/* Writes graph iframe page. */ -}}\n<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n  <meta charset=\"utf-8\">\n  {{.CSPMeta}}\n  <meta name=\"robots\" content=\"noindex, nofollow\">\n  <title>graph</title>\n{{- range .ScriptURLs}}\n  <script src=\"{{.}}\"></script>\n{{end}}\n{{- range .InlineScripts}}\n  <script>{{.}}</script>\n{{end}}\n  <style>{{.InlineStyle}}</style>\n</head>\n<body>\n  <a id=\"graph-node\"></a>\n</body>\n</html>\n",
	"image_block.tmpl": "{{/* Writes <figure> and <img> for \"image\" code block. */ -}}\n{{template \"figure_start\" .}}\n{{if .Href}}<a href=\"{{.Href}}\">{{end}}\n{{template \"img\" .}}\n{{if .Href}}</a>{{end}}\n{{template \"figure_end\" .}}\n",
	"img.tmpl":         "{{/* Writes <picture> containing <source> and <img>. */}}\n{{/* If rendering for AMP, calls the amp-img template instead. */}}\n{{define \"img\" -}}\n{{if amp}}{{template \"amp-img\" .}}{{else -}}\n<picture>{{/**/ -}}\n  {{if .WebPSrc -}}\n  <source type=\"image/webp\" sizes=\"{{.Sizes}}\" srcset=\"{{.WebPSrcset}}\">{{/**/ -}}\n  {{end -}}\n  <img {{if .ID}}id=\"{{.ID}}\" {{end}}{{range .Attr}}{{.}} {{end}}{{/**/ -}}\n      src=\"{{.Src}}\" sizes=\"{{.Sizes}}\" srcset=\"{{.Srcset}}\" {{/**/ -}}\n      width=\"{{.Width}}\" height=\"{{.Height}}\" alt=\"{{.Alt}}\">{{/**/ -}}\n</picture>{{end -}}\n{{end}}\n\n{{/* Writes <amp-img></amp-img> and a fallback. */}}\n{{define \"amp-img\" -}}\n{{if .WebPSrc -}}\n<amp-img {{if .ID}}id=\"{{.ID}}\" {{end}}{{range .Attr}}{{.}} {{end}}{{/**/ -}}\n    src=\"{{.WebPSrc}}\" sizes=\"{{.Sizes}}\" srcset=\"{{.WebPSrcset}}\" {{/**/ -}}\n    width=\"{{.Width}}\" height=\"{{.Height}}\" alt=\"{{.Alt}}\">{{/**/ -}}\n{{end -}}\n<amp-img {{if .WebPSrc}}fallback {{end}}{{range .Attr}}{{.}} {{end}}{{/**/ -}}\n    src=\"{{.Src}}\" sizes=\"{{.Sizes}}\" srcset=\"{{.Srcset}}\" {{/**/ -}}\n    width=\"{{.Width}}\" height=\"{{.Height}}\" alt=\"{{.Alt}}\"></amp-img>{{/**/ -}}\n</amp-img>{{/**/ -}}\n{{end}}\n",
	"map.tmpl":         "{{/* Writes <iframe></iframe> for \"map\" code block. */ -}}\n<div class=\"mapbox\">\n  {{if amp}}<amp-iframe {{else}}<iframe {{end -}}\n  id=\"map\" width=\"{{.Width}}\" height=\"{{.Height}}\" {{/**/ -}}\n  {{if amp}}layout=\"responsive\" frameborder=\"0\" {{end -}}\n  referrerpolicy=\"unsafe-url\" {{/* referrer used by iframe to construct links */ -}}\n  sandbox=\"{{if not amp}}allow-same-origin {{end}}allow-scripts allow-top-navigation\" {{/**/ -}}\n  src=\"{{.Href}}\">{{/**/ -}}\n  {{if amp}}\n  {{template \"img\" .}}\n  {{end}}\n  {{if amp}}</amp-iframe>{{else}}</iframe>{{end}}\n</div>\n",
	"map_page.tmpl":    "{{/* Writes map iframe page. */ -}}\n<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n  <meta charset=\"utf-8\">\n  <meta name=\"robots\" content=\"noindex, nofollow\">\n  <title>map</title>\n{{- range .ScriptURLs}}\n  <script src=\"{{.}}\"></script>\n{{end}}\n{{- range .InlineScripts}}\n  <script>{{.}}</script>\n{{end}}\n  <style>{{.InlineStyle}}</style>\n</head>\n<body>\n  <div id=\"map-div\"></div>\n</body>\n</html>\n",
	"page.tmpl":        "{{/* Writes the top of a normal (AMP or non-AMP) page. */}}\n{{define \"start\" -}}\n<!DOCTYPE html>\n{{/* TODO: Make language configurable. */ -}}\n<html {{if amp}}amp {{end}}lang=\"en\">\n  <head>\n    <meta charset=\"utf-8\">\n    <link rel=\"{{.LinkRel}}\" href=\"{{.LinkHref}}\">\n    <link rel=\"alternate\" type=\"application/atom+xml\" href=\"{{.FeedHref}}\">\n    {{.CSPMeta}}\n    <meta name=\"viewport\" content=\"width=device-width, initial-scale=1, minimum-scale=1\">\n    <meta name=\"description\" content=\"{{.Desc}}\">\n    <meta name=\"robots\" content=\"NOODP\">\n\n    <title>{{.FullTitle}}</title>\n\n    {{if .SiteInfo.FaviconPath -}}\n    <link rel=\"apple-touch-icon-precomposed\" href=\"{{.SiteInfo.FaviconPath}}\">\n    <link rel=\"icon\" sizes=\"{{.SiteInfo.FaviconWidth}}x{{.SiteInfo.FaviconHeight}}\" href=\"{{.SiteInfo.FaviconPath}}\">\n    {{end -}}\n\n    <script type=\"application/ld+json\">{{.StructData}}</script>\n    {{if amp}}\n      <style amp-boilerplate>{{.AMPStyle}}</style>\n      <noscript><style amp-boilerplate>{{.AMPNoscriptStyle}}</style></noscript>\n      <style amp-custom>{{.AMPCustomStyle}}</style>\n      <script async custom-element=\"amp-sidebar\" src=\"https://cdn.ampproject.org/v0/amp-sidebar-0.1.js\"></script>\n      {{if or .HasGraph .HasMap -}}\n      <script async custom-element=\"amp-iframe\" src=\"https://cdn.ampproject.org/v0/amp-iframe-0.1.js\"></script>\n      {{end -}}\n      {{if .SiteInfo.GoogleAnalyticsCode -}}\n      <script async custom-element=\"amp-analytics\" src=\"https://cdn.ampproject.org/v0/amp-analytics-0.1.js\"></script>\n      {{end -}}\n      <script async src=\"https://cdn.ampproject.org/v0.js\"></script>\n    {{else}}{{/* non-AMP */}}\n      <style>{{.HTMLStyle}}</style>\n      {{range .HTMLScripts}}<script>{{.}}</script>\n      {{end -}}\n    {{end}}\n  </head>\n\n  <body id=\"top\">\n    {{if amp}}{{template \"header_amp\" .}}{{end}}\n    <div id=\"container\">\n    {{if not amp}}{{template \"header_html\" .}}{{end}}\n{{end}}\n\n{{/* Writes top-of-<body> data for non-AMP pages. */}}\n{{/* For desktop and responsive mobile, the logo and navbox are at the top of the page. */}}\n{{define \"header_html\"}}\n<header>\n  <nav id=\"nav-wrapper\">\n    {{template \"img\" .LogoHTML}}\n    {{/* This mirrors the box_header and box_footer templates. */ -}}\n    <div id=\"nav-box\" class=\"box\">\n      <div id=\"nav-title\" class=\"title\">\n        {{.SiteInfo.NavText}}\n        {{template \"img\" .NavToggle}}\n      </div>\n      <div class=\"body\">\n        {{/* On mobile, collapse the navbox if the page isn't the index and doesn't have subpages. */ -}}\n        <ul id=\"nav-list\"{{- if and (not .NavItem.IsIndex) (not .NavItem.VisibleChildren)}} class=\"collapsed-mobile\"{{end}}>\n          {{range .SiteInfo.NavItems}}{{template \"nav_item\" .}}{{end}}\n        </ul>\n      </div>\n    </div>\n  </nav>\n</header>\n{{end}}\n\n{{/* Writes start-of-<body> data for AMP pages. */}}\n{{/* For AMP, just the logo and a menu button go at the top. The navbox ends up in a sidebar. */}}\n{{define \"header_amp\"}}\n{{/* The validator barfs if the <amp-analytics> <script> tag doesn't have the \"type\" attribute. */ -}}\n{{if .SiteInfo.GoogleAnalyticsCode -}}\n<amp-analytics type=\"googleanalytics\">\n  <script type=\"application/json\">\n    {\n      \"vars\": {\n        \"account\": \"{{.SiteInfo.GoogleAnalyticsCode}}\"\n      },\n      \"triggers\": {\n        \"trackPageview\": {\n          \"on\": \"visible\",\n          \"request\": \"pageview\"\n        }\n      }\n    }\n  </script>\n</amp-analytics>\n{{end -}}\n\n<amp-sidebar id=\"sidebar\" layout=\"nodisplay\" side=\"right\">\n  {{/* This mirrors the box_header and box_footer templates. */ -}}\n  <nav id=\"nav-box\" class=\"box\">\n    <div id=\"nav-title\" class=\"title\">\n      {{.SiteInfo.NavText}}\n    </div>\n    <div class=\"body\">\n      <ul id=\"nav-list\">\n        {{range .SiteInfo.NavItems}}{{template \"nav_item\" .}}{{end}}\n      </ul>\n    </div>\n  </nav>\n</amp-sidebar>\n\n<header>\n  <div id=\"top-container\">\n    {{template \"img\" .LogoAMP}}\n    {{template \"img\" .MenuButton}}\n  </div>\n</header>\n{{end}}\n\n{{/* Writes the bottom of a normal page. */}}\n{{define \"end\" -}}\n      {{if or (not .HideBackToTop) (and (not .HideDates) (or .Created .Modified)) -}}\n      <footer>\n        {{/* TODO: Make text configurable. */ -}}\n        {{if not .HideBackToTop}}<div class=\"back-to-top\"><a href=\"#top\">Back to top</a></div>{{end}}\n        {{if not .HideDates}}<div class=\"dates\">\n          {{if .Created}}<div class=\"created\">Page created in {{formatDate .Created \"2006\"}}.</div>{{end}}\n          {{if .Modified}}<div class=\"modified\"> Last modified {{formatDate .Modified \"Jan 2, 2006\"}}.</div>{{end}}\n        </div>{{end}}\n      </footer>{{/**/ -}}\n      {{end}}\n    </div>{{/* #container */}}\n    {{if and .SiteInfo.CloudflareAnalyticsToken (not amp)}}<!-- Cloudflare Web Analytics --><script defer src=\"{{.SiteInfo.CloudflareAnalyticsScriptURL}}\" data-cf-beacon=\"{&quot;token&quot;:&quot;{{.SiteInfo.CloudflareAnalyticsToken}}&quot;}\"></script><!-- End Cloudflare Web Analytics -->\n    {{end}}\n  </body>\n</html>\n{{end}}\n\n{{/* Writes an <li> for a navigation item and its children. */}}\n{{define \"nav_item\" -}}\n<li>\n{{- if .HasID current.ID}}<span class=\"selected\">{{.Name}}</span>\n{{- else}}<a href=\"{{if amp}}{{.AMPURL}}{{else}}{{.URL}}{{end}}\">{{.Name}}</a>\n{{- end}}\n{{- if and .VisibleChildren (.FindID current.ID) (not current.OmitFromMenu)}}\n<ul>\n{{range .VisibleChildren}}{{template \"nav_item\" .}}{{end}}\n</ul>\n{{end -}}\n</li>\n{{end}}\n"}
