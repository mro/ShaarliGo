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
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v2"
)

type BanPenalties struct {
	Penalties map[string]struct {
		Penalty int
		End     time.Time
	}
}

func isBanned(r *http.Request, now time.Time) (bool, error) {
	return isRemoteAddrBannedFromYamlFileName(r.RemoteAddr, now, filepath.Join("app", "var", "session.yaml"))
}

func isRemoteAddrBannedFromYamlFileName(r string, now time.Time, yamlFileName string) (bool, error) {
	if data, err := ioutil.ReadFile(yamlFileName); err != nil {
		bans := BanPenalties{}
		if err := yaml.Unmarshal(data, &bans); err != nil {
			return true, err
		} else {
			return bans.isRemoteAddrBanned(r, now), nil
		}
	} else {
		return true, err
	}
}

func (bans BanPenalties) isRemoteAddrBanned(key string, now time.Time) bool {
	pen := bans.Penalties[key]
	if pen.Penalty <= 4 { // allow for some failed tries
		return false
	}
	return pen.End.After(now)
}

func squealFailure(r *http.Request, now time.Time) error {
	return squealFailureToYamlFileName(r.RemoteAddr, now, filepath.Join("app", "var", "session.yaml"))
}

func squealFailureToYamlFileName(key string, now time.Time, yamlFileName string) error {
	var err error
	var data []byte
	if data, err = ioutil.ReadFile(yamlFileName); err == nil {
		bans := BanPenalties{}
		if err = yaml.Unmarshal(data, &bans); err == nil {
			if bans.squealFailure(key, now) {
				if data, err = yaml.Marshal(bans); err == nil {
					tmpFileName := fmt.Sprintf("%s~%d", yamlFileName, os.Getpid()) // want the mv to be atomic, so use the same dir
					if err = ioutil.WriteFile(tmpFileName, data, os.ModePerm); err == nil {
						err = os.Rename(tmpFileName, yamlFileName)
					}
				}
			}
		}
	}
	return err
}

func (bans *BanPenalties) squealFailure(key string, now time.Time) bool {
	pen := bans.Penalties[key]
	if pen.Penalty < 0 {
		return false // we're known and welcome. So we do not ban.
	}

	if pen.End.Before(now) {
		pen.End = now
	}

	if pen.End.After(now.Add(time.Hour)) {
		// already banned for more than an hour left, so don't bother adding to the ban.
		// But rather reduce I/O load a bit.
		return false
	}

	pen.End.Add(4 * time.Hour)
	pen.Penalty += 1
	bans.Penalties[key] = pen

	for ip, pen := range bans.Penalties {
		if pen.End.Before(now) {
			delete(bans.Penalties, ip)
		}
	}

	return true
}

// // // // // // // // // // // // // // // // // // // // // // // // //

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
