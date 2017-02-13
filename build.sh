#!/bin/sh
# https://golang.org/doc/install/source#environment
#

cd "$(dirname "${0}")"
# $ uname -s -m
# Darwin x86_64
# Linux x86_64
# Linux armv6l

# go get gopkg.in/yaml.v2

PROG_NAME="AtomicShaarli"
VERSION="0.0.1"

rm "${PROG_NAME}"-*-"${VERSION}" 2>/dev/null

# http://dave.cheney.net/2015/08/22/cross-compilation-with-go-1-5
# env GOOS=linux GOARCH=arm GOARM=6 go build -o "${PROG_NAME}-linux-arm-${VERSION}"
env GOOS=linux GOARCH=amd64 go build -o "${PROG_NAME}-linux-amd64-${VERSION}" || { echo "Aua" 1>&2 && exit 1; }
# env GOOS=linux GOARCH=386 GO386=387 go build -o "${PROG_NAME}-linux-386-${VERSION}" # https://github.com/golang/go/issues/11631
# env GOOS=darwin GOARCH=amd64 go build -o "${PROG_NAME}-darwin-amd64-${VERSION}"

scp "${PROG_NAME}-linux-amd64-${VERSION}" vario:~/mro.name/vorschau.blog/"${PROG_NAME}".cgi

curl --data-urlencode "url=wall" --dump-header head.txt "http://vorschau.blog.mro.name/${PROG_NAME}.cgi"
echo "===="
cat head.txt
