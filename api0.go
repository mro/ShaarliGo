//
// Copyright (C) 2017-2017 Marcus Rohrmoser, http://purl.mro.name/AtomicShaarli
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
//

package main

import (
	"errors"
	"html/template"
	"io"
	"net/http"
)

func (app *App) handleDoLogin(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	//
	// and https://github.com/mro/ShaarliOS/blob/master/ios/ShaarliOS/API/ShaarliCmd.m#L386
	case "GET":
		if tmpl, err := template.New("login").Parse(`<html xmlns="http://www.w3.org/1999/xhtml">
<body>
  <form method="post" name="loginform" id="loginform">
    <input type="text" name="login" />
    <input type="password" name="password" />
    <input type="submit" value="Login" />
    <input type="checkbox" name="longlastingsession" id="longlastingsession" />
    <input type="hidden" name="token" value="{{index . "token"}}" />
  </form>
</body>
</html>
`); err == nil {
			w.Header().Set("Content-Type", "text/xml; charset=utf-8")
			io.WriteString(w, `<?xml version="1.0" encoding="UTF-8"?>
<?xml-stylesheet type='text/xsl' href='./assets/default/de/do-login.xslt'?>
<!--
  must be compatible with https://github.com/mro/Shaarli-API-test/blob/master/tests/test-login-ok.sh
  https://github.com/mro/ShaarliOS/blob/master/ios/ShaarliOS/API/ShaarliCmd.m#L386
-->
`)
			return tmpl.Execute(w, map[string]string{
				"token": "ff13e7eaf9541ca2ba30fd44e864c3ff014d2bc9",
			})
		}
	case "POST":
	}
	//     NSString *xpath = [NSString stringWithFormat:@"/html/body//form[@name='%1$@']//input[(@type='text' or @type='password' or @type='hidden' or @type='checkbox') and @name] | /html/body//form[@name='%1$@']//textarea[@name]

	// 'POST' validate, respond error (and squeal) or set session and redirect
	return errors.New("'login' not implemented yet.")
}

func (app *App) handleDoLogout(w http.ResponseWriter, r *http.Request) error {
	io.WriteString(w, "logout"+"\n")
	//  invalidate session and redirect
	return errors.New("'login' not implemented yet.")
}

func (app *App) handleDoPost(w http.ResponseWriter, r *http.Request) error {
	io.WriteString(w, "post"+"\n")
	// 'GET': send a form to the client
	// must be compatible to https://github.com/mro/Shaarli-API-Test/...

	// 'POST' validate, respond error (and squeal) or post and redirect
	return errors.New("'login' not implemented yet.")
}
