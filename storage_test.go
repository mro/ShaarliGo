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
	"os"
	"time"

	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCleanup(t *testing.T) {

	subdirs := [2]string{"app", "~me"}
	for _, n := range subdirs {
		os.Rename(n, n+"~")
	}
	for _, n := range subdirs {
		os.Rename(n, n+"~")
	}

	assert.Nil(t, nil, "soso")
}

func TestSeed(t *testing.T) {
	assert.Nil(t, nil, "soso")
}

func (feed *Feed) SplitByAudience() []Feed {
	f := *feed          // make a copy
	return []Feed{f, f} // fake it for now
}

func (feed *Feed) SplitByTags() []Feed {
	f := *feed       // make a copy
	return []Feed{f} // fake it for now
}

func (feed *Feed) SplitByDays(tz *time.Location) []Feed {
	f := *feed       // make a copy
	return []Feed{f} // fake it for now
}

// Build a paged feed. https://tools.ietf.org/html/rfc5005
func (feed *Feed) Paginate(namebase string, maxEntries int) []Feed {
	f := *feed       // make a copy
	return []Feed{f} // fake it for now
}

// generated feeds should go into a channel, so we could split in goroutines
func (feed *Feed) Dispatch() []Feed {
	num := 50
	var here *time.Location = nil
	ret := make([]Feed, 40) // may be waaaay more.
	for _, audi := range feed.SplitByAudience() {
		// pub/posts
		// pub/posts-1
		ret = append(ret, audi.Paginate("posts", num)...)

		for _, tag := range audi.SplitByTags() {
			// pub/tags/Design
			// pub/tags/Design-1
			ret = append(ret, tag.Paginate("tags", num)...)
		}

		ret = append(ret, audi.SplitByDays(here)...)

		// write each single Entry for a permanent url? add backlinks to the feed(s)
		// pub/posts/aZ6453Gf
	}
	// channel close ?
	return ret
}

// Result from Dispatch
func SaveAsAtom(feeds []Feed) error {
	// save to ~
	// remove .
	// rename to .
	return nil
}

func TestGeneratePosts(t *testing.T) {
	assert.Nil(t, nil, "soso")

	// Paged feed https://tools.ietf.org/html/rfc5005#section-3
	feed := Feed{
		Generator: &Generator{Uri: "http://purl.mro.name/AtomicShaarli"},
		Title:     HumanText{Body: "foo", Type: "text/plain"},
		Id:        "http://purl.mro.name/AtomicShaarli",
		Updated:   iso8601{Time: time.Now()},
		Authors:   []Person{Person{Name: "Jon Doe"}},
		Links: []Link{
			Link{Rel: "self", Href: "http://purl.mro.name/AtomicShaarli"},
			Link{Rel: "first", Href: "http://purl.mro.name/AtomicShaarli"},
			Link{Rel: "last", Href: "http://purl.mro.name/AtomicShaarli"},
			// Link{Rel: "previous", Href: "http://purl.mro.name/AtomicShaarli"},
			// Link{Rel: "next", Href: "http://purl.mro.name/AtomicShaarli"},
			Link{Rel: "AddUri", Href: "../../atom.cgi/posts"},
		},
		Entries: []Entry{},
	}

	file, err := os.Create("testdata/posts.xml~")
	if err == nil {
		defer file.Close()
		enc := xml.NewEncoder(file)
		enc.Indent("", "  ")
		enc.EncodeToken(xml.ProcInst{"xml", []byte(`version="1.0" encoding="UTF-8"`)})
		enc.EncodeToken(xml.CharData("\n"))
		enc.EncodeToken(xml.ProcInst{"xml-stylesheet", []byte("type='text/xsl' href='../../assets/atom2html.xslt'")})
		enc.EncodeToken(xml.CharData("\n"))
		enc.EncodeToken(xml.Comment(lengthyAtomPreambleComment))
		enc.EncodeToken(xml.CharData("\n"))
		if err := enc.Encode(feed); err != nil {
			fmt.Printf("error: %v\n", err)
		}
		enc.EncodeToken(xml.CharData("\n"))
		enc.Flush()
	}

	// iterate over all posts (assumes sorted descending update)
	// open outputs:
	// - pub/posts
	// - pub/posts/DK0BTg
	// - ~me/posts-1
	// - pub/tags/a
	// - ~me/tags/b
	// - pub/2017-07-13

}
