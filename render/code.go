// Copyright 2022 Daniel Erat <dan@erat.org>.
// All rights reserved.

package render

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

var chromaFmt = html.New(html.WithClasses(true), html.WithPreWrapper(&preWrapper{}))

// Fallbacks for Chroma token types missing in a style.
// See https://pygments.org/docs/tokens/ for descriptions.
var chromaTokenMap = map[chroma.TokenType]chroma.TokenType{
	chroma.LiteralStringChar:  chroma.LiteralString, // e.g. C 'a'
	chroma.NameBuiltin:        chroma.NameVariable,  // builtins in global namespace
	chroma.NameBuiltinPseudo:  chroma.NameVariable,  // e.g. Ruby self or Java this
	chroma.NameEntity:         chroma.NameConstant,  // e.g. HTML &nbsp;
	chroma.NameFunctionMagic:  chroma.NameFunction,  // e.g. Python __init__
	chroma.NameLabel:          chroma.NameFunction,  // e.g. C goto label
	chroma.NameVariableGlobal: chroma.NameVariable,  // e.g. Ruby global variables
	chroma.NameVariableMagic:  chroma.NameVariable,  // e.g. Python __doc__
}

// getCodeCSS generates CSS class definitions for the named Chroma style.
// baseStyle should contain the Chroma style that is passed to writeCode().
// If selPrefix is non-empty (e.g. "body.dark "), it will be prefixed to each selector.
// Colors will be adjusted to have a brightness within minBrightness and maxBrightness
// in the range [0, 1]. If forcePlain is true, bold, italic, and underline formatting
// will be removed from all style entries.
func getCodeCSS(style, baseStyle, selPrefix string, minBrightness, maxBrightness float64,
	forcePlain bool) (string, error) {
	cs := styles.Get(style)
	if cs == nil {
		return "", fmt.Errorf("couldn't find chroma style %q", style)
	}

	// Build a new style with the token types defined in the base style.
	// The base style determines the classes that will be used by writeCode(),
	// so we want this CSS to cover the same token types.
	bs := cs
	if baseStyle != style {
		if bs = styles.Get(baseStyle); bs == nil {
			return "", fmt.Errorf("couldn't find base chroma style %q", baseStyle)
		}
	}

	sb := chroma.NewStyleBuilder(style + "-adjusted")
	for _, tt := range bs.Types() {
		var se chroma.StyleEntry
		if cs.Has(tt) {
			se = cs.Get(tt) // exact match
		} else if mt, ok := chromaTokenMap[tt]; ok && cs.Has(mt) {
			se = cs.Get(mt) // custom mapping
		} else {
			se = cs.Get(tt) // builtin fallback
		}
		se.Colour = se.Colour.ClampBrightness(minBrightness, maxBrightness)
		if forcePlain {
			se.Bold = chroma.Pass
			se.Italic = chroma.Pass
			se.Underline = chroma.Pass
		}
		sb.Add(tt, se.String())
	}
	var err error
	if cs, err = sb.Build(); err != nil {
		return "", err
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
