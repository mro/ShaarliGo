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
	"compress/gzip"
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

const uriPub = "o"
const uriPosts = "p"
const uriDays = "d"
const uriTags = "t"

const relSelf = Relation("self")            // https://www.iana.org/assignments/link-relations/link-relations.xhtml
const relAlternate = Relation("alternate")  // https://www.iana.org/assignments/link-relations/link-relations.xhtml
const relVia = Relation("via")              // Atom https://tools.ietf.org/html/rfc4287
const relEnclosure = Relation("enclosure")  // Atom https://tools.ietf.org/html/rfc4287
const relFirst = Relation("first")          // paged feeds https://tools.ietf.org/html/rfc5005#section-3
const relLast = Relation("last")            // paged feeds https://tools.ietf.org/html/rfc5005#section-3
const relNext = Relation("next")            // paged feeds https://tools.ietf.org/html/rfc5005#section-3
const relPrevious = Relation("previous")    // paged feeds https://tools.ietf.org/html/rfc5005#section-3
const relEdit = Relation("edit")            // AtomPub https://tools.ietf.org/html/rfc5023
const relEditMedia = Relation("edit-media") // AtomPub https://tools.ietf.org/html/rfc5023
const relUp = Relation("up")                // https://www.iana.org/assignments/link-relations/link-relations.xhtml
const relSearch = Relation("search")        // http://www.opensearch.org/Specifications/OpenSearch/1.1#Autodiscovery_in_RSS.2FAtom

const newDirPerms = 0775

var rexPath = regexp.MustCompile("[^/]+")

const uriPubPosts = uriPub + "/" + uriPosts + "/"
const uriPubTags = uriPub + "/" + uriTags + "/"
const uriPubDays = uriPub + "/" + uriDays + "/"

func uri2subtitle(subtitle *HumanText, uri string) *HumanText {
	if strings.HasPrefix(uri, uriPubTags) {
		return &HumanText{Body: "#" + strings.TrimRight(uri[len(uriPubTags):], "/")}
	}
	if strings.HasPrefix(uri, uriPubDays) {
		return &HumanText{Body: "📅 " + strings.TrimRight(uri[len(uriPubDays):], "/")}
	}
	return subtitle
}

func (entry Entry) FeedFilters(uri2filter map[string]func(*Entry) bool) map[string]func(*Entry) bool {
	// defer un(trace("Entry.FeedFilters " + entry.Id))
	if nil == uri2filter {
		uri2filter = make(map[string]func(*Entry) bool, 10)
	}

	uri2filter[uriPubPosts] = func(*Entry) bool { return true }
	uri2filter[uriPubPosts+string(entry.Id)+"/"] = func(iEntry *Entry) bool { return entry.Id == iEntry.Id }

	uri2filter[uriPubTags] = func(*Entry) bool { return false } // dummy to get an (empty) feed
	for _, cat := range entry.Categories {
		trm := cat.Term
		uri2filter[uriPubTags+trm+"/"] = func(iEntry *Entry) bool {
			for _, iCat := range iEntry.Categories {
				if trm == iCat.Term { // && cat.Scheme == iCat.Scheme {
					return true
				}
			}
			return false
		}
	}

	// uri2filter["pub/days/", func(*Entry) bool { return false })
	dayStr := entry.Published.Format(time.RFC3339[:10])
	uri2filter[uriPubDays+dayStr+"/"] = func(iEntry *Entry) bool {
		return dayStr == iEntry.Published.Format(time.RFC3339[:10])
	}

	return uri2filter
}

func LinkRel(rel Relation, links []Link) Link {
	for _, l := range links {
		for _, r := range strings.Fields(string(l.Rel)) { // may be worth caching
			if rel == Relation(r) {
				return l
			}
		}
	}
	return Link{}
}

func LinkRelSelf(links []Link) Link {
	return LinkRel(relSelf, links)
}

