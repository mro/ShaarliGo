//
// Copyright (C) 2017-2021 Marcus Rohrmoser, http://purl.mro.name/ShaarliGo
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

	"github.com/yhat/scrape"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
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

func formValuesFromReader(r io.Reader, name string) (ret url.Values, err error) {
	root, err := html.Parse(r) // assumes r is UTF8
	if err != nil {
		return ret, err
	}

	for _, form := range scrape.FindAll(root, func(n *html.Node) bool {
		return atom.Form == n.DataAtom &&
			(name == scrape.Attr(n, "name") || name == scrape.Attr(n, "id"))
	}) {
		ret := url.Values{}
		for _, inp := range scrape.FindAll(form, func(n *html.Node) bool {
			return atom.Input == n.DataAtom || atom.Textarea == n.DataAtom
		}) {
			n := scrape.Attr(inp, "name")
			if n == "" {
				n = scrape.Attr(inp, "id")
			}

			ty := scrape.Attr(inp, "type")
			v := scrape.Attr(inp, "value")
			if atom.Textarea == inp.DataAtom {
				v = scrape.Text(inp)
			} else if v == "" && ty == "checkbox" {
				v = scrape.Attr(inp, "checked")
			}
			ret.Set(n, v)
		}
		return ret, err // return on first occurrence
	}
	return ret, err
}
