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
// app/var/session.yaml
// app/var/stage/
// app/var/old/
// assets/
// pub/
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
	"time"

	"github.com/gorilla/sessions"
)

const myselfNamespace = "http://purl.mro.name/AtomicShaarli"

// evtl. as a server, too: http://www.dav-muz.net/blog/2013/09/how-to-use-go-and-fastcgi/
func main() {
	// log.Print("I am AtomicShaarli log.Print") stderr
	// fmt.Print("I am AtomicShaarli fmt.Print") http 500
	// fmt.Fprint(os.Stderr, "I am AtomicShaarli fmt.Fprint(os.Stderr, ...)") stderr

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

func (app App) IsLoggedIn(now time.Time) bool {
	// https://gowebexamples.com/sessions/
	// or https://stackoverflow.com/questions/28616830/gorilla-sessions-how-to-automatically-update-cookie-expiration-on-request
	timeout, ok := app.ses.Values["timeout"].(int64)
	return ok && now.Before(time.Unix(timeout, 0))
}

func (app *App) startSession(w http.ResponseWriter, r *http.Request, now time.Time) error {
	app.ses.Values["timeout"] = now.Add(30 * time.Minute).Unix()
	return app.ses.Save(r, w)
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
		if !ifErrRespond500(err, w, r) {
			respond(http.StatusNotAcceptable, "Sorry, banned", w, r)
		}
		return
	}

	// get config and session
	app := App{}
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
		if app.ses, err = sessions.NewCookieStore(buf).Get(r, "AtomicShaarli"); err != nil {
			// what if the cookie has changed?
			http.Error(w, "Couldn't get session: "+err.Error(), http.StatusInternalServerError)
			return
		} else {
			if app.IsLoggedIn(now) {
				if err = app.startSession(w, r, now); err != nil { // keep the session alive
					http.Error(w, "Couldn't save session: "+err.Error(), http.StatusInternalServerError)
					return
				}
			}
		}
	}

	script_name := os.Getenv("SCRIPT_NAME")
	path_info := os.Getenv("PATH_INFO")

	switch {
	case "/config" == path_info:
		// make a 404 if already configured but not currently logged in
		if !app.cfg.IsConfigured() || app.IsLoggedIn(now) {
			ifErrRespond500(app.handleSettings(w, r), w, r)
			return
		}
	case "/session" == path_info:
		if !app.IsLoggedIn(now) {
			http.NotFound(w, r)
		}
		return
	case "/tools" == path_info:
		ifErrRespond500(app.handleTools(w, r), w, r)
	case "" == path_info:
		param := r.URL.Query()
		switch {
		case "" == r.URL.RawQuery && !app.cfg.IsConfigured():
			http.Redirect(w, r, path.Join(script_name, "config"), http.StatusSeeOther)
			return
		// legacy API, https://github.com/mro/Shaarli-API-Test
		case "login" == param["do"][0]:
			ifErrRespond500(app.handleDoLogin(w, r), w, r)
			return
		case "logout" == param["do"][0]:
			ifErrRespond500(app.handleDoLogout(w, r), w, r)
			return
		case 1 == len(param["post"]):
			ifErrRespond500(app.handleDoPost(w, r), w, r)
			return
		}
	case "/search" == path_info:
		ifErrRespond500(app.handleSearch(w, r), w, r)
		return
	}
	squealFailure(r, now)
	http.NotFound(w, r)
}

func (app *App) handleTools(w http.ResponseWriter, r *http.Request) error {
	io.WriteString(w, "tools"+"\n")
	return nil
}

func (app *App) handleSearch(w http.ResponseWriter, r *http.Request) error {
	io.WriteString(w, "search"+"\n")
	return nil
}
