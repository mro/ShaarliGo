#!/bin/sh
cd "$(dirname "${0}")" || exit 1

for fo in Regular Bold Italic BoldItalic
do
  for fmt in woff2 woff
  do
    f="B612-${fo}.${fmt}"
    curl --output "${f}" --remote-time --time-cond "${f}" "http://b612-font.com/fonts/${f}"
  done
done
