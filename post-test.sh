#!/bin/sh

BASE_URL="https://.../shaarligo.cgi"
BASE_URL="https://demo.shaarli.org/"
USERNAME="demo"
PASSWORD="demo"
# link
# title
# description
# tags
# private
# (linkdate)

# Check preliminaries
curl --version >/dev/null       || { echo "apt-get install curl" && exit 1; }
xmllint --version 2> /dev/null  || { echo "apt-get install libxml2-utils" && exit 1; }
ruby --version > /dev/null      || { echo "apt-get install ruby" && exit 1; }

cd "$(dirname "${0}")"

rm curl.* 2> /dev/null
LOCATION=$(curl --location --get --url "${BASE_URL}" \
  --cookie curl.cook --cookie-jar curl.cook \
  --location --output curl.tmp.html \
  --trace-ascii curl.tmp.trace --dump-header curl.tmp.head \
  --data-urlencode "post=https://github.com/sebsauvage/Shaarli/commit/450342737ced8ef2864b4f83a4107a7fafcc4add" \
  --data-urlencode "title=Mein Titel" \
  --data-urlencode "source=ShaarliGo" \
  --write-out '%{url_effective}' 2>/dev/null)
echo "next: ${LOCATION}"

fgrep -v '<?xml ' curl.tmp.html | xmllint --html --nowarning --xmlout --encode utf-8 --nodefdtd - \
| /usr/bin/env ruby loginform.rb "${USERNAME}" "${PASSWORD}" \
> "curl.post.login"

[ -s "curl.post.login" ] && {
  rm curl.tmp.*
  LOCATION=$(curl --location --url "${LOCATION}" \
    --data-urlencode "login=${USERNAME}" \
    --data-urlencode "password=${PASSWORD}" \
    --data "@curl.post.login" \
    --cookie curl.cook --cookie-jar curl.cook \
    --output curl.tmp.html \
    --trace-ascii curl.tmp.trace --dump-header curl.tmp.head \
    --write-out '%{url_effective}' 2>/dev/null)
  rm "curl.post.login"
  echo "next: ${LOCATION}"
}

fgrep -v '<?xml ' curl.tmp.html | xmllint --html --nowarning --xmlout --encode utf-8 --nodefdtd - \
| /usr/bin/env ruby linkform.rb \
> "curl.post.link"

if [ ! -s "curl.post.link" ] ; then
  echo "login failed"
  cat curl.tmp.head
  cat curl.tmp.html
else
  rm curl.tmp.*
  LOCATION=$(curl --location --url "${LOCATION}" \
    --cookie curl.cook --cookie-jar curl.cook \
    --output curl.tmp.html \
    --trace-ascii curl.tmp.trace --dump-header curl.tmp.head \
    --data-urlencode "lf_tags=t1 t2 t3" \
    --data "@curl.post.link" \
    --write-out '%{url_effective}' 2>/dev/null)
  echo "finish: ${LOCATION}"
fi

rm curl.*
