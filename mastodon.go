//
// Copyright (C) 2020-2021 Marcus Rohrmoser, http://purl.mro.name/ShaarliGo
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
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// https://docs.joinmastodon.org/methods/statuses/
// https://docs.joinmastodon.org/client/token/
//
func mastodonStatusPost(base url.URL, token string, maxlen int, en Entry, foot string) (err error) {
	if 0 == len(en.Links) {
		return fmt.Errorf("need an url.")
	}
	body := func(t *HumanText) string {
		if t == nil {
			return ""
		}
		return t.Body
	}

	form := url.Values{}
	{
		txt := body(&en.Title)
		txt += "\n" + en.Links[0].Href
		txt += "\n" + body(en.Content)
		if "" != txt && !strings.HasSuffix(txt, "\n") {
			foot = "\n" + foot
		}
		txt = limit(maxlen-len(foot), "…", txt)
		txt += foot
		form.Add("status", txt)
	}

	httpPostBody := func(ur *url.URL, token string, form url.Values, timeout time.Duration) (io.Reader, error) {
		defer un(trace(strings.Join([]string{"HttpPostBody", ur.String()}, " ")))
		client := &http.Client{Timeout: timeout}

		req, _ := http.NewRequest(http.MethodPost, ur.String(), strings.NewReader(form.Encode()))
		req.Header.Set("Accept-Encoding", "gzip, deflate")
		req.Header.Set("User-Agent", strings.Join([]string{myselfNamespace, version}, ""))
		if "" != token {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
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

	if !strings.HasSuffix(base.Path, "/") {
		base.Path += "/"
	}
	base.Path += "statuses"
	_, err = httpPostBody(&base, token, form, 4*time.Second)
	return
}
