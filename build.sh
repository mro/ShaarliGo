#!/bin/sh
# get the go toolchain https://golang.org/dl/
# https://golang.org/doc/install/source#environment
#

cd "$(dirname "${0}")" || exit 1

if say -v '?' 1>/dev/null 2>/dev/null ; then
  say="say"
else
  say="echo"
fi

# prevent GOPATH pointing outside userland (especially on Alpine)
go env -w GOPATH="${HOME}/go"
mkdir -p "${GOPATH}" 2>/dev/null
go env -w GOBIN="$(go env GOPATH)/bin"

parm="" # "-u"
{
  "${say}" "go get (install dependencies)"
  go get "${parm}" github.com/gorilla/sessions \
    github.com/jteeuwen/go-bindata/... \
    golang.org/x/crypto/bcrypt \
    golang.org/x/net/html \
    golang.org/x/net/html/atom \
    golang.org/x/text/language \
    golang.org/x/text/search \
    gopkg.in/yaml.v2 \
    github.com/stretchr/testify \
    github.com/yhat/scrape \
    golang.org/x/tools/blog/atom
}

"$(go env GOBIN)/go-bindata" -ignore="\\.DS_Store" -ignore="assets/dark/" -ignore=".+\\.woff" -prefix static static/... tpl/...

VERSION="$(grep -F 'version = ' version.go | cut -d \" -f 2)"
LDFLAGS="-s -w -X main.GitSHA1=$(git rev-parse --short HEAD)"

rm shaarligo-*.cgi 2>/dev/null

"${say}" "test"
umask 0022
go fmt || { exit $?; }
go vet || { exit $?; }
go test || { exit $?; }
"${say}" "ok"

"${say}" "build localhost"
go build -ldflags "${LDFLAGS}" -o "shaarligo.cgi" || { echo "Aua" 1>&2 && exit 1; }
mv "shaarligo.cgi" "shaarligo-${VERSION}-$(uname -s)-$(uname -m).cgi"
"${say}" "ok"
# open "http://localhost/~$(whoami)/b/shaarligo.cgi"

"${say}" bench
go test -bench=.
"${say}" ok

"${say}" "linux build"
rm -rf build
mkdir -p "build/${VERSION}-Linux-x86_64/"
mkdir -p "build/${VERSION}-Linux-armv6l/"
# http://dave.cheney.net/2015/08/22/cross-compilation-with-go-1-5
env GOOS=linux GOARCH=amd64       go build -ldflags="${LDFLAGS}" -o "build/${VERSION}-Linux-x86_64/shaarligo.cgi" || { echo "Aua" 1>&2 && exit 1; }
env GOOS=linux GOARCH=arm GOARM=6 go build -ldflags="${LDFLAGS}" -o "build/${VERSION}-Linux-armv6l/shaarligo.cgi" || { echo "Aua" 1>&2 && exit 1; }

"${say}" "deploy to dev server"

if [ "${1}" = "prod" ] ; then
  rsync -avPz "build/" s0:"/var/www/lighttpd/darknet.mro.name/public_html/dev/shaarligo/"
else
  rsync -avPz "build/${VERSION}-Linux-x86_64/shaarligo.cgi" s0:"/var/www/lighttpd/demo.0x4c.de/"
fi

