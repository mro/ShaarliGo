//
// Copyright (C) 2017-2020 Marcus Rohrmoser, http://purl.mro.name/ShaarliGo
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
	"net/url"
	"sort"
	"strings"
	"time"
)

func yesno(yes bool) string {
	if !yes {
		return "no"
	}
	return "yes"
}

// limit bytes brute force, breaks multibyte at the end in case.
func limit(mx int, eli, str string) string {
	if len(str) > mx {
		return str[:mx-len(eli)] + eli
	}
	return str
}

// https://pinboard.in/api/#posts_add
// https://api.pinboard.in/v1/posts/add
//
// url	url	the URL of the item
// description	title	Title of the item. This field is unfortunately named 'description' for backwards compatibility with the delicious API
// extended	text	Description of the item. Called 'extended' for backwards compatibility with delicious API
// tags	tag	List of up to 100 tags
// dt	datetime	creation time for this bookmark. Defaults to current time. Datestamps more than 10 minutes ahead of server time will be reset to current server time
// replace	yes/no	Replace any existing bookmark with this URL. Default is yes. If set to no, will throw an error if bookmark exists
// shared	yes/no	Make bookmark public. Default is "yes" unless user has enabled the "save all bookmarks as private" user setting, in which case default is "no"
// toread	yes/no	Marks the bookmark as unread. Default is "no"
func pinboardPostsAdd(base url.URL, en Entry, foot string) (url url.URL, err error) {
	if 0 == len(en.Links) {
		return url, fmt.Errorf("need an url.")
	}
	body := func(t *HumanText) string {
		if t == nil {
			return ""
		}
		return t.Body
	}
	pars := base.Query()
	pars.Add("url", en.Links[0].Href)
	{
		ti := limit(255, "…", body(&en.Title))
		pars.Add("description", ti)
	}
	{
		de := body(en.Content)
		if "" != de && !strings.HasSuffix(de, "\n") {
			foot = "\n" + foot
		}
		de = limit(65536-len(foot), "…", de)
		de += foot
		pars.Add("extended", de)
	}
	{
		tgs := make([]string, 0, len(en.Categories))
		for _, ca := range en.Categories {
			tgs = append(tgs, ca.Term)
		}
		sort.Strings(tgs)
		ta := limit(255, "…", strings.Join(tgs, " "))
		pars.Add("tags", ta)
	}
	// pars.Add("replace", yesno(replace))
	// pars.Add("shared", yesno(shared))
	// pars.Add("toread", yesno(toread))
	pars.Add("dt", en.Published.Format(time.RFC3339))
	url = base
	if !strings.HasSuffix(url.Path, "/") {
		url.Path += "/"
	}
	url.Path += "posts/add"
	url.RawQuery = pars.Encode()
	return
}
