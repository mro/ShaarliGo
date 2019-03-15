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
	"net/url"
	"strings"
	"time"
)

func yesno(yes bool) string {
	if !yes {
		return "no"
	}
	return "yes"
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
func pinboardAddUrl(base *url.URL, auth_token string, ur *url.URL, description, extended string, tags []string, dt time.Time) *url.URL {
	// _, err := HttpGetBody(&url, timeout)
	pars := url.Values{}
	pars.Add("url", ur.String())
	pars.Add("auth_token", auth_token)
	pars.Add("description", description)
	pars.Add("extended", extended)
	pars.Add("tags", strings.Join(tags, " "))
	// pars.Add("replace", yesno(replace))
	// pars.Add("shared", yesno(shared))
	// pars.Add("toread", yesno(toread))
	pars.Add("dt", dt.Format(time.RFC3339))
	ret := &(*base)
	ret.Path += "/posts/add"
	ret.RawQuery = pars.Encode()
	return ret
}
