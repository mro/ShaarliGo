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
	"compress/gzip"
	"encoding/gob"
	"encoding/xml"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFeedFromFileName_LinksAtom(t *testing.T) {
	feed, err := FeedFromFileName("testdata/links.atom")

	assert.Nil(t, err, "soso")
	assert.Equal(t, "🔗 mro", feed.Title.Body, "soso")
	assert.Equal(t, "2017-02-09T22:44:52+01:00", feed.Updated.Format(time.RFC3339), "soso")
	assert.Equal(t, 2, len(feed.Links), "soso")
	assert.Equal(t, "self", feed.Links[0].Rel, "soso")
	assert.Equal(t, "https://links.mro.name/?do=atom&nb=all", feed.Links[0].Href, "soso")
	assert.Equal(t, "hub", feed.Links[1].Rel, "soso")
	assert.Equal(t, "http://blog.mro.name/?pushpress=hub", feed.Links[1].Href, "soso")
	assert.Equal(t, "https://links.mro.name/", feed.Authors[0].Name, "soso")
	assert.Equal(t, "https://links.mro.name/", feed.Authors[0].Uri, "soso")
	assert.Equal(t, "https://links.mro.name/", feed.Id, "soso")

	assert.Equal(t, 3618, len(feed.Entries), "soso")

	if false {
		file, err := os.Create("testdata/links.gob")
		if err == nil {
			defer file.Close()
			enc := gob.NewEncoder(file)
			enc.Encode(feed)
		}
	}
	if false {
		file, err := os.Create("testdata/links.atom.bak")
		if err == nil {
			defer file.Close()
			enc := xml.NewEncoder(file)
			enc.Indent("", "  ")
			enc.Encode(feed)
		}
	}
}

func TestFeedFromFileName_PhotosAtom(t *testing.T) {
	feed, err := FeedFromFileName("testdata/photos.atom")
	assert.Nil(t, err, "soso")
	assert.Equal(t, "Demo Album", feed.Title.Body, "soso")
	assert.Equal(t, "2016-11-27T12:32:57+01:00", feed.Updated.Format(time.RFC3339), "soso")
	assert.Equal(t, 2, len(feed.Links), "soso")
	assert.Equal(t, "self", feed.Links[0].Rel, "soso")
	assert.Equal(t, "https://lager.mro.name/galleries/demo/", feed.Links[0].Href, "soso")
	assert.Equal(t, "alternate", feed.Links[1].Rel, "soso")
	assert.Equal(t, "https://lager.mro.name/galleries/demo/", feed.Links[1].Href, "soso")
	assert.Equal(t, "Marcus Rohrmoser", feed.Authors[0].Name, "soso")
	assert.Equal(t, "http://mro.name/me", feed.Authors[0].Uri, "soso")
	assert.Equal(t, "https://lager.mro.name/galleries/demo/", feed.Id, "soso")

	assert.Equal(t, 127, len(feed.Entries), "soso")

	assert.Equal(t, "https://lager.mro.name/galleries/demo/200p/fbb6669a533054da3747fb71790dc515bbf76da2.jpeg", feed.Entries[0].MediaThumbnail.Url, "soso")
	assert.Equal(t, float32(48.047504), feed.Entries[0].GeoRssPoint.Lat, "soso")
	assert.Equal(t, float32(10.871933), feed.Entries[0].GeoRssPoint.Lon, "soso")

	if false {
		enc := xml.NewEncoder(os.Stdout)
		enc.Indent("", "  ")
		enc.EncodeToken(xml.ProcInst{"xml", []byte(`version="1.0" encoding="UTF-8"`)})
		enc.EncodeToken(xml.CharData("\n"))
		enc.EncodeToken(xml.ProcInst{"xml-stylesheet", []byte("type='text/xsl' href='assets/default/de/posts.xslt'")})
		enc.EncodeToken(xml.CharData("\n"))
		enc.EncodeToken(xml.Comment(" Am Anfang war das Licht! "))
		enc.EncodeToken(xml.CharData("\n"))
		if err := enc.Encode(feed); err != nil {
			fmt.Printf("error: %v\n", err)
		}
	}
}

func TestFeedFromFileName_LinksGobGz(t *testing.T) {
	file, err := os.Open("testdata/links.gob.gz")
	assert.Nil(t, err, "soso")
	defer file.Close()

	gunzip, err := gzip.NewReader(file)
	assert.Nil(t, err, "soso")
	defer gunzip.Close()

	feed := Feed{}
	err = gob.NewDecoder(gunzip).Decode(&feed)
	assert.Nil(t, err, "soso")

	assert.Nil(t, err, "soso")
	assert.Equal(t, "🔗 mro", feed.Title.Body, "soso")
	assert.Equal(t, "2017-02-09T22:44:52+01:00", feed.Updated.Format(time.RFC3339), "soso")
	assert.Equal(t, 2, len(feed.Links), "soso")
	assert.Equal(t, "self", feed.Links[0].Rel, "soso")
	assert.Equal(t, "https://links.mro.name/?do=atom&nb=all", feed.Links[0].Href, "soso")
	assert.Equal(t, "hub", feed.Links[1].Rel, "soso")
	assert.Equal(t, "http://blog.mro.name/?pushpress=hub", feed.Links[1].Href, "soso")
	assert.Equal(t, "https://links.mro.name/", feed.Authors[0].Name, "soso")
	assert.Equal(t, "https://links.mro.name/", feed.Authors[0].Uri, "soso")
	assert.Equal(t, "https://links.mro.name/", feed.Id, "soso")

	assert.Equal(t, 3618, len(feed.Entries), "soso")
}
