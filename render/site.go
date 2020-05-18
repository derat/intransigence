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

// SiteInfo specifies high-level information about the site.
type SiteInfo struct {
	// BaseURL is the base site URL with a trailing slash, e.g. "https://www.example.org/".
	BaseURL string `yaml:"base_url"`
	// TitleSuffix is appended to page titles that don't explicitly forbid it.
	TitleSuffix string `yaml:"title_suffix"`
	// DefaultDesc is used as the default meta description for pages.
	DefaultDesc string `yaml:"default_desc"`
	// NavText is displayed near the logo in the navigation area.
	NavText string `yaml:"nav_text"`

	// LogoPrefix, LogoSuffix, and LogoPath specify the image displayed near the top of the page.
	LogoPrefix string `yaml:"logo_prefix"`
	LogoSuffix string `yaml:"logo_suffix"`
	LogoPath   string `yaml:"logo_path"`
	// LogoAlt contains the alt text for the logo.
	LogoAlt string `yaml:"logo_alt"`
	// NavTogglePath is the path to the image used to expand or collapse the navbox on non-AMP mobile.
	NavTogglePath string `yaml:"nav_toggle_path"`
	// MenuButtonPath is the path to the image used to show the menu on AMP.
	MenuButtonPath string `yaml:"menu_button_path"`

	// The following fields are used in structured data.
	AuthorName          string `yaml:"author_name"`
	AuthorEmail         string `yaml:"author_email"`
	PublisherName       string `yaml:"publisher_name"`
	PublisherLogoURL    string `yaml:"publisher_logo_url"`
	PublisherLogoWidth  int    `yaml:"publisher_logo_width"`
	PublisherLogoHeight int    `yaml:"publisher_logo_height"`

	// GoogleAnalyticsCode uniquely identifies the site, e.g. "UA-123456-1".
	GoogleAnalyticsCode string `yaml:"google_analytics_code"`
	// GoogleMapsAPIKey is used for Google Maps API billing.
	GoogleMapsAPIKey string `yaml:"google_maps_api_key"`
	// D3ScriptURL is the URL of the minified d3.js to use for graphs.
	D3ScriptURL string `yaml:"d3_script_url"`
	// ExtraStaticDirs contains extra dirs to copy into the output dir.
	// Keys are paths relative to the site dir and values are paths relative to the output dir.
	ExtraStaticDirs map[string]string `yaml:"extra_static_dirs"`

	// NavItems specifies the site's navigation hierarchy.
	NavItems []*NavItem `yaml:"nav_items"`

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
	if err := yaml.NewDecoder(f).Decode(&si); err != nil {
		return nil, err
	}

	ip := filepath.Join(si.PageDir(), "index.md")
	if _, err := os.Stat(ip); err != nil {
		return nil, err
	}

	return &si, nil
}

// ReadInline reads and returns the contents of the named file in the inline dir.
// It panics if the file cannot be read.
func (si *SiteInfo) ReadInline(fn string) string {
	b, err := ioutil.ReadFile(filepath.Join(si.InlineDir(), fn))
	if err != nil {
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
	return si.BaseURL + u, nil
}
