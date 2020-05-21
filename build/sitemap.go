// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package build

import (
	"encoding/xml"
	"io"
	"os"

	"github.com/derat/plaingen/render"
)

// writeSitemap generates a sitemap listing all normal pages in si and writes it to path p.
func writeSitemap(p string, si *render.SiteInfo) error {
	var sm sitemap
	url, err := si.AbsURL(render.IndexPage)
	if err != nil {
		return err
	}
	sm.add(url, changesWeekly)

	// Recursively walk nav items and add all pages.
	var add func(it *render.NavItem) error
	add = func(it *render.NavItem) error {
		if it.IsBarePage() {
			url, err := si.AbsURL(it.URL)
			if err != nil {
				return err
			}
			sm.add(url, changesWeekly)
		}
		for _, c := range it.Children {
			if err := add(c); err != nil {
				return err
			}
		}
		return nil
	}
	for _, it := range si.NavItems {
		if err := add(it); err != nil {
			return err
		}
	}

	f, err := os.Create(p)
	if err != nil {
		return err
	}
	if err := sm.write(f); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

// sitemap holds the contents of a sitemap XML document.
type sitemap struct {
	XMLName string `xml:"http://www.sitemaps.org/schemas/sitemap/0.9 urlset"`
	URLs    []sitemapURL
}

// add adds an entry to s.
func (s *sitemap) add(loc string, cf changeFreq) {
	s.URLs = append(s.URLs, sitemapURL{Loc: loc, ChangeFreq: string(cf)})
}

// write writes s to w as XML.
func (s *sitemap) write(w io.Writer) error {
	// Write the <?xml...?> tag first.
	if _, err := w.Write([]byte(xml.Header)); err != nil {
		return err
	}
	// Write <urlset> containing <url> elements.
	enc := xml.NewEncoder(w)
	enc.Indent("", "  ")
	if err := enc.Encode(s); err != nil {
		return err
	}
	_, err := io.WriteString(w, "\n")
	return err
}

// changeFreq describes how frequently a URL in a sitemap is expected to change.
type changeFreq string

const (
	changesWeekly changeFreq = "weekly"
)

// sitemapURL is an entry in a sitemap.
type sitemapURL struct {
	XMLName    string `xml:"url"`
	Loc        string `xml:"loc"`
	ChangeFreq string `xml:"changefreq"`
}
