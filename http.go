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

package main

import (
	"compress/gzip"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func contains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if needle == s {
			return true
		}
	}
	return false
}

func HttpGetBody(url *url.URL, timeout time.Duration) (io.Reader, error) {
	defer un(trace(strings.Join([]string{"HttpGetBody", url.String()}, " ")))
	client := &http.Client{Timeout: timeout}
	req, _ := http.NewRequest("GET", url.String(), nil)
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("User-Agent", strings.Join([]string{myselfNamespace, version}, ""))
	if resp, err := client.Do(req); nil == resp && nil != err {
		return nil, err
	} else {
		encs := resp.Header["Content-Encoding"]
		switch {
		case contains(encs, "gzip"), contains(encs, "deflate"):
			return gzip.NewReader(resp.Body)
		case 0 == len(encs):
			// NOP
		default:
			log.Printf("Strange compression: %s\n", encs)
		}
		return resp.Body, err
	}
}
