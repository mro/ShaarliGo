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
# go env -w GOPATH="${HOME}/go"
# mkdir -p "${GOPATH}" 2>/dev/null
# go env -w GOBIN="$(go env GOPATH)/bin"

parm="" # "-u"
{
  "${say}" "go get (install dependencies)"
	# go get "${parm}" github.com/jteeuwen/go-bindata/...
	go get "${parm}" github.com/kevinburke/go-bindata/...
}

"$(go env GOBIN)/go-bindata" -ignore="\\.DS_Store" -ignore=".+\\.woff" -prefix static static/... tpl/...

VERSION="$(grep -F 'version = ' version.go | cut -d \" -f 2)"
LDFLAGS="-s -w -X main.GitSHA1=$(git rev-parse --short HEAD)"

rm shaarligo*.cgi 2>/dev/null

"${say}" "test"
umask 0022
go fmt || { exit $?; }
go vet || { exit $?; }
go test || { exit $?; }

"${say}" "build localhost"
go build -ldflags "${LDFLAGS}" -o "shaarligo.cgi" || { echo "Aua" 1>&2 && exit 1; }
mv "shaarligo.cgi" "shaarligo-${VERSION}-$(uname -s)-$(uname -m).cgi"
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

"${say}" "deploy"

if [ "${1}" = "prod" ] ; then
  rsync -avPz "build/" c1:"/var/www/vhosts/darknet.mro.name/pages/dev/shaarligo/"
else
  rsync -avPz "build/${VERSION}-Linux-x86_64/shaarligo.cgi" c0:"/var/www/vhosts/demo.mro.name/"
fi

"${say}" "done"

