```page
title: Scottish Fold
id: scottish
created: 2020-05-20
modified: 2020-05-21
```

# Scottish Fold

This page describes the [Scottish Fold] breed of cat.

It's listed as a child of the [Cats page](cats.html) in `site.yaml`, so it
appears nested under it in the navigation menu.

The following table of contents is inserted using a `contents` code block.

```contents
heading: Table of contents
```

## Characteristics {#chars}

Scottish Folds possess a genetic mutation that makes their ears bend forward,
giving them what the [Cat Fanciers' Association] describes as an owl-like
appearance.

The "Characteristics" title up there was created using a second-level heading,
and an anchor was assigned to it so that an entry in the navigation menu can
link directly to it.

[Scottish Fold]: https://en.wikipedia.org/wiki/Scottish_Fold
[Cat Fanciers' Association]: https://cfa.org/scottish-fold/scottish-fold-article/

## Fame {#fame}

The most famous Scottish Fold on the Internet is probably [Maru]. Maru lives in
Japan and is fascinated by cardboard boxes:

```image
path: scottish_fold/maru-*.jpg
alt: Maru the cat sitting in a small cardboard box
align: desktop_left
caption: Maru
class: custom-class
```

```clear
```

The above image was inserted using an `image` fenced code block with a YAML
dictionary specifying image attributes:

````md
```image
path: scottish_fold/maru-*.jpg
alt: Maru the cat sitting in a small cardboard box
[other attributes]
```
````

The `path` pattern matches multiple copies of the image at different
resolutions, and an `srcset` attribute is automatically created so the browser
can choose the most appropriate size. Additionally, WebP versions of the image
are automatically generated for browsers that support them, and the image
automatically links to the high-resolution version of itself.

Images can also be inserted inline via `<image path="..." alt="..."></image>`
tags, as you can see here: <image path="scottish_fold/nyan.gif" alt="Nyan Cat"></image>

Images that are already in the WebP format are also supported:

```image
path: scottish_fold/christmas.webp
alt: Scottish Fold cat under a Christmas tree
```

Ditto for data URLs:

```image
url: data:image/gif;base64,R0lGODlhBAAEAIMAAHRPLY9uXD5VQ048L5ljR5lmTzVVSD1cTbWnnIJ1VEBXTyhRQ8bGvXp2Sk5fVjRYUCwAAAAABAAEAAAEC4AgFkpqwig3znoiADs=
alt: A blurry, unintelligible image
width: 200
height: 200
```

[Maru]: https://en.wikipedia.org/wiki/Maru_(cat)

## Controversy _(Really!)_ {#controversy/}

[This BBC article](https://www.bbc.com/news/uk-scotland-39717634) reports on
arguments that breeding of Scottish Folds should be banned due to concerns about
ear disorders and hearing problems. I have no opinion on this topic.

### Some inline tags

Instead, I'll use this space to describe how `<only-amp>` and `<only-nonamp>`
tags can be used to wrap content that's only relevant for the AMP or non-AMP
version of a page.

You can also suffix a page link with `!force_amp` or `!force_nonamp` to force a
link to go to one version of the page or the other.

<only-amp>You're viewing the AMP version of this page. Here's the [non-AMP
version](scottish_fold.html!force_nonamp).</only-amp>
<only-nonamp>You're viewing the non-AMP version of this page. Here's the [AMP
version](scottish_fold.html!force_amp).</only-nonamp>

`<text-size small>` <text-size small>makes text small</text-size>, and
`<text-size tiny>` <text-size tiny>makes it even smaller</text-size>.

```
Text can also be ‹marked as non-selectable› within a code block.
```

Syntax highlighting of code blocks is supported using the [Chroma] package:

```go
package main

import "fmt"

func main() {
	fmt.Println("Hello, world!")
}
```

Here's an example map (which won't display since a Google Maps API key is not
specified):

```map
href: iframes/map.html
width: 640
height: 480
path: scottish_fold/map_light.png
path_dark: scottish_fold/map_dark.png
```

And an example graph:

```graph
href: iframes/graph.html
name: line
width: 300
height: 200
```

[Chroma]: https://github.com/alecthomas/chroma
