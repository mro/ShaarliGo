//
// Copyright (C) 2017-2017 Marcus Rohrmoser, http://purl.mro.name/AtomicShaarli
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
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func mustParseURL(u string) *url.URL {
	if ret, err := url.Parse(u); err != nil {
		panic("Cannot parse URL '" + u + "' " + err.Error())
	} else {
		return ret
	}
}

const cgiName = "atom.cgi"
const fileName = "index.xml" // could be 'index.atom' but xml may have a proper mimetype out of the box
const dirAssets = "assets"

const themeDefault = "default"
const langDefault = "de"

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

const newDirPerms = os.ModeDir | os.ModePerm

var rexPath = regexp.MustCompile("[^/]+")

type feedWriter interface {
	Write(feedOrEntry interface{}, self *url.URL, xsltFileName string) error
}

// maybe replace the CONTENT of pub rather than pub itself. So . could remain readonly.
//
func (feed *Feed) replaceFeeds() error {
	if path.Join("a", "b") != filepath.Join("a", "b") {
		return errors.New("Go, get an OS.")
	}
	// check race: if .lock exists kill pid?
	if byPid, err := ioutil.ReadFile("./app/var/lock"); err == nil {
		if pid, err := strconv.Atoi(string(byPid)); err == nil {
			if proc, err := os.FindProcess(pid); err == nil {
				err = proc.Kill()
			}
		}
		if err != nil {
			return err
		}
		if err = os.Remove("./app/var/lock"); err != nil {
			return err
		}
	}
	// create .lock file with pid
	var err error
	if err = os.MkdirAll("./app/var", newDirPerms); err == nil {
		if err = ioutil.WriteFile("./app/var/lock", []byte(string(os.Getpid())), os.ModeExclusive); err == nil {
			defer os.Remove("./app/var/lock")
			os.RemoveAll("./app/var/stage")
			os.RemoveAll("./app/var/old")
			if err = os.MkdirAll("./app/var", newDirPerms); err == nil {
				if err = feed.writeFeeds(100, fileFeedWriter{baseDir: "./app/var/stage"}); err == nil {
					if _, err = os.Stat("./pub"); os.IsNotExist(err) {
						err = nil // ignore nonexisting pub dir. That's fine for first launch.
					} else {
						err = os.Rename("./pub", "./app/var/old")
					}
					if err == nil {
						if err = os.Rename("./app/var/stage/pub", "./pub"); err == nil {
							os.RemoveAll("./app/var/stage")
							if err = os.RemoveAll("./app/var/old"); err == nil {
								err = os.Remove("./app/var/lock")
							}
						}
					}
				}
			}
		}
	}
	return err
}

func (feed *Feed) writeFeeds(entriesPerPage int, fw feedWriter) error {
	xmlBase := mustParseURL(feed.XmlBase)
	if !xmlBase.IsAbs() || !strings.HasSuffix(xmlBase.Path, "/") {
		return errors.New("feed/@xml:base must be set to an absolute URL with a trailing slash")
	}
	// load template feed, set Id and birthday.
	// we need to know the total pages per each feed in order to know the 'last' uri.
	// So no concurrency here :-(
	catScheme := xmlBase.ResolveReference(mustParseURL(path.Join(uriPub, uriTags))).String() + "/"
	uri2entries := make(map[string][]*Entry)
	for _, item := range feed.Entries {
		item.XmlBase = xmlBase.String()
		// change entries for output but don't save the change:
		selfURL := mustParseURL(path.Join(uriPub, uriPosts, item.Id))
		editURL := path.Join(cgiName, selfURL.String())
		item.Id = xmlBase.ResolveReference(selfURL).String() // expand XmlBase as required by https://validator.w3.org/feed/check.cgi?url=
		item.Links = append(item.Links,
			Link{Rel: relSelf, Href: selfURL.String()},
			Link{Rel: relEdit, Href: editURL},
			// Link{Rel: relEditMedia, Href: editURL},
			Link{Rel: relUp, Href: "..", Title: feed.Title.Body}, // we need the feed-name somewhere.
		)
		for i, _ := range item.Categories {
			item.Categories[i].Scheme = catScheme
		}
		for _, uri := range feedUrlsForEntry(item) {
			uri2entries[uri] = append(uri2entries[uri], item)
		}

		fw.Write(item, selfURL, "posts.xslt")
	}

	for uri, entries := range uri2entries {
		// ... if not for the error handling. Maybe https://godoc.org/golang.org/x/sync/errgroup
		feed.writeFeed(uri, entries, entriesPerPage, fw)
	}

	return nil
}

