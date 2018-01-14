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
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

func mustParseURL(u string) *url.URL {
	if ret, err := url.Parse(u); err != nil {
		panic("Cannot parse URL '" + u + "' " + err.Error())
	} else {
		return ret
	}
}

const cgiName = "shaarligo.cgi"
const dirAssets = "assets"
const dirApp = "app"

const uriPub = "pub"
const uriPosts = "posts"
const uriDays = "days"
const uriTags = "tags"

const relSelf = "self"            // https://www.iana.org/assignments/link-relations/link-relations.xhtml
const relAlternate = "alternate"  // https://www.iana.org/assignments/link-relations/link-relations.xhtml
const relVia = "via"              // Atom https://tools.ietf.org/html/rfc4287
const relEnclosure = "enclosure"  // Atom https://tools.ietf.org/html/rfc4287
const relFirst = "first"          // paged feeds https://tools.ietf.org/html/rfc5005#section-3
const relLast = "last"            // paged feeds https://tools.ietf.org/html/rfc5005#section-3
const relNext = "next"            // paged feeds https://tools.ietf.org/html/rfc5005#section-3
const relPrevious = "previous"    // paged feeds https://tools.ietf.org/html/rfc5005#section-3
const relEdit = "edit"            // AtomPub https://tools.ietf.org/html/rfc5023
const relEditMedia = "edit-media" // AtomPub https://tools.ietf.org/html/rfc5023
const relUp = "up"                // https://www.iana.org/assignments/link-relations/link-relations.xhtml
const relSearch = "search"        // http://www.opensearch.org/Specifications/OpenSearch/1.1#Autodiscovery_in_RSS.2FAtom

const newDirPerms = 0775

var rexPath = regexp.MustCompile("[^/]+")

const uriPubPosts = uriPub + "/" + uriPosts + "/"
const uriPubTags = uriPub + "/" + uriTags + "/"

func addEntryFilter(m map[string][]func(*Entry) bool, k string, f func(*Entry) bool) map[string][]func(*Entry) bool {
	m[k] = append(m[k], f)
	return m
}

func (entry Entry) FeedFilters(uri2filter map[string][]func(*Entry) bool) map[string][]func(*Entry) bool {
	defer un(trace("Entry.FeedFilters " + entry.Id))
	if nil == uri2filter {
		uri2filter = make(map[string][]func(*Entry) bool, 10)
	}
	a := func(k string, f func(*Entry) bool) {
		addEntryFilter(uri2filter, k, f)
	}

	a(uriPubPosts, func(*Entry) bool { return true })
	a(uriPubPosts+entry.Id+"/", func(iEntry *Entry) bool { return entry.Id == iEntry.Id })

	a(uriPubTags, func(*Entry) bool { return false }) // dummy to get an (empty) feed
	for _, cat := range entry.Categories {
		a(uriPubTags+cat.Term+"/", func(iEntry *Entry) bool {
			for _, iCat := range iEntry.Categories {
				if cat.Term == iCat.Term && cat.Scheme == iCat.Scheme {
					return true
				}
			}
			return false
		})
	}

	// a("pub/days/", func(*Entry) bool { return false })
	dayStr := entry.Published.Format(time.RFC3339[:10])
	a("pub/days/"+dayStr+"/", func(iEntry *Entry) bool {
		return dayStr == iEntry.Published.Format(time.RFC3339[:10])
	})

	return uri2filter
}

func LinkRelSelf(links []Link) Link {
	for _, l := range links {
		if relSelf == l.Rel {
			return l
		}
	}
	return Link{}
}

// collect all entries into all (unpaged, complete) feeds to publish
func (seed Feed) CompleteFeeds(uri2filter map[string][]func(*Entry) bool) []Feed {
	defer un(trace("Feed.CompleteFeeds"))
	ret := make([]Feed, 0, len(uri2filter))
	for uri, entryFilters := range uri2filter {
		feed := seed // clone
		feed.Id = uri
		feed.Entries = nil // save reallocs?
		for _, entry := range seed.Entries {
			for _, entryFilter := range entryFilters {
				if entryFilter(entry) {
					feed.Entries = append(feed.Entries, entry)
					break
				}
			}
		}
		if 0 < len(feed.Entries) {
			feed.Updated = feed.Entries[0].Updated // that's just a guess.
		}
		if uriPubTags == uri {
			feed.Categories = AggregateCategories(seed.Entries) // rather the ones from pub/posts
		}
		ret = append(ret, feed)
	}
	return ret
}

