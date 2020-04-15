#!/bin/sh
# https://golang.org/doc/install/source#environment
#

cd "$(dirname "${0}")" || exit 1
# $ uname -s -m
# Darwin x86_64
# Linux x86_64
# Linux armv6l

say="say"
parm="" # "-u"
{
  "${say}" "go get"
  go get "${parm}" github.com/gorilla/sessions \
    github.com/jteeuwen/go-bindata/... \
    golang.org/x/crypto/bcrypt \
    golang.org/x/net/html \
    golang.org/x/net/html/atom \
    golang.org/x/text/language \
    golang.org/x/text/search \
    gopkg.in/yaml.v2 \
    \
    github.com/stretchr/testify \
    github.com/yhat/scrape \
    golang.org/x/tools/blog/atom
}

"$(go env GOPATH)/bin/go-bindata" -ignore="\\.DS_Store" -ignore="assets/dark/" -ignore=".+\\.woff" -prefix static static/... tpl/...

VERSION="$(grep -F 'version = ' version.go | cut -d \" -f 2)"
LDFLAGS="-s -w -X main.GitSHA1=$(git rev-parse --short HEAD)"

rm shaarligo-*.cgi 2>/dev/null

"${say}" "test"
umask 0022
go fmt && go vet && go test --short || { exit $?; }
"${say}" "ok"

tar -czf testdata.tar.gz testdata/*.html testdata/*.atom testdata/*.gob

"${say}" "build localhost"
go build -ldflags "${LDFLAGS}" -o "shaarligo.cgi" || { echo "Aua" 1>&2 && exit 1; }
mv "shaarligo.cgi" "shaarligo-${VERSION}-$(uname -s)-$(uname -m).cgi"
"${say}" "ok"
# open "http://localhost/~$(whoami)/b/shaarligo.cgi"

"${say}" bench
go test -bench=.
"${say}" ok

"${say}" "linux build"
# http://dave.cheney.net/2015/08/22/cross-compilation-with-go-1-5
env GOOS=linux GOARCH=amd64       go build -ldflags="${LDFLAGS}" -o "shaarligo-${VERSION}-Linux-x86_64.cgi" || { echo "Aua" 1>&2 && exit 1; }
env GOOS=linux GOARCH=arm GOARM=6 go build -ldflags="${LDFLAGS}" -o "shaarligo-${VERSION}-Linux-armv6l.cgi" || { echo "Aua" 1>&2 && exit 1; }

"${say}" "s0"
rsync -avPz "shaarligo-${VERSION}-Linux-"*.cgi "s0:/var/www/lighttpd/demo.0x4c.de/"

