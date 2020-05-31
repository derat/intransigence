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

// Process inline/*.scss into inline/*.css and inline/*.js into inline/*.js.min.
//go:generate gen/gen_css.sh
//go:generate gen/gen_js_min.sh

// Generate an std_inline.go file that defines a map[string]string named stdInline.
//go:generate sh -c "go run gen/gen_filemap.go stdInline inline/*.css inline/*.js.min | gofmt -s >std_inline.go"

// SiteInfo specifies high-level information about the site.
type SiteInfo struct {
	// BaseURL is the base site URL with a trailing slash, e.g. "https://www.example.org/".
	BaseURL string `yaml:"base_url"`
	// TitleSuffix is appended to page titles that don't explicitly forbid it.
	TitleSuffix string `yaml:"title_suffix"`
	// DefaultDesc is used as the default meta description for pages.
	DefaultDesc string `yaml:"default_desc"`
	// FaviconPath is the path to the favicon image, e.g. "resources/favicon.png".
	FaviconPath string `yaml:"favicon_path"`
	// NavText is displayed near the logo in the navigation area.
	NavText string `yaml:"nav_text"`

	// LogoPathHTML is the image at the top of the page for non-AMP, e.g. "resources/logo-*.png".
	// The image's intrinsic dimensions for larger displays come from the smallest image.
	// The second-smallest image is used for mobile if it isn't a 2x version of the smallest image.
	LogoPathHTML string `yaml:"logo_path_html"`
	// LogoPathAMP is the AMP analogue to LogoPathHTML.
	LogoPathAMP string `yaml:"logo_path_amp"`
	// LogoAlt contains the alt text for the logo image.
	LogoAlt string `yaml:"logo_alt"`
	// NavTogglePath is the image used to expand or collapse the navbox on non-AMP mobile.
	NavTogglePath string `yaml:"nav_toggle_path"`
	// MenuButtonPath is the image used to show the menu on AMP.
	MenuButtonPath string `yaml:"menu_button_path"`

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
	// D3ScriptURL is the URL of the minified d3.js to use for graphs.
	// The CDN-hosted version of the file is used by default.
	D3ScriptURL string `yaml:"d3_script_url"`
	// ExtraStaticDirs contains extra dirs to copy into the output dir.
	// Keys are paths relative to the site dir and values are paths relative to the output dir.
	ExtraStaticDirs map[string]string `yaml:"extra_static_dirs"`

	// NavItems specifies the site's navigation hierarchy.
	NavItems []*NavItem `yaml:"nav_items"`

	// FaviconWidth and FaviconHeight are automatically inferred from FaviconPath.
	FaviconWidth  int `yaml:"-"`
	FaviconHeight int `yaml:"-"`

	// dir contains the path to the base site directory (i.e. containing the "pages" subdirectory).
	// It is assumed to be the directory that the SiteInfo was loaded from.
	dir string
}

// NewSiteInfo constructs a new SiteInfo from the YAML file at p.
func NewSiteInfo(p string) (*SiteInfo, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	si := SiteInfo{dir: filepath.Dir(p)}
	dec := yaml.NewDecoder(f)
	dec.KnownFields(true)
	if err := dec.Decode(&si); err != nil {
		return nil, err
	}

	ip := filepath.Join(si.PageDir(), "index.md")
	if _, err := os.Stat(ip); err != nil {
		return nil, err
	}

	if si.D3ScriptURL == "" {
		si.D3ScriptURL = "https://d3js.org/d3.v3.min.js"
	}

	if si.FaviconPath != "" {
		if si.FaviconWidth, si.FaviconHeight, err = imageSize(
			filepath.Join(si.StaticDir(), si.FaviconPath)); err != nil {
			return nil, fmt.Errorf("failed getting favicon dimensions: %v", err)
		}
	}

	return &si, nil
}

// ReadInline reads and returns the contents of the named file in the inline dir.
// It returns an empty string if the file does not exist and panics if the file cannot be read.
func (si *SiteInfo) ReadInline(fn string) string {
	b, err := ioutil.ReadFile(filepath.Join(si.InlineDir(), fn))
	if os.IsNotExist(err) {
		return ""
	} else if err != nil {
		panic(fmt.Sprint("Failed reading file: ", err))
	}
	return strings.TrimSpace(string(b)) // sassc doesn't remove trailing newline
}

// CheckStatic returns an error if p (e.g. "foo/bar.png") doesn't exist
// in si.StaticDir or in the matching si.ExtraStaticDirs source dir.
func (si *SiteInfo) CheckStatic(p string) error {
	for src, dst := range si.ExtraStaticDirs {
		if p == dst || strings.HasPrefix(dst+"/", p) {
			_, err := os.Stat(filepath.Join(si.dir, src, p[len(dst):]))
			return err
		}
	}
	_, err := os.Stat(filepath.Join(si.StaticDir(), p))
	return err
}

func (si *SiteInfo) InlineDir() string {
	return filepath.Join(si.dir, "inline")
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
