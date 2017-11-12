#!/bin/sh

export cgi="shaarligo.cgi"

go fmt *.go && go build -ldflags "-s" -o "${cgi}" || exit 1

export SCRIPT_NAME="/sub/${cgi}"
export SERVER_PROTOCOL="HTTP/1.1"
export HTTP_HOST="example.com"

get() {
  export REQUEST_METHOD="GET"
  export PATH_INFO="${1}"
  time "../${cgi}"
}

post() {
  export REQUEST_METHOD="POST"
  export CONTENT_TYPE="application/x-www-form-urlencoded"
  export PATH_INFO="${1}"
  export CONTENT_LENGTH="${2}"
  time "../${cgi}"
}

rm -rf "tmp" 2>/dev/null
mkdir "tmp" && cd "tmp" && time "../${cgi}"

post "/config" 119 <<EOF
title=A&setlogin=B&setpassword=123456789012&import_shaarli_url=&import_shaarli_setlogin=&import_shaarli_setpassword=
EOF

export QUERY_STRING="?post=www.heise.de"
get

# get "/config"
