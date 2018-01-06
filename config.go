//
// Copyright (C) 2017-2018 Marcus Rohrmoser, http://purl.mro.name/ShaarliGo
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
	"html/template"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func xmlBaseFromRequestURL(r *url.URL, scriptName string) *url.URL {
	dir := path.Dir(scriptName)
	if dir[len(dir)-1:] != "/" {
		dir = dir + "/"
	}
	return mustParseURL(r.Scheme + "://" + r.Host + dir)
}

func mustParseRFC3339(str string) time.Time {
	if ret, err := time.Parse(time.RFC3339, str); err != nil {
		panic(err)
	} else {
		return ret
	}
}

func feedFromLegacyShaarli(urlbase string, uid string, pwd string) (feed Feed, err error) {
	return
}

func (app *App) handleSettings(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	if app.cfg.IsConfigured() && !app.IsLoggedIn(now) {
		http.Error(w, "double check failed.", http.StatusInternalServerError)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "couldn't parse form: "+err.Error(), http.StatusInternalServerError)
		return
	}

	switch r.Method {
	case http.MethodPost:
		uid := strings.TrimSpace(r.FormValue("setlogin"))
		pwd := strings.TrimSpace(r.FormValue("setpassword"))
		title := strings.TrimSpace(r.FormValue("title"))
		// https://astaxie.gitbooks.io/build-web-application-with-golang/en/09.5.html
		// $GLOBALS['salt'] = sha1(uniqid('',true).'_'.mt_rand()); // Salt renders rainbow-tables attacks useless.
		// original shaarli did $hash = sha1($password.$login.$GLOBALS['salt']);
		if pwdBcrypt, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost); err != nil {
			http.Error(w, "couldn't crypt pwd: "+err.Error(), http.StatusInternalServerError)
			return
		} else {
			if len(uid) < 1 || len([]rune(pwd)) < 12 {
				app.cfg.renderSettingsPage(w, http.StatusBadRequest)
				return
			}
			app.cfg.Title = title
			app.cfg.Uid = uid
			app.cfg.PwdBcrypt = string(pwdBcrypt)
			err = app.cfg.Save()
		}

		if feed, err := app.LoadFeed(); err != nil {
			http.Error(w, "couldn't load seed feed feeds: "+err.Error(), http.StatusInternalServerError)
			return
		} else {
			feed.XmlBase = xmlBaseFromRequestURL(r.URL, os.Getenv("SCRIPT_NAME")).String()
			feed.Id = feed.XmlBase // expand XmlBase as required by https://validator.w3.org/feed/check.cgi?url=
			feed.Title = HumanText{Body: title}
			feed.Authors = []Person{Person{Name: uid}}
			feed.Generator = &Generator{Uri: myselfNamespace, Version: version + "+" + GitSHA1, Body: "ShaarliGo"}
			feed.Links = []Link{
				Link{Rel: relEdit, Href: path.Join(cgiName, uriPub, uriPosts), Title: "PostURI, maybe better a app:collection https://tools.ietf.org/html/rfc5023#section-8.3.3"},
			}

			// fork that one?
			//_, err := feedFromLegacyShaarli(r.FormValue("import_shaarli_url"), r.FormValue("import_shaarli_uid"), r.FormValue("import_shaarli_pwd"))
			// log.Println(err.Error())

			// if process is running: add a hint about the running background task into the response,
			// e.g. as a refresh timer. <meta http-equiv="refresh" content="5; URL=http://www.yourdomain.com/yoursite.html">

			if err := app.SaveFeed(feed); err != nil {
				http.Error(w, "couldn't store feed data: "+err.Error(), http.StatusInternalServerError)
				return
			}
			if err = app.replaceFeeds(feed); err != nil {
				http.Error(w, "couldn't write feeds: "+err.Error(), http.StatusInternalServerError)
				return
			}

			app.startSession(w, r, now)
			http.Redirect(w, r, path.Join("..", "..", uriPub, uriPosts)+"/", http.StatusFound)
		}
	case http.MethodGet:
		app.cfg.renderSettingsPage(w, http.StatusOK)
	default:
		http.Error(w, "MethodNotAllowed", http.StatusMethodNotAllowed)
	}
}

func (cfg Config) renderSettingsPage(w http.ResponseWriter, code int) {
	tmpl, err := template.New("settings").Parse(`<html xmlns="http://www.w3.org/1999/xhtml">
  <head/>
  <body>
    <form method="post" name="installform" id="installform">
      <input type="text" name="setlogin" value="{{index . "setlogin"}}"/>
      <input type="password" name="setpassword" />
      <input type="text" name="title" value="{{index . "title"}}"/>
      <input type="submit" name="Save" value="Save config" />
    </form>
  </body>
</html>
`)
	if err == nil {
		w.Header().Set("Content-Type", "text/xml; charset=utf-8")
		w.WriteHeader(code)

		io.WriteString(w, "<?xml version='1.0' encoding='UTF-8'?>\n"+
			"<?xml-stylesheet type='text/xsl' href='"+path.Join("..", "..", "assets", cfg.Skin, "config.xslt")+"'?>\n")
		io.WriteString(w, `<!--
  The html you see here is for compatibilty with vanilla shaarli.

  The main reason is backward compatibility for e.g. https://github.com/mro/ShaarliOS and
  https://github.com/dimtion/Shaarlier as tested via
  https://github.com/mro/Shaarli-API-test
-->
`)
		err = tmpl.Execute(w, map[string]string{
			"title":    cfg.Title,
			"setlogin": cfg.Uid,
		})
	}
	if err != nil {
		http.Error(w, "couldn't restore assets: "+err.Error(), http.StatusInternalServerError)
	}
}
