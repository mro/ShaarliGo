#!/bin/sh
cd "$(dirname "${0}")/.."

chmod a-w *.cgi .htaccess app/.htaccess

ls -l *.cgi .htaccess app/.htaccess
