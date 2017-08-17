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
	"fmt"
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
	log.Print("I am AtomicShaarli log.Print")
	// fmt.Print("I am AtomicShaarli fmt.Print") http 500
	fmt.Fprint(os.Stderr, "I am AtomicShaarli fmt.Fprint(os.Stderr, ...)")

	// - check non-write perm of program?
	// - check non-http read perm on ./app
	err := cgi.Serve(http.HandlerFunc(handleMux))

	if err != nil {
		log.Fatal(err)
	}
}

func respond(w http.ResponseWriter, r *http.Request, code int, msg string) {
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	io.WriteString(w, msg)
}

func ifErrRespond500(err error, w http.ResponseWriter, r *http.Request) bool {
	if err != nil {
		respond(w, r, http.StatusInternalServerError, "error:\n")
		io.WriteString(w, err.Error())
	}
	return err != nil
}

func handleMux(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	mgr := GetManager()
	{
		banned, err := mgr.IsBanned(r, now)
		if ifErrRespond500(err, w, r) {
			return
		}
		if banned {
			respond(w, r, http.StatusNotAcceptable, "Sorry, banned")
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
		if !mgr.IsLoggedIn(r, now) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.Header().Set("Location", script_name+"/login"+"?url="+r.URL.String())
			w.WriteHeader(http.StatusSeeOther)
			io.WriteString(w, "I need a login first, redirecting to "+script_name+"/settings"+"\n")
			return
		} else {
			ifErrRespond500(handleSettings(mgr, w, r), w, r)
			return
		}
	case !mgr.IsConfigured():
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		// http.Redirect(w, r, script_name+"/config", http.StatusSeeOther) adds html :-(
		w.Header().Set("Location", script_name+"/settings")
		w.WriteHeader(http.StatusSeeOther)
		io.WriteString(w, "configure first, redirecting to "+script_name+"/settings"+"\n")
	case "/login" == path_info:
		io.WriteString(w, "session\n")
	case "/logout" == path_info:
		io.WriteString(w, "session\n")
	case r.URL.Path == "":
		switch {
		case r.URL.RawQuery == "do=login":
			io.WriteString(w, "login\n")
		case r.URL.RawQuery == "do=logout":
			io.WriteString(w, "logout\n")
		case strings.HasPrefix(r.URL.RawQuery, "q="):
			io.WriteString(w, "search\n")
		}
	default:
		mgr.SquealFailure(r, now)
		http.NotFoundHandler().ServeHTTP(w, r)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Debug-Pfad", r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
		io.WriteString(w, "not found: "+r.URL.String()+"\n")
	}
}
