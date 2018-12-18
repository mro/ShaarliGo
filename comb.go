//
// Copyright (C) 2017-2018 Marcus Rohrmoser, http://purl.mro.name/ShaarliGo
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
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/yhat/scrape"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var serverLocation *time.Location

func init() {
	// TODO rather use app settings?
	serverLocation, _ = time.LoadLocation("Europe/Berlin")
}

func entryFromURL(ur *url.URL, timeout time.Duration) (Entry, error) {
	if r, err := HttpGetBody(ur, timeout); err != nil {
		return Entry{}, err
	} else {
		return entryFromReader(r, ur)
	}
}

func entryFromReader(r io.Reader, ur *url.URL) (Entry, error) {
	if root, err := html.Parse(r); err != nil {
		return Entry{}, err
	} else {
		return entryFromNode(root, ur)
	}
}

func entryFromNode(root *html.Node, ur *url.URL) (Entry, error) {
	ret := Entry{}
	for _, node := range scrape.FindAll(root, func(n *html.Node) bool {
		return n.Parent == root && n.Type == html.ElementNode && atom.Html == n.DataAtom
	}) {
		ret.XmlLang = Lang(scrape.Attr(node, "lang"))
		break
	}

	for _, node := range scrape.FindAll(root, func(n *html.Node) bool { return n.Type == html.ElementNode && atom.Meta == n.DataAtom }) {
		strName := scrape.Attr(node, "name")
		strProp := scrape.Attr(node, "property")
		strContent := scrape.Attr(node, "content")
		switch {
		case "title" == strName:
			ret.Title = HumanText{Body: strContent}

		case "description" == strName:
			ret.Summary = &HumanText{Body: strContent}

		case "author" == strName:
			ret.Authors = append(ret.Authors, Person{Name: strContent})

		case "date" == strName:
			var t time.Time
			var err error
			if t, err = time.Parse(time.RFC3339, strContent); err != nil {
				if t, err = time.ParseInLocation("2006-01-02T15:04:05Z0700", strContent, serverLocation); err != nil {
					if t, err = time.ParseInLocation("2006-01-02T15:04:05", strContent, serverLocation); err != nil {
						//panic(err)
					}
				}
			}
			if err == nil {
				ret.Published = iso8601(t)
			}

		case "keywords" == strName:
			for _, txt := range strings.Split(strContent, ",") {
				if t := strings.Replace(strings.TrimSpace(txt), " ", "_", -1); "" != t {
					ret.Categories = append(ret.Categories, Category{Term: t})
				}
			}

		case "og:title" == strProp:
			ret.Title = HumanText{Body: strContent}

		case "og:description" == strProp:
			ret.Summary = &HumanText{Body: strContent}

		case nil == ret.MediaThumbnail && "og:image" == strProp:
			ret.MediaThumbnail = &MediaThumbnail{Url: Iri(strContent)}
		}
	}
	return ret, nil
}
