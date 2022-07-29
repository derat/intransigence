// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package render

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

// Process inline/*.scss into inline/*.css.
//go:generate gen/gen_css.sh

// Generate an std_inline.go file that defines a map[string]string named stdInline.
//go:generate sh -c "go run gen/gen_filemap.go stdInline inline/*.css inline/*.js | gofmt -s >std_inline.go"

// SiteInfo specifies high-level information about the site.
type SiteInfo struct {
	// BaseURL is the base site URL with a trailing slash, e.g. "https://www.example.org/".
	BaseURL string `yaml:"base_url"`
	// TitleSuffix is appended to page titles that don't explicitly forbid it.
	TitleSuffix string `yaml:"title_suffix"`
	// DefaultDesc is used as the default meta description for pages.
	DefaultDesc string `yaml:"default_desc"`
	// ManifestPath contains the path to the web manifest file, e.g. "site.webmanifest".
	ManifestPath string `yaml:"manifest_path"`
	// AppleTouchIconPath is the path to the apple-touch-icon image.
	// Per https://realfavicongenerator.net/faq, this should be a 180x180 PNG named "apple-touch-icon.png".
	AppleTouchIconPath string `yaml:"apple_touch_icon_path"`
	// FaviconPaths contains paths to multiple favicon images in display order.
	// Per https://realfavicongenerator.net/faq, this should contain "favicon-32x32.png" and "favicon-16x16.png".
	FaviconPaths []string `yaml:"favicon_paths"`
	// NavText is displayed near the logo in the navigation area.
	NavText string `yaml:"nav_text"`
	// FeedTitle is used as the title for the RSS (Atom) feed.
	FeedTitle string `yaml:"feed_title"`
	// FeedDesc is used as the description for the RSS (Atom) feed.
	FeedDesc string `yaml:"feed_desc"`

	// LogoPathHTML is the image at the top of the page for non-AMP, e.g. "resources/logo-*.png".
	// The image's intrinsic dimensions for larger displays come from the smallest image.
	// The second-smallest image is used for mobile if it isn't a 2x version of the smallest image.
	LogoPathHTML string `yaml:"logo_path_html"`
	// LogoWidthHTML contains LogoPathHTML's width in pixels. Only needed for SVG.
	LogoWidthHTML int `yaml:"logo_width_html"`
	// LogoHeightHTML contains LogoPathHTML's height in pixels. Only needed for SVG.
	LogoHeightHTML int `yaml:"logo_height_html"`
	// LogoPathAMP is the AMP analogue to LogoPathHTML.
	LogoPathAMP string `yaml:"logo_path_amp"`
	// LogoWidthAMP contains LogoPathAMP's width in pixels. Only needed for SVG.
	LogoWidthAMP int `yaml:"logo_width_amp"`
	// LogoHeightAMP contains LogoPathAMP's height in pixels. Only needed for SVG.
	LogoHeightAMP int `yaml:"logo_height_amp"`
	// LogoAlt contains the alt text for the logo image.
	LogoAlt string `yaml:"logo_alt"`

	// NavTogglePath is the image used to expand or collapse the navbox on non-AMP mobile.
	NavTogglePath string `yaml:"nav_toggle_path"`
	// NavToggleWidth contains NavTogglePath's width in pixels. Only needed for SVG.
	NavToggleWidth int `yaml:"nav_toggle_width"`
	// NavToggleHeight contains NavTogglePath's height in pixels. Only needed for SVG.
	NavToggleHeight int `yaml:"nav_toggle_height"`

	// MenuButtonPath is the image used to show the menu on AMP.
	MenuButtonPath string `yaml:"menu_button_path"`
	// MenuButtonWidth contains MenuButtonPath's width in pixels. Only needed for SVG.
	MenuButtonWidth int `yaml:"menu_button_width"`
	// MenuButtonHeight contains MenuButtonPath's height in pixels. Only needed for SVG.
	MenuButtonHeight int `yaml:"menu_button_height"`

	// DarkButtonPath is the image used to toggle between light and dark mode.
	DarkButtonPath string `yaml:"dark_button_path"`
	// DarkButtonWidth contains DarkButtonPath's width in pixels. Only needed for SVG.
	DarkButtonWidth int `yaml:"dark_button_width"`
	// DarkButtonHeight contains DarkButtonPath's height in pixels. Only needed for SVG.
	DarkButtonHeight int `yaml:"dark_button_height"`

	// The following fields are used in structured data.
	AuthorName        string `yaml:"author_name"`
	AuthorEmail       string `yaml:"author_email"`
	PublisherName     string `yaml:"publisher_name"`
	PublisherLogoPath string `yaml:"publisher_logo_path"`

	// GoogleAnalyticsCode uniquely identifies the site, e.g. "UA-123456-1".
	// Google Analytics is only used for the AMP version of the page (typically served from a CDN),
	// and only if this field is non-empty.
	GoogleAnalyticsCode string `yaml:"google_analytics_code"`
	// GoogleMapsAPIKey is used for Google Maps API billing.
	// It is only needed if maps are embedded in the site.
	GoogleMapsAPIKey string `yaml:"google_maps_api_key"`
	// CloudflareAnalyticsToken identifies the site for Cloudflare Web Analytics,
	// e.g. "4d65822107fcfd524d65822107fcfd52". This is only used for the non-AMP version of the page,
	// and only if this field is non-empty.
	CloudflareAnalyticsToken string `yaml:"cloudflare_analytics_token"`
	// D3ScriptURL is the URL of the minified d3.js to use for graphs.
	// The CDN-hosted version of the file is used by default.
	D3ScriptURL string `yaml:"d3_script_url"`
	// ExtraStaticDirs contains extra dirs to copy into the output dir.
	// Keys are paths relative to the site dir and values are paths relative to the output dir.
	ExtraStaticDirs map[string]string `yaml:"extra_static_dirs"`

	// CodeStyleLight contains the Chroma style to use when highlighting code in the light theme.
	// See https://xyproto.github.io/splash/docs/all.html for available styles.
	CodeStyleLight string `yaml:"code_style_light"`
	// CodeStyleDark is like CodeStyleLight, but used for the dark theme.
	CodeStyleDark string `yaml:"code_style_dark"`

	// NavItems specifies the site's navigation hierarchy.
	NavItems []*NavItem `yaml:"nav_items"`

	// AppleTouchIconWidth and AppleTouchIconHeight are inferred from AppleTouchIconPath.
	AppleTouchIconWidth  int `yaml:"-"`
	AppleTouchIconHeight int `yaml:"-"`
	// Favicons is automatically generated from FaviconPaths.
	Favicons []faviconInfo `yaml:"-"`

	// CloudflareAnalyticsScriptURL is sourced for Cloudflare Web Analytics.
	CloudflareAnalyticsScriptURL string `yaml:"-"`
	// CloudflareAnalyticsConnectPattern is a CSP connect-src pattern matching connections
	// made by CloudflareAnalyticsScriptURL.
	CloudflareAnalyticsConnectPattern string `yaml:"-"`

	// CompressPages specifies whether a .html.gz, gzip-compressed file should be generated
	// alongside every .html file. This is useful for web hosts that don't automatically compress
	// pages when they are being served.
	CompressPages bool `yaml:"compress_pages"`

	// dir contains the path to the base site directory (i.e. containing the "pages" subdirectory).
	// It is assumed to be the directory that the SiteInfo was loaded from.
	dir string

	codeCSS string // CSS class definitions for code syntax highlighting
}

