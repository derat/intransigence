#!/bin/sh -e
for i in inline/*.scss; do
  # This matches generateCSS() in the build package.
  sassc --style compressed "$i" "${i%scss}css"
done
