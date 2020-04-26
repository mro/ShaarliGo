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

			byt, _ := tplToolsHtmlBytes()
			if tmpl, err := template.New("tools").Parse(string(byt)); err == nil {
				w.Header().Set("Content-Type", "text/xml; charset=utf-8")
				io.WriteString(w, xml.Header)
				io.WriteString(w, `<?xml-stylesheet type='text/xsl' href='../../themes/current/tools.xslt'?>
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
							feed, _ := LoadFeed()
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
