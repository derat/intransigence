// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package render

import (
	"fmt"
	"regexp"
)

const (
	indexID   = "index" // used for index page
	indexFile = "index.html"
)

type navItem struct {
	Name     string     `yaml:"name"`     // name displayed in the nav box
	URL      string     `yaml:"url"`      // nav URL, e.g. "page.html" or "page.html#frag"
	ID       string     `yaml:"id"`       // corresponds to page ID; can be empty
	Children []*navItem `yaml:"children"` // items nested under this one in the menu
}

// AMPURL returns the the AMP version of n.URL.
// If n does not represent a page, n.URL is returned.
func (n *navItem) AMPURL() string {
	if isPage(n.URL) {
		return ampPage(n.URL)
	}
	// Special case: URL is empty for the synthetic item corresponding to the index,
	// but we need to use a real filename to link to the AMP version.
	if n.ID == indexID {
		return ampPage(indexFile)
	}
	return n.URL
}

// HasID returns true if n has the supplied ID.
func (n *navItem) HasID(id string) bool {
	return n.ID != "" && n.ID == id
}

// FindID returns the item with the supplied ID (possibly n itself) rooted at n.
// Returns nil if the item is not found.
func (n *navItem) FindID(id string) *navItem {
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

var pageRegexp = regexp.MustCompile("^([a-z_]+)\\.html(#.+)?$")

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
	return base + ".amp.html" + frag
}
