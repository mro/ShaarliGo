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

// Files & Directories
//
// .htaccess
// shaarligo.cgi
// app/.htaccess
// app/config.yaml
// app/posts.gob.gz
// app/posts.xml.gz
// app/var/bans.yaml
// app/var/error.log
// app/var/stage/
// app/var/old/
// assets/default/de/
// pub/posts/
//
package main

import (
	"encoding/base64"
	"encoding/gob"
	"encoding/xml"
	"io"
	"log"
	"net/http"
	"net/http/cgi"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/sessions"
)

const toSession = 30 * time.Minute

const myselfNamespace = "http://purl.mro.name/ShaarliGo/"

var GitSHA1 = "Please set -ldflags \"-X main.GitSHA1=$(git rev-parse --short HEAD)\"" // https://medium.com/@joshroppo/setting-go-1-5-variables-at-compile-time-for-versioning-5b30a965d33e
var fileFeedStorage string

func init() {
	fileFeedStorage = filepath.Join(dirApp, "var", uriPub+".atom")
	gob.Register(Id("")) // http://www.gorillatoolkit.org/pkg/sessions
}

// even cooler: https://stackoverflow.com/a/8363629
//
// inspired by // https://coderwall.com/p/cp5fya/measuring-execution-time-in-go
func trace(name string) (string, time.Time) { return name, time.Now() }
func un(name string, start time.Time)       { log.Printf("%s took %s", name, time.Since(start)) }

// evtl. as a server, too: http://www.dav-muz.net/blog/2013/09/how-to-use-go-and-fastcgi/
func main() {
	if false {
		// lighttpd doesn't seem to like more than one (per-vhost) server.breakagelog
		log.SetOutput(os.Stderr)
	} else { // log to custom logfile rather than stderr (may not be reachable on shared hosting)
		dst := filepath.Join(dirApp, "var", "log", "error.log")
		if err := os.MkdirAll(filepath.Dir(dst), 0770); err != nil {
			log.Fatal("Couldn't create app/var/log dir: " + err.Error())
			return
		}
		if fileLog, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0660); err != nil {
			log.Fatal("Couldn't open logfile: " + err.Error())
			return
		} else {
			defer fileLog.Close()
			log.SetOutput(fileLog)
		}
	}

	// - check non-write perm of program?
	// - check non-http read perm on ./app
	if err := cgi.Serve(http.HandlerFunc(handleMux)); err != nil {
		log.Fatal(err)
	}
}

type App struct {
	cfg Config
	ses *sessions.Session
	tz  *time.Location
}

func (app *App) startSession(w http.ResponseWriter, r *http.Request, now time.Time) error {
	app.ses.Values["timeout"] = now.Add(toSession).Unix()
	return app.ses.Save(r, w)
}

func (app *App) stopSession(w http.ResponseWriter, r *http.Request) error {
	delete(app.ses.Values, "timeout")
	return app.ses.Save(r, w)
}

func (app *App) KeepAlive(w http.ResponseWriter, r *http.Request, now time.Time) error {
	if app.IsLoggedIn(now) {
		return app.startSession(w, r, now)
	}
	return nil
}

func (app App) IsLoggedIn(now time.Time) bool {
	// https://gowebexamples.com/sessions/
	// or https://stackoverflow.com/questions/28616830/gorilla-sessions-how-to-automatically-update-cookie-expiration-on-request
	timeout, ok := app.ses.Values["timeout"].(int64)
	return ok && now.Before(time.Unix(timeout, 0))
}

func (app App) LoadFeed() (Feed, error) {
	defer un(trace("App.LoadFeed"))
	if feed, err := FeedFromFileName(fileFeedStorage); err != nil {
		return feed, err
	} else {
		for _, ent := range feed.Entries {
			if 6 == len(ent.Id) {
				if id, err := base64ToBase24x7(string(ent.Id)); err != nil {
					log.Printf("Error converting id \"%s\": %s\n", ent.Id, err)
				} else {
					log.Printf("shaarli_go_path_0 + \"?(%[1]s|\\?)%[2]s/?$\" => \"%[1]s%[3]s/\",\n", uriPubPosts, ent.Id, id)
					ent.Id = Id(id)
				}
			}
		}
		return feed, nil
	}
}

