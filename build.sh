#!/bin/sh
# https://golang.org/doc/install/source#environment
#

cd "$(dirname "${0}")"
# $ uname -s -m
# Darwin x86_64
# Linux x86_64
# Linux armv6l

say="say"
false && {
  "${say}" "go get"
  go get -u github.com/gorilla/sessions
  go get -u github.com/jteeuwen/go-bindata/...
  go get -u golang.org/x/crypto/bcrypt
  go get -u golang.org/x/net/html
  go get -u golang.org/x/net/html/atom
  go get -u golang.org/x/text/language
  go get -u golang.org/x/text/search
  go get -u gopkg.in/yaml.v2
  # for testing only:
  go get -u github.com/stretchr/testify
  go get -u github.com/yhat/scrape
  go get -u golang.org/x/tools/blog/atom
  "${say}" "ok"
}

go-bindata -ignore=\\.DS_Store -prefix static static/...

PROG_NAME="ShaarliGo"
VERSION=`fgrep 'version = ' version.go | cut -d '"' -f 2`

rm "${PROG_NAME}"-*-"${VERSION}" 2>/dev/null

"${say}" "test"
umask 0022
go fmt && go test --short || { exit $?; }
"${say}" "ok"

"${say}" "build localhost"
go build -ldflags "-s -w -X main.GitSHA1=$(git rev-parse --short HEAD)" -o ~/public_html/b/shaarligo.cgi || { echo "Aua" 1>&2 && exit 1; }
"${say}" "ok"
# open "http://localhost/~$(whoami)/b/shaarligo.cgi"

"${say}" bench
go test -bench=.
"${say}" ok

"${say}" "linux build"
# http://dave.cheney.net/2015/08/22/cross-compilation-with-go-1-5
# env GOOS=linux GOARCH=arm GOARM=6 go build -o "${PROG_NAME}-linux-arm-${VERSION}"
env GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.GitSHA1=$(git rev-parse --short HEAD)" -o "${PROG_NAME}-linux-amd64-${VERSION}" || { echo "Aua" 1>&2 && exit 1; }
# env GOOS=linux GOARCH=386 GO386=387 go build -o "${PROG_NAME}-linux-386-${VERSION}" # https://github.com/golang/go/issues/11631
# env GOOS=darwin GOARCH=amd64 go build -o "${PROG_NAME}-darwin-amd64-${VERSION}"
"${say}" "ok"

"${say}" "simply"
# scp "ServerInfo.cgi" simply:/var/www/lighttpd/h4u.r-2.eu/public_html/"info.cgi"
gzip --best < "${PROG_NAME}-linux-amd64-${VERSION}" \
| ssh simply "cd /var/www/lighttpd/b.mro.name/public_html/u/ && tee shaarligo_cgi.gz | gunzip > shaarligo.cgi && ls -l shaarligo?cgi*"
"${say}" "ok"

"${say}" "vario"
# scp "ServerInfo.cgi" vario:~/mro.name/webroot/b/"info.cgi"
ssh vario "cd mro.name/webroot/b/ && curl https://b.mro.name/u/shaarligo_cgi.gz | tee shaarligo_cgi.gz | gunzip > shaarligo.cgi && chmod a+x shaarligo.cgi && ls -l shaarligo?cgi*"
"${say}" "ok"

