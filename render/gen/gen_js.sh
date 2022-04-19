#!/bin/sh -e
for i in inline/*.js; do
  intransigence -minify "$i" >"${i}.min"
done