// Internal storage, not publishing.
func (app App) SaveFeed(feed Feed) error {
	defer un(trace("App.SaveFeed"))
	feed.Id = ""
	feed.XmlBase = ""
	feed.Generator = nil
	feed.Updated = iso8601{}
	feed.Categories = nil
	return feed.SaveToFile(fileFeedStorage)
}

func handleMux(w http.ResponseWriter, r *http.Request) {
	defer un(trace(strings.Join([]string{"v", version, "+", GitSHA1, " ", r.RemoteAddr, " ", r.Method, " ", r.URL.String()}, "")))
	// w.Header().Set("Server", strings.Join([]string{myselfNamespace, CurrentShaarliGoVersion}, "#"))
	// w.Header().Set("X-Powered-By", strings.Join([]string{myselfNamespace, CurrentShaarliGoVersion}, "#"))
	now := time.Now()

	// check if the request is from a banned client
	if banned, err := isBanned(r, now); err != nil || banned {
		if err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
		} else {
			http.Error(w, "Sorry, banned", http.StatusNotAcceptable)
		}
		return
	}

	path_info := os.Getenv("PATH_INFO")
	//	script_name :=
	urlBase := mustParseURL(string(xmlBaseFromRequestURL(r.URL, os.Getenv("SCRIPT_NAME"))))

	// unpack (nonexisting) static files
	func() {
		if _, err := os.Stat(filepath.Join(dirApp, "delete_me_to_restore")); !os.IsNotExist(err) {
			return
		}
		defer un(trace("RestoreAssets"))
		for _, filename := range AssetNames() {
			if _, err := os.Stat(filename); os.IsNotExist(err) {
				if err := RestoreAsset(".", filename); err != nil {
					http.Error(w, "failed "+filename+": "+err.Error(), http.StatusInternalServerError)
					return
				} else {
					log.Printf("create %s\n", filename)
				}
			} else {
				log.Printf("keep   %s\n", filename)
			}
		}
		// os.Chmod(dirApp, os.FileMode(0750)) // not sure if this is a good idea.
	}()

	// get config and session
	app := App{}
	{
		var err error
		if app.cfg, err = LoadConfig(); err != nil {
			http.Error(w, "Couldn't load config: "+err.Error(), http.StatusInternalServerError)
			return
		}
		var buf []byte
		if buf, err = base64.StdEncoding.DecodeString(app.cfg.CookieStoreSecret); err != nil {
			http.Error(w, "Couldn't get seed: "+err.Error(), http.StatusInternalServerError)
			return
		} else {
			// what if the cookie has changed? Ignore cookie errors, especially on new/changed keys.
			app.ses, _ = sessions.NewCookieStore(buf).Get(r, "ShaarliGo")
			app.ses.Options = &sessions.Options{
				Path:     urlBase.EscapedPath(), // to match all requests
				MaxAge:   int(toSession / time.Second),
				HttpOnly: true,
			}
		}
		if app.tz, err = time.LoadLocation(app.cfg.TimeZone); err != nil {
			http.Error(w, "Invalid timezone '"+app.cfg.TimeZone+"': "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	switch path_info {
	case "/about":
		base := *r.URL
		base.Path = path.Join(base.Path[0:len(base.Path)-len(path_info)], "about") + "/"
		http.Redirect(w, r, base.Path, http.StatusFound)

		return
	case "/about/":
		w.Header().Set("Content-Type", "text/xml; charset=utf-8")
		io.WriteString(w, xml.Header)
		io.WriteString(w, `<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
   xmlns:rfc="https://tools.ietf.org/html/"
   xmlns="http://usefulinc.com/ns/doap#">
  <Project>
    <name>🌺 ShaarliGo</name>
    <audience>Self-hosting Microbloggers</audience>
    <short-description xml:lang="en">🌺 self-hosted microblogging inspired by http://sebsauvage.net/wiki/doku.php?id=php:shaarli. Destilled down to the bare minimum, with easy hosting and security in mind. No PHP, no DB, no server-side templating, JS optional.</short-description>
    <implements rdf:resource="https://sebsauvage.net/wiki/doku.php?id=php:shaarli"/>
    <implements rdf:resource="https://tools.ietf.org/html/rfc4287"/>
    <implements rdf:resource="https://tools.ietf.org/html/rfc5005"/>
    <!-- implements rdf:resource="https://tools.ietf.org/html/rfc5023"/ -->
    <service-endpoint rdf:resource="https://demo.mro.name/shaarligo"/>
    <blog rdf:resource="https://demo.mro.name/shaarligo"/>
    <platform rdf:resource="https://httpd.apache.org/"/>
    <platform rdf:resource="https://www.lighttpd.net/"/>
    <platform rdf:resource="https://tools.ietf.org/html/rfc3875"/>
    <homepage rdf:resource="http://purl.mro.name/ShaarliGo"/>
    <wiki rdf:resource="https://code.mro.name/mro/ShaarliGo/wiki"/>
    <bug-database rdf:resource="https://code.mro.name/mro/ShaarliGo/issues"/>
    <maintainer rdf:resource="http://mro.name/~me"/>
    <programming-language>golang</programming-language>
    <programming-language>xslt</programming-language>
    <programming-language>js</programming-language>
    <category>self-hosting</category>
    <category>microblogging</category>
    <category>shaarli</category>
    <category>nodb</category>
    <category>static</category>
    <category>atom</category>
    <category>cgi</category>
    <repository>
      <GitRepository>
        <browse rdf:resource="https://code.mro.name/mro/ShaarliGo"/>
        <location rdf:resource="https://code.mro.name/mro/ShaarliGo.git"/>
      </GitRepository>
    </repository>
    <release>
      <Version>
        <name>`+version+"+"+GitSHA1+`</name>
        <revision>`+GitSHA1+`</revision>
        <description xml:lang="en">…</description>
      </Version>
    </release>
  </Project>
</rdf:RDF>`)

		return
	case "/config/":
		// make a 404 (fallthrough) if already configured but not currently logged in
		if !app.cfg.IsConfigured() || app.IsLoggedIn(now) {
			app.KeepAlive(w, r, now)
			app.handleSettings(w, r)
			return
		}
	case "/session/":
		// maybe cache a bit, but never KeepAlive
		if app.IsLoggedIn(now) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			// w.Header().Set("Etag", r.URL.Path)
			// w.Header().Set("Cache-Control", "max-age=59") // 59 Seconds
			io.WriteString(w, app.cfg.Uid)
		} else {
			// don't squeal to ban.
			http.NotFound(w, r)
		}
		return
	case "":
		app.KeepAlive(w, r, now)
		params := r.URL.Query()
		switch {
		case "" == r.URL.RawQuery && !app.cfg.IsConfigured():
			http.Redirect(w, r, path.Join(r.URL.Path, "config")+"/", http.StatusSeeOther)
			return

		// legacy API, https://code.mro.name/mro/Shaarli-API-test
		case 1 == len(params["post"]):
			app.handleDoPost(w, r)
			return
		case (1 == len(params["do"]) && "login" == params["do"][0]) ||
			(http.MethodPost == r.Method && "" != r.FormValue("login")): // really. https://github.com/sebsauvage/Shaarli/blob/master/index.php#L402
			app.handleDoLogin(w, r)
			return
		case 1 == len(params["do"]) && "logout" == params["do"][0]:
			app.handleDoLogout(w, r)
			return
		case 1 == len(params["do"]) && "changepasswd" == params["do"][0]:
			app.handleDoCheckLoginAfterTheFact(w, r)
			return
		case 1 == len(params):
			// redirect legacy Ids [A-Za-z0-9_-]{6} in case
			for k, v := range params {
				if 1 == len(v) && "" == v[0] && len(k) == 6 {
					if id, err := base64ToBase24x7(k); err != nil {
						http.Error(w, "Invalid Id '"+k+"': "+err.Error(), http.StatusNotAcceptable)
					} else {
						log.Printf("shaarli_go_path_0 + \"?(%[1]s|\\?)%[2]s/?$\" => \"%[1]s%[3]s/\",\n", uriPubPosts, k, id)
						http.Redirect(w, r, path.Join(r.URL.Path, "..", uriPub, uriPosts, id)+"/", http.StatusMovedPermanently)
					}
					return
				}
			}
		}
	case "/search/":
		app.handleSearch(w, r)
		return
	case "/tools/":
		app.handleTools(w, r)
		return
	}
	squealFailure(r, now, "404")
	http.NotFound(w, r)
}
