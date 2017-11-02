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
	"encoding/xml"
	"errors"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// unused
func _urlFromString(raw string) (url *url.URL, err error) {
	url, err = url.Parse(raw)
	if err != nil {
		url, err = url.Parse("https://" + raw)
	}
	return
}

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

func encodeValueElement(enc *xml.Encoder, name string, value string) *xml.Encoder {
	elm := xml.Name{Local: name}
	enc.EncodeToken(xml.StartElement{Name: elm})
	enc.EncodeToken(xml.CharData(value))
	enc.EncodeToken(xml.EndElement{Name: elm})
	return enc
}

func encodeMetaNameContent(enc *xml.Encoder, name string, content string) {
	a := []xml.Attr{xml.Attr{Name: xml.Name{Local: "name"}, Value: name}, xml.Attr{Name: xml.Name{Local: "content"}, Value: content}}
	n := xml.Name{Local: "meta"}
	enc.EncodeToken(xml.StartElement{Name: n, Attr: a})
	enc.EncodeToken(xml.EndElement{Name: n})
}

func feedFromLegacyShaarli(urlbase string, uid string, pwd string) (feed Feed, err error) {
	return
}

func (app *App) handleSettings(w http.ResponseWriter, r *http.Request) error {
	isAlreadyConfigured := app.cfg.IsConfigured()

	now := time.Now()
	if isAlreadyConfigured && !app.IsLoggedIn(now) {
		return errors.New("double check failed.")
	}

	// unpack (nonexisting) static files
	for _, filename := range AssetNames() {
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			if err := RestoreAsset(".", filename); err != nil {
				return err
			}
		}
	}
	// os.Chmod("app", os.FileMode(0750)) // not sure if this is a good idea.

	if err := r.ParseForm(); err != nil {
		return err
	}

	switch r.Method {
	case http.MethodPost:
		app.cfg.AuthorName = strings.TrimSpace(r.FormValue("setlogin"))
		app.cfg.Title = strings.TrimSpace(r.FormValue("title"))
		pwd := strings.TrimSpace(r.FormValue("setpassword"))

		if !app.cfg.IsConfigured() || len([]rune(pwd)) < 12 {
			return renderSettingsPage(&app.cfg, http.StatusBadRequest, w)
		}

		// https://astaxie.gitbooks.io/build-web-application-with-golang/en/09.5.html
		// $GLOBALS['salt'] = sha1(uniqid('',true).'_'.mt_rand()); // Salt renders rainbow-tables attacks useless.
		// original shaarli did $hash = sha1($password.$login.$GLOBALS['salt']);
		pwdBcrypt, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
		if err == nil {
			app.cfg.PwdBcrypt = string(pwdBcrypt)
			err = app.cfg.Save()
		}
		if err != nil {
			return err
		}

		// fork that one?
		//_, err := feedFromLegacyShaarli(r.FormValue("import_shaarli_url"), r.FormValue("import_shaarli_uid"), r.FormValue("import_shaarli_pwd"))
		// log.Println(err.Error())

		// if process is running: add a hint about the running background task into the response,
		// e.g. as a refresh timer. <meta http-equiv="refresh" content="5; URL=http://www.yourdomain.com/yoursite.html">

		urlBase := xmlBaseFromRequestURL(r.URL, os.Getenv("SCRIPT_NAME"))

		if !isAlreadyConfigured {
			// idxPost := idxBase + len(os.Getenv("SCRIPT_NAME"))
			// urlPost, err := url.Parse(strURL[:idxPost] + "/" + uriPub + "/" + uriPosts)
			// load template feed, set Id and birthday.
			// tagScheme := baseURL.ResolveReference(mustParseURL(uriPub, uriTags, "#")).String()
			authors := []Person{Person{Name: app.cfg.AuthorName}}
			feed := &Feed{
				XmlBase:   urlBase.String(),
				Id:        urlBase.String(), // expand XmlBase as required by https://validator.w3.org/feed/check.cgi?url=
				Title:     HumanText{Body: app.cfg.Title},
				Generator: &Generator{Uri: myselfNamespace, Version: "0.0.1", Body: "AtomicShaarli"},
				Links: []Link{
					Link{Rel: relEdit, Href: path.Join(cgiName, uriPub, uriPosts), Title: "PostURI, maybe better a app:collection https://tools.ietf.org/html/rfc5023#section-8.3.3"},
				},
			}

			feed.Append(&Entry{
				Title: HumanText{Body: "Hello, #Atom!"},
				Id:    "voo8Uo",
				Links: []Link{
					Link{Href: mustParseURL("http://www.loremipsum.de/").String()},
				},
				Authors: authors,
				Categories: []Category{
					Category{Term: "🐳"},
					Category{Term: "Atom"},
					Category{Term: "opensource"},
					Category{Term: "ipsum"},
				},
				Content: &HumanText{Body: `Lorem #ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet.

			   Duis autem vel eum iriure dolor in hendrerit in vulputate velit esse molestie consequat, vel illum dolore eu feugiat nulla facilisis at vero eros et accumsan et iusto odio dignissim qui blandit praesent luptatum zzril delenit augue duis dolore te feugait nulla facilisi. Lorem ipsum dolor sit amet, consectetuer adipiscing elit, sed diam nonummy nibh euismod tincidunt ut laoreet dolore magna aliquam erat volutpat.

			   Ut wisi enim ad minim veniam, quis nostrud exerci tation ullamcorper suscipit lobortis nisl ut aliquip ex ea commodo consequat. Duis autem vel eum iriure dolor in hendrerit in vulputate velit esse molestie consequat, vel illum dolore eu feugiat nulla facilisis at vero eros et accumsan et iusto odio dignissim qui blandit praesent luptatum zzril delenit augue duis dolore te feugait nulla facilisi.`},
				Updated: iso8601{mustParseRFC3339("2012-12-31T02:02:02+01:00")},
			}).Append(&Entry{
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
			   - Tag #Cloud,
			   - Tagesansicht,
			   - Shaarli Import,
			   - Komplett-Feed (Archiv),
			   - neue Posts vorbelegen (optional: extern per http GET),
			   - Suche,
			   - clickbare Links im Text (client-seitig),
			   - Referer-Anonymisierer,
			   - Bilder Cache/Proxy,
			   - QR-Code pro Post,
			   - Skinning/Themeing (asset dir),
			   - private Posts,
			   - Kommentare,
			   - PuSH,
			   - Bilder hochladen,
			   - #i18n,
			   - Html/Markdown (client-seitig),
			   - Atom Aggregator?`},
				Updated: iso8601{mustParseRFC3339("2012-12-31T01:01:01+01:00")},
			}).Append(&Entry{
				Title: HumanText{Body: "Shaarli — sebsauvage.net"},
				Id:    "kaJ9Rw",
				Links: []Link{
					Link{Href: mustParseURL("http://sebsauvage.net/wiki/doku.php?id=php:shaarli").String()},
				},
				Authors:        authors,
				Categories:     []Category{Category{Term: "opensource"}, Category{Term: "Software"}},
				Updated:        iso8601{mustParseRFC3339("2011-09-13T15:45:00+02:00")},
				Content:        &HumanText{Body: "Welcome to Shaarli ! This is a bookmark. To edit or delete me, you must first login."},
				MediaThumbnail: &MediaThumbnail{Url: mustParseURL("http://cdn.rawgit.com/mro/ShaarliOS/master/shaarli-petal.svg").String()},
			})

			sort.Reverse(feed.Entries)
			// TODO: make persistent

			if err = feed.replaceFeeds(); err != nil {
				return err
			}
		}

		app.startSession(w, r, now)

		// all went well:
		http.Redirect(w, r, path.Join("..", uriPub, uriPosts)+"/", http.StatusFound)
	case http.MethodGet:
		return renderSettingsPage(&app.cfg, http.StatusOK, w)
	default:
		http.Error(w, "MethodNotAllowed", http.StatusMethodNotAllowed)
	}

	return nil
}

func renderSettingsPage(c *Config, code int, w http.ResponseWriter) error {
	if tmpl, err := template.New("settings").Parse(`<html xmlns="http://www.w3.org/1999/xhtml">
  <head/>
  <body>
    <form method="post" action="#" name="installform" id="installform">
      <input type="text" name="setlogin" value="{{index . "setlogin"}}"/>
      <input type="password" name="setpassword" />
      <input type="text" name="title" />
      <input type="submit" name="Save" value="Save config" />
    </form>
  </body>
</html>
`); err == nil {
		w.Header().Set("Content-Type", "text/xml; charset=utf-8")
		w.WriteHeader(code)

		io.WriteString(w, "<?xml version='1.0' encoding='UTF-8'?>\n"+
			"<?xml-stylesheet type='text/xsl' href='"+path.Join("..", "assets", "default", "de", "config.xslt")+"'?>\n")
		io.WriteString(w, `<!--
	The html you see here is for compatibilty with vanilla shaarli.
	
	The main reason is backward compatibility for e.g. https://github.com/mro/ShaarliOS and
	https://github.com/dimtion/Shaarlier as tested via
	https://github.com/mro/Shaarli-API-test
-->
`)
		return tmpl.Execute(w, map[string]string{
			"setlogin": c.AuthorName,
		})
	} else {
		return err
	}
}
