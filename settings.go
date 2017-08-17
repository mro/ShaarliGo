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
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func urlFromString(raw string) (url *url.URL, err error) {
	url, err = url.Parse(raw)
	if err != nil {
		url, err = url.Parse("https://" + raw)
	}
	return
}

func writeElementKeyValue(enc *xml.Encoder, element string, key string, value string) {
	a := []xml.Attr{xml.Attr{Name: xml.Name{Local: "name"}, Value: key}, xml.Attr{Name: xml.Name{Local: "content"}, Value: value}}
	n := xml.Name{Local: "meta"}
	enc.EncodeToken(xml.StartElement{Name: n, Attr: a})
	enc.EncodeToken(xml.EndElement{Name: n})
}

func feedFromLegacyShaarli(urlbase string, uid string, pwd string) (feed Feed, err error) {
	return
}

func handleSettings(mgr *SessionManager, w http.ResponseWriter, r *http.Request) error {
	isAlreadyConfigured := mgr.IsConfigured()
	w.Header().Set("Server", myselfNamespace)
	w.Header().Set("Handler", "handleSettings")
	w.Header().Set("Content-Type", "application/xhtml+xml; charset=utf-8")

	if isAlreadyConfigured && !mgr.IsLoggedIn(r, time.Now()) {
		panic("double check failed.")
	}

	// unpack (nonexisting) static files
	for _, filename := range AssetNames() {
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			err := RestoreAsset(".", filename)
			if err != nil {
				return err
			}
		}
	}
	// os.Chmod("app", os.FileMode(0750)) // not sure if this is a good idea.

	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	switch r.Method {
	case "POST":
		uid := strings.TrimSpace(r.FormValue("author/name"))
		if uid != "" {
			mgr.config.AuthorName = uid // any conditions? suitable for author: https://gist.github.com/tjdett/4617547#file-atom-rng-xml-L64
		}

		tit := strings.TrimSpace(r.FormValue("title"))
		if tit != "" {
			mgr.config.Title = tit
		}
		if mgr.config.Title == "" {
			mgr.config.Title = "Shared links on …"
		}

		pwd := strings.TrimSpace(r.FormValue("password"))

		if !mgr.IsConfigured() || len([]rune(pwd)) < 12 {
			renderPage(http.StatusBadRequest, w, &mgr.config)
			return nil
		}

		// https://astaxie.gitbooks.io/build-web-application-with-golang/en/09.5.html
		// $GLOBALS['salt'] = sha1(uniqid('',true).'_'.mt_rand()); // Salt renders rainbow-tables attacks useless.
		// original shaarli did $hash = sha1($password.$login.$GLOBALS['salt']);
		pwdBcrypt, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
		if err == nil {
			mgr.config.PwdBcrypt = string(pwdBcrypt)
			err = mgr.config.Save()
		}
		if err != nil {
			return err
		}

		// fork that one?
		//_, err := feedFromLegacyShaarli(r.FormValue("import_shaarli_url"), r.FormValue("import_shaarli_uid"), r.FormValue("import_shaarli_pwd"))
		// log.Println(err.Error())

		// if process is running: add a hint about the running background task into the response,
		// e.g. as a refresh timer. <meta http-equiv="refresh" content="5; URL=http://www.yourdomain.com/yoursite.html">

		if !isAlreadyConfigured {
			err = mgr.generateAtomFeeds()
			if err != nil {
				return err
			}
		}

		// all went well:
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Location", "../pub/posts/")
		w.WriteHeader(http.StatusSeeOther)
		io.WriteString(w, "let's go to "+"../pub/posts/"+"\n")
	case "GET":
		renderPage(http.StatusOK, w, &mgr.config)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}

	return nil
}

func renderPage(code int, w http.ResponseWriter, c *Config) {
	w.WriteHeader(code)
	renderPageXml(w, c)
}

func renderPageXml(w io.Writer, c *Config) {
	enc := xml.NewEncoder(w)
	enc.Indent("", "  ")
	enc.EncodeToken(xml.ProcInst{"xml", []byte(`version="1.0" encoding="UTF-8"`)})
	enc.EncodeToken(xml.CharData("\n"))
	enc.EncodeToken(xml.ProcInst{"xml-stylesheet", []byte("type='text/xsl' href='../assets/settings.xslt'")})
	enc.EncodeToken(xml.CharData("\n"))

	n := xml.Name{Local: "as:setup"}
	enc.EncodeToken(xml.StartElement{Name: n, Attr: []xml.Attr{
		xml.Attr{Name: xml.Name{Local: "xmlns:as"}, Value: myselfNamespace},
		xml.Attr{Name: xml.Name{Local: "xmlns"}, Value: atomNamespace}},
	})
	// todo: link/@rel nach https://martinfowler.com/articles/richardsonMaturityModel.html

	{
		title := xml.Name{Local: "title"}
		enc.EncodeToken(xml.StartElement{Name: title})
		enc.EncodeToken(xml.CharData(c.Title))
		enc.EncodeToken(xml.EndElement{Name: title})
	}

	{
		author := xml.Name{Local: "author"}
		enc.EncodeToken(xml.StartElement{Name: author})
		{
			name := xml.Name{Local: "name"}
			enc.EncodeToken(xml.StartElement{Name: name})
			enc.EncodeToken(xml.CharData(c.AuthorName))
			enc.EncodeToken(xml.EndElement{Name: name})
		}
		enc.EncodeToken(xml.EndElement{Name: author})
	}

	enc.EncodeToken(xml.EndElement{Name: n})
	enc.EncodeToken(xml.CharData("\n"))
	enc.Flush()
}
