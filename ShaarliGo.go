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

// Files & Directories
//
// .htaccess
// shaarli.cgi
// app/.htaccess
// app/config.yaml
// app/posts.gob.gz
// app/posts.xml.gz
// app/var/bans.yaml
// app/var/error.log
// app/var/stage/
// app/var/old/
// themes/current/
// o/p/
//
package main

import (
	"encoding/base64"
	"encoding/gob"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cgi"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
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

func LoadFeed() (Feed, error) {
	defer un(trace("LoadFeed"))
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

// are we running cli
func runCli() bool {
	if 0 != len(os.Getenv("REQUEST_METHOD")) {
		return false
	}
	fmt.Printf("%sv%s+%s#:\n", myselfNamespace, version, GitSHA1)

	cfg, err := LoadConfig()
	if err != nil {
		panic(err)
	}
	fmt.Printf("  timezone: %s\n", cfg.TimeZone)

	feed, err := LoadFeed()
	if os.IsNotExist(err) {
		cwd, _ := os.Getwd()
		fmt.Fprintf(os.Stderr, "%s: cannot access %s: No such file or directory\n", filepath.Base(os.Args[0]), filepath.Join(cwd, fileFeedStorage))
		os.Exit(1)
		return true
	}
	if err != nil {
		panic(err)
	}
	fmt.Printf("  posts: %d\n", len(feed.Entries))
	//fmt.Printf("  tags:  %d\n", len(feed.Categories))
	if 0 < len(feed.Entries) {
		fmt.Printf("  first: %v\n", feed.Entries[len(feed.Entries)-1].Published.Format(time.RFC3339))
		fmt.Printf("  last:  %v\n", feed.Entries[0].Published.Format(time.RFC3339))
	}

	return true
}

// evtl. as a server, too: http://www.dav-muz.net/blog/2013/09/how-to-use-go-and-fastcgi/
func main() {
	if runCli() {
		return
	}

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

	wg := &sync.WaitGroup{}
	// - check non-write perm of program?
	// - check non-http read perm on ./app
	if err := cgi.Serve(http.HandlerFunc(handleMux(wg))); err != nil {
		log.Fatal(err)
	}
	wg.Wait()
}

type Server struct {
	cfg Config
	ses *sessions.Session
	tz  *time.Location
	url url.URL
	cgi url.URL
}

func (app *Server) startSession(w http.ResponseWriter, r *http.Request, now time.Time) error {
	app.ses.Values["timeout"] = now.Add(toSession).Unix()
	return app.ses.Save(r, w)
}

func (app *Server) stopSession(w http.ResponseWriter, r *http.Request) error {
	delete(app.ses.Values, "timeout")
	return app.ses.Save(r, w)
}

func (app *Server) KeepAlive(w http.ResponseWriter, r *http.Request, now time.Time) error {
	if app.IsLoggedIn(now) {
		return app.startSession(w, r, now)
	}
	return nil
}

func (app Server) IsLoggedIn(now time.Time) bool {
	// https://gowebexamples.com/sessions/
	// or https://stackoverflow.com/questions/28616830/gorilla-sessions-how-to-automatically-update-cookie-expiration-on-request
	timeout, ok := app.ses.Values["timeout"].(int64)
	return ok && now.Before(time.Unix(timeout, 0))
}

// Internal storage, not publishing.
func (app Server) SaveFeed(feed Feed) error {
	defer un(trace("Server.SaveFeed"))
	feed.Id = ""
	feed.XmlBase = ""
	feed.Generator = nil
	feed.Updated = iso8601{}
	feed.Categories = nil
	return feed.SaveToFile(fileFeedStorage)
}

func (app Server) Posse(en Entry) {
	defer un(trace("Server.Posse"))
	to := 4 * time.Second
	for _, pi := range app.cfg.Posse {
		if ep, err := url.Parse(pi.Endpoint); err != nil {
			log.Printf("- posse %s error %s\n", pi, err)
		} else {
			foot := pi.Prefix
			if "" == foot {
				foot = "¹ " + app.url.String() + uriPubPosts
			}
			if url, err := pinboardPostsAdd(*ep, en, foot+string(en.Id)); err != nil {
				log.Printf("- posse %s error %s\n", ep, err)
			} else {
				if _, err := HttpGetBody(&url, to); err != nil {
					log.Printf("- posse %s error %s\n", url.String(), err)
				} else {
					// TODO: check response
					log.Printf("- posse %s\n", url.String())
				}
			}
		}
	}
}

func handleMux(wg *sync.WaitGroup) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer un(trace(strings.Join([]string{"v", version, "+", GitSHA1, " ", r.RemoteAddr, " ", r.Method, " ", r.URL.String()}, "")))

		// w.Header().Set("Server", strings.Join([]string{myselfNamespace, CurrentShaarliGoVersion}, "#"))
		// w.Header().Set("X-Powered-By", strings.Join([]string{myselfNamespace, CurrentShaarliGoVersion}, "#"))
		now := time.Now()

		// check if the request is from a banned client
		if banned, err := isBanned(r, now); err != nil || banned {
			if err != nil {
				http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
			} else {
				w.Header().Set("Retry-After", "14400") // https://www.w3.org/Protocols/rfc2616/rfc2616-sec14.html#sec14.37
				// evtl. 429 StatusTooManyRequests?
				// or    503 StatusServiceUnavailable?
				http.Error(w, "Sorry, banned", http.StatusTooManyRequests)
			}
			return
		}

		if !r.URL.IsAbs() {
			log.Printf("request URL not absolute >>> %s <<<", r.URL)
		}

		path_info := os.Getenv("PATH_INFO")

		// unpack (nonexisting) static files
		func() {
			if _, err := os.Stat(filepath.Join(dirApp, "delete_me_to_restore")); !os.IsNotExist(err) {
				return
			}
			defer un(trace("RestoreAssets"))
			for _, filename := range AssetNames() {
				if filepath.Dir(filename) == "tpl" {
					continue
				}
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

		cfg, err := LoadConfig()
		if err != nil {
			log.Printf("Couldn't load config: %s", err.Error())
			http.Error(w, "Couldn't load config: "+err.Error(), http.StatusInternalServerError)
			return
		}
		tz, err := time.LoadLocation(cfg.TimeZone)
		if err != nil {
			http.Error(w, "Invalid timezone '"+cfg.TimeZone+"': "+err.Error(), http.StatusInternalServerError)
			return
		}

		// get config and session
		app := Server{cfg: cfg, tz: tz}
		{
			app.cgi = func(u url.URL, cgi string) url.URL {
				u.Path = cgi
				u.RawQuery = ""
				return u
			}(*r.URL, os.Getenv("SCRIPT_NAME"))
			app.url = app.cgi
			app.url.Path = path.Dir(app.cgi.Path)
			if !strings.HasSuffix(app.url.Path, "/") {
				app.url.Path += "/"
			}

			var err error
			var buf []byte
			if buf, err = base64.StdEncoding.DecodeString(app.cfg.CookieStoreSecret); err != nil {
				http.Error(w, "Couldn't get seed: "+err.Error(), http.StatusInternalServerError)
				return
			} else {
				// what if the cookie has changed? Ignore cookie errors, especially on new/changed keys.
				app.ses, _ = sessions.NewCookieStore(buf).Get(r, "ShaarliGo")
				app.ses.Options = &sessions.Options{
					Path:     app.url.EscapedPath(), // to match all requests
					MaxAge:   int(toSession / time.Second),
					HttpOnly: true,
				}
			}
		}

		switch path_info {
		case "/about":
			http.Redirect(w, r, "about/", http.StatusFound)

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
    <service-endpoint rdf:resource="https://demo.0x4c.de/shaarligo"/>
    <blog rdf:resource="https://demo.0x4c.de/shaarligo"/>
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
				app.handleSettings()(w, r)
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
			case 1 == len(params["post"]) ||
				("" == r.URL.RawQuery && r.Method == http.MethodPost && r.FormValue("save_edit") == "Save"):
				app.handleDoPost(app.Posse)(w, r)
				return
			case (1 == len(params["do"]) && "login" == params["do"][0]) ||
				(http.MethodPost == r.Method && "" != r.FormValue("login")): // really. https://github.com/sebsauvage/Shaarli/blob/master/index.php#L402
				app.handleDoLogin()(w, r)
				return
			case 1 == len(params["do"]) && "logout" == params["do"][0]:
				app.handleDoLogout()(w, r)
				return
			case 1 == len(params["do"]) && "configure" == params["do"][0]:
				http.Redirect(w, r, path.Join(r.URL.Path, "config")+"/", http.StatusSeeOther)
				return
			case 1 == len(params["do"]) && "changepasswd" == params["do"][0]:
				app.handleDoCheckLoginAfterTheFact()(w, r)
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
			app.handleSearch()(w, r)
			return
		case "/tools/":
			app.handleTools()(w, r)
			return
		}
		squealFailure(r, now, "404")
		http.NotFound(w, r)
	}
}
