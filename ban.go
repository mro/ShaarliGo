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
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

var banFileName string

func init() {
	banFileName = filepath.Join(dirApp, "var", "bans.yaml")
}

type Penalty struct {
	Badness int
	End     time.Time
}

type BanPenalties struct {
	Penalties map[string]Penalty
}

func remoteAddressToKey(addr string) string {
	// strip port number
	if idx := strings.LastIndex(addr, ":"); idx > -1 {
		addr = addr[:idx]
	}
	return addr
}

func isBanned(r *http.Request, now time.Time) (bool, error) {
	key := remoteAddressToKey(r.RemoteAddr)
	if data, err := ioutil.ReadFile(banFileName); err == nil || os.IsNotExist(err) {
		bans := BanPenalties{}
		if err := yaml.Unmarshal(data, &bans); err == nil {
			return bans.isRemoteAddrBanned(key, now), nil
		} else {
			return true, err
		}
	} else {
		return true, err
	}
}

func squealFailure(r *http.Request, now time.Time, reason string) error {
	key := remoteAddressToKey(r.RemoteAddr)
	var err error
	var data []byte
	if data, err = ioutil.ReadFile(banFileName); err == nil || os.IsNotExist(err) {
		bans := BanPenalties{Penalties: map[string]Penalty{}}
		if err = yaml.Unmarshal(data, &bans); err == nil {
			if bans.squealFailure(key, now, reason) {
				if data, err = yaml.Marshal(bans); err == nil {
					tmpFileName := fmt.Sprintf("%s~%d", banFileName, os.Getpid()) // want the mv to be atomic, so use the same dir
					if err = ioutil.WriteFile(tmpFileName, data, 0600); err == nil {
						err = os.Rename(tmpFileName, banFileName)
					}
				}
			}
		}
	}
	return err
}

const banThreshold = 4

func (bans BanPenalties) isRemoteAddrBanned(key string, now time.Time) bool {
	pen := bans.Penalties[key]
	if pen.Badness <= banThreshold { // allow for some failed tries
		return false
	}
	return pen.End.After(now)
}

func (bans *BanPenalties) squealFailure(key string, now time.Time, reason string) bool {
	pen := bans.Penalties[key]
	if pen.Badness < 0 {
		return false // we're known and welcome. So we do not ban.
	}

	if pen.End.Before(now) {
		pen.End = now
	}

	if pen.Badness > banThreshold && pen.End.After(now.Add(1*time.Hour)) {
		// already banned for more than an hour left, so don't bother adding to the ban.
		// But rather reduce I/O load a bit.
		return false
	}

	pen.End = pen.End.Add(4 * time.Hour)
	pen.Badness += 1
	bans.Penalties[key] = pen

	log.Printf("squeal %d %s %s", pen.Badness, key, reason)

	for ip, pen := range bans.Penalties {
		if pen.End.Before(now) {
			delete(bans.Penalties, ip)
		}
	}

	return true
}
