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
	"regexp"
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
<head><title>{{.title}}</title></head>
<body>
  <form method="post" name="loginform">
    <input type="text" name="login" />
    <input type="password" name="password" />
    <input type="submit" value="Login" />
    <input type="checkbox" name="longlastingsession" />
    <input type="hidden" name="token" value="{{.token}}" />
    <input type="hidden" name="returnurl" value="{{.returnurl}}" />
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

func (app *App) handleDoLogout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("ses_timeout_logout", fmt.Sprintf("%v", app.ses.Values["timeout"]))
	if err := app.stopSession(w, r); err != nil {
		http.Error(w, "Couldn't end session: "+err.Error(), http.StatusInternalServerError)
	} else {
		http.Redirect(w, r, path.Join(uriPub, uriPosts)+"/", http.StatusFound)
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

func (app *App) handleDoPost(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	switch r.Method {
	case http.MethodGet:
		// 'GET': send a form to the client
		// must be compatible with https://github.com/mro/Shaarli-API-Test/...
		// and https://github.com/mro/ShaarliOS/blob/master/ios/ShaarliOS/API/ShaarliCmd.m#L386

		if !app.IsLoggedIn(now) {
			http.Redirect(w, r, cgiName+"?do=login&returnurl="+url.QueryEscape(r.URL.String()), http.StatusFound)
			return
		}
		params := r.URL.Query()
		if 1 != len(params["post"]) {
			http.Error(w, "StatusBadRequest", http.StatusBadRequest)
			return
		}

		app.KeepAlive(w, r, now)
		feed, _ := FeedFromFileName(fileFeedStorage)
		post := sanitiseURLString(params["post"][0], app.cfg.UrlCleaner)

		_, ent := feed.findEntry(post)
		if nil == ent {
			// nothing found, so we need a new (dangling, unsaved) entry:
			if url := urlFromPostParam(post); url != nil {
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
				ent.Links = []Link{Link{Href: url.String()}}
			} else {
				ent = &Entry{}
				ent.Title = HumanText{Body: post}
			}
			ent.Updated = iso8601{now}
			if ent.Published.IsZero() {
				ent.Published = ent.Updated
			}
			// do not append to feed yet, keep dangling
		}

		if 1 == len(params["title"]) && "" != params["description"][0] {
			ent.Title = HumanText{Body: params["title"][0]}
		}
		if 1 == len(params["description"]) && "" != params["description"][0] {
			ent.Content = &HumanText{Body: params["description"][0]}
		}
		if 1 == len(params["source"]) {
			// data["lf_source"] = params["source"][0]
		}

		if tmpl, err := template.New("linkform").Parse(`<html xmlns="http://www.w3.org/1999/xhtml">
<head><title>{{.title}}</title></head>
<body>
  <ul id="taglist" style="display:none">{{ range $idx, $cat := .categories }}<li>#{{ $cat.Term }}</li>{{ end }}</ul>
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
    <input name="lf_image" type="hidden" value="{{.lf_image}}"/>
    <input name="lf_identifier" type="hidden" value="{{.lf_identifier}}"/>
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
			data["categories"] = feed.Categories
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
			squealFailure(r, now, "Unauthorised")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		app.KeepAlive(w, r, now)
		location := path.Join(uriPub, uriPosts) + "/"

		// https://github.com/sebsauvage/Shaarli/blob/master/index.php#L1479
		if "" != r.FormValue("save_edit") {
			if lf_linkdate, err := time.ParseInLocation(fmtTimeLfTime, strings.TrimSpace(r.FormValue("lf_linkdate")), app.tz); err != nil {
				squealFailure(r, now, "BadRequest: "+err.Error())
				http.Error(w, "Looks like a forged request: "+err.Error(), http.StatusBadRequest)
				return
			} else {
				token := r.FormValue("token")
				log.Println("todo: check token ", token)
				if returnurl, err := url.Parse(r.FormValue("returnurl")); err != nil {
					log.Println("Error parsing returnurl: ", err.Error())
					http.Error(w, "couldn't parse returnurl: "+err.Error(), http.StatusInternalServerError)
					return
				} else {
					log.Println("todo: use returnurl ", returnurl)

					feed, _ := FeedFromFileName(fileFeedStorage)

					lf_url := r.FormValue("lf_url")
					_, ent := feed.findEntry(lf_url)
					if nil == ent {
						ent = &Entry{
							Authors:   []Person{Person{Name: app.cfg.AuthorName}},
							Published: iso8601{lf_linkdate},
							Id:        smallDateHash(lf_linkdate),
						}
						if err := feed.Append(ent); err != nil {
							http.Error(w, "couldn't add entry: "+err.Error(), http.StatusInternalServerError)
							return
						}
					}
					ent.Updated = iso8601{now}
					ent.Title = HumanText{Body: strings.TrimSpace(r.FormValue("lf_title")), Type: "text"}
					url := mustParseURL(lf_url)
					if url.IsAbs() && "" != url.Host {
						ent.Links = []Link{Link{Href: lf_url}}
					} else {
						ent.Links = []Link{}
					}
					ent.Content = &HumanText{Body: strings.TrimSpace(r.FormValue("lf_description")), Type: "text"}
					if img := strings.TrimSpace(r.FormValue("lf_image")); "" != img {
						ent.MediaThumbnail = &MediaThumbnail{Url: img}
					}

					{
						tags := strings.Split(r.FormValue("lf_tags"), " ")
						a := make([]Category, 0, len(tags))
						for _, tg := range tags {
							if "" != tg {
								a = append(a, Category{Term: tg})
							}
						}
						ent.Categories = a // discard old categories and only use from POST.
						ent.Categories = ent.CategoriesMerged()
					}
					if err := ent.Validate(); err != nil {
						http.Error(w, "couldn't add entry: "+err.Error(), http.StatusInternalServerError)
						return
					}
					location = strings.Join([]string{location, ent.Id}, "?#")
					feed.XmlBase = xmlBaseFromRequestURL(r.URL, os.Getenv("SCRIPT_NAME")).String()
					feed.Id = feed.XmlBase
					feed.Save(fileFeedStorage)

					if err := feed.replaceFeeds(); err != nil {
						log.Println("couldn't write feeds: ", err.Error())
						http.Error(w, "couldn't write feeds: "+err.Error(), http.StatusInternalServerError)
						return
					}
				}
			}
		} else if "" != r.FormValue("cancel_edit") {

		} else if "" != r.FormValue("delete_edit") {
			token := r.FormValue("token")
			log.Println("todo: check token ", token)
			feed, _ := FeedFromFileName(fileFeedStorage)
			id := strings.TrimSpace(r.FormValue("lf_url"))
			if entry := feed.deleteEntry(id); nil != entry {
				feed.XmlBase = xmlBaseFromRequestURL(r.URL, os.Getenv("SCRIPT_NAME")).String()
				feed.Id = feed.XmlBase
				feed.Save(fileFeedStorage)

				if err := feed.replaceFeeds(); err != nil {
					log.Println("couldn't write feeds: ", err.Error())
					http.Error(w, "couldn't write feeds: "+err.Error(), http.StatusInternalServerError)
					return
				}
			} else {
				squealFailure(r, now, "Not Found")
				log.Println("entry not found: ", id)
				http.Error(w, "Not Found", http.StatusNotFound)
				return
			}
		} else {
			squealFailure(r, now, "BadRequest")
			http.Error(w, "BadRequest", http.StatusBadRequest)
			return
		}
		if "bookmarklet" == r.FormValue("source") {
			// io.WriteString(w, "<script>self.close(); // close bookmarklet popup</script>")
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/javascript")
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
func (entry Entry) api0LinkFormMap() map[string]interface{} {
	data := map[string]interface{}{
		"lf_identifier": entry.Id,
		"lf_linkdate":   entry.Published.Format(fmtTimeLfTime),
		"lf_title":      entry.Title.Body,
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
	if "" == data["lf_url"] && "" != entry.Id {
		// todo: also if it's not a note
		data["lf_url"] = entry.Id
	}
	if nil != entry.Content {
		data["lf_description"] = entry.Content.Body
	}
	if nil != entry.MediaThumbnail && len(entry.MediaThumbnail.Url) > 0 {
		data["lf_image"] = entry.MediaThumbnail.Url
	}
	return data
}