type faviconInfo struct {
	Path          string
	Width, Height int
}

// NewSiteInfo constructs a new SiteInfo from the YAML file at p.
func NewSiteInfo(p string) (*SiteInfo, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	si := SiteInfo{
		CodeStyleLight:                    "github",
		CodeStyleDark:                     "dracula",
		D3ScriptURL:                       "https://d3js.org/d3.v3.min.js",
		CloudflareAnalyticsScriptURL:      "https://static.cloudflareinsights.com/beacon.min.js",
		CloudflareAnalyticsConnectPattern: "https://cloudflareinsights.com",
		dir:                               filepath.Dir(p),
	}
	dec := yaml.NewDecoder(f)
	dec.KnownFields(true)
	if err := dec.Decode(&si); err != nil {
		return nil, err
	}

	ip := filepath.Join(si.PageDir(), "index.md")
	if _, err := os.Stat(ip); err != nil {
		return nil, err
	}

	if si.codeCSS, err = getCodeCSS(si.CodeStyleLight, ""); err != nil {
		return nil, err
	}
	if css, err := getCodeCSS(si.CodeStyleDark, "body.dark "); err != nil {
		return nil, err
	} else {
		si.codeCSS += css
	}

	if si.AppleTouchIconPath != "" {
		if si.AppleTouchIconWidth, si.AppleTouchIconHeight, err = imageSize(
			filepath.Join(si.StaticDir(), si.AppleTouchIconPath)); err != nil {
			return nil, fmt.Errorf("failed getting apple-touch-icon dimensions: %v", err)
		}
	}
	for _, fp := range si.FaviconPaths {
		width, height, err := imageSize(filepath.Join(si.StaticDir(), fp))
		if err != nil {
			return nil, fmt.Errorf("failed getting favicon dimensions: %v", err)
		}
		si.Favicons = append(si.Favicons, faviconInfo{fp, width, height})
	}

	return &si, nil
}

