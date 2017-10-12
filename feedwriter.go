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
	"fmt"
	"io"
	"os"
)

const uriPub = "pub"
const uriPosts = "posts"
const uriTags = "tags"
const fileName = "index.xml"

func (m *SessionManager) replaceFeeds() error {
	// create stage dir

	var feed *Feed

	base_dir := "/t.b.d./"
	fctWriteCloser := func(uri string, page int) (io.WriteCloser, error) {
		fileName := fmt.Sprintf("%s/%s/%s", base_dir, appendPageNumber(uri, page), fileName)
		return os.Open(fileName)
	}

	err := feed.writeFeeds(100, fctWriteCloser)

	// move pub -> old
	// move stage -> pub
	// remove old

	return err
}

func (feed *Feed) writeFeeds(entriesPerPage int, fctWriteCloser func(string, int) (io.WriteCloser, error)) error {
	base := fmt.Sprintf("%s/%s/%s", feed.Id, uriPub, uriPosts)
	// we need to know the total pages per each feed in order to know the 'last' uri.
	// So no concurrency here :-(
	uri2entries := make(map[string][]*Entry)
	for _, item := range feed.Entries {
		// change entries for output but don't save the change: normalize Id:
		realId := item.Id
		item.Id = appendPageNumber(base, 0) + "/" + realId
		for _, uri := range feedUrlsForEntry(item) {
			uri2entries[uri] = append(uri2entries[uri], item)
		}
	}

	for uri, entries := range uri2entries {
		// but here we could go parallel...
		// ... if not for the error handling. Maybe https://godoc.org/golang.org/x/sync/errgroup
		feed.writeFeed(uri, entries, entriesPerPage, fctWriteCloser)
	}

	return nil
}

// where does the post have to go.
//
// The results have to be turned into (relative) feed uris down at writeEntries()
// and may well be a int index into a LUT
func feedUrlsForEntry(itm *Entry) []string {
	audience := uriPub
	date := itm.Published
	if date == nil {
		date = &itm.Updated
	}
	ret := make([]string, 0, 2+len(itm.Categories))
	ret = append(ret,
		fmt.Sprintf("%s/%s", audience, uriPosts), // default feed
		// fmt.Sprintf("%s/%s/%s", audience, uriPosts, itm.Id), // standalone TODO?
		fmt.Sprintf("%s/%s", audience, date.Format("2006-01-02")), // daily feed
	)
	for _, cat := range itm.Categories {
		ret = append(ret, fmt.Sprintf("%s/%s/%s", audience, uriTags, cat.Term)) // category feeds
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
func (seed Feed) writeFeed(uri string, entries []*Entry, entriesPerPage int, fctWriteCloser func(string, int) (io.WriteCloser, error)) error {
	seed.Id = fmt.Sprintf("%s/%s", seed.Id, uri)

	totalEntries := len(entries)
	lastPage := computeLastPage(totalEntries, entriesPerPage)

	if totalEntries > 0 {
		seed.Updated = entries[0].Updated // most recent sets the date
	}

	for page := 0; page <= lastPage; page++ {
		lower := page * entriesPerPage
		upper := lower + entriesPerPage
		if upper > totalEntries {
			upper = totalEntries
		}
		pageEntries := entries[lower:upper]

		if err := func() error {
			// and here we could go even more parallel...
			if w, err := fctWriteCloser(uri, page); err != nil {
				return err
			} else {
				defer w.Close()
				enc := xml.NewEncoder(w)
				enc.Indent("", "  ")
				seed.writePage(page, lastPage, pageEntries, enc)
				return enc.Flush()
			}
		}(); err != nil {
			return err
		}
	}
	return nil
}

// feed is cloned on each call
func (feed Feed) writePage(page int, lastPage int, entries []*Entry, enc *xml.Encoder) error {
	feed.Links = append(feed.Links, Link{Rel: "self", Href: appendPageNumber(feed.Id, page)})
	// https://tools.ietf.org/html/rfc5005#section-3
	if lastPage > 0 {
		feed.Links = append(feed.Links, Link{Rel: "first", Href: appendPageNumber(feed.Id, 0)})
		if page > 0 {
			feed.Links = append(feed.Links, Link{Rel: "previous", Href: appendPageNumber(feed.Id, page-1)})
		}
		if page < lastPage {
			feed.Links = append(feed.Links, Link{Rel: "next", Href: appendPageNumber(feed.Id, page+1)})
		}
		feed.Links = append(feed.Links, Link{Rel: "last", Href: appendPageNumber(feed.Id, lastPage)})
	}
	feed.Entries = entries
	return feed.write(enc)
}

func (feed *Feed) write(enc *xml.Encoder) error {
	// preamble
	if err := enc.EncodeToken(xml.ProcInst{"xml", []byte(`version="1.0" encoding="UTF-8"`)}); err != nil {
		return err
	}
	if err := enc.EncodeToken(xml.CharData("\n")); err != nil {
		return err
	}
	if err := enc.EncodeToken(xml.ProcInst{"xml-stylesheet", []byte("type='text/xsl' href='../../assets/atom2html.xslt'")}); err != nil {
		return err
	}
	if err := enc.EncodeToken(xml.CharData("\n")); err != nil {
		return err
	}
	// space comment
	if err := enc.EncodeToken(xml.Comment(lengthyAtomPreambleComment)); err != nil {
		return err
	}
	if err := enc.EncodeToken(xml.CharData("\n")); err != nil {
		return err
	}
	// feed content
	if err := enc.Encode(feed); err != nil {
		return err
	}
	return enc.EncodeToken(xml.CharData("\n"))
}