func uriSliceSorted(uri2filter map[string]func(*Entry) bool) []string {
	keys := make([]string, len(uri2filter))
	{
		i := 0
		for k := range uri2filter {
			keys[i] = k
			i++
		}
	}
	sort.Strings(keys) // I don't care too much how they're sorted, I just want them to be stable.
	return keys
}

// collect all entries into all (unpaged, complete) feeds to publish.
//
// return sorted by Id
func (seed Feed) CompleteFeeds(uri2filter map[string]func(*Entry) bool) []Feed {
	defer un(trace("Feed.CompleteFeeds"))
	ret := make([]Feed, 0, len(uri2filter))
	for _, uri := range uriSliceSorted(uri2filter) {
		entryFilter := uri2filter[uri]
		feed := seed // clone
		feed.Id = Id(uri)
		feed.Subtitle = uri2subtitle(feed.Subtitle, uri)
		feed.Entries = nil // save reallocs?
		for _, entry := range seed.Entries {
			if entryFilter(entry) {
				feed.Entries = append(feed.Entries, entry)
			}
		}
		if uriPubTags == uri {
			feed.Categories = AggregateCategories(seed.Entries) // rather the ones from o/p
		}
		ret = append(ret, feed)
	}
	return ret
}

func appendPageNumber(prefix string, page, pageCount int) string {
	if !strings.HasSuffix(prefix, "/") {
		panic("invalid input: appendPageNumber('" + prefix + "', " + string(page) + ") needs a trailing slash")
	}
	if page == pageCount-1 {
		return prefix
	}
	return fmt.Sprintf("%s"+"-"+"%d"+"/", prefix[:len(prefix)-1], page)
}

func computePageCount(count int, entriesPerPage int) int {
	if count == 0 {
		// even 0 entries need one (empty) page
		return 1
	}
	return 1 + (count-1)/entriesPerPage
}

func (seed Feed) Pages(entriesPerPage int) []Feed {
	// defer un(trace("Feed.Pages " + seed.Id))
	entriesPerPage = max(1, entriesPerPage)
	totalEntries := len(seed.Entries)
	pageCount := computePageCount(totalEntries, entriesPerPage)
	ret := make([]Feed, 0, pageCount)
	uri := string(seed.Id)

	link := func(rel Relation, page int) Link {
		return Link{Rel: rel, Href: appendPageNumber(uri, page, pageCount), Title: strconv.Itoa(page + 1)}
	}

	lower := totalEntries // start past the oldest entry
	for page := 0; page < pageCount; page++ {
		feed := seed

		{
			upper := lower
			step := entriesPerPage
			if page == pageCount-2 {
				// only the page BEFORE the last one has variable length (if needed)
				step = totalEntries % entriesPerPage
				if 0 == step {
					step = entriesPerPage
				}
			}
			lower = max(0, upper-step)
			feed.Entries = seed.Entries[lower:upper]
			feed.Updated = iso8601(time.Time{}) // start with zero
			for _, ent := range feed.Entries {  // max of entries
				if feed.Updated.Before(ent.Updated) {
					feed.Updated = ent.Updated
				}
			}
		}
		ls := append(make([]Link, 0, len(feed.Links)+5), feed.Links...)
		ls = append(ls, link(relSelf, page))
		// https://tools.ietf.org/html/rfc5005#section-3
		if pageCount > 1 {
			ls = append(ls, link(relLast, 0)) // oldest, i.e. lowest page number
			if page > 0 {
				ls = append(ls, link(relNext, page-1)) // older, i.e. smaller page number
			}
			if page < pageCount-1 {
				ls = append(ls, link(relPrevious, page+1)) // newer, i.e. higher page number
			}
			ls = append(ls, link(relFirst, pageCount-1)) // newest, i.e. largest page number
		} else {
			// TODO https://tools.ietf.org/html/rfc5005#section-2
			// xmlns:fh="http://purl.org/syndication/history/1.0" <fh:complete/>
		}
		feed.Links = ls
		ret = append(ret, feed)
	}
	return ret
}

