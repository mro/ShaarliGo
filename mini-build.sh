#!/bin/sh
cd "$(dirname "${0}")"

go-bindata -ignore=\\.DS_Store -prefix static static/... \
&& go fmt \
&& go test --short \
&& go build -ldflags "-s -w -X main.GitSHA1=$(git rev-parse --short HEAD)" -o ~/Sites/b/shaarligo.cgi \
|| { echo "Aua" 1>&2 && exit 1; }

say "na los"
ls -Al ~/Sites/b/shaarligo.cgi