func appendPageNumber(prefix string, page int) string {
	if !strings.HasSuffix(prefix, "/") {
		panic("invalid input: appendPageNumber('" + prefix + "', " + string(page) + ") needs a trailing slash")
	}
	if page == 0 {
		return prefix
	}
	return fmt.Sprintf("%s-%d/", prefix[:len(prefix)-1], page)
}

func computeLastPage(count int, entriesPerPage int) int {
	if count == 0 {
		return 0
	}
	return (count - 1) / entriesPerPage
}

func (seed Feed) Pages(entriesPerPage int) []Feed {
	defer un(trace("Feed.Pages " + seed.Id))
	entriesPerPage = max(1, entriesPerPage)
	totalEntries := len(seed.Entries)
	lastPage := computeLastPage(totalEntries, entriesPerPage)
	ret := make([]Feed, 0, 1+lastPage)
	uri := seed.Id

	for page := 0; page <= lastPage; page++ {
		feed := seed
		{
			lower := page * entriesPerPage
			upper := lower + entriesPerPage
			if upper > totalEntries {
				upper = totalEntries
			}
			feed.Entries = seed.Entries[lower:upper]
		}
		pagedUri := appendPageNumber(uri, page)
		feed.Links = append(feed.Links, Link{Rel: relSelf, Href: pagedUri, Title: fmt.Sprintf("%d", page+1)})
		// https://tools.ietf.org/html/rfc5005#section-3
		if lastPage > 0 {
			feed.Links = append(feed.Links, Link{Rel: relFirst, Href: appendPageNumber(uri, 0), Title: fmt.Sprintf("%d", 0+1)})
			if page > 0 {
				feed.Links = append(feed.Links, Link{Rel: relPrevious, Href: appendPageNumber(uri, page-1), Title: fmt.Sprintf("%d", page-1+1)})
			}
			if page < lastPage {
				feed.Links = append(feed.Links, Link{Rel: relNext, Href: appendPageNumber(uri, page+1), Title: fmt.Sprintf("%d", page+1+1)})
			}
			feed.Links = append(feed.Links, Link{Rel: relLast, Href: appendPageNumber(uri, lastPage), Title: fmt.Sprintf("%d", lastPage+1)})
		} else {
			// TODO https://tools.ietf.org/html/rfc5005#section-2
			// xmlns:fh="http://purl.org/syndication/history/1.0" <fh:complete/>
		}
		ret = append(ret, feed)
	}
	return ret
}

func (feed Feed) CompleteFeedsForModifiedEntries(entries []*Entry) []Feed {
	defer un(trace("Feed.CompleteFeedsForModifiedEntries"))
	var uri2filter map[string][]func(*Entry) bool
	for _, entry := range entries {
		uri2filter = entry.FeedFilters(uri2filter)
	}

	return feed.CompleteFeeds(uri2filter)
}

func (feed Feed) PagedFeeds(complete []Feed, linksPerPage int) ([]Feed, error) {
	defer un(trace("Feed.PagedFeeds"))
	xmlBase := mustParseURL(feed.XmlBase)
	if !xmlBase.IsAbs() || !strings.HasSuffix(xmlBase.Path, "/") {
		log.Printf("xml:base is '%s'\n", xmlBase)
		return []Feed{}, errors.New("feed/@xml:base must be set to an absolute URL with a trailing slash")
	}

	pages := make([]Feed, 0, 2*len(complete))
	for _, complete := range complete {
		pages = append(pages, complete.Pages(linksPerPage)...)
	}

	// do before writing but after all matching is done:
	catScheme := xmlBase.ResolveReference(mustParseURL(path.Join(uriPub, uriTags))).String() + "/"
	for _, entry := range feed.Entries {
		entry.XmlBase = xmlBase.String()
		if entry.Updated.IsZero() {
			entry.Updated = entry.Published
		}
		// change entries for output but don't save the change:
		upURL := mustParseURL(path.Join(uriPub, uriPosts) + "/")
		selfURL := mustParseURL(path.Join(uriPub, uriPosts, entry.Id) + "/")
		editURL := strings.Join([]string{cgiName, "?post=", selfURL.String()}, "")
		entry.Id = xmlBase.ResolveReference(selfURL).String() // expand XmlBase as required by https://validator.w3.org/feed/check.cgi?url=
		entry.Links = append(entry.Links,
			Link{Rel: relSelf, Href: selfURL.String()},
			Link{Rel: relEdit, Href: editURL},
			// Link{Rel: relEditMedia, Href: editURL},
			Link{Rel: relUp, Href: upURL.String(), Title: feed.Title.Body}, // we need the feed-name somewhere.
		)
		for i, _ := range entry.Categories {
			entry.Categories[i].Scheme = catScheme
		}
	}

	return pages, nil
}

