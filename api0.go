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
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"hash/crc32"
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

const fmtTimeLfTime = "20060102_150405"

func parseLinkUrl(raw string) *url.URL {
	if ret, err := url.Parse(raw); err == nil {
		if !ret.IsAbs() {
			if ret, err = url.Parse("http://" + raw); err != nil {
				return nil
			}
		}
		return ret
	} else {
		return nil
	}
}

/* Returns the small hash of a string, using RFC 4648 base64url format
   eg. smallHash('20111006_131924') --> yZH23w
   Small hashes:
     - are unique (well, as unique as crc32, at last)
     - are always 6 characters long.
     - only use the following characters: a-z A-Z 0-9 - _ @
     - are NOT cryptographically secure (they CAN be forged)
   In Shaarli, they are used as a tinyurl-like link to individual entries.

   https://github.com/sebsauvage/Shaarli/blob/master/index.php#L228
*/
func smallHash(text string) string {
	// ret:= rtrim(base64_encode(hash('crc32',$text,true)),'=');
	crc := crc32.ChecksumIEEE([]byte(text))
	bs := make([]byte, 4) // https://stackoverflow.com/a/16889357
	binary.LittleEndian.PutUint32(bs, crc)
	return base64.RawURLEncoding.EncodeToString(bs)
}

func smallDateHash(tt time.Time) string {
	bs := make([]byte, 4) // https://stackoverflow.com/a/16889357
	// unix time in seconds as uint32
	binary.LittleEndian.PutUint32(bs, uint32(tt.Unix()&0xFFFFFFFF))
	return base64.RawURLEncoding.EncodeToString(bs)
}