// where does the post have to go.
//
// The results have to be turned into (relative) feed uris down at writeEntries()
// and may well be a int index into a LUT
func feedUrlsForEntry(itm *Entry) []string {
	day := itm.Published
	if day == nil {
		day = &itm.Updated
	}
	ret := make([]string, 0, 2+len(itm.Categories))

	audience := uriPub
	ret = append(ret,
		path.Join(audience, uriPosts),                          // default feed
		path.Join(audience, uriDays, day.Format("2006-01-02")), // daily feed
	)
	for _, cat := range itm.Categories {
		ret = append(ret, path.Join(audience, uriTags, cat.Term)) // category feeds
	}

	return ret
}

func appendPageNumber(prefix string, page int) string {
	if page == 0 {
		return prefix
	}
	return fmt.Sprintf("%s-%d", prefix, page)
}

func computeLastPage(count int, entriesPerPage int) int {
	if count == 0 {
		return 0
	}
	return (count - 1) / entriesPerPage
}

// seed is cloned on each call
func (seed Feed) writeFeed(uri string, entries []*Entry, entriesPerPage int, fw feedWriter) error {
	totalEntries := len(entries)
	lastPage := computeLastPage(totalEntries, entriesPerPage)

	if totalEntries > 0 {
		seed.Updated = entries[0].Updated // most recent sets the date
	}

	switch {
	case strings.HasPrefix(uri, path.Join(uriPub, uriPosts)):
		// leave as is
	case strings.HasPrefix(uri, path.Join(uriPub, uriTags)):
		seed.Subtitle = &HumanText{Body: "#" + uri[len(uriPub)+len(uriTags)+2:]}
	case strings.HasPrefix(uri, path.Join(uriPub, uriDays)):
		seed.Subtitle = &HumanText{Body: uri[len(uriPub)+len(uriDays)+2:]}
	default:
		seed.Subtitle = &HumanText{Body: "Oha"}
	}

	for page := 0; page <= lastPage; page++ {
		lower := page * entriesPerPage
		upper := lower + entriesPerPage
		if upper > totalEntries {
			upper = totalEntries
		}
		seed.Entries = entries[lower:upper]

		if err := seed.writePage(uri, page, lastPage, fw); err != nil {
			return err
		}
	}
	return nil
}

// feed is cloned on each call
func (feed Feed) writePage(uri string, page, lastPage int, fw feedWriter) error {
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
	return fw.Write(&feed, mustParseURL(pagedUri), "posts.xslt")
}

type fileFeedWriter struct {
	baseDir string
}

func (ffw fileFeedWriter) Write(feedOrEntry interface{}, self *url.URL, xsltFileName string) error {
	uri := self.Path
	pathPrefix := rexPath.ReplaceAllString(uri, "..")
	xslt := path.Join(pathPrefix, dirAssets, themeDefault, langDefault, xsltFileName)
	dstDirName := filepath.Join(ffw.baseDir, filepath.FromSlash(uri))
	dstFileName := filepath.Join(dstDirName, fileName)
	var err error
	if err = os.MkdirAll(dstDirName, newDirPerms); err == nil {
		var w *os.File
		if w, err = os.Create(dstFileName); err == nil {
			defer w.Close() // just to be sure
			enc := xml.NewEncoder(w)
			enc.Indent("", "  ")
			if err = xmlEncodeWithXslt(feedOrEntry, xslt, enc); err == nil {
				if err = enc.Flush(); err == nil {
					return w.Close()
					// TODO: set timestamp (Updated)
				}
			}
		}
	}
	return err
}