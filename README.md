# intransigence

## Description

This repository contains Go code for generating a basic-but-modern static
website from Markdown files. Various features are included (and hardcoded):

*   Both [HTML5] and [AMP] versions of pages are generated.
*   Generated pages are [pretty-printed] and [validated].
*   All generated pages live at the top level of the site, but a hierarchical
    navigation menu is automatically created.
*   [srcset] attributes are automatically added when images are provided in
    multiple resolutions.
*   [WebP] versions of image files are automatically created.
*   [Content Security Policy] `<meta>` tags are automatically generated.
*   Interactive [Google Maps] and simple [D3.js] line graphs can be embedded via
    iframes.
*   Textual files is automatically compressed via [gzip] for efficient serving.
*   File modification timestamps are preserved when regenerating the site to
    minimize the data that needs to be sent when pushing the site to a remote
    web server via [rsync].
*   Page display is (somewhat) configurable via custom [Sass] SCSS files.
*   A [Sitemap] XML file listing the site's pages is automatically created.
*   When regenerating the site, a web server is started to allow viewing the new
    pages and a diff is displayed to make it easy to see changes from the
    previous version.

This codebase started out in 2010 as a [Ruby] library and associated [eRuby]
templates that I used to generate [my website]. The original system is described
in detail in a document that I wrote about [generating AMP pages].

After spending a decade living in terror of making changes to the code, I
finally replaced it with this Go program in 2020. After getting the code to the
point where it could produce a close-to-identical copy of my website and adding
tests, I tried to generalize it somewhat. This repository was created from a
subset of my website's repository, so its commit history is weird.

There are a zillion systems for static site generation, and I'm skeptical that
anyone will choose to use this one, due in large part to its lack of
customizability. However, I'm of the opinion that it's generally better to
open-source stuff instead of leaving it private, and maybe someone will find
some part of this to be useful or informative. At the very least, generalizing
and documenting the code increases the chance that I'll reuse it myself in the
future.

[HTML5]: https://developer.mozilla.org/en-US/docs/Web/Guide/HTML/HTML5
[AMP]: https://amp.dev/
[pretty-printed]: https://github.com/derat/htmlpretty
[validated]: https://github.com/derat/validate
[srcset]: https://developer.mozilla.org/en-US/docs/Learn/HTML/Multimedia_and_embedding/Responsive_images
[WebP]: https://developers.google.com/speed/webp
[Content Security Policy]: https://developer.mozilla.org/en-US/docs/Web/HTTP/CSP
[Google Maps]: https://cloud.google.com/maps-platform/
[D3.js]: https://d3js.org/
[gzip]: https://www.gzip.org/
[rsync]: https://rsync.samba.org/
[Sass]: https://sass-lang.com/
[Sitemap]: https://www.sitemaps.org/
[Ruby]: https://www.ruby-lang.org/
[eRuby]: https://en.wikipedia.org/wiki/ERuby
[my website]: https://www.erat.org/
[generating AMP pages]: https://www.erat.org/amp.html#my-site

## Usage

### Installation

The `intransigence` executable can be installed to `$GOPATH/bin` by running `go
install ./...`.

### Dependencies

The following tools must be available in `$PATH`:

*   `cwebp` and `gif2webp` to create WebP images.
*   `sassc` to process SASS `.scss` files into standard CSS files.
*   `yui-compressor` to minify JavaScript files.

These tools can be installed on a Debian system by running `apt install sassc
webp yui-compressor` as root.

### Site directory

A site is defined via a directory with the following structure:

*   `site.yaml` - YAML representation of the `SiteInfo` struct struct from
    [render/site.go](render/site.go) containing high-level information about the
    site, along with a hierarchy of `NavItem` structs from
    [render/nav.go](render/nav.go) defining the structure of the site's
    navigation menu.
*   `pages/` - Subdirectory containing Markdown files specifying the content of
    individual pages. These files are described in more detail below.
*   `inline/` - Subdirectory containing `.scss` files with custom CSS rules:
    *   `base.scss` - Rules included in all generated pages.
    *   `desktop.scss` - Rules applying only to large screen sizes.
    *   `mobile.scss` - Rules applying only to small screen sizes.
    *   `amp.scss` - Rules included only in AMP pages.
    *   `nonamp.scss` - Rules included only in non-AMP pages.
*   `static/` - Subdirectory containing files that are copied unchanged to the
    top level of the output directory (e.g. images).

