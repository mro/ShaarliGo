//
// Copyright (C) 2017-2019 Marcus Rohrmoser, http://purl.mro.name/ShaarliGo
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
	"crypto/rand"
	"encoding/hex"
	"encoding/xml"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
)

const fmtTimeLfTime = "20060102_150405"

func parseLinkUrl(raw string) *url.URL {
	if ret, err := url.Parse(raw); err == nil {
		if !ret.IsAbs() {
			if ret, err = url.Parse("http://" + raw); err != nil {
				return nil
			}
		}
		return ret
	}
	return nil
}

func (app *Server) handleDoLogin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()
		switch r.Method {
		// and https://code.mro.name/mro/ShaarliOS/src/1d124e012933d1209d64071a90237dc5ec6372fc/ios/ShaarliOS/API/ShaarliCmd.m#L386
		case http.MethodGet:
			returnurl := r.Referer()
			if ru := r.URL.Query()["returnurl"]; ru != nil && 1 == len(ru) && "" != ru[0] {
				returnurl = ru[0]
			}

			byt, _ := tplLoginHtmlBytes()
			if tmpl, err := template.New("login").Parse(string(byt)); err == nil {
				w.Header().Set("Content-Type", "text/xml; charset=utf-8")
				io.WriteString(w, xml.Header)
				io.WriteString(w, `<?xml-stylesheet type='text/xsl' href='./assets/`+app.cfg.Skin+`/do-login.xslt'?>
<!--
  must be compatible with https://code.mro.name/mro/Shaarli-API-test/src/master/tests/test-post.sh
  https://code.mro.name/mro/ShaarliOS/src/1d124e012933d1209d64071a90237dc5ec6372fc/ios/ShaarliOS/API/ShaarliCmd.m#L386
-->
`)
				if err := tmpl.Execute(w, map[string]string{
					"skin":      app.cfg.Skin,
					"title":     app.cfg.Title,
					"token":     "ff13e7eaf9541ca2ba30fd44e864c3ff014d2bc9",
					"returnurl": returnurl,
				}); err != nil {
					http.Error(w, "Couldn't send login form: "+err.Error(), http.StatusInternalServerError)
				}
			}
		case http.MethodPost:
			val := func(key string) string { return strings.TrimSpace(r.FormValue(key)) }
			// todo: verify token
			uid := val("login")
			pwd := val("password")
			returnurl := val("returnurl")
			// compute anyway (a bit more time constantness)
			err := bcrypt.CompareHashAndPassword([]byte(app.cfg.PwdBcrypt), []byte(pwd))
			if uid != app.cfg.Uid || err == bcrypt.ErrMismatchedHashAndPassword {
				squealFailure(r, now, "Unauthorised.")
				// http.Error(w, "<script>alert(\"Wrong login/password.\");document.location='?do=login&returnurl='"+url.QueryEscape(returnurl)+"';</script>", http.StatusUnauthorized)
				w.WriteHeader(http.StatusUnauthorized)
				w.Header().Set("Content-Type", "application/javascript")
				io.WriteString(w, "<script>alert(\"Wrong login/password.\");document.location='?do=login&returnurl='"+url.QueryEscape(returnurl)+"';</script>")
				return
			}
			if err == nil {
				err = app.startSession(w, r, now)
			}
			if err == nil {
				if "" == returnurl { // TODO restrict to local urls within app scope
					returnurl = path.Join(uriPub, uriPosts) + "/"
				}
				http.Redirect(w, r, returnurl, http.StatusFound)
				return
			}
			http.Error(w, "Fishy post: "+err.Error(), http.StatusInternalServerError)
		default:
			squealFailure(r, now, "MethodNotAllowed "+r.Method)
			http.Error(w, "MethodNotAllowed", http.StatusMethodNotAllowed)
		}
		//     NSString *xpath = [NSString stringWithFormat:@"/html/body//form[@name='%1$@']//input[(@type='text' or @type='password' or @type='hidden' or @type='checkbox') and @name] | /html/body//form[@name='%1$@']//textarea[@name]

		// 'POST' validate, respond error (and squeal) or set session and redirect
	}
}

func (app *Server) handleDoLogout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := app.stopSession(w, r); err != nil {
			http.Error(w, "Couldn't end session: "+err.Error(), http.StatusInternalServerError)
		} else {
			http.Redirect(w, r, path.Join(uriPub, uriPosts)+"/", http.StatusFound)
		}
	}
}

func sanitiseURLString(raw string, lst []RegexpReplaceAllString) string {
	for idx, row := range lst {
		if rex, err := regexp.Compile(row.Regexp); err != nil {
			log.Printf("Invalid regular expression #%d '%s': %s", idx, row.Regexp, err)
		} else {
			raw = rex.ReplaceAllString(raw, row.ReplaceAllString)
		}
	}
	return raw
}

