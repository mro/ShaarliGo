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
	"net/url"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/language"
	"golang.org/x/text/search"
)

// better: https://stackoverflow.com/questions/24836044/case-insensitive-string-search-in-golang
func rankEntryTerms(entry *Entry, terms []string, matcher *search.Matcher) int {
	// defer un(trace("ranker"))
	parts := [2]string{"", ""}
	if nil != entry {
		if nil != entry.Content {
			parts[0] = entry.Content.Body
		}
		parts[1] = entry.Title.Body
	}
	rank := 0
	for _, term := range terms {
		if strings.HasPrefix(term, "#") {
			t := term[1:]
			for _, cat := range entry.Categories {
				if idx, _ := matcher.IndexString(t, cat.Term); idx >= 0 {
					rank += 5
				}
			}
		}
		for weight, txt := range parts {
			if idx, _ := matcher.IndexString(txt, term); idx >= 0 {
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
		if q := query["q"]; q != nil && 0 < len(q) {
			terms := strings.Fields(strings.TrimSpace(strings.Join(q, " ")))
			if 0 == len(terms) {
				http.Redirect(w, r, path.Join("..", "..", uriPub, uriPosts)+"/", http.StatusFound)
				return
			}
			limit := max(1, app.cfg.LinksPerPage)
			offset := 0
			if o := query["offset"]; o != nil {
				offset, _ = strconv.Atoi(o[0]) // just ignore conversion errors. 0 is a fine fallback
			}
			qu := cgiName + "/search/" + "?" + "q" + "=" + url.QueryEscape(strings.Join(terms, " "))

			xmlBase := xmlBaseFromRequestURL(r.URL, os.Getenv("SCRIPT_NAME"))
			catScheme := xmlBase.ResolveReference(mustParseURL(path.Join(uriPub, uriTags))).String() + "/"

			feed, _ := app.LoadFeed()
			feed.XmlBase = xmlBase.String()
			feed.Id = xmlBase.ResolveReference(mustParseURL(qu)).String()

			feed.XmlNSShaarliGo = "http://purl.mro.name/ShaarliGo/2018/"
			feed.SearchTerms = strings.Join(q, " ") // rather use http://www.opensearch.org/Specifications/OpenSearch/1.1#Example_of_OpenSearch_response_elements_in_Atom_1.0
			feed.XmlNSOpenSearch = "http://a9.com/-/spec/opensearch/1.1/"

			lang := language.Make("de") // should come from the entry, feed, settings, default (in that order)
			matcher := search.New(lang, search.IgnoreDiacritics, search.IgnoreCase)
			ret := feed.Search(func(entry *Entry) int { return rankEntryTerms(entry, terms, matcher) })

			// paging / RFC5005
			clamp := func(x int) int { return min(len(ret.Entries), x) }
			offset = clamp(max(0, offset))
			count := len(ret.Entries)
			ret.Links = append(ret.Links, Link{Rel: relSelf, Href: qu + "&" + "offset" + "=" + strconv.Itoa(offset), Title: strconv.Itoa(1 + offset/limit)})
			if count > limit {
				ret.Links = append(ret.Links, Link{Rel: relFirst, Href: qu, Title: strconv.Itoa(1 + 0)})
				ret.Links = append(ret.Links, Link{Rel: relLast, Href: qu + "&" + "offset" + "=" + strconv.Itoa(count-(count%limit)), Title: strconv.Itoa(1 + count/limit)})
				if intPrev := offset - limit; intPrev >= 0 {
					ret.Links = append(ret.Links, Link{Rel: relPrevious, Href: qu + "&" + "offset" + "=" + strconv.Itoa(intPrev), Title: strconv.Itoa(1 + intPrev/limit)})
				}
				if intNext := offset + limit; intNext < count {
					ret.Links = append(ret.Links, Link{Rel: relNext, Href: qu + "&" + "offset" + "=" + strconv.Itoa(intNext), Title: strconv.Itoa(1 + intNext/limit)})
				}
				ret.Entries = ret.Entries[offset:clamp(offset+limit)]
			}
			// prepare entries for Atom publication
			for _, item := range ret.Entries {
				// change entries for output but don't save the change:
				selfURL := mustParseURL(path.Join(uriPub, uriPosts, item.Id) + "/")
				editURL := strings.Join([]string{cgiName, "?post=", selfURL.String()}, "")
				item.Id = xmlBase.ResolveReference(selfURL).String() // expand XmlBase as required by https://validator.w3.org/feed/check.cgi?url=
				item.Links = append(item.Links,
					Link{Rel: relSelf, Href: selfURL.String()},
					Link{Rel: relEdit, Href: editURL},
				)
				for i, _ := range item.Categories {
					item.Categories[i].Scheme = catScheme
				}
				if item.Updated.IsZero() {
					item.Updated = item.Published
				}
				if item.Updated.After(ret.Updated.Time) {
					ret.Updated = item.Updated
				}
			}
			if ret.Updated.IsZero() {
				ret.Updated = iso8601{now}
			}

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

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func (feed Feed) Search(ranker func(*Entry) int) Feed {
	defer un(trace("Feed.Search"))
	feed.Entries = searchEntries(feed.Entries, ranker)
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
