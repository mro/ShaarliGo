#!/bin/sh
# https://golang.org/doc/install/source#environment
#

cd "$(dirname "${0}")"
# $ uname -s -m
# Darwin x86_64
# Linux x86_64
# Linux armv6l

go get golang.org/x/tools/blog/atom
go get golang.org/x/crypto/bcrypt
go get gopkg.in/yaml.v2
go get github.com/jteeuwen/go-bindata/...
# ssh vario find mro.name/vorschau.blog/assets -type f

# rsync -aPz --delete --delete-excluded --exclude jquery* --exclude *.zip --exclude *.html vario:~/mro.name/vorschau.blog/assets/ static/assets
go-bindata -ignore=\\.DS_Store -prefix static static/...

PROG_NAME="AtomicShaarli"
VERSION="0.0.1"

rm "${PROG_NAME}"-*-"${VERSION}" 2>/dev/null

go test || exit $?

# http://dave.cheney.net/2015/08/22/cross-compilation-with-go-1-5
# env GOOS=linux GOARCH=arm GOARM=6 go build -o "${PROG_NAME}-linux-arm-${VERSION}"
env GOOS=linux GOARCH=amd64 go build -ldflags "-s" -o "${PROG_NAME}-linux-amd64-${VERSION}" || { echo "Aua" 1>&2 && exit 1; }
# env GOOS=linux GOARCH=386 GO386=387 go build -o "${PROG_NAME}-linux-386-${VERSION}" # https://github.com/golang/go/issues/11631
# env GOOS=darwin GOARCH=amd64 go build -o "${PROG_NAME}-darwin-amd64-${VERSION}"

# https://lager.mro.name/as/atom.cgi
scp "${PROG_NAME}-linux-amd64-${VERSION}" simply:/var/www/lighttpd/lager.mro.name/public_html/as/"atom.cgi"
scp "ServerInfo.cgi" simply:/var/www/lighttpd/lager.mro.name/public_html/as/"info.cgi"
ssh simply rm -vrf /var/www/lighttpd/lager.mro.name/public_html/as/assets
ssh simply rm -vrf /var/www/lighttpd/lager.mro.name/public_html/as/app

# http://vorschau.blog.mro.name/atom.cgi
scp "${PROG_NAME}-linux-amd64-${VERSION}" vario:~/mro.name/vorschau.blog/"atom.cgi"
scp "ServerInfo.cgi" vario:~/mro.name/vorschau.blog/"info.cgi"
ssh vario rm -vrf mro.name/vorschau.blog/assets
ssh vario rm -vrf mro.name/vorschau.blog/app

# curl --data-urlencode "url=wall" --dump-header head.txt "http://vorschau.blog.mro.name/${PROG_NAME}.cgi"
# curl --location --dump-header head.txt "http://vorschau.blog.mro.name/"
# echo "===="
#cat head.txt
