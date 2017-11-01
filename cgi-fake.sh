#!/bin/sh

export cgi="atom.cgi"
export SCRIPT_NAME="/sub/${cgi}"
export PATH_INFO="/config"

go fmt *.go && go build -ldflags "-s" -o "atom.cgi" || exit 1

# export SERVER_PROTOCOL="HTTP/1.1"
# export REQUEST_METHOD="GET"
# export HTTP_HOST=example.com
# 
# rm -rf tmp 2>/dev/null
# mkdir tmp && cd tmp && time ../${cgi}
#
# exit 0

export SERVER_PROTOCOL="HTTP/1.1"
export REQUEST_METHOD="POST"
export HTTP_HOST=example.com
export CONTENT_TYPE="application/x-www-form-urlencoded"
export CONTENT_LENGTH="119"

rm -rf tmp 2>/dev/null
mkdir tmp && cd tmp && time ../${cgi} <<EOF
title=A&setlogin=B&setpassword=123456789012&import_shaarli_url=&import_shaarli_setlogin=&import_shaarli_setpassword=
EOF
