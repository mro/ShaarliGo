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

"$(go env GOPATH)/bin/go-bindata" -ignore="\\.DS_Store" -ignore=".+\\.woff" -prefix static static/... tpl/...

VERSION="$(grep -F 'version = ' version.go | cut -d \" -f 2)"
LDFLAGS="-s -w -X main.GitSHA1=$(git rev-parse --short HEAD)"

rm "shaarligo"-*-*".cgi"* 2>/dev/null

"${say}" "test"
umask 0022
go fmt && go vet && go test --short || { exit $?; }
"${say}" "ok"

tar -czf testdata.tar.gz testdata/*.html testdata/*.atom testdata/*.gob

"${say}" "build localhost"
go build -ldflags "${LDFLAGS}" -o "shaarligo.cgi" || { echo "Aua" 1>&2 && exit 1; }
cp "shaarligo.cgi" ~/"public_html/c/shaarligo.cgi"
"${say}" "ok"
# open "http://localhost/~$(whoami)/b/shaarligo.cgi"

"${say}" bench
go test -bench=.
"${say}" ok

"${say}" "linux build"
# http://dave.cheney.net/2015/08/22/cross-compilation-with-go-1-5
env GOOS=linux GOARCH=amd64       go build -ldflags="${LDFLAGS}" -o "shaarligo-Linux-x86_64.cgi" || { echo "Aua" 1>&2 && exit 1; }
env GOOS=linux GOARCH=arm GOARM=6 go build -ldflags="${LDFLAGS}" -o "shaarligo-Linux-armv6l.cgi" || { echo "Aua" 1>&2 && exit 1; }
# env GOOS=linux GOARCH=386 GO386=387 go build -o "shaarligo-linux-386-${VERSION}" # https://github.com/golang/go/issues/11631
# env GOOS=darwin GOARCH=amd64 go build -o "shaarligo-darwin-amd64-${VERSION}"


"${say}" "s0"
gzip --force --best "shaarligo-"*-*".cgi" \
&& rsync -vp --bwlimit=1234 "shaarligo-"*-*".cgi.gz" "s0:/var/www/lighttpd/l.mro.name/public_html/" \
&& ssh s0 "sh -c 'cd /var/www/lighttpd/l.mro.name/public_html/ && gunzip < shaarligo-$(uname -s)-$(uname -m).cgi.gz > shaarligo.cgi && ls -l shaarligo*cgi*'" \
&& ssh s0 "sh -c 'cd /var/www/lighttpd/demo.mro.name/public_html/shaarligo/ && cp /var/www/lighttpd/l.mro.name/public_html/shaarligo?cgi* . && ls -l shaarligo*cgi*'" \
&& ssh s0 "sh -c 'cd /var/www/lighttpd/b.r-2.eu/public_html/u/ && cp /var/www/lighttpd/l.mro.name/public_html/shaarligo?cgi* . && ls -l shaarligo*cgi*'" \
&& ssh s0 "sh -c 'cd /var/www/lighttpd/b.mro.name/public_html/u/ && cp /var/www/lighttpd/l.mro.name/public_html/shaarligo?cgi* . && ls -l shaarligo*cgi*'"
"${say}" "ok"

"${say}" "vario"
# scp "ServerInfo.cgi" vario:~/mro.name/webroot/b/"info.cgi"
ssh vario "sh -c 'cd ~/mro.name/webroot/b/ && curl -L http://purl.mro.name/shaarligo-$(uname -s)-$(uname -m).cgi.gz | tee shaarligo.cgi.gz | gunzip > shaarligo.cgi && chmod a+x shaarligo.cgi && ls -l shaarligo?cgi*'"
"${say}" "ok"

