// Copyright 2022 Daniel Erat <dan@erat.org>.
// All rights reserved.

package render

import (
	"strings"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/js"
)

var minifier *minify.M

const jsType = "application/javascript"

func init() {
	minifier = minify.New()
	minifier.AddFunc(jsType, js.Minify)
}

// minifyData minifies in based on its extension.
func minifyData(in, ext string) (string, error) {
	switch ext {
	case ".css":
		// sassc already minifies, but it doesn't remove trailing newlines.
		return strings.TrimSpace(in), nil
	case ".js":
		return minifier.String(jsType, in)
	default:
		return in, nil
	}
}