func urlFromPostParam(post string) *url.URL {
	if url, err := url.Parse(post); err == nil && url != nil && url.IsAbs() && "" != url.Hostname() {
		return url
	} else {
		if nil != url && !url.IsAbs() {
			if !strings.ContainsRune(post, '.') {
				return nil
			}
			post = strings.Join([]string{"http://", post}, "")
			if url, err := url.Parse(post); err == nil && url != nil && url.IsAbs() && "" != url.Hostname() {
				return url
			}
		}
		return nil
	}
}

func termsVisitor(entries ...*Entry) func(func(string)) {
	return func(callback func(string)) {
		for _, ee := range entries {
			for _, ca := range ee.Categories {
				callback(ca.Term)
			}
		}
	}
}

/* Store identifier of edited entry in cookie.
 */
func (app *Server) handleDoPost(posse func(Entry)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()
		switch r.Method {
		case http.MethodGet:
			// 'GET': send a form to the client
			// must be compatible with https://code.mro.name/mro/Shaarli-API-Test/...
			// and https://code.mro.name/mro/ShaarliOS/src/1d124e012933d1209d64071a90237dc5ec6372fc/ios/ShaarliOS/API/ShaarliCmd.m#L386

			if !app.IsLoggedIn(now) {
				http.Redirect(w, r, cgiName+"?do=login&returnurl="+url.QueryEscape(r.URL.String()), http.StatusFound)
				return
			}

			params := r.URL.Query()
			if 1 != len(params["post"]) {
				http.Error(w, "StatusBadRequest", http.StatusBadRequest)
				return
			}

			feed, _ := LoadFeed()
			post := sanitiseURLString(params["post"][0], app.cfg.UrlCleaner)

			feed.XmlBase = Iri(app.url.String())
			_, ent := feed.findEntryByIdSelfOrUrl(post)
			if nil == ent {
				// nothing found, so we need a new (dangling, unsaved) entry:
				if url := urlFromPostParam(post); url == nil {
					// post parameter doesn't look like an url, so we treat it as a note.
					ent = &Entry{}
					ent.Title = HumanText{Body: post}
				} else {
					// post parameter looks like an url, so we try to GET it
					{
						ee, err := entryFromURL(url, time.Second*3/2)
						if nil != err {
							ee.Title.Body = err.Error()
						}
						ent = &ee
					}
					if nil == ent.Content || "" == ent.Content.Body {
						ent.Content = ent.Summary
					}
					ent.Links = []Link{{Href: url.String()}}
				}
				ent.Updated = iso8601(now)
				const SetPublishedToNowInitially = true
				if SetPublishedToNowInitially || ent.Published.IsZero() {
					ent.Published = ent.Updated
				}
				// do not append to feed yet, keep dangling
			} else {
				log.Printf("storing Id in cookie: %v", ent.Id)
				app.ses.Values["identifier"] = ent.Id
			}
			app.KeepAlive(w, r, now)

			if 1 == len(params["title"]) && "" != params["title"][0] {
				ent.Title = HumanText{Body: params["title"][0]}
			}
			if 1 == len(params["description"]) && "" != params["description"][0] {
				ent.Content = &HumanText{Body: params["description"][0]}
			}
			if 1 == len(params["source"]) {
				// data["lf_source"] = params["source"][0]
			}

			byt, _ := tplLinkformHtmlBytes() // todo: err => 500
			if tmpl, err := template.New("linkform").Parse(string(byt)); err == nil {
				w.Header().Set("Content-Type", "text/xml; charset=utf-8")
				io.WriteString(w, xml.Header)
				io.WriteString(w, `<?xml-stylesheet type='text/xsl' href='./assets/`+app.cfg.Skin+`/do-post.xslt'?>
<!--
  must be compatible with https://code.mro.name/mro/Shaarli-API-test/src/master/tests/test-post.sh
  https://code.mro.name/mro/ShaarliOS/src/1d124e012933d1209d64071a90237dc5ec6372fc/ios/ShaarliOS/API/ShaarliCmd.m#L386
-->
`)
				data := ent.api0LinkFormMap()
				data["skin"] = app.cfg.Skin
				data["title"] = feed.Title.Body
				data["categories"] = feed.Categories
				bTok := make([]byte, 20) // keep in local session or encrypted cookie
				io.ReadFull(rand.Reader, bTok)
				data["token"] = hex.EncodeToString(bTok)
				data["returnurl"] = ""
				data["xml_base"] = feed.XmlBase

				if err := tmpl.Execute(w, data); err != nil {
					http.Error(w, "Coudln't send linkform: "+err.Error(), http.StatusInternalServerError)
				}
			}
		case http.MethodPost:
			val := func(key string) string { return strings.TrimSpace(r.FormValue(key)) }
			// 'POST' validate, respond error (and squeal) or post and redirect
			if !app.IsLoggedIn(now) {
				squealFailure(r, now, "Unauthorised")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			identifier, ok := app.ses.Values["identifier"].(Id)
			if ok {
				delete(app.ses.Values, "identifier")
			}

			log.Printf("pulled Id from cookie: %v", identifier)
			app.KeepAlive(w, r, now)
			location := path.Join(uriPub, uriPosts) + "/"

			// https://github.com/sebsauvage/Shaarli/blob/master/index.php#L1479
			if "" != val("save_edit") {
				if lf_linkdate, err := time.ParseInLocation(fmtTimeLfTime, val("lf_linkdate"), app.tz); err != nil {
					squealFailure(r, now, "BadRequest: "+err.Error())
					http.Error(w, "Looks like a forged request: "+err.Error(), http.StatusBadRequest)
					return
				} else {
					log.Println("todo: check token ", val("token"))
					if returnurl, err := url.Parse(val("returnurl")); err != nil {
						log.Println("Error parsing returnurl: ", err.Error())
						http.Error(w, "couldn't parse returnurl: "+err.Error(), http.StatusInternalServerError)
						return
					} else {
						log.Println("todo: use returnurl ", returnurl)

						// make persistent
						feed, _ := LoadFeed()
						feed.XmlBase = Iri(app.url.String())

						lf_url := val("lf_url")
						_, ent := feed.findEntryById(identifier)
						if nil == ent {
							ent = feed.newEntry(lf_linkdate)
							if _, err := feed.Append(ent); err != nil {
								http.Error(w, "couldn't add entry: "+err.Error(), http.StatusInternalServerError)
								return
							}
						}
						ent0 := *ent

						// prepare redirect
						location = strings.Join([]string{location, string(ent.Id)}, "?#")

						// human := func(key string) HumanText { return HumanText{Body: val(key), Type: "text"} }
						// humanP := func(key string) *HumanText { t := human(key); return &t }

						ent.Updated = iso8601(now)

						url := mustParseURL(lf_url)
						if url.IsAbs() && "" != url.Host {
							ent.Links = []Link{{Href: lf_url}}
						} else {
							ent.Links = []Link{}
						}

						ds, ex, tags := tagsNormalise(
							val("lf_title"),
							val("lf_description"),
							tagsVisitor(strings.Split(val("lf_tags"), " ")...),
							termsVisitor(feed.Entries...),
						)
						ent.Title = HumanText{Body: ds, Type: "text"}
						ent.Content = &HumanText{Body: ex, Type: "text"}
						{
							a := make([]Category, 0, len(tags))
							for _, tag := range tags {
								a = append(a, Category{Term: tag})
							}
							ent.Categories = a
						}

						if img := val("lf_image"); "" != img {
							ent.MediaThumbnail = &MediaThumbnail{Url: Iri(img)}
						}

						if err := ent.Validate(); err != nil {
							http.Error(w, "couldn't add entry: "+err.Error(), http.StatusInternalServerError)
							return
						}

						if err := app.SaveFeed(feed); err != nil {
							http.Error(w, "couldn't store feed data: "+err.Error(), http.StatusInternalServerError)
							return
						}
						// todo: waiting group? fire and forget go function?
						// we should, however, lock re-entrancy
						posse(*ent)
						// refresh feeds
						if err := app.PublishFeedsForModifiedEntries(feed, []*Entry{ent, &ent0}); err != nil {
							log.Println("couldn't write feeds: ", err.Error())
							http.Error(w, "couldn't write feeds: "+err.Error(), http.StatusInternalServerError)
							return
						}
					}
				}
			} else if "" != val("cancel_edit") {

			} else if "" != val("delete_edit") {
				token := val("token")
				log.Println("todo: check token ", token)
				// make persistent
				feed, _ := LoadFeed()
				if ent := feed.deleteEntryById(identifier); nil != ent {
					if err := app.SaveFeed(feed); err != nil {
						http.Error(w, "couldn't store feed data: "+err.Error(), http.StatusInternalServerError)
						return
					}
					// todo: POSSE
					// refresh feeds
					feed.XmlBase = Iri(app.url.String())
					if err := app.PublishFeedsForModifiedEntries(feed, []*Entry{ent}); err != nil {
						log.Println("couldn't write feeds: ", err.Error())
						http.Error(w, "couldn't write feeds: "+err.Error(), http.StatusInternalServerError)
						return
					}
				} else {
					squealFailure(r, now, "Not Found")
					log.Println("entry not found: ", identifier)
					http.Error(w, "Not Found", http.StatusNotFound)
					return
				}
			} else {
				squealFailure(r, now, "BadRequest")
				http.Error(w, "BadRequest", http.StatusBadRequest)
				return
			}
			if "bookmarklet" == val("source") {
				w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Type", "application/javascript")
				// CSP script-src 'sha256-hGqewLn4csF93PEX/0TCk2jdnAytXBZFxFBzKt7wcgo='
				// echo -n "self.close(); // close bookmarklet popup" | openssl dgst -sha256 -binary | base64
				io.WriteString(w, "<script>self.close(); // close bookmarklet popup</script>")
			} else {
				http.Redirect(w, r, location, http.StatusFound)
			}
			return
		default:
			squealFailure(r, now, "MethodNotAllowed: "+r.Method)
			http.Error(w, "MethodNotAllowed", http.StatusMethodNotAllowed)
			return
		}
	}
}

