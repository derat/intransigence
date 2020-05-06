// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package render

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
)

type SiteInfo struct {
	// Dir contains the path to the base site directory (i.e. containing the "pages" subdirectory).
	Dir string
	// BaseURL is the base site URL with a trailing slash, e.g. "https://www.example.org/".
	BaseURL string
	// TitleSuffix is appended to page titles that don't explicitly forbid it.
	TitleSuffix string
	// DefaultDesc is used as the default meta description for pages.
	DefaultDesc string

	// The following fields are used in structured data.
	AuthorName          string
	AuthorEmail         string
	PublisherName       string
	PublisherLogoURL    string
	PublisherLogoWidth  int
	PublisherLogoHeight int

	// GoogleAnalyticsCode uniquely identifies the site, e.g. "UA-123456-1".
	GoogleAnalyticsCode string
	// GoogleMapsAPIKey is used for Google Maps API billing.
	GoogleMapsAPIKey string
	// D3ScriptURL is the URL of the minified d3.js to use for graphs.
	D3ScriptURL string
}

func (si *SiteInfo) ReadInline(fn string) string {
	b, err := ioutil.ReadFile(filepath.Join(si.Dir, "inline", fn))
	if err != nil {
		panic(fmt.Sprint("Failed reading file: ", err))
	}
	return string(b)
}

func (si *SiteInfo) IframeDir() string {
	return filepath.Join(si.Dir, "iframes")
}
func (si *SiteInfo) PageDir() string {
	return filepath.Join(si.Dir, "pages")
}
func (si *SiteInfo) StaticDir() string {
	return filepath.Join(si.Dir, "static")
}
func (si *SiteInfo) TemplateDir() string {
	return filepath.Join(si.Dir, "templates")
}
