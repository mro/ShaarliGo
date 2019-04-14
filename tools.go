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
	"encoding/xml"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const timeoutShaarliImportFetch = time.Minute

func (app *Server) handleTools() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
    <li id="disclosure">
      <b>Responsible Disclosure:</b> In case you are reluctant to <a
      href="http://purl.mro.name/ShaarliGo/">file a public issue</a>, feel free to
      email <a href="mailto:ShaarliGo@mro.name?subject=">ShaarliGo@mro.name</a>.
    </li>

    <li id="update">
      <b>Update:</b> Just replace the file <code>shaarligo.cgi</code>. To update the assets, delete them and
      <code>app/delete_me_to_restore</code>, then clear your browser cache and visit the CGI, e.g.
      the <a href="../search/?q=foo">search</a>.
      <br class="br"/>
      <code>$ ssh <kbd>myserver.example.com</kbd><br class="br"/>
$ cd <kbd>filesystem/path/to/</kbd><br class="br"/>
$ curl -R -L -o shaarligo.cgi.gz <a href="http://purl.mro.name/shaarligo_cgi.gz">http://purl.mro.name/shaarligo_cgi.gz</a> &amp;&amp; gunzip shaarligo.cgi.gz<br class="br"/>
$ chmod a+x shaarligo.cgi<br class="br"/>
$ ls -l shaarligo?cgi*<br class="br"/>
$ rm -rf .htaccess assets app/delete_me_to_restore</code>
    </li>

    <li id="config"><a href="../config/">Config</a></li>

    <li>
      <form class="form-inline" name="tag_rename">
        <div class="form-group">
          <label for="tag_rename_old">Rename Tag:</label>
          <input type="text" class="form-control" id="tag_rename_old" placeholder="#before" value="{{ .tag_rename_old }}"/>
        </div>
        <div class="form-group">
          <label for="tag_rename_new" class="sr-only">To:</label>
          <input type="text" class="form-control" id="tag_rename_new" placeholder="#after" value="{{ .tag_rename_new }}"/>
        </div>
        <button type="submit" class="btn btn-primary">Rename</button>
      </form>    
    </li>

    <li>
      <form class="form-inline" name="shaarli_import" method="post">
        <div class="form-group">
          <label for="shaarli_import_url">Import Other Shaarli:</label>
          <input type="url" class="form-control" name="shaarli_import_url" placeholder="https://demo.shaarli.org/?" value="{{ .other_shaarli_url }}"/>
        </div>
        <div class="form-group">
          <label for="shaarli_import_tag" class="sr-only">#MarkerForThisImport</label>
          <input type="text" class="form-control" name="shaarli_import_tag" placeholder="#MarkerTagForThisImport" value="#{{ .other_shaarli_tag }}"/>
        </div>
        <button name="shaarli_import_submit" type="submit" value="shaarli_import_submit" class="btn btn-primary">Import</button>
      </form>    
    </li>

    <li id="bookmarklet">
      <b>Bookmarklet:</b> <a
        onclick="alert('Drag this link to your bookmarks toolbar, or right-click it and choose Bookmark This Link...');return false;"
        href="javascript:javascript:(function(){var%20url%20=%20location.href;var%20title%20=%20document.title%20||%20url;window.open('{{.xml_base}}?post='%20+%20encodeURIComponent(url)+'&amp;title='%20+%20encodeURIComponent(title)+'&amp;description='%20+%20encodeURIComponent(document.getSelection())+'&amp;source=bookmarklet','_blank','menubar=no,height=450,width=600,toolbar=no,scrollbars=no,status=no,dialog=1');})();"
      >✚ShaarliGo 🌺</a>
      <span>⇐ Drag this link to your bookmarks toolbar (or right-click it and choose Bookmark This Link…).
      Then click "✚ShaarliGo 🌺" button in any page you want to share.</span>
    </li>

    <li id="version">
    	<b>Version:</b> <span id="number">v{{.version}}</span>+<span id="gitsha1">{{.gitsha1}}</span>
    </li>
  </ol>
