// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package render

type navItem struct {
	Name     string     `yaml:"name"` // name displayed in the nav box
	URL      string     `yaml:"url"`  // TODO: different for amp; see emit()
	ID       string     `yaml:"id"`   // corresponds to page ID; can be empty
	Children []*navItem `yaml:"children"`

	Current  bool `yaml:"-"` // true if nav item corresponds to current page
	Expanded bool `yaml:"-"` // true if nav item's children should be displayed

	HTMLURL string `yaml:"-"` // canonical URL for HTML page
	AMPURL  string `yaml:"-"` // non-CDN URL for AMP page
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
		n.HTMLURL = baseURL + n.URL
		amp, err := ampPage(n.URL)
		if err != nil {
			return err
		}
		n.AMPURL = baseURL + amp
	}

	return nil
}
