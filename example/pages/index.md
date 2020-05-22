```page
title: Welcome to an example site
id: index
created: 2020-05-20
modified: 2020-05-20
hide_title_suffix: true
hide_back_to_top: true
hide_dates: true
```

# Welcome

This is the site's landing page.

There's nothing really interesting on it. This page is listed in the navigation
menu due to its inclusion in the `nav_items` array in `site.yaml`, but you can
leave it out of the array if you'd rather it not show up in the menu.

The "Welcome" box was creating using a Markdown first-level heading, i.e.
`# Welcome`.

Additional attributes can be included after a first-level heading's name.

# Mobile-only {#/mobile_only}

This box only displays at mobile resolutions.

# Desktop-only {#/desktop_only/narrow}

This box only displays at "desktop" (i.e. bigger than mobile) resolutions, and
it is narrower than normal boxes.