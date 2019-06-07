#!/bin/sh
cd "$(dirname "${0}")/.."

chmod a-w shaarligo.cgi .htaccess app/.htaccess

ls -l shaarligo.cgi .htaccess app/.htaccess