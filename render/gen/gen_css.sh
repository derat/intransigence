#!/bin/sh
for i in inline/*.scss; do
  # This matches generateCSS() in the build package.
  sassc --style compressed "$i" "$(echo "$i" | sed -e 's/\.scss$/.css/')"
done
