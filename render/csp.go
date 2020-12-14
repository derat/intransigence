// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package render

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
)

type cspDirective string

const (
	cspDefault cspDirective = "default-src"
	cspChild                = "child-src"
	cspConnect              = "connect-src"
	cspImg                  = "img-src"
	cspScript               = "script-src"
	cspStyle                = "style-src"
	cspFrame                = "frame-src" // deprecated by "child-src" in CSP 2
)

type cspSource string

const (
	cspNone cspSource = "'none'"
	cspSelf           = "'self'"
)

// cspBuilder constructs a Content-Security-Policy <meta> tag.
// See https://www.w3.org/TR/CSP2/ for more info.
type cspBuilder map[cspDirective][]cspSource

// add adds a raw source expression (e.g. cspNone, cspSelf, "'sha256-...'",
// "https://example.com/foo.js" -- note quoting) to the supplied directive (cspDefault, cspScript, etc).
func (c cspBuilder) add(dir cspDirective, src cspSource) {
	c[dir] = append(c[dir], src)

	// child-src was introducted by CSP Level 2 and deprecates frame-src.
	// Duplicate it in frame-src since https://caniuse.com/#feat=contentsecuritypolicy2
	// seems to indicate that Firefox for Android still doesn't support it.
	if dir == cspChild {
		c[cspFrame] = append(c[cspFrame], src)
	}
}

// hash hashes the supplied data (e.g. JavaScript or CSS) and adds a hash source to the supplied directive.
func (c cspBuilder) hash(dir cspDirective, val string) {
	b := sha256.Sum256([]byte(val))
	c.add(dir, cspSource(fmt.Sprintf("'sha256-%s'", base64.StdEncoding.EncodeToString(b[:]))))
}

// tag returns a full <meta> tag with the previously-specified directives.
func (c cspBuilder) tag() string {
	var dirs []string
	for _, dir := range []cspDirective{cspDefault, cspChild, cspConnect, cspImg, cspScript, cspStyle, cspFrame} {
		var srcs []string
		for _, s := range c[dir] {
			srcs = append(srcs, string(s)) // this is dumb
		}
		if len(srcs) == 0 {
			continue
		}
		policy := strings.Join(srcs, " ")

		// Add 'unsafe-inline' for fallback on browsers that don't support script and
		// style hashes (introduced in CSP Level 2):
		// http://stackoverflow.com/questions/31720023/microsoft-edge-not-accepting-hashes-for-content-security-policy
		// TODO: Determine if this is still needed.
		if dir == cspScript || dir == cspStyle {
			if strings.Contains(policy, "'sha256-") {
				policy += " 'unsafe-inline'"
			}
		}

		dirs = append(dirs, string(dir)+" "+policy)
	}
	return fmt.Sprintf(`<meta http-equiv="Content-Security-Policy" content="%s">`, strings.Join(dirs, "; "))
}
