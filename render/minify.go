// Copyright 2022 Daniel Erat <dan@erat.org>.
// All rights reserved.

package render

import (
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/js"
)

const (
	cssType = "text/css"
	jsType  = "application/javascript"
)

var minifier *minify.M

func init() {
	minifier = minify.New()
	minifier.AddFunc(cssType, css.Minify)
	minifier.AddFunc(jsType, js.Minify)
}

// MinifyData minifies data based on its extension.
func MinifyData(data, ext string) (string, error) {
	switch ext {
	case ".css":
		return minifier.String(cssType, data)
	case ".js":
		return minifier.String(jsType, data)
	default:
		return data, nil
	}
}
