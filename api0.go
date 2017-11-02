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
	"fmt"
	"html/template"
	"io"
	"net/http"
	"path"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func (app *App) handleDoLogin(w http.ResponseWriter, r *http.Request) error {
	now := time.Now()
	switch r.Method {
	// and https://github.com/mro/ShaarliOS/blob/master/ios/ShaarliOS/API/ShaarliCmd.m#L386
	case http.MethodGet:
		if tmpl, err := template.New("login").Parse(`<html xmlns="http://www.w3.org/1999/xhtml">
<head><title>{{index . "title"}}</title></head>
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
				"title": app.cfg.Title,
				"token": "ff13e7eaf9541ca2ba30fd44e864c3ff014d2bc9",
			})
		}
	case http.MethodPost:
		// todo: verify token
		uid := strings.TrimSpace(r.FormValue("login"))
		pwd := strings.TrimSpace(r.FormValue("password"))
		// compute anyway (a bit more time constantness)
		err := bcrypt.CompareHashAndPassword([]byte(app.cfg.PwdBcrypt), []byte(pwd))
		if uid != app.cfg.AuthorName || err == bcrypt.ErrMismatchedHashAndPassword {
			squealFailure(r, now)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return nil
		}
		if err == nil {
			err = app.startSession(w, r, now)
		}
		if err == nil {
			http.Redirect(w, r, path.Join(uriPub, uriPosts)+"/", http.StatusSeeOther)
		}
		return err
	default:
		squealFailure(r, now)
		http.Error(w, "MethodNotAllowed", http.StatusMethodNotAllowed)
	}
	//     NSString *xpath = [NSString stringWithFormat:@"/html/body//form[@name='%1$@']//input[(@type='text' or @type='password' or @type='hidden' or @type='checkbox') and @name] | /html/body//form[@name='%1$@']//textarea[@name]

	// 'POST' validate, respond error (and squeal) or set session and redirect
	return nil
}

func (app *App) handleDoLogout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("ses_timeout_logout", fmt.Sprintf("%v", app.ses.Values["timeout"]))
	if err := app.stopSession(w, r); err != nil {
		http.Error(w, "Couldn't end session: "+err.Error(), http.StatusInternalServerError)
	} else {
		http.Redirect(w, r, path.Join(uriPub, uriPosts)+"/", http.StatusSeeOther)
	}
}

func (app *App) handleDoPost(w http.ResponseWriter, r *http.Request) error {
	io.WriteString(w, "post"+"\n")
	// 'GET': send a form to the client
	// must be compatible to https://github.com/mro/Shaarli-API-Test/...

	// 'POST' validate, respond error (and squeal) or post and redirect
	return errors.New("'login' not implemented yet.")
}