</body>
</html>
`); err == nil {
				w.Header().Set("Content-Type", "text/xml; charset=utf-8")
				io.WriteString(w, xml.Header)
				io.WriteString(w, `<?xml-stylesheet type='text/xsl' href='../../assets/`+app.cfg.Skin+`/tools.xslt'?>
`)
				data := map[string]string{
					"title":             app.cfg.Title,
					"xml_base":          app.cgi.String(),
					"tag_rename_old":    "",
					"tag_rename_new":    "",
					"other_shaarli_url": "",
					"other_shaarli_tag": time.Now().Format(time.RFC3339[:16]),
					"version":           version,
					"gitsha1":           GitSHA1,
				}

				if err := tmpl.Execute(w, data); err != nil {
					http.Error(w, "Coudln't render tools: "+err.Error(), http.StatusInternalServerError)
				}
			}
		case http.MethodPost:
			app.KeepAlive(w, r, now)
			if "" != r.FormValue("shaarli_import_submit") {
				if url, err := url.Parse(strings.TrimSpace(r.FormValue("shaarli_import_url")) + "?do=atom&nb=all"); err != nil {
					http.Error(w, "Coudln't parse shaarli_import_url "+err.Error(), http.StatusBadRequest)
				} else {
					if rq, err := HttpGetBody(url, timeoutShaarliImportFetch); err != nil {
						http.Error(w, "Coudln't fetch shaarli_import_url "+err.Error(), http.StatusBadRequest)
					} else {
						if importedFeed, err := FeedFromReader(rq); err != nil {
							http.Error(w, "Coudln't parse feed from shaarli_import_url "+err.Error(), http.StatusBadRequest)
						} else {
							log.Printf("Import %d entries from %v\n", len(importedFeed.Entries), url)
							cat := Category{Term: strings.TrimSpace(strings.TrimPrefix(r.FormValue("shaarli_import_tag"), "#"))}
							feed, _ := app.LoadFeed()
							feed.XmlBase = Iri(app.url.String())
							// feed.Id = feed.XmlBase
							impEnt := make([]*Entry, 0, len(importedFeed.Entries))
							for _, entry := range importedFeed.Entries {
								if et, err := entry.NormaliseAfterImport(); err != nil {
									log.Printf("Error with %v: %v\n", entry.Id, err.Error())
								} else {
									// log.Printf("done entry: %s\n", et.Id)
									if "" != cat.Term {
										et.Categories = append(et.Categories, cat)
									}
									if _, err := feed.Append(&et); err == nil {
										impEnt = append(impEnt, &et)
									} else {
										log.Printf("couldn't add entry: %s\n", err.Error())
									}
								}
							}
							if err := app.SaveFeed(feed); err != nil {
								http.Error(w, "couldn't store feed data: "+err.Error(), http.StatusInternalServerError)
								return
							}
							if err := app.PublishFeedsForModifiedEntries(feed, feed.Entries); err != nil {
								log.Println("couldn't write feeds: ", err.Error())
								http.Error(w, "couldn't write feeds: "+err.Error(), http.StatusInternalServerError)
								return
							}
						}
					}
				}
			}
			http.Redirect(w, r, "../..", http.StatusFound)
		}
	}
}

func (entry Entry) NormaliseAfterImport() (Entry, error) {
	// log.Printf("process entry: %s\n", entry.Id)
	// normalise Id
	if idx := strings.Index(string(entry.Id), "?"); idx >= 0 {
		entry.Id = entry.Id[idx+1:]
	}
	if entry.Published.IsZero() {
		entry.Published = entry.Updated
	}
	// normalise Links
	if nil != entry.Content {
		entry.Content = &HumanText{Body: cleanLegacyContent(entry.Content.Body)}
	}
	err := entry.Validate()
	return entry, err
}
