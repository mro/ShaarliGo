#!/bin/sh
cd "$(dirname "${0}")"

say="say"
umask 0022

go-bindata -ignore=\\.DS_Store -prefix static static/... \
&& go fmt \
&& go vet \
&& go test --short \
&& go build -ldflags "-s -w -X main.GitSHA1=$(git rev-parse --short HEAD)" -o ~/Sites/b/shaarligo.cgi \
|| { echo "Aua" 1>&2 && exit 1; }

"${say}" "pack mas"
ls -Al ~/Sites/b/shaarligo.cgi
echo "http://$(hostname)/~$(whoami)/b/"
