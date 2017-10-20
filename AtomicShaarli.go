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
	// "fmt"
	"io"
	"log"
	"net/http"
	"net/http/cgi"
	"os"
	"strings"
	"time"
)

const myselfNamespace = "http://purl.mro.name/AtomicShaarli"

// evtl. as a server, too: http://www.dav-muz.net/blog/2013/09/how-to-use-go-and-fastcgi/
func main() {
	// log.Print("I am AtomicShaarli log.Print") stderr
	// fmt.Print("I am AtomicShaarli fmt.Print") http 500
	// fmt.Fprint(os.Stderr, "I am AtomicShaarli fmt.Fprint(os.Stderr, ...)") stderr

	// - check non-write perm of program?
	// - check non-http read perm on ./app
	err := cgi.Serve(http.HandlerFunc(handleMux))

	if err != nil {
		log.Fatal(err)
	}
}

func respond(code int, msg string, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	io.WriteString(w, msg)
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
	now := time.Now()
	mgr := GetManager()
	{
		banned, err := mgr.IsBanned(r, now)
		if ifErrRespond500(err, w, r) {
			return
		}
		if banned {
			respond(http.StatusNotAcceptable, "Sorry, banned", w, r)
			return
		}
	}
	if ifErrRespond500(mgr.LoadConfig(), w, r) {
		return
	}

	// session/login state
	// cooks := r.Cookies()

	script_name := os.Getenv("SCRIPT_NAME")
	path_info := os.Getenv("PATH_INFO")

	switch {
	case "/settings" == path_info:
		if mgr.IsConfigured() && !mgr.IsLoggedIn(r, now) {
			// we need a login first.
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.Header().Set("Location", script_name+"/login"+"?url="+r.URL.String())
			w.WriteHeader(http.StatusSeeOther)
			io.WriteString(w, "I need a login first, redirecting to "+script_name+"/settings"+"\n")
		} else {
			ifErrRespond500(mgr.handleSettings(w, r), w, r)
		}
	case !mgr.IsConfigured():
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		// http.Redirect(w, r, script_name+"/config", http.StatusSeeOther) adds html :-(
		w.Header().Set("Location", script_name+"/settings")
		w.WriteHeader(http.StatusSeeOther)
		io.WriteString(w, "configure first, redirecting to "+script_name+"/settings"+"\n")
	case "/login" == path_info:
		ifErrRespond500(mgr.handleLogin(w, r), w, r)
	case "/logout" == path_info:
		ifErrRespond500(mgr.handleLogout(w, r), w, r)
	case r.URL.Path == "":
		switch {
		case r.URL.RawQuery == "do=login":
			ifErrRespond500(mgr.handleLogin(w, r), w, r)
		case r.URL.RawQuery == "do=logout":
			ifErrRespond500(mgr.handleLogout(w, r), w, r)
		case strings.HasPrefix(r.URL.RawQuery, "post="):
			ifErrRespond500(mgr.handlePost(w, r), w, r)
		case strings.HasPrefix(r.URL.RawQuery, "q="):
			ifErrRespond500(mgr.handleSearch(w, r), w, r)
		}
	default:
		mgr.SquealFailure(r, now)
		// http.NotFoundHandler().ServeHTTP(w, r)
		w.WriteHeader(http.StatusNotFound)
		io.WriteString(w, "not found: "+r.URL.String()+"\n")
	}
}

func (mgr *SessionManager) handleLogin(w http.ResponseWriter, r *http.Request) error {
	io.WriteString(w, "login"+"\n")
	// 'GET': send a login form to the client
	// 'POST' validate, respond error (and squeal) or set session and redirect
	return nil
}

func (mgr *SessionManager) handleLogout(w http.ResponseWriter, r *http.Request) error {
	io.WriteString(w, "logout"+"\n")
	//  invalidate session and redirect
	return nil
}

func (mgr *SessionManager) handlePost(w http.ResponseWriter, r *http.Request) error {
	io.WriteString(w, "post"+"\n")
	// 'GET': send a form to the client
	// 'POST' validate, respond error (and squeal) or post and redirect
	return nil
}

func (mgr *SessionManager) handleSearch(w http.ResponseWriter, r *http.Request) error {
	io.WriteString(w, "search"+"\n")
	return nil
}