The [example](example) directory defines an example site and is a good starting point.

When the `intransigence` executable is run within a site directory, it reads the
`site.yaml` file and then builds the site into an `out/` subdirectory. Flags can
be passed to control the executable's behavior:

```
% intransigence -help
Usage of intransigence:
  -dir string
        Site directory (defaults to working dir)
  -out string
        Destination directory (site is built under -dir if empty)
  -pretty
        Pretty-print HTML (default true)
  -prompt
        Prompt with a diff before replacing dest dir (only if -out is empty) (default true)
  -serve
        Serve output over HTTP while displaying diff (default true)
  -validate
        Validate generated files (default true)
```

### Pages

Each Markdown file in the site directory's `pages/` subdirectory must start with
a [fenced code block] of type `page`. The block contains a YAML representation
of the `pageInfo` struct defined in [render/page.go](render/page.go), e.g.

````md
```page
title: My New Page
id: new_page
desc: My page's meta description.
image_path: new_page.png
created: 2020-05-19
modified: 2020-05-20
```
````

[fenced code block]: https://www.markdownguide.org/extended-syntax/#fenced-code-blocks

### Boxes

Boxes are started by defining level-1 headings:

```md
# Box title {#id/desktop_only/narrow}
```

The (optional) ID can be followed with slash-separated arguments to further
customize the box. See `renderHeading` in [render/page.go](render/page.go) for
available arguments.

The box is automatically closed when a new box is started or the document ends.

### Iframes

Graph and map iframes are inserted using fenced code blocks of type `graph` and
`map`, respectively, containing YAML dictionaries. See the `"graph"` and `"map"`
cases in `renderCodeBlock` in [render/page.go](render/page.go) for available
options.

### Images

Block-style images are inserted using fenced code blocks of type `image`
containing YAML dictionaries. See the `"image"` case in `renderCodeBlock` in
[render/page.go](render/page.go) for available options.

Inline images are inserted using `<image></image>`, with details specified via
attributes on the opening tag. See the `"image"` case in `renderHTMLSpan` in
[render/page.go](render/page.go) for available options.

### Clearing floats

Left and right floats can be cleared using an empty fenced code block of type
`clear`.

### URLs

URLs can be placed within `<code-url></code-url>` to let them wrap on mobile
even if they don't contain characters that would usually trigger wrapping. The
contents are not escaped. Add a backslash in the scheme to prevent auto-linking.

### Text size

`<text-size></text-size>` can be used to change the size of the contained text.
Valid boolean attributes for the opening tag include `small` and `tiny`.

### AMP-only content

`<only-amp></only-amp>` and `<only-nonamp></only-nonamp>can be used to denote
content that's only relevant for the AMP or non-AMP version of the page. The
enclosed markup will be wrapped in an HTML comment in the other version.

### Links

A link to a page can be suffixed with `!force\_amp` or `!force\_nonamp` to force
it to always be rewritten to the AMP or non-AMP version of the page.

## Implementation

### `render` package

The [render](render) package is responsible for rendering an individual page in
either HTML or AMP format. It uses Go's [html/template package] and the
[Blackfriday Markdown library]. A custom Markdown renderer is used to execute
HTML templates in response to specific Markdown data.

Hardcoded HTML templates are located in [render/templates](render/templates) and
copied into [render/std_templates.go](render/std_templates.go) so they can be
included in the `intransigence` executable. The start and end of each page are
rendered by [render/templates/page.tmpl](render/templates/page.tmpl). Other
templates are used to render individual elements of the page.

Hardcoded Javascript and CSS files are located in [render/inline](render/inline)
and copied into [render/std_inline.go](render/std_inline.go).

[html/template package]: https://golang.org/pkg/html/template/
[Blackfriday Markdown library]: https://github.com/russross/blackfriday

### `build` package

The [build](build) package is responsible for building the whole site. The
`Build` function in [build/build.go](build/build.go) provides the implementation
of the [intransigence executable](cmd/intransigence/main.go) by:

*   generating and minifying CSS, JS, and WebP files as needed,
*   generating HTML and AMP versions of all pages,
*   generating HTML iframe pages,
*   validating generated files,
*   copying static data into the output directory,
*   writing a `sitemap.xml` file,
*   compressing all textual files, and
*   starting a web server and showing a diff to let the user manually verify the
    new version of the site.

[build/build_test.go](build/build_test.go) performs end-to-end testing of the
whole process.
