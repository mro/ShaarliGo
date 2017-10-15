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
	"bytes"
	"encoding/xml"
	"io"
	"net/http"
	"net/url"
	"os"
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

func mustParseRFC3339(str string) time.Time {
	ret, err := time.Parse(time.RFC3339, str)
	if err != nil {
		panic(err)
	}
	return ret
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

func (mgr *SessionManager) handleSettings(w http.ResponseWriter, r *http.Request) error {
	isAlreadyConfigured := mgr.IsConfigured()

	if isAlreadyConfigured && !mgr.IsLoggedIn(r, time.Now()) {
		panic("double check failed.")
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
	case "POST":

		if uid := strings.TrimSpace(r.FormValue("author/name")); uid != "" {
			mgr.config.AuthorName = uid // any conditions? suitable for author: https://gist.github.com/tjdett/4617547#file-atom-rng-xml-L64
		}
		if tit := strings.TrimSpace(r.FormValue("title")); tit != "" {
			mgr.config.Title = tit
		}
		if mgr.config.Title == "" {
			mgr.config.Title = "Shared links on …"
		}

		pwd := strings.TrimSpace(r.FormValue("password"))
		if !mgr.IsConfigured() || len([]rune(pwd)) < 12 {
			renderPage(&mgr.config, http.StatusBadRequest, w)
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
			strURL := r.URL.String()
			idx := bytes.LastIndex([]byte(strURL), []byte(os.Getenv("SCRIPT_NAME")+os.Getenv("PATH_INFO")))
			if idx < 0 {
				panic("cannot happen.")
			}
			baseURL := strURL[:idx]
			feed := &Feed{
				Id:      baseURL,
				Title:   HumanText{Body: "Hello, Atom!"},
				Authors: []Person{Person{Name: "John Doe"}},
				Entries: []*Entry{
					&Entry{
						Id:         "e2",
						Title:      HumanText{Body: "Hello, Entry 2!"},
						Links:      []Link{Link{Rel: "via", Href: "http://www.loremipsum.de/"}},
						Categories: []Category{Category{Term: "🐳", Scheme: baseURL + "/" + uriPub + "/" + uriTags + "#"}, Category{Term: "self-hosted", Scheme: baseURL + "/" + uriPub + "/" + uriTags + "#"}},
						Content:    &HumanText{Type: "text", Body: "Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet."},
						Updated:    iso8601{mustParseRFC3339("1990-12-31T02:02:02+01:00")},
					},
					&Entry{
						Id:         "e1",
						Title:      HumanText{Body: "Hello, Entry 1!"},
						Links:      []Link{Link{Rel: "via", Href: "http://blog.fefe.de/"}},
						Categories: []Category{Category{Term: "🐳", Scheme: baseURL + "/" + uriPub + "/" + uriTags + "#"}},
						Content:    &HumanText{Body: ""},
						Updated:    iso8601{mustParseRFC3339("1990-12-31T01:01:01+01:00")},
					},
					&Entry{
						Id:         "e0",
						Title:      HumanText{Body: "Hello, Entry 0!"},
						Categories: []Category{Category{Term: "self-hosted", Scheme: baseURL + "/" + uriPub + "/" + uriTags + "#"}},
						Content:    &HumanText{Body: ""},
						Updated:    iso8601{mustParseRFC3339("1990-12-30T00:00:00+01:00")},
					},
				},
			}
			if err = feed.replaceFeeds(); err != nil {
				return err
			}
		}

		mgr.startSession(w, r, mgr.config.AuthorName)

		// all went well:
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Location", "../pub/posts/")
		w.WriteHeader(http.StatusSeeOther)
		io.WriteString(w, "let's go to "+"../pub/posts/"+"\n")
	case "GET":
		renderPage(&mgr.config, http.StatusOK, w)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}

	return nil
}

func renderPage(c *Config, code int, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/xml; charset=utf-8")
	w.WriteHeader(code)
	renderPageXml(c, w)
}

func renderPageXml(c *Config, w io.Writer) {
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

	encodeValueElement(enc, "title", c.Title)

	{
		author := xml.Name{Local: "author"}
		enc.EncodeToken(xml.StartElement{Name: author})
		encodeValueElement(enc, "name", c.AuthorName)
		enc.EncodeToken(xml.EndElement{Name: author})
	}

	enc.EncodeToken(xml.EndElement{Name: n})
	enc.EncodeToken(xml.CharData("\n"))
	enc.Flush()
}
