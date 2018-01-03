//
// Copyright (C) 2018-2018 Marcus Rohrmoser, http://purl.mro.name/ShaarliGo
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
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

func rankEntryTerms(entry *Entry, terms []string) int {
	// defer un(trace("ranker"))
	var parts = [2]string{entry.Content.Body, entry.Title.Body}
	rank := 0
	for _, term := range terms {
		for weight, txt := range parts {
			// better: https://stackoverflow.com/questions/24836044/case-insensitive-string-search-in-golang
			if strings.Contains(txt, term) {
				rank += 1 + weight
			}
		}
	}
	return rank
}

func (app *App) handleSearch(w http.ResponseWriter, r *http.Request) {
	now := time.Now()

	// evtl. check permission to search (non-logged-in visitor?)
	if !app.cfg.IsConfigured() {
		http.Redirect(w, r, cgiName+"/config", http.StatusPreconditionFailed)
		return
	}

	switch r.Method {
	case http.MethodGet:
		app.KeepAlive(w, r, now)
		// pull out parameters q, offset, limit
		query := r.URL.Query()
		if rq := query["q"]; rq != nil && 0 < len(rq) {
			terms := regexp.MustCompile("\\s+").Split(rq[0], -1)
			limit := 50
			offset := 0

			feed, _ := app.LoadFeed()
			feed.XmlBase = xmlBaseFromRequestURL(r.URL, os.Getenv("SCRIPT_NAME")).String()
			feed.Id = feed.XmlBase // expand XmlBase as required by https://validator.w3.org/feed/check.cgi?url=

			ret := feed.Search(func(entry *Entry) int { return rankEntryTerms(entry, terms) }, offset, limit)

			w.Header().Set("Content-Type", "text/xml; charset=utf-8")
			enc := xml.NewEncoder(w)
			enc.Indent("", "  ")
			if err := xmlEncodeWithXslt(ret, "../../assets/default/de/posts.xslt", enc); err == nil {
				if err := enc.Flush(); err == nil {
					return
				}
			}
		}
	}
}

func (feed Feed) Search(ranker func(*Entry) int, offset, limit int) Feed {
	defer un(trace("Feed.Search"))
	feed.Entries = searchEntries(feed.Entries, ranker)
	// paging: adjust links
	return feed
}

type search_results struct {
	Ranks   []int
	Entries []*Entry
}

func (r search_results) Len() int { return len(r.Ranks) }
func (r search_results) Less(i, j int) bool {
	if r.Ranks[i] == r.Ranks[j] {
		return ByPublishedDesc(r.Entries).Less(i, j)
	}
	return r.Ranks[i] > r.Ranks[j]
}
func (r search_results) Swap(i, j int) {
	r.Ranks[i], r.Ranks[j] = r.Ranks[j], r.Ranks[i]
	r.Entries[i], r.Entries[j] = r.Entries[j], r.Entries[i]
}

func searchEntries(entries []*Entry, ranker func(*Entry) int) []*Entry {
	defer un(trace("searchEntries"))
	r := search_results{
		Ranks:   make([]int, len(entries)),
		Entries: entries,
	}
	// could be concurrent:
	for idx, ent := range entries {
		r.Ranks[idx] = ranker(ent)
	}
	// sort entries according to rank
	sort.Sort(r)
	cut := sort.Search(len(r.Ranks), func(idx int) bool { return r.Ranks[idx] <= 0 })
	return r.Entries[0:cut]
}

//