func (app App) PublishFeedsForModifiedEntries(feed Feed, entries []*Entry) error {
	defer un(trace("App.PublishFeedsForModifiedEntries"))

	sort.Sort(ByPublishedDesc(feed.Entries))

	complete := feed.CompleteFeedsForModifiedEntries(entries)
	if pages, err := feed.PagedFeeds(complete, app.cfg.LinksPerPage); err == nil {
		return app.PublishFeeds(pages)
	} else {
		return err
	}
}

// create a lock file to avoid races and then call PublishFeed in loop
func (app App) PublishFeeds(feeds []Feed) error {
	defer un(trace("App.PublishFeeds"))
	strFileLock := filepath.Join("app", "var", "lock")

	// check race: if .lock exists kill pid?
	if byPid, err := ioutil.ReadFile(strFileLock); err == nil {
		if pid, err := strconv.Atoi(string(byPid)); err == nil {
			if proc, err := os.FindProcess(pid); err == nil {
				err = proc.Kill()
			}
		}
		if err != nil {
			return err
		}
		if err = os.Remove(strFileLock); err != nil {
			return err
		}
	}

	// create .lock file with pid
	if err := ioutil.WriteFile(strFileLock, []byte(string(os.Getpid())), os.ModeExclusive); err == nil {
		defer os.Remove(strFileLock)
		for _, feed := range feeds {
			if err := app.PublishFeed(feed); err != nil {
				return err
			}
			if uriPubTags == LinkRelSelf(feed.Links).Href {
				// write additional index.json with all (public) category terms
				const jsonFileName = "index.json"
				tags := make([]string, 0, len(feed.Categories))
				for _, cat := range feed.Categories {
					tags = append(tags, "#"+cat.Term)
				}
				dstDirName := filepath.FromSlash(uriPubTags)
				dstFileName := filepath.Join(dstDirName, jsonFileName)
				tmpFileName := dstFileName + "~"
				var w *os.File
				if w, err = os.Create(tmpFileName); err == nil {
					defer w.Close() // just to be sure
					enc := json.NewEncoder(w)
					if err = enc.Encode(tags); err == nil {
						if err = w.Close(); err == nil {
							if err := os.Rename(tmpFileName, dstFileName); err != nil {
								return err
							}
						}
					}
				}
			}
		}
	}

	return nil
}

func (app App) PublishFeed(feed Feed) error {
	const feedFileName = "index.xml"
	const xsltFileName = "posts.xslt"
	defer un(trace(strings.Join([]string{"App.PublishFeed", feed.Id}, " ")))

	uri := LinkRelSelf(feed.Links).Href
	pathPrefix := rexPath.ReplaceAllString(uri, "..")
	dstDirName := filepath.FromSlash(uri)
	dstFileName := filepath.Join(dstDirName, feedFileName)

	remove := ((1 == len(feed.Entries) && feed.Entries[0].Published.IsZero()) ||
		0 == len(feed.Entries)) &&
		"../../../" == pathPrefix
	if remove {
		log.Printf("remove %s", dstFileName)
		err := os.Remove(dstFileName)
		os.Remove(dstDirName)
		return err
	}

	if feed.Updated.IsZero() {
		log.Println("repairing Feed.Updated time")
		if 0 < len(feed.Entries) {
			feed.Updated = feed.Entries[0].Updated
		} else {
			feed.Updated = iso8601{time.Now()}
		}
	}

	var feedOrEntry interface{} = feed
	if "../../../" == pathPrefix && strings.HasPrefix(uri, uriPubPosts) {
		if 1 != len(feed.Entries) {
			return errors.New("Invalid feed")
		}
		feedOrEntry = feed.Entries[0]
	}

	log.Printf("write %s", dstFileName)
	tmpFileName := dstFileName + "~"
	xslt := path.Join(pathPrefix, dirAssets, app.cfg.Skin, xsltFileName)

	var err error
	if err = os.MkdirAll(dstDirName, newDirPerms); err == nil {
		var w *os.File
		if w, err = os.Create(tmpFileName); err == nil {
			defer w.Close() // just to be sure
			enc := xml.NewEncoder(w)
			enc.Indent("", "  ")
			if err = xmlEncodeWithXslt(feedOrEntry, xslt, enc); err == nil {
				if err = enc.Flush(); err == nil {
					if err = w.Close(); err == nil {
						mTime := feed.Updated.Time
						os.Chtimes(tmpFileName, mTime, mTime)
						return os.Rename(tmpFileName, dstFileName)
					}
				}
			}
		}
	}
	return err
}
