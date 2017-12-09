//
// Copyright (C) 2017-2017 Marcus Rohrmoser, http://purl.mro.name/ShaarliGo
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
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"sort"
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
	isAlreadyConfigured := app.cfg.IsConfigured()
	if isAlreadyConfigured && !app.IsLoggedIn(now) {
		http.Error(w, "double check failed.", http.StatusInternalServerError)
		return
	}

	// unpack (nonexisting) static files
	for _, filename := range AssetNames() {
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			if err := RestoreAsset(".", filename); err != nil {
				http.Error(w, "failed "+filename+": "+err.Error(), http.StatusInternalServerError)
				return
			} else {
				log.Println(strings.Join([]string{"create ", filename}, ""))
			}
		} else {
			log.Println(strings.Join([]string{"keep   ", filename}, ""))
		}
	}
	// os.Chmod("app", os.FileMode(0750)) // not sure if this is a good idea.

	if err := r.ParseForm(); err != nil {
		http.Error(w, "couldn't parse form: "+err.Error(), http.StatusInternalServerError)
		return
	}

	switch r.Method {
	case http.MethodPost:
		app.cfg.Title = strings.TrimSpace(r.FormValue("title"))
		app.cfg.AuthorName = strings.TrimSpace(r.FormValue("setlogin"))
		pwd := strings.TrimSpace(r.FormValue("setpassword"))
		// https://astaxie.gitbooks.io/build-web-application-with-golang/en/09.5.html
		// $GLOBALS['salt'] = sha1(uniqid('',true).'_'.mt_rand()); // Salt renders rainbow-tables attacks useless.
		// original shaarli did $hash = sha1($password.$login.$GLOBALS['salt']);
		pwdBcrypt, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
		if err == nil {
			app.cfg.PwdBcrypt = string(pwdBcrypt)
			if len(app.cfg.AuthorName) < 1 || len([]rune(pwd)) < 12 {
				app.cfg.renderSettingsPage(w, http.StatusBadRequest)
				return
			}
			err = app.cfg.Save()
		}
		if err != nil {
			http.Error(w, "couldn't crypt pwd: "+err.Error(), http.StatusInternalServerError)
			return
		}
		authors := []Person{Person{Name: app.cfg.AuthorName}}

		// fork that one?
		//_, err := feedFromLegacyShaarli(r.FormValue("import_shaarli_url"), r.FormValue("import_shaarli_uid"), r.FormValue("import_shaarli_pwd"))
		// log.Println(err.Error())

		// if process is running: add a hint about the running background task into the response,
		// e.g. as a refresh timer. <meta http-equiv="refresh" content="5; URL=http://www.yourdomain.com/yoursite.html">

		urlBase := xmlBaseFromRequestURL(r.URL, os.Getenv("SCRIPT_NAME"))
		var feed Feed
		if isAlreadyConfigured {
			feed, _ = FeedFromFileName(fileFeedStorage)
		} else {
			// idxPost := idxBase + len(os.Getenv("SCRIPT_NAME"))
			// urlPost, err := url.Parse(strURL[:idxPost] + "/" + uriPub + "/" + uriPosts)
			// load template feed, set Id and birthday.
			// tagScheme := baseURL.ResolveReference(mustParseURL(uriPub, uriTags, "#")).String()
			feed = Feed{}

			feed.Append(&Entry{
				Title: HumanText{Body: "Hello, #Atom!"},
				Id:    "voo8Uo",
				Links: []Link{
					Link{Href: mustParseURL("http://www.loremipsum.de/").String()},
				},
				Categories: []Category{
					Category{Term: "🐳"},
					Category{Term: "Atom"},
					Category{Term: "opensource"},
					Category{Term: "ipsum"},
				},
				Content: &HumanText{Body: `Lorem #ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet.

Duis autem vel eum iriure dolor in hendrerit in vulputate velit esse molestie consequat, vel illum dolore eu feugiat nulla facilisis at vero eros et accumsan et iusto odio dignissim qui blandit praesent luptatum zzril delenit augue duis dolore te feugait nulla facilisi. Lorem ipsum dolor sit amet, consectetuer adipiscing elit, sed diam nonummy nibh euismod tincidunt ut laoreet dolore magna aliquam erat volutpat.

Ut wisi enim ad minim veniam, quis nostrud exerci tation ullamcorper suscipit lobortis nisl ut aliquip ex ea commodo consequat. Duis autem vel eum iriure dolor in hendrerit in vulputate velit esse molestie consequat, vel illum dolore eu feugiat nulla facilisis at vero eros et accumsan et iusto odio dignissim qui blandit praesent luptatum zzril delenit augue duis dolore te feugait nulla facilisi.`},
				Published: iso8601{mustParseRFC3339("2012-12-31T02:02:02+01:00")},
			})
			feed.Append(&Entry{
				Title:   HumanText{Body: "Was noch alles fehlt"},
				Id:      "Naev8k",
				Authors: authors,
				Categories: []Category{
					Category{Term: "🐳"},
					Category{Term: "Cloud"},
					Category{Term: "i18n"},
				},
				Content: &HumanText{Body: `- Posten, Löschen, Bookmarklet,
- 'API'-Kompatibilität mit Vanilla Shaarli (=> ShaarliOS, Shaarlier)
- neue Posts vorbelegen (optional: extern per http GET, rewrite URLs à la heise),
- Tag #Cloud,
- Shaarli Import,
- Tagesansicht,
- Suche,
- PuSH,
- Komplett-Feed (Archiv),
- clickbare Links im Text (client-seitig),
- Referer-Anonymisierer,
- Bilder Cache/Proxy,
- QR-Code pro Post,
- Skinning/Themeing (asset dir),
- #i18n,
- private Posts,
- Kommentare,
- Bilder hochladen,
- AtomPub, https://tools.ietf.org/html/rfc5023#section-9.2
- Html/Markdown (client-seitig),
- Atom Aggregator?`},
				Published: iso8601{mustParseRFC3339("2012-12-31T01:01:01+01:00")},
			})
			feed.Append(&Entry{
				Title: HumanText{Body: "Shaarli — sebsauvage.net"},
				Id:    "kaJ9Rw",
				Links: []Link{
					Link{Href: mustParseURL("http://sebsauvage.net/wiki/doku.php?id=php:shaarli").String()},
				},
				Authors:        authors,
				Categories:     []Category{Category{Term: "opensource"}, Category{Term: "Software"}},
				Published:      iso8601{mustParseRFC3339("2011-09-13T15:45:00+02:00")},
				Content:        &HumanText{Body: "Welcome to Shaarli ! This is a bookmark. To edit or delete me, you must first login."},
				MediaThumbnail: &MediaThumbnail{Url: mustParseURL("https://cdn.rawgit.com/mro/ShaarliOS/master/shaarli-petal.svg").String()},
			})
		}

		feed.Title = HumanText{Body: app.cfg.Title}
		feed.XmlBase = urlBase.String()
		feed.Id = urlBase.String() // expand XmlBase as required by https://validator.w3.org/feed/check.cgi?url=
		feed.Authors = authors
		feed.Generator = &Generator{Uri: myselfNamespace, Version: "0.0.2", Body: "ShaarliGo"}
		feed.Links = []Link{
			Link{Rel: relEdit, Href: path.Join(cgiName, uriPub, uriPosts), Title: "PostURI, maybe better a app:collection https://tools.ietf.org/html/rfc5023#section-8.3.3"},
		}
		sort.Sort(ByPublishedDesc(feed.Entries))
		feed.Save(fileFeedStorage)
		// TODO: make persistent

		if err = feed.replaceFeeds(); err != nil {
			http.Error(w, "couldn't write feeds: "+err.Error(), http.StatusInternalServerError)
			return
		}

		app.startSession(w, r, now)
		http.Redirect(w, r, path.Join("..", "..", uriPub, uriPosts)+"/", http.StatusFound)
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
			"<?xml-stylesheet type='text/xsl' href='"+path.Join("..", "..", "assets", "default", "de", "config.xslt")+"'?>\n")
		io.WriteString(w, `<!--
  The html you see here is for compatibilty with vanilla shaarli.

  The main reason is backward compatibility for e.g. https://github.com/mro/ShaarliOS and
  https://github.com/dimtion/Shaarlier as tested via
  https://github.com/mro/Shaarli-API-test
-->
`)
		err = tmpl.Execute(w, map[string]string{
			"setlogin": cfg.AuthorName,
			"title":    cfg.Title,
		})
	}
	if err != nil {
		http.Error(w, "couldn't restore assets: "+err.Error(), http.StatusInternalServerError)
	}
}
