# erat.org site generator

This code generates the site using a disgusting mix of Go's [html/template
package](https://golang.org/pkg/html/template/) and the [Blackfriday Markdown
library](https://github.com/russross/blackfriday).

Markup that is common across all pages is defined in
`templates/page_header.tmpl` and `templates/page_footer.tmpl`. Page content is
defined via Markdown in the `pages/` directory. A custom Markdown renderer is
used to execute additional templates from the `templates/` directory in response
to specific Markdown data as described below.

## Page info

Every page starts with a fenced code block of type `page_info` that contains a
single YAML dictionary with high-level information about the page. See
`renderer.RenderHeader`.

## Boxes

Boxes are started by defining level-1 headings:

```md
# Box title {#id/desktop_only/narrow}
```

The (optional) ID can be followed with slash-separated arguments to further
customize the box. In `renderer.RenderNode`, see the `md.Heading` case.

The box is automatically closed when a new box is started or the document ends.

## Block images

Block-style images are inserted using fenced code blocks of type `image`. In
`renderer.RenderNode`, see the `"image"` case under `md.CodeBlock`.

## Inline images

Inline images are inserted using `<img-inline>` tags, with options specified via
attributes.

## URLs

URLs can be specified within `<code-url>` tags.
`<code-url>http\://www.example.com/</code-url>` produces `<code
class="url">http://www.example.com/</code>`. Note that the contents are not
escaped. Add a backslash in the scheme to prevent auto-linking.

## Clearing floats

A `<clear-floats>` tag can be used to clear all floats.
