# erat.org site generator

This code generates the site using a disgusting mix of Go's [html/template
package](https://golang.org/pkg/html/template/) and the [Blackfriday Markdown
library](https://github.com/russross/blackfriday).

Markup that is common across all pages is defined in
`templates/page_header.tmpl` and `templates/page_footer.tmpl`. Page content is
defined via Markdown in the `pages/` directory.

A custom Markdown renderer is used to execute additional templates from the
`templates/` directory in response to specific Markdown data as described below.
See the `renderCodeBlock`, `renderHeading`, and `renderHTMLSpan` functions in
[render/renderer.go](./render/renderer.go) for details.

## Page info

Every page starts with a fenced code block of type `page_info` that contains a
single YAML dictionary with high-level information about the page.

## Boxes

Boxes are started by defining level-1 headings:

```md
# Box title {#id/desktop_only/narrow}
```

The (optional) ID can be followed with slash-separated arguments to further
customize the box.

The box is automatically closed when a new box is started or the document ends.

## Iframes

Graph and map iframes are inserted using fenced code blocks of type `graph` and
`map`, respectively.

## Images

Block-style images are inserted using fenced code blocks of type `image`.

Inline images are inserted using `<img-inline></img-inline>`, with details
specified via attributes on the opening tag.

## Clearing floats

Left and right floats can be cleared using an empty fenced code block of type
`clear`.

## URLs

URLs can be specified within `<code-url></code-url>`. The contents are not
escaped. Add a backslash in the scheme to prevent auto-linking.

## Text size

`<text-size></text-size>` can be used to change the size of the contained text.
Valid boolean attributes for the opening tag include `small` and `tiny`.

## AMP-only content

`<only-amp></only-amp>` and `<only-nonamp></only-nonamp>can be used to denote
content that's only relevant for the AMP or non-AMP version of the page. The
enclosed markup will be wrapped in an HTML comment in the other version.

## Links

A link to a page can be suffixed with `!force_amp` or `!force_nonamp` to force
it to always be rewritten to the AMP or non-AMP version of the page.
