// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package render

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
)

type cspHasher map[string][]string

func (c cspHasher) hash(dir, val string) {
	b := sha256.Sum256([]byte(val))
	h := fmt.Sprintf("'sha256-%s'", base64.StdEncoding.EncodeToString(b[:]))
	c[dir] = append(c[dir], h)

	// child-src was introducted by CSP Level 2 and deprecates frame-src.
	// Duplicate it in frame-src since https://caniuse.com/#feat=contentsecuritypolicy2
	// seems to indicate that Firefox for Android still doesn't support it.
	if dir == "child" {
		c["frame"] = append(c["frame"], h)
	}
}

func (c cspHasher) tag() string {
	var dirs []string
	for _, dir := range strings.Fields("default child img script style frame") {
		hs := c[dir]
		if len(hs) == 0 {
			continue
		}
		policy := strings.Join(hs, " ")

		// Add 'unsafe-inline' for fallback on browsers that don't support script and
		// style hashes (introduced in CSP Level 2):
		// http://stackoverflow.com/questions/31720023/microsoft-edge-not-accepting-hashes-for-content-security-policy
		// TODO: Determine if this is still needed.
		if dir == "script" || dir == "style" {
			if strings.Contains(policy, "'sha256-") {
				policy += " 'unsafe-inline'"
			}
		}

		dirs = append(dirs, dir+"-src "+policy)
	}
	return fmt.Sprintf(`<meta http-equiv="Content-Security-Policy" content="%s">`, strings.Join(dirs, "; "))
}