func (app *Server) handleDoCheckLoginAfterTheFact() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()
		switch r.Method {
		case http.MethodGet:
			if !app.IsLoggedIn(now) {
				http.Redirect(w, r, cgiName+"?do=login&returnurl="+url.QueryEscape(r.URL.String()), http.StatusFound)
				return
			}
			app.KeepAlive(w, r, now)

			byt, _ := tplChangepasswordformHtmlBytes()
			if tmpl, err := template.New("changepasswordform").Parse(string(byt)); err == nil {
				w.Header().Set("Content-Type", "text/xml; charset=utf-8")
				io.WriteString(w, xml.Header)
				io.WriteString(w, `<?xml-stylesheet type='text/xsl' href='./assets/`+app.cfg.Skin+`/do-changepassword.xslt'?>
<!--
  must be compatible with https://code.mro.name/mro/Shaarli-API-test/src/master/tests/test-post.sh
  https://code.mro.name/mro/ShaarliOS/src/1d124e012933d1209d64071a90237dc5ec6372fc/ios/ShaarliOS/API/ShaarliCmd.m#L386
-->
`)
				data := make(map[string]string)
				data["skin"] = app.cfg.Skin
				data["title"] = app.cfg.Title
				bTok := make([]byte, 20) // keep in local session or encrypted cookie
				io.ReadFull(rand.Reader, bTok)
				data["token"] = hex.EncodeToString(bTok)
				data["returnurl"] = ""

				if err := tmpl.Execute(w, data); err != nil {
					http.Error(w, "Coudln't send changepasswordform: "+err.Error(), http.StatusInternalServerError)
				}
			}
		}
	}
}

