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
	"net/http"
	"net/url"
	"time"
)

func (app *App) handleEditPost(w http.ResponseWriter, r *http.Request) {
	now := time.Now()

	if !app.IsLoggedIn(now) {
		http.Redirect(w, r, cgiName+"?do=login&returnurl="+url.QueryEscape(r.URL.String()), http.StatusUnauthorized)
		return
	}

	if !app.cfg.IsConfigured() {
		http.Redirect(w, r, cgiName+"/config", http.StatusPreconditionFailed)
		return
	}

	// feed, _ := FeedFromFileName(fileFeedStorage)

	switch r.Method {
	case http.MethodGet:
		// return the atom xml
	case http.MethodPost:
		// create the atom xml
	case http.MethodPut:
		// change but do not create the atom xml
	case http.MethodDelete:
		// delete the atom xml
	case http.MethodHead:
		// delete the atom xml
	}
}
