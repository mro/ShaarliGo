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
	"crypto/rand"
	"encoding/base64"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type SessionManager struct {
	baseDir        string
	config         Config
	cookieName     string
	maxLifeSeconds int
}

func GetManager() *SessionManager {
	b := SessionManager{baseDir: "app/var/cache", cookieName: "AtomicShaarli", maxLifeSeconds: 30 * 60}
	return &b
}

func (m *SessionManager) IsLoggedIn(r *http.Request, t time.Time) bool {
	return true
}

func (m *SessionManager) startSession(w http.ResponseWriter, r *http.Request, uid string) {
	// https://astaxie.gitbooks.io/build-web-application-with-golang/en/06.2.html
	b := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return
	}
	sid := base64.URLEncoding.EncodeToString(b)

	cookie := http.Cookie{Name: m.cookieName, Value: url.QueryEscape(sid), Path: "/", HttpOnly: true, MaxAge: m.maxLifeSeconds}
	// todo: store session locally
	// ---
	// sessions:
	//   abc:
	//     expire: '2018-02-19T09:08:27Z'
	//     uid: mro
	// tokens:
	//   a:
	//     expire: '2018-02-19T09:08:27Z'
	//   b:
	//     expire: '2018-02-19T09:10:27Z'
	// bans:
	//   127.0.0.1:
	//     penalty: 12
	//     expire: '2018-02-19T09:10:27Z'
	http.SetCookie(w, &cookie)
}

func (m *SessionManager) PrepareDirs() error {
	return nil
}

func (m *SessionManager) IsBanned(r *http.Request, t time.Time) (bool, error) {
	if r == nil {
		return m.isBanned(nil, t)
	}
	return m.isBanned(&r.RemoteAddr, t)
}

func (m *SessionManager) isBanned(r *string, t0 time.Time) (bool, error) {
	if r == nil {
		return true, nil
	}
	file := m.banMarkerPath(r)
	byt, err := ioutil.ReadFile(*file)
	banEndUnix := int64(-9223372036854775808) // very far past
	if err != nil {
		pe := err.(*os.PathError)
		if pe != nil && pe.Op == "open" {

			//		pe := os.PathError{Op: "open", Path: *file, Err: nil}
			//https://davidnix.io/post/error-handling-in-go/
			//https://www.goinggo.net/2014/10/error-handling-in-go-part-i.html
			//		if &pe == nil {
			//			return true, err
			//		}
		}
	} else {
		// turn byte slice (UTC unix time as a decimal string) into time
		banEndUnix, err = strconv.ParseInt(string(byt[:]), 10, 64)
		if err != nil {
			return true, err
		}
	}
	banned := time.Unix(banEndUnix, 0).After(t0)
	if !banned {
		err = os.Remove(*file)
	}
	return banned, nil
}

func (m *SessionManager) SquealFailure(r *http.Request, t time.Time) error {
	return m.squealFailure(&r.RemoteAddr, t)
}

func (m *SessionManager) squealFailure(r *string, t time.Time) error {
	// load number of tries
	// increment
	// is above ban threshold?
	// yes: add ban, remove failures
	// no: increment failures
	return nil
}

func (m *SessionManager) LiftBanAndFailures(r *http.Request) error {
	return m.liftBanAndFailures(&r.RemoteAddr)
}

func (m *SessionManager) liftBanAndFailures(r *string) error {
	// remove ban
	// remove failures
	return nil
}

func (m *SessionManager) banMarkerPath(r *string) *string {
	ret := filepath.Join(m.baseDir, "ban", "banned", *r)
	return &ret
}
func (m *SessionManager) failureMarkerPath(r *string) *string {
	ret := filepath.Join(m.baseDir, "ban", "failed", *r)
	return &ret
}
