// Copyright 2022 Daniel Erat <dan@erat.org>.
// All rights reserved.

package render

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
)

var chromaFmt = html.New(html.WithClasses(true), html.WithPreWrapper(&preWrapper{}))

// getCodeCSS generates CSS class definitions for the named Chroma style.
// If selPrefix is non-empty (e.g. "body.dark "), it will be prefixed to each selector.
func getCodeCSS(style, selPrefix string) (string, error) {
	cs := styles.Get(style)
	if cs == nil {
		return "", fmt.Errorf("couldn't find chroma style %q", style)
	}
	var b bytes.Buffer
	if err := chromaFmt.WriteCSS(&b, cs); err != nil {
		return "", err
	}
	s := b.String()
	if selPrefix != "" {
		// The unminified CSS consists of lines like
		// "/* Background */ .chroma { color: #f8f8f2; ...".
		// Adding prefixes like this is awful, but wrapping the whole thing in a selector
		// and then shelling out to sassc feels gross too (especially since we'll be doing
		// this once per page).
		s = strings.ReplaceAll(s, " .chroma ", selPrefix+" .chroma ")
	}
	return minifyData(s, ".css")
}

// writeCode performs syntax highlighting on the supplied code and writes the corresponding HTML
// (using the CSS classes from codeCSS, and wrapped in <pre> and <code> elements) to w.
// lang is a Chroma syntax ID: https://github.com/alecthomas/chroma#identifying-the-language
func writeCode(w io.Writer, code, lang, style string) error {
	lexer := lexers.Get(lang)
	if lexer == nil {
		return fmt.Errorf("no lexer for language %q", lang)
	}
	lexer = chroma.Coalesce(lexer)
	it, err := lexer.Tokenise(nil, code)
	if err != nil {
		return err
	}
	cs := styles.Get(style)
	if cs == nil {
		return fmt.Errorf("couldn't find chroma style %q", style)
	}
	return chromaFmt.Format(w, cs, it)
}

// preWrapper implements html.PreWrapper. The default implementation only
// wraps code in <pre> and </pre>, but we want to include <code> as well.
type preWrapper struct{}

func (w *preWrapper) Start(code bool, styleAttr string) string {
	if code {
		return fmt.Sprintf(`<pre><code%s>`, styleAttr)
	}
	return fmt.Sprintf(`<pre%s>`, styleAttr)
}
func (w *preWrapper) End(code bool) string {
	if code {
		return `</code></pre>`
	}
	return `</pre>`
}
