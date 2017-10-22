#!/bin/sh

export SCRIPT_NAME="/atom.cgi"
export PATH_INFO="/settings"

export SERVER_PROTOCOL="HTTP/1.1"
export REQUEST_METHOD="POST"
export HTTP_HOST=example.com
export CONTENT_TYPE="application/x-www-form-urlencoded"
export CONTENT_LENGTH="119"

go fmt *.go && go build -ldflags "-s" -o "atom.cgi" || exit 1

rm -rf tmp 2>/dev/null
mkdir tmp && cd tmp && time ..${SCRIPT_NAME} <<EOF
title=A&author%2Fname=B&password=123456789012&import_shaarli_url=&import_shaarli_setlogin=&import_shaarli_setpassword=
EOF