// ReadInline reads and returns the contents of the named file in si.InlineDir or si.InlineGenDir.
// It returns an empty string if the file does not exist and panics if the file cannot be read.
func (si *SiteInfo) ReadInline(fn string) string {
	b, err := ioutil.ReadFile(filepath.Join(si.InlineDir(), fn))
	if os.IsNotExist(err) {
		b, err = ioutil.ReadFile(filepath.Join(si.InlineGenDir(), fn))
	}
	if os.IsNotExist(err) {
		return ""
	} else if err != nil {
		panic(fmt.Sprint("Failed reading file: ", err))
	}
	min, err := minifyData(string(b), filepath.Ext(fn))
	if err != nil {
		panic(fmt.Sprintf("Failed minifying %v: %v", fn, err))
	}
	return min
}

// CheckStatic returns an error if p (e.g. "foo/bar.png") doesn't exist
// in si.StaticDir, si.StaticGenDir, or in the matching si.ExtraStaticDirs source dir.
func (si *SiteInfo) CheckStatic(p string) error {
	for src, dst := range si.ExtraStaticDirs {
		if p == dst || strings.HasPrefix(dst+"/", p) {
			_, err := os.Stat(filepath.Join(si.dir, src, p[len(dst):]))
			return err
		}
	}
	if _, err := os.Stat(filepath.Join(si.StaticGenDir(), p)); err == nil {
		return nil
	}
	_, err := os.Stat(filepath.Join(si.StaticDir(), p))
	return err
}

func (si *SiteInfo) InlineDir() string {
	return filepath.Join(si.dir, "inline")
}
func (si *SiteInfo) InlineGenDir() string {
	return filepath.Join(si.dir, "gen/inline")
}
func (si *SiteInfo) IframeDir() string {
	return filepath.Join(si.dir, "iframes")
}
func (si *SiteInfo) PageDir() string {
	return filepath.Join(si.dir, "pages")
}
func (si *SiteInfo) StaticDir() string {
	return filepath.Join(si.dir, "static")
}
func (si *SiteInfo) StaticGenDir() string {
	return filepath.Join(si.dir, "gen/static")
}
func (si *SiteInfo) TemplateDir() string {
	return filepath.Join(si.dir, "templates")
}

// AbsURL converts the supplied string into an absolute URL by appending it to si.BaseURL.
// Returns the unchanged string if it's already absolute.
// Returns an error if the URL is absolute but not prefixed by si.BaseURL.
func (si *SiteInfo) AbsURL(u string) (string, error) {
	ur, err := url.Parse(u)
	if err != nil {
		return "", fmt.Errorf("failed to parse %q: %v", u, err)
	}
	if ur.IsAbs() {
		if !strings.HasPrefix(u, si.BaseURL) {
			return "", fmt.Errorf("URL %q doesn't have prefix %q", u, si.BaseURL)
		}
		return u, nil
	}
	// Just link to the main URL instead of appending index.html.
	if u == IndexPage {
		return si.BaseURL, nil
	}
	return si.BaseURL + u, nil
}
