// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package render

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func parseCSPTag(tag string) (map[cspDirective][]cspSource, error) {
	tk, err := parseTag([]byte(tag))
	if err != nil {
		return nil, err
	}
	if tk.Data != "meta" {
		return nil, fmt.Errorf("tag is %q; want \"meta\"", tk.Data)
	}
	if tk.Type != html.StartTagToken {
		return nil, errors.New("not start tag")
	}

	var attr = struct {
		HTTPEquiv string `html:"http-equiv"`
		Content   string `html:"content"`
	}{}
	if err := unmarshalAttrs(tk.Attr, &attr); err != nil {
		return nil, err
	}
	if want := "Content-Security-Policy"; attr.HTTPEquiv != want {
		return nil, fmt.Errorf("http-equiv attr is %q; want %q", attr.HTTPEquiv, want)
	}

	// This is sloppy compared to https://www.w3.org/TR/CSP2/#syntax-and-algorithms.
	policy := make(map[cspDirective][]cspSource)
	for i, dir := range strings.Split(attr.Content, ";") {
		fs := strings.Fields(strings.TrimSpace(dir))
		if len(fs) == 0 {
			return nil, fmt.Errorf("directive %d is empty", i)
		}
		dir := cspDirective(fs[0])
		for _, f := range fs[1:] {
			policy[dir] = append(policy[dir], cspSource(f))
		}
	}
	return policy, nil
}

func TestCSPBuilderRaw(t *testing.T) {
	csp := cspBuilder{}
	csp.add(cspDefault, cspNone)
	csp.add(cspStyle, cspSelf)
	csp.add(cspScript, cspSelf)

	tag := csp.tag()
	if policy, err := parseCSPTag(tag); err != nil {
		t.Fatalf("Bad tag %q: %v", tag, err)
	} else if want := map[cspDirective][]cspSource{
		cspDefault: []cspSource{cspNone},
		cspStyle:   []cspSource{cspSelf},
		cspScript:  []cspSource{cspSelf},
	}; !reflect.DeepEqual(policy, want) {
		t.Errorf("Tag %q defines policy %v; want %v", tag, policy, want)
	}
}

func TestCSPBuilderList(t *testing.T) {
	const (
		url = "https://www.example.com/foo.js"
		js  = "alert('foo');"
		// echo -n "alert('foo');" | openssl dgst -binary -sha256 | openssl base64.
		jsHash = "WOdSzz11/3cpqOdrm89LBL2UPwEU9EhbDtMy2OciEhs="
	)

	csp := cspBuilder{}
	csp.add(cspScript, cspSelf)
	csp.add(cspScript, cspSource(url))
	csp.hash(cspScript, js)

	tag := csp.tag()
	if policy, err := parseCSPTag(tag); err != nil {
		t.Fatalf("Bad tag %q: %v", tag, err)
	} else if want := map[cspDirective][]cspSource{
		cspScript: []cspSource{
			cspSelf,
			cspSource(url),
			cspSource(fmt.Sprintf("'sha256-%s'", jsHash)),
			cspSource("'unsafe-inline'"), // added for fallback
		},
	}; !reflect.DeepEqual(policy, want) {
		t.Errorf("Tag %q defines policy %v; want %v", tag, policy, want)
	}
}

func TestCSPBuilderCopyChildToFrame(t *testing.T) {
	csp := cspBuilder{}
	csp.add(cspChild, cspSelf)

	tag := csp.tag()
	if policy, err := parseCSPTag(tag); err != nil {
		t.Fatalf("Bad tag %q: %v", tag, err)
	} else if want := map[cspDirective][]cspSource{
		cspChild: []cspSource{cspSelf},
		cspFrame: []cspSource{cspSelf},
	}; !reflect.DeepEqual(policy, want) {
		t.Errorf("Tag %q defines policy %v; want %v", tag, policy, want)
	}
}
