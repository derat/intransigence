// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package render

import "testing"

func TestUnmarshalAttrs(t *testing.T) {
	const tag = `<foo string-a="foo" string-b="bar" bool int="123" other="456">`
	tk, err := parseTag([]byte(tag))
	if err != nil {
		t.Fatalf("parseTag(%q) failed: %v", tag, err)
	}

	type st struct {
		OtherInt int
		StringA  string `html:"string-a"`
		Int      int    `html:"int"`
		Bool     bool   `html:"bool"`
		StringB  string `html:"string-b"`
	}
	var s st
	if err := unmarshalAttrs(tk.Attr, &s); err != nil {
		t.Fatalf("unmarshalAttrs(%v, ...) failed: %v", tk.Attr, err)
	}
	if want := (st{StringA: "foo", Int: 123, Bool: true, StringB: "bar"}); s != want {
		t.Fatalf("unmarshalAttrs(%v, ...) produced %+v; want %+v", tk.Attr, s, want)
	}
}

func TestUnmarshalAttrs_Bad(t *testing.T) {
	for _, tc := range []struct {
		tag string
		dst interface{}
	}{
		{`<duplicate-field a="123" a=456">`, &struct {
			A int `html:"a"`
		}{}},
		{`<bad-value a="abc">`, &struct {
			A int `html:"a"`
		}{}},
		{`<unsupported-type a="1.5">`, &struct {
			A float32 `html:"a"`
		}{}},
	} {
		if tk, err := parseTag([]byte(tc.tag)); err != nil {
			t.Errorf("parseTag(%q) failed: %v", tc.tag, err)
		} else if err := unmarshalAttrs(tk.Attr, tc.dst); err == nil {
			t.Fatalf("unmarshalAttrs on %q unexpectedly succeeded (%+v)", tc.tag, tc.dst)
		}
	}
}
