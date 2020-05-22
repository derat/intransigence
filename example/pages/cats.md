```page
title: Cats
id: cats
created: 2020-05-20
modified: 2020-05-20
```

# Cats

This page lists the various cats that this site has to offer.

A `page` fenced code block must appear at the top of every page Markdown file.
It contains metadata about the page like its title and the page's ID in
`site.yaml`:

````md
```page
title: Cats
id: cats
created: 2020-05-20
modified: 2020-05-20
```
````

This page appears at the top level of the navigation menu due to its appearance
at the top level of the `nav_items` array in `site.yaml`.

When this page is active, it is automatically expanded in the navigation menu so
that its children are visible. As such, you can see that it has a subpage named
[Scottish Fold](scottish_fold.html).
