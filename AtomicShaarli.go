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

// Files & Directories
//
// .htaccess
// atom.cgi
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
	"io"
	"log"
	"net/http"
	"net/http/cgi"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/gorilla/sessions"
)

const myselfNamespace = "http://purl.mro.name/AtomicShaarli"
const toSession = 30 * time.Minute

// evtl. as a server, too: http://www.dav-muz.net/blog/2013/09/how-to-use-go-and-fastcgi/
func main() {
	{ // log to custom logfile rather than stderr (which may not accessible on shared hosting)
		dst := filepath.Join("app", "var", "error.log")
		if err := os.MkdirAll(filepath.Dir(dst), 0700); err != nil {
			log.Fatal("Couldn't create app/var dir: " + err.Error())
			return
		}
		if w, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600); err != nil {
			log.Fatal("Couldn't create open logfile: " + err.Error())
			return
		} else {
			defer w.Close()
			log.SetOutput(w)
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

func respond(code int, msg string, w http.ResponseWriter, r *http.Request) {
	http.Error(w, msg, code)
}

func ifErrRespond500(err error, w http.ResponseWriter, r *http.Request) bool {
	if err != nil {
		respond(http.StatusInternalServerError, "error:\n", w, r)
		io.WriteString(w, err.Error())
	}
	return err != nil
}

func handleMux(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Server", myselfNamespace)
	w.Header().Set("CGI-Server", myselfNamespace)

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
	urlBase := xmlBaseFromRequestURL(r.URL, os.Getenv("SCRIPT_NAME"))

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
			app.ses, _ = sessions.NewCookieStore(buf).Get(r, "AtomicShaarli")
			app.ses.Options = &sessions.Options{
				Path:     urlBase.Path, // to match all requests
				MaxAge:   int(toSession / time.Second),
				HttpOnly: true,
			}
		}
	}

	switch path_info {
	case "/config":
		// make a 404 if already configured but not currently logged in
		if !app.cfg.IsConfigured() || app.IsLoggedIn(now) {
			app.handleSettings(w, r)
			return
		}
	case "/session":
		if app.IsLoggedIn(now) {
			// maybe cache, but don't KeepAlive
			// w.Header().Set("Etag", r.URL.Path)
			// w.Header().Set("Cache-Control", "max-age=60") // 60 Seconds
			io.WriteString(w, app.cfg.AuthorName)
		} else {
			// don't squeal to ban.
			http.NotFound(w, r)
		}
		return
	case "":
		params := r.URL.Query()
		switch {
		case "" == r.URL.RawQuery && !app.cfg.IsConfigured():
			http.Redirect(w, r, path.Join(r.URL.Path, "config"), http.StatusSeeOther)
			return

		// legacy API, https://github.com/mro/Shaarli-API-Test
		case 1 == len(params["post"]):
			app.handleDoPost(w, r)
			return
		case 1 == len(params["do"]) && "login" == params["do"][0]:
			app.handleDoLogin(w, r)
			return
		case 1 == len(params["do"]) && "logout" == params["do"][0]:
			app.handleDoLogout(w, r)
			return
		}
	case "/tools":

	case "/search":

	}
	squealFailure(r, now)
	http.NotFound(w, r)
}
