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
	"html/template"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

func (app *App) handleTools(w http.ResponseWriter, r *http.Request) {
	now := time.Now()

	if !app.IsLoggedIn(now) {
		http.Redirect(w, r, cgiName+"?do=login&returnurl="+url.QueryEscape(r.URL.String()), http.StatusUnauthorized)
		return
	}

	if !app.cfg.IsConfigured() {
		http.Redirect(w, r, cgiName+"/config", http.StatusPreconditionFailed)
		return
	}

	switch r.Method {
	case http.MethodGet:
		app.KeepAlive(w, r, now)

		if tmpl, err := template.New("tools").Parse(`<html xmlns="http://www.w3.org/1999/xhtml">
<head><title>{{.title}}</title></head>
<body>
  <ol>
    <li class="config"><a href="../config/">Config</a></li>
    
    <li class="bookmarklet">
      <a
        onclick="alert('Drag this link to your bookmarks toolbar, or right-click it and choose Bookmark This Link...');return false;"
        href="javascript:javascript:(function(){var%20url%20=%20location.href;var%20title%20=%20document.title%20||%20url;window.open('{{.xml_base}}?post='%20+%20encodeURIComponent(url)+'&amp;title='%20+%20encodeURIComponent(title)+'&amp;description='%20+%20encodeURIComponent(document.getSelection())+'&amp;source=bookmarklet','_blank','menubar=no,height=450,width=600,toolbar=no,scrollbars=no,status=no,dialog=1');})();"
      >✚ShaarliGo</a>
      <span>⇐ Drag this link to your bookmarks toolbar (or right-click it and choose Bookmark This Link…).
      Then click "✚ShaarliGo" button in any page you want to share.</span>
    </li>
  </ol>
</body>
</html>
`); err == nil {
			w.Header().Set("Content-Type", "text/xml; charset=utf-8")
			io.WriteString(w, `<?xml version="1.0" encoding="UTF-8"?>
<?xml-stylesheet type='text/xsl' href='../../assets/default/de/tools.xslt'?>
`)
			data := map[string]string{
				"title":    app.cfg.Title,
				"xml_base": xmlBaseFromRequestURL(r.URL, os.Getenv("SCRIPT_NAME")).String() + cgiName,
			}

			if err := tmpl.Execute(w, data); err != nil {
				http.Error(w, "Coudln't send linkform: "+err.Error(), http.StatusInternalServerError)
			}
		}
	}
}
