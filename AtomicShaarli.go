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
)

const myselfNamespace = "http://purl.mro.name/AtomicShaarli"

// evtl. as a server, too: http://www.dav-muz.net/blog/2013/09/how-to-use-go-and-fastcgi/
func main() {
	// - check non-write perm auf Programm
	// - check non-http read perm auf private
	err := cgi.Serve(&muxHandler{})
	if err != nil {
		log.Fatal(err)
	}
}

type muxHandler struct{}

func (h *muxHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	{
		banned, err := GetBanManager().IsBanned(r, nil)
		if banned || err != nil {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				io.WriteString(w, "error.\n")
				fmt.Fprintf(os.Stderr, "%s: %s\n", r.RemoteAddr, err)
			} else {
				w.WriteHeader(http.StatusNotAcceptable)
				io.WriteString(w, "banned.\n")
			}
			return
		}
	}

	cfg, err := LoadConfig()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		io.WriteString(w, "error:\n")
		io.WriteString(w, err.Error())
		return
	}

	// session/login state

	script_name := os.Getenv("SCRIPT_NAME")
	path_info := os.Getenv("PATH_INFO")

	switch {
	case "/settings" == path_info:
		newSettingsHandler(cfg).ServeHTTP(w, r)
	case !cfg.IsConfigured():
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		// http.Redirect(w, r, script_name+"/config", http.StatusSeeOther) adds html :-(
		w.Header().Set("Location", script_name+"/settings")
		w.WriteHeader(http.StatusSeeOther)
		io.WriteString(w, "configure first, redirecting to "+script_name+"/settings"+"\n")
	case "/session" == path_info:
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
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Debug-Pfad", r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
		fmt.Print(w, "Gibt's ja gar nicht\n")
	}
}
