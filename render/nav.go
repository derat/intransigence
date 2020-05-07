// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package render

import (
	"fmt"
	"regexp"
)

const (
	// Extensions for non-AMP and AMP pages.
	HTMLExt = ".html"
	AMPExt  = ".amp.html"

	// Used by a synthetic NavItem created when the index page is active.
	indexID       = "index"
	indexFileBase = "index"
)

// pageRegexp matches regular page filenames with an optional fragment.
var pageRegexp = regexp.MustCompile("^([a-z_]+)" + regexp.QuoteMeta(HTMLExt) + "(#.+)?$")

// NavItem describes a single item displayed in the site navigation menu.
type NavItem struct {
	// Name contains the short, human-readable name for the item that is displayed in the menu.
	Name string `yaml:"name"`
	// URL contains the non-root-relative URL for the item, e.g. "page.html" or "page.html#frag".
	// It may also be e.g. a "mailto:" link.
	URL string `yaml:"url"`
	// ID corresponds to the ID specified in the page linked to by the item.
	// It may be empty if the item doesn't correspond to the top of a page
	// (e.g. if it contains a fragment).
	ID string `yaml:"id"`
	// Children contains items nested under this one in the menu.
	Children []*NavItem `yaml:"children"`
}

// AMPURL returns the the AMP version of n.URL.
// If n does not represent a page, n.URL is returned.
func (n *NavItem) AMPURL() string {
	if isPage(n.URL) {
		return ampPage(n.URL)
	}
	// Special case: URL is empty for the synthetic item corresponding to the index,
	// but we need to use a real filename to link to the AMP version.
	if n.ID == indexID {
		return ampPage(indexFileBase + HTMLExt)
	}
	return n.URL
}

// HasID returns true if n has the supplied ID.
func (n *NavItem) HasID(id string) bool {
	return n.ID != "" && n.ID == id
}

// FindID returns the item with the supplied ID (possibly n itself) rooted at n.
// Returns nil if the item is not found.
func (n *NavItem) FindID(id string) *NavItem {
	if n.HasID(id) {
		return n
	}
	for _, c := range n.Children {
		if m := c.FindID(id); m != nil {
			return m
		}
	}
	return nil
}

// splitPage splits a string like "foo.html#frag" into "foo" and "#frag".
// Returns empty strings if p isn't a page URL.
func splitPage(p string) (base, fragment string) {
	m := pageRegexp.FindStringSubmatch(p)
	if m == nil {
		return "", ""
	}
	return m[1], m[2]
}

// isPage returns true if p is an HTML page (e.g. "foo.html") with an optional fragment.
func isPage(p string) bool {
	base, _ := splitPage(p)
	return base != ""
}

// ampPage returns the AMP page corresponding to p, an HTML page.
// If a fragment is present, it is preserved.
func ampPage(p string) string {
	base, frag := splitPage(p)
	if base == "" {
		panic(fmt.Sprintf("Can't get AMP version of non-page %q", p))
	}
	return base + AMPExt + frag
}
