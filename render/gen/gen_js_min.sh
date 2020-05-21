#!/bin/sh
for i in inline/*.js; do
  # This matches minifyInline() in the build package.
  yui-compressor --type js -o "${i}.min" "$i"
done