func (app *App) handleDoLogin(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	switch r.Method {
	// and https://github.com/mro/ShaarliOS/blob/master/ios/ShaarliOS/API/ShaarliCmd.m#L386
	case http.MethodGet:
		returnurl := r.Referer()
		if ru := r.URL.Query()["returnurl"]; ru != nil && 1 == len(ru) && "" != ru[0] {
			returnurl = ru[0]
		}
		if tmpl, err := template.New("login").Parse(`<html xmlns="http://www.w3.org/1999/xhtml">
<head><title>{{index . "title"}}</title></head>
<body>
  <form method="post" name="loginform" id="loginform">
    <input type="text" name="login" />
    <input type="password" name="password" />
    <input type="submit" value="Login" />
    <input type="checkbox" name="longlastingsession" id="longlastingsession" />
    <input type="hidden" name="token" value="{{index . "token"}}" />
    <input type="hidden" name="returnurl" value="{{index . "returnurl"}}" />
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
			if err := tmpl.Execute(w, map[string]string{
				"title":     app.cfg.Title,
				"token":     "ff13e7eaf9541ca2ba30fd44e864c3ff014d2bc9",
				"returnurl": returnurl,
			}); err != nil {
				http.Error(w, "Coudln't send login form: "+err.Error(), http.StatusInternalServerError)
			}
		}
	case http.MethodPost:
		// todo: verify token
		uid := strings.TrimSpace(r.FormValue("login"))
		pwd := strings.TrimSpace(r.FormValue("password"))
		returnurl := strings.TrimSpace(r.FormValue("returnurl"))
		// compute anyway (a bit more time constantness)
		err := bcrypt.CompareHashAndPassword([]byte(app.cfg.PwdBcrypt), []byte(pwd))
		if uid != app.cfg.AuthorName || err == bcrypt.ErrMismatchedHashAndPassword {
			squealFailure(r, now)
			http.Error(w, "<script>alert(\"Wrong login/password.\");document.location='?do=login&returnurl='"+url.QueryEscape(returnurl)+"';</script>", http.StatusUnauthorized)
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
		squealFailure(r, now)
		http.Error(w, "MethodNotAllowed", http.StatusMethodNotAllowed)
	}
	//     NSString *xpath = [NSString stringWithFormat:@"/html/body//form[@name='%1$@']//input[(@type='text' or @type='password' or @type='hidden' or @type='checkbox') and @name] | /html/body//form[@name='%1$@']//textarea[@name]

	// 'POST' validate, respond error (and squeal) or set session and redirect
}

func (app *App) handleDoLogout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("ses_timeout_logout", fmt.Sprintf("%v", app.ses.Values["timeout"]))
	if err := app.stopSession(w, r); err != nil {
		http.Error(w, "Couldn't end session: "+err.Error(), http.StatusInternalServerError)
	} else {
		http.Redirect(w, r, path.Join(uriPub, uriPosts)+"/", http.StatusFound)
	}
}

func (app *App) handleDoPost(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	switch r.Method {
	case http.MethodGet:
		if !app.IsLoggedIn(now) {
			http.Redirect(w, r, cgiName+"?do=login&returnurl="+url.QueryEscape(r.URL.String()), http.StatusFound)
			return
		}
		app.KeepAlive(w, r, now)

		// 'GET': send a form to the client
		// must be compatible to https://github.com/mro/Shaarli-API-Test/...
		// and https://github.com/mro/ShaarliOS/blob/master/ios/ShaarliOS/API/ShaarliCmd.m#L386

		// 1. pull post= title= and source= and from GET
		// 2. url or note?
		// 3. if url:
		//   a. already there? findEntryForLink
		//   b. map post, title and source to a new (sparse?) atom entry - feasible in JavaScript as well?
		//   c. turn atom entry to form map

		params := r.URL.Query()
		feed, _ := FeedFromFileName(fileFeedStorage)

		var ent *Entry = nil
		if 1 == len(params["post"]) {
			lf_url := parseLinkUrl(params["post"][0])
			if nil == lf_url || !lf_url.IsAbs() || "" == lf_url.Hostname() {
				lf_url = nil
			}
			ent = feed.findOrCreateEntryForURL(lf_url, now, false)
			if lf_url == nil {
				ent.Title = HumanText{Body: params["post"][0]}
			}
		}
		if 1 == len(params["title"]) {
			ent.Title = HumanText{Body: params["title"][0]}
		}
		if 1 == len(params["description"]) {
			ent.Content = &HumanText{Body: params["description"][0]}
		}
		if 1 == len(params["source"]) {
			// data["lf_source"] = params["source"][0]
		}

		if tmpl, err := template.New("linkform").Parse(`<html xmlns="http://www.w3.org/1999/xhtml">
<head><title>{{.title}}</title></head>
<body>
  <form method="post" name="linkform">
    <input name="lf_linkdate" type="hidden" value="{{.lf_linkdate}}"/>
    <input name="lf_url" type="text" value="{{.lf_url}}"/>
    <input name="lf_title" type="text" value="{{.lf_title}}"/>
    <textarea name="lf_description" rows="4" cols="25">{{.lf_description}}</textarea>
    <input name="lf_tags" type="text" data-multiple="data-multiple" value="{{.lf_tags}}"/>
    <input name="lf_private" type="checkbox" value="{{.lf_private}}"/>
    <input name="save_edit" type="submit" value="Save"/>
    <input name="cancel_edit" type="submit" value="Cancel"/>
    <input name="token" type="hidden" value="{{.token}}"/>
    <input name="returnurl" type="hidden" value="{{.returnurl}}"/>
  </form>
</body>
</html>
`); err == nil {
			w.Header().Set("Content-Type", "text/xml; charset=utf-8")
			io.WriteString(w, `<?xml version="1.0" encoding="UTF-8"?>
<?xml-stylesheet type='text/xsl' href='./assets/default/de/do-post.xslt'?>
<!--
  must be compatible with https://github.com/mro/Shaarli-API-test/blob/master/tests/test-post.sh
  https://github.com/mro/ShaarliOS/blob/master/ios/ShaarliOS/API/ShaarliCmd.m#L386
-->
`)
			data := ent.api0LinkFormMap()
			data["title"] = app.cfg.Title
			bTok := make([]byte, 20) // keep in local session or encrypted cookie
			io.ReadFull(rand.Reader, bTok)
			data["token"] = hex.EncodeToString(bTok)
			data["returnurl"] = ""

			if err := tmpl.Execute(w, data); err != nil {
				http.Error(w, "Coudln't send linkform: "+err.Error(), http.StatusInternalServerError)
			}
		}
	case http.MethodPost:
		// 'POST' validate, respond error (and squeal) or post and redirect
		if !app.IsLoggedIn(now) {
			squealFailure(r, now)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		app.KeepAlive(w, r, now)
		location := path.Join(uriPub, uriPosts)

		// https://github.com/sebsauvage/Shaarli/blob/master/index.php#L1479
		log.Println("save_edit: '" + r.FormValue("save_edit") + "'")
		log.Println("cancel_edit: '" + r.FormValue("cancel_edit") + "'")
		if r.FormValue("save_edit") != "" {
			if lf_linkdate, err := time.ParseInLocation(fmtTimeLfTime, strings.TrimSpace(r.FormValue("lf_linkdate")), app.tz); err != nil {
				squealFailure(r, now)
				http.Error(w, "Looks like a forged request: "+err.Error(), http.StatusBadRequest)
				return
			} else {
				if lf_url, err := url.Parse(strings.TrimSpace(r.FormValue("lf_url"))); err != nil {
					squealFailure(r, now)
					http.Error(w, "Looks like a forged request: "+err.Error(), http.StatusBadRequest)
					return
				} else {
					lf_title := strings.TrimSpace(r.FormValue("lf_title"))
					lf_description := strings.TrimSpace(r.FormValue("lf_description"))
					lf_tags := strings.TrimSpace(r.FormValue("lf_tags"))
					token := strings.TrimSpace(r.FormValue("token"))
					if returnurl, err := url.Parse(strings.TrimSpace(r.FormValue("returnurl"))); err != nil {
						log.Printf("Error parsing returnurl: %s", err.Error())
					} else {
						if nil == lf_url || !lf_url.IsAbs() || "" == lf_url.Hostname() {
							lf_url = nil
						}
						log.Println("err", err)
						log.Println("lf_linkdate", lf_linkdate)
						log.Println("lf_url", lf_url)
						log.Println("lf_title", lf_title)
						log.Println("lf_description", lf_description)
						log.Println("lf_tags", lf_tags)
						log.Println("token", token)
						log.Println("returnurl", returnurl)

						// todo: check token.

						feed, _ := FeedFromFileName(fileFeedStorage)
						feed.XmlBase = xmlBaseFromRequestURL(r.URL, os.Getenv("SCRIPT_NAME")).String()
						feed.Id = feed.XmlBase
						ent := feed.findOrCreateEntryForURL(lf_url, now, true)
						ent.Authors = []Person{Person{Name: app.cfg.AuthorName}}
						ent.Published = iso8601{lf_linkdate}
						if "" == ent.Id {
							ent.Id = smallDateHash(ent.Published.Time)
						}
						if nil == lf_url {
							ent.Links = []Link{}
						} else {
							ent.Links = []Link{Link{Href: lf_url.String()}}
						}
						ent.Title = HumanText{Body: lf_title}
						ent.Content = &HumanText{Body: lf_description}
						{
							tags := strings.Split(lf_tags, " ")
							ent.Categories = make([]Category, 0, len(tags)) // discard old categories and only use from POST.
							for _, tg := range tags {
								ent.Categories = append(ent.Categories, Category{Term: tg})
							}
							ent.Categories = ent.CategoriesMerged()
						}
						location = strings.Join([]string{location, ent.Id}, "/?#")

						feed.Save(fileFeedStorage)

						if err := feed.replaceFeeds(); err != nil {
							log.Println("couldn't write feeds: " + err.Error())
							http.Error(w, "couldn't write feeds: "+err.Error(), http.StatusInternalServerError)
							return
						}
					}
				}
			}
		} else if r.FormValue("cancel_edit") != "" {

		} else {
			squealFailure(r, now)
			http.Error(w, "StatusBadRequest", http.StatusBadRequest)
			return
		}
		// close bookmarklet popup in case!
		http.Redirect(w, r, location, http.StatusFound)
		return
	default:
		squealFailure(r, now)
		http.Error(w, "MethodNotAllowed", http.StatusMethodNotAllowed)
		return
	}
}

func (app *App) handleDoCheckLoginAfterTheFact(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	switch r.Method {
	case http.MethodGet:
		if !app.IsLoggedIn(now) {
			http.Redirect(w, r, cgiName+"?do=login&returnurl="+url.QueryEscape(r.URL.String()), http.StatusFound)
			return
		}
		app.KeepAlive(w, r, now)

		if tmpl, err := template.New("changepasswordform").Parse(`<html xmlns="http://www.w3.org/1999/xhtml">
<head><title>{{.title}}</title></head>
<body>
  <a href="?do=logout">Logout</a>
  <form method="post" name="changepasswordform">
    <input type="password" name="oldpassword" />
    <input type="password" name="setpassword" />
    <input type="hidden" name="token" value="{{.token}}" />
    <input type="submit" name="Save" value="Save password" />
  </form>
</body>
</html>
`); err == nil {
			w.Header().Set("Content-Type", "text/xml; charset=utf-8")
			io.WriteString(w, `<?xml version="1.0" encoding="UTF-8"?>
<?xml-stylesheet type='text/xsl' href='./assets/default/de/do-changepassword.xslt'?>
<!--
  must be compatible with https://github.com/mro/Shaarli-API-test/blob/master/tests/test-post.sh
  https://github.com/mro/ShaarliOS/blob/master/ios/ShaarliOS/API/ShaarliCmd.m#L386
-->
`)
			data := make(map[string]string)
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

// Aggregate all tags from #title, #description and <category and remove the first two groups from the set.
func (entry Entry) api0LinkFormMap() map[string]string {
	data := map[string]string{
		"lf_linkdate": entry.Published.Format(fmtTimeLfTime),
		"lf_title":    entry.Title.Body,
	}
	{
		// 1. get all atom categories
		set := make(map[string]struct{}, len(entry.Categories))
		for _, c := range entry.Categories {
			set[c.Term] = struct{}{}
		}
		// 2. minus #tags from title
		for _, tag := range tagsFromString(entry.Title.Body) {
			delete(set, tag)
		}
		// 2. minus #tags from content
		if entry.Content != nil {
			for _, tag := range tagsFromString(entry.Content.Body) {
				delete(set, tag)
			}
		}
		// turn map keys into sorted array
		tags := make([]string, 0, len(set))
		for key, _ := range set {
			tags = append(tags, key)
		}
		sort.Slice(tags, func(i, j int) bool { return strings.Compare(tags[i], tags[j]) < 0 })
		data["lf_tags"] = strings.Join(tags, " ")
	}
	for _, li := range entry.Links {
		if "" == li.Rel {
			data["lf_url"] = li.Href
			break
		}
	}
	if entry.Content != nil {
		data["lf_description"] = entry.Content.Body
	}
	return data
}
