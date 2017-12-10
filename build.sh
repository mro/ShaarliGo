#!/bin/sh
# https://golang.org/doc/install/source#environment
#

cd "$(dirname "${0}")"
# $ uname -s -m
# Darwin x86_64
# Linux x86_64
# Linux armv6l

say "go get"
go get -u golang.org/x/tools/blog/atom
go get -u golang.org/x/crypto/bcrypt
go get -u gopkg.in/yaml.v2
go get -u github.com/jteeuwen/go-bindata/...
go get -u github.com/gorilla/sessions
# for testing only:
go get -u github.com/yhat/scrape
go get -u golang.org/x/net/html
go get -u golang.org/x/net/html/atom
say "ok"

# ssh vario find mro.name/vorschau.blog/assets -type f

# rsync -aPz --delete --delete-excluded --exclude jquery* --exclude *.zip --exclude *.html vario:~/mro.name/vorschau.blog/assets/ static/assets
go-bindata -ignore=\\.DS_Store -prefix static static/...

PROG_NAME="ShaarliGo"
VERSION=`fgrep 'version = ' version.go | cut -d '"' -f 2`

rm "${PROG_NAME}"-*-"${VERSION}" 2>/dev/null

say "test"
go fmt && go test --short || { exit $?; }
say "ok"

say "ithneen"
go build -ldflags "-v=2 -s -w -X main.GitSHA1=$(git rev-parse --short HEAD)" -o ~/Sites/b/shaarligo.cgi || { echo "Aua" 1>&2 && exit 1; }
say "ok"
# open "http://localhost/~$(whoami)/b/shaarligo.cgi"

say "linux"
# http://dave.cheney.net/2015/08/22/cross-compilation-with-go-1-5
# env GOOS=linux GOARCH=arm GOARM=6 go build -o "${PROG_NAME}-linux-arm-${VERSION}"
env GOOS=linux GOARCH=amd64 go build -ldflags="-v=2 -s -w -X main.GitSHA1=$(git rev-parse --short HEAD)" -o "${PROG_NAME}-linux-amd64-${VERSION}" || { echo "Aua" 1>&2 && exit 1; }
# env GOOS=linux GOARCH=386 GO386=387 go build -o "${PROG_NAME}-linux-386-${VERSION}" # https://github.com/golang/go/issues/11631
# env GOOS=darwin GOARCH=amd64 go build -o "${PROG_NAME}-darwin-amd64-${VERSION}"
say "ok"

# https://lager.mro.name/as/shaarligo.cgi
# scp "${PROG_NAME}-linux-amd64-${VERSION}" simply:/var/www/lighttpd/lager.mro.name/public_html/as/"shaarligo.cgi"
# scp "ServerInfo.cgi" simply:/var/www/lighttpd/lager.mro.name/public_html/as/"info.cgi"
# ssh simply rm -vrf /var/www/lighttpd/lager.mro.name/public_html/as/assets
# ssh simply rm -vrf /var/www/lighttpd/lager.mro.name/public_html/as/app

# ssh vario rm -vrf mro.name/webroot/b/.htaccess
# ssh vario rm -vrf mro.name/webroot/b/app
# ssh vario rm -vrf mro.name/webroot/b/assets
# ssh vario rm -vrf mro.name/webroot/b/pub

say "simply"
scp "${PROG_NAME}-linux-amd64-${VERSION}" simply:/var/www/lighttpd/h4u.r-2.eu/public_html/"shaarligo.cgi"
say "ok"
# scp "ServerInfo.cgi" simply:/var/www/lighttpd/h4u.r-2.eu/public_html/"info.cgi"

# scp "${PROG_NAME}-linux-amd64-${VERSION}" vario:~/mro.name/webroot/b/"shaarligo.cgi"
# scp "ServerInfo.cgi" vario:~/mro.name/webroot/b/"info.cgi"

say "vario"
ssh simply scp /var/www/lighttpd/h4u.r-2.eu/public_html/shaarligo.cgi vario:~/mro.name/webroot/b/
ssh vario ls -Al mro.name/webroot/b/shaarligo.cgi
say "ok"

exit 0

# curl --data-urlencode "url=wall" --dump-header head.txt "http://vorschau.blog.mro.name/${PROG_NAME}.cgi"
# curl --location --dump-header head.txt "http://vorschau.blog.mro.name/"
# echo "===="
#cat head.txt

# curl --location 'http://vorschau.blog.mro.name/shaarligo.cgi/settings?foo' ; say 'aha, aha, soso'

curl --dump-header head0.txt --location 'http://mro.name/b/shaarligo.cgi/config' \
  --data-urlencode 'title=🔗 My Bookmärks' \
  --data-urlencode 'setlogin=Bö' \
  --data-urlencode 'setpassword=123456789012' \
  --data-urlencode 'import_shaarli_url=' \
  --data-urlencode 'import_shaarli_setlogin=' \
  --data-urlencode 'import_shaarli_setpassword=' \
> body0.txt ; cat head0.txt body0.txt ; say 'aha, aha, soso'

# curl --dump-header head1.txt --location 'http://mro.name/b/shaarligo.cgi?do=login' \
# > body1.txt ; cat head1.txt body1.txt ; say 'aha, aha, soso'