func (feed Feed) CompleteFeedsForModifiedEntries(entries []*Entry) []Feed {
	// defer un(trace("Feed.CompleteFeedsForModifiedEntries"))
	var uri2filter map[string]func(*Entry) bool
	for _, entry := range entries {
		uri2filter = entry.FeedFilters(uri2filter)
	}

	if feed.Updated.IsZero() {
		feed.Updated = func() iso8601 {
			if len(feed.Entries) > 0 {
				ent := feed.Entries[0]
				if !ent.Updated.IsZero() {
					return ent.Updated
				}
				if !ent.Published.IsZero() {
					return ent.Published
				}
			}
			return iso8601(time.Now())
		}()
	}

	return feed.CompleteFeeds(uri2filter)
}

func (feed Feed) PagedFeeds(complete []Feed, linksPerPage int) ([]Feed, error) {
	defer un(trace("Feed.PagedFeeds"))
	xmlBase := mustParseURL(string(feed.XmlBase))
	if !xmlBase.IsAbs() || !strings.HasSuffix(xmlBase.Path, "/") {
		log.Printf("xml:base is '%s'\n", xmlBase)
		return []Feed{}, errors.New("feed/@xml:base must be set to an absolute URL with a trailing slash")
	}

	pages := make([]Feed, 0, 2*len(complete))
	for _, comp := range complete {
		pages = append(pages, comp.Pages(linksPerPage)...)
	}

	// do before writing but after all matching is done:
	catScheme := Iri(xmlBase.ResolveReference(mustParseURL(path.Join(uriPub, uriTags))).String() + "/")
	for _, entry := range feed.Entries {
		entry.XmlBase = Iri(xmlBase.String())
		if entry.Updated.IsZero() {
			entry.Updated = entry.Published
		}
		// change entries for output but don't save the change:
		upURL := mustParseURL(path.Join(uriPub, uriPosts) + "/")
		selfURL := mustParseURL(path.Join(uriPub, uriPosts, string(entry.Id)) + "/")
		editURL := strings.Join([]string{cgiName, "?post=", selfURL.String()}, "")
		entry.Id = Id(xmlBase.ResolveReference(selfURL).String()) // expand XmlBase as required by https://validator.w3.org/feed/check.cgi?url=
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

func (app Server) PublishFeedsForModifiedEntries(feed Feed, entries []*Entry) error {
	defer un(trace("App.PublishFeedsForModifiedEntries"))

	feed.Generator = &Generator{Uri: myselfNamespace, Version: version + "+" + GitSHA1, Body: "🌺 ShaarliGo"}
	sort.Sort(ByPublishedDesc(feed.Entries))
	// entries = feed.Entries // force write all entries. Every single one.
	complete := feed.CompleteFeedsForModifiedEntries(entries)
	if pages, err := feed.PagedFeeds(complete, app.cfg.LinksPerPage); err == nil {
		if err = app.PublishFeeds(pages, true); err != nil {
			return err
		} else {
			// just assure ALL entries index.xml.gz exist and are up to date
			for _, ent := range feed.Entries {
				if err = app.PublishEntry(ent, false); err != nil { // only if newer
					return err
				}
			}
			return nil
		}
	} else {
		return err
	}
}

// create a lock file to avoid races and then call PublishFeed in loop
func (app Server) PublishFeeds(feeds []Feed, force bool) error {
	defer un(trace("App.PublishFeeds"))
	strFileLock := filepath.Join(dirApp, "var", "lock")

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
			if err := app.PublishFeed(feed, force); err != nil {
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

func (app Server) PublishFeed(feed Feed, force bool) error {
	const feedFileName = "index.xml.gz"
	const xsltFileName = "posts.xslt"
	uri := LinkRelSelf(feed.Links).Href
	ti, to := trace(strings.Join([]string{"App.PublishFeed", uri}, " "))

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
		defer un(ti, to)
		return err
	}

	feed.Id = Id(string(feed.XmlBase) + string(feed.Id))
	mTime := time.Time(feed.Updated)
	var feedOrEntry interface{} = feed
	if "../../../" == pathPrefix && strings.HasPrefix(uri, uriPubPosts) {
		if 0 == len(feed.Entries) {
			return fmt.Errorf("Invalid feed, self: %v len(entries): %d", uri, len(feed.Entries))
		}
		if 1 < len(feed.Entries) {
			log.Printf("%d entries with Id: %v, keeping just one.", len(feed.Entries), uri)
		}
		ent := feed.Entries[0]
		feedOrEntry = ent
		mTime = time.Time(ent.Updated)
	}

	if fi, err := os.Stat(dstFileName); !force && (fi != nil && !fi.ModTime().Before(mTime)) && !os.IsNotExist(err) {
		// log.Printf("skip %s, still up to date.", dstFileName)
		return err
	}

	defer un(ti, to)
	tmpFileName := dstFileName + "~"
	xslt := path.Join(pathPrefix, dirAssets, app.cfg.Skin, xsltFileName)

	var err error
	if err = os.MkdirAll(dstDirName, newDirPerms); err == nil {
		var gz *os.File
		if gz, err = os.Create(tmpFileName); err == nil {
			defer gz.Close() // just to be sure
			var w *gzip.Writer
			if w, err = gzip.NewWriterLevel(gz, gzip.BestCompression); err == nil {
				defer w.Close() // just to be sure
				enc := xml.NewEncoder(w)
				enc.Indent("", "  ")
				if err = xmlEncodeWithXslt(feedOrEntry, xslt, enc); err == nil {
					if err = enc.Flush(); err == nil {
						if err = w.Close(); err == nil {
							os.Chtimes(tmpFileName, mTime, mTime)
							return os.Rename(tmpFileName, dstFileName)
						}
					}
				}
			}
		}
	}
	return err
}

func (app Server) PublishEntry(ent *Entry, force bool) error {
	const feedFileName = "index.xml.gz"
	const xsltFileName = "posts.xslt"
	uri := LinkRelSelf(ent.Links).Href
	ti, to := trace(strings.Join([]string{"App.PublishEntry", uri}, " "))

	pathPrefix := rexPath.ReplaceAllString(uri, "..")
	dstDirName := filepath.FromSlash(uri)
	dstFileName := filepath.Join(dstDirName, feedFileName)

	var feedOrEntry interface{} = ent
	ent.Id = Id(string(ent.XmlBase) + string(ent.Id))
	mTime := time.Time(ent.Updated)

	if fi, err := os.Stat(dstFileName); !force && (fi != nil && !fi.ModTime().Before(mTime)) && !os.IsNotExist(err) {
		// log.Printf("skip %s, still up to date.", dstFileName)
		return err
	}

	defer un(ti, to)
	tmpFileName := dstFileName + "~"
	xslt := path.Join(pathPrefix, dirAssets, app.cfg.Skin, xsltFileName)

	var err error
	if err = os.MkdirAll(dstDirName, newDirPerms); err == nil {
		var gz *os.File
		if gz, err = os.Create(tmpFileName); err == nil {
			defer gz.Close() // just to be sure
			var w *gzip.Writer
			if w, err = gzip.NewWriterLevel(gz, gzip.BestCompression); err == nil {
				defer w.Close() // just to be sure
				enc := xml.NewEncoder(w)
				enc.Indent("", "  ")
				if err = xmlEncodeWithXslt(feedOrEntry, xslt, enc); err == nil {
					if err = enc.Flush(); err == nil {
						if err = w.Close(); err == nil {
							os.Chtimes(tmpFileName, mTime, mTime)
							return os.Rename(tmpFileName, dstFileName)
						}
					}
				}
			}
		}
	}
	return err
}
