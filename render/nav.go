// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package render

import (
	"fmt"
	"net/url"
	"regexp"
)

type navItem struct {
	Name     string     `yaml:"name"`     // name displayed in the nav box
	URL      string     `yaml:"url"`      // nav URL, e.g. "page.html" or "page.html#frag"
	ID       string     `yaml:"id"`       // corresponds to page ID; can be empty
	Children []*navItem `yaml:"children"` // items nested under this one in the menu

	Current  bool `yaml:"-"` // true if nav item corresponds to current page
	Expanded bool `yaml:"-"` // true if nav item's children should be displayed

	AMPURL    string `yaml:"-"` // AMP analog of URL field
	AbsURL    string `yaml:"-"` // absolute version of URL
	AbsAMPURL string `yaml:"-"` // absolute version of AMPURL
}

func (n *navItem) update(currentID string, m map[string]*navItem) error {
	if n.ID != "" {
		m[n.ID] = n
	}

	n.Current = n.ID != "" && n.ID == currentID
	n.Expanded = n.Current
	for _, c := range n.Children {
		c.update(currentID, m)
		if c.Expanded {
			n.Expanded = true
		}
	}

	if isPage(n.URL) {
		n.AMPURL = ampPage(n.URL)
		n.AbsURL = baseURL + n.URL
		n.AbsAMPURL = baseURL + n.AMPURL
	} else {
		n.AMPURL = n.URL
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

// absURL converts the supplied string into an absolute URL by appending it to baseURL.
// Returns the unchanged string if it's already absolute.
func absURL(us string) (string, error) {
	u, err := url.Parse(us)
	if err != nil {
		return "", fmt.Errorf("failed to parse %q: %v", us, err)
	}
	if u.IsAbs() {
		return us, nil
	}
	return baseURL + us, nil
}