// Aggregate all tags from #title, #description and <category and remove the first two groups from the set.
func (entry Entry) api0LinkFormMap() map[string]interface{} {
	body := func(t *HumanText) string {
		if t == nil {
			return ""
		}
		return t.Body
	}

	data := map[string]interface{}{
		"lf_linkdate": entry.Published.Format(fmtTimeLfTime),
	}
	{
		ti, de, ta := tagsNormalise(
			body(&entry.Title),
			body(entry.Content),
			termsVisitor(&entry),
			termsVisitor(&entry), // rather all the feed's tags, but as we don't have them it's ok, too.
		)
		data["lf_title"] = ti
		data["lf_description"] = de
		data["lf_tags"] = strings.Join(ta, " ")
	}
	for _, li := range entry.Links {
		if "" == li.Rel {
			data["lf_url"] = li.Href
			break
		}
	}
	if "" == data["lf_url"] && "" != entry.Id {
		// todo: also if it's not a note
		data["lf_url"] = entry.Id
	}
	if nil != entry.MediaThumbnail && len(entry.MediaThumbnail.Url) > 0 {
		data["lf_image"] = entry.MediaThumbnail.Url
	}

	for key, value := range data {
		if s, ok := value.(string); ok && !utf8.ValidString(s) {
			data[key] = "Invalid UTF8"
		}
	}
	return data
}

func (feed *Feed) findEntryByIdSelfOrUrl(id_self_or_link string) (int, *Entry) {
	defer un(trace(strings.Join([]string{"Feed.findEntryByIdSelfOrUrl('", id_self_or_link, "')"}, "")))
	if "" != id_self_or_link {
		if parts := strings.SplitN(id_self_or_link, "/", 4); 4 == len(parts) && "" == parts[3] && uriPub == parts[0] && uriPosts == parts[1] {
			// looks like an internal id, so treat it as such.
			id_self_or_link = parts[2]
		}

		doesMatch := func(entry *Entry) bool {
			if id_self_or_link == string(entry.Id) {
				return true
			}
			for _, l := range entry.Links {
				if ("" == l.Rel || "self" == l.Rel) && (id_self_or_link == l.Href /* todo: url equal */) {
					return true
				}
			}
			return false
		}

		return feed.findEntry(doesMatch)
	}
	return feed.findEntry(nil)
}
