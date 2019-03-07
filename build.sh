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
  go get ${parm} github.com/gorilla/sessions \
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

"$(go env GOPATH)/bin/go-bindata" -ignore="\\.DS_Store" -ignore=".+\\.woff" -prefix static static/...

PROG_NAME="ShaarliGo"
VERSION="$(grep -F 'version = ' version.go | cut -d \" -f 2)"

rm "${PROG_NAME}"-*-"${VERSION}" 2>/dev/null

"${say}" "test"
umask 0022
go fmt && go vet && go test --short || { exit $?; }
"${say}" "ok"

tar -czf testdata.tar.gz testdata/*.html testdata/*.atom testdata/*.gob

"${say}" "build localhost"
go build -ldflags "-s -w -X main.GitSHA1=$(git rev-parse --short HEAD)" -o ~/public_html/b/shaarligo.cgi || { echo "Aua" 1>&2 && exit 1; }
"${say}" "ok"
# open "http://localhost/~$(whoami)/b/shaarligo.cgi"

"${say}" bench
go test -bench=.
"${say}" ok

"${say}" "linux build"
# http://dave.cheney.net/2015/08/22/cross-compilation-with-go-1-5
env GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.GitSHA1=$(git rev-parse --short HEAD)" -o "${PROG_NAME}-linux-amd64-${VERSION}" || { echo "Aua" 1>&2 && exit 1; }
env GOOS=linux GOARCH=arm GOARM=6 go build -ldflags="-s -w -X main.GitSHA1=$(git rev-parse --short HEAD)" -o "${PROG_NAME}-linux-arm-${VERSION}" || { echo "Aua" 1>&2 && exit 1; }
# env GOOS=linux GOARCH=386 GO386=387 go build -o "${PROG_NAME}-linux-386-${VERSION}" # https://github.com/golang/go/issues/11631
# env GOOS=darwin GOARCH=amd64 go build -o "${PROG_NAME}-darwin-amd64-${VERSION}"


"${say}" "simply"
# scp "ServerInfo.cgi" simply:/var/www/lighttpd/h4u.r-2.eu/public_html/"info.cgi"
gzip --force --best "${PROG_NAME}"-*-"${VERSION}" \
&& chmod a-x "${PROG_NAME}"-*-"${VERSION}.gz" \
&& rsync -vp --bwlimit=1234 "${PROG_NAME}"-*-"${VERSION}.gz" "simply:/tmp/" \
&& ssh simply "sh -c 'cd /var/www/lighttpd/l.mro.name/public_html/ && cp "/tmp/${PROG_NAME}-linux-amd64-${VERSION}.gz" shaarligo_cgi.gz && gunzip < shaarligo_cgi.gz > shaarligo.cgi && chmod a+x shaarligo.cgi && ls -l shaarligo?cgi*'" \
&& ssh simply "sh -c 'cd /var/www/lighttpd/b.r-2.eu/public_html/u/ && cp /var/www/lighttpd/l.mro.name/public_html/shaarligo?cgi* .'"

ssh simply "sh -c 'cd /var/www/lighttpd/b.mro.name/public_html/u/ && cp /var/www/lighttpd/l.mro.name/public_html/shaarligo?cgi* . && ls -l shaarligo?cgi*'"
"${say}" "ok"

"${say}" "vario"
# scp "ServerInfo.cgi" vario:~/mro.name/webroot/b/"info.cgi"
ssh vario "sh -c 'cd ~/mro.name/webroot/b/ && curl -L http://purl.mro.name/shaarligo_cgi.gz | tee shaarligo_cgi.gz | gunzip > shaarligo.cgi && chmod a+x shaarligo.cgi && ls -l shaarligo?cgi*'"
"${say}" "ok"

