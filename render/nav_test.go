// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package render

import "testing"

func TestNavItem_AMPURL(t *testing.T) {
	for _, tc := range []struct {
		ni   NavItem
		want string
	}{
		{NavItem{URL: "page" + HTMLExt}, "page" + AMPExt},
		{NavItem{URL: "page" + HTMLExt + "#frag"}, "page" + AMPExt + "#frag"},
		{NavItem{URL: "mailto:me@example.org"}, "mailto:me@example.org"},
		{NavItem{ID: indexID}, indexFileBase + AMPExt},
	} {
		if got := tc.ni.AMPURL(); got != tc.want {
			t.Errorf("NavItem%+v.AMPURL() = %q; want %q", tc.ni, got, tc.want)
		}
	}
}

func TestNavItem_FindID(t *testing.T) {
	n := &NavItem{ID: "a", Children: []*NavItem{
		&NavItem{ID: "b"},
		&NavItem{ID: "c", Children: []*NavItem{
			&NavItem{ID: "d"},
		}},
	}}

	find := func(id string) {
		f := n.FindID(id)
		if f == nil {
			t.Errorf("FindID(%q) returned nil", id)
		} else if f.ID != id {
			t.Errorf("FindID(%q) returned item %q", id, f.ID)
		}
	}
	find("a")
	find("b")
	find("c")
	find("d")

	if f := n.FindID("e"); f != nil {
		t.Errorf(`FindID("e") unexpectedly returned item %q`, f.ID)
	}
}
