//
// Copyright (C) 2017-2017 Marcus Rohrmoser, http://purl.mro.name/ShaarliGo
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
	"crypto/rand"
	"encoding/hex"
	"fmt"
	// "hash/crc32"
	"io"
	"net/url"
	"time"

	"github.com/stretchr/testify/assert"
	"testing"
)

func TestURLQuery(t *testing.T) {
	t.Parallel()

	par := mustParseURL("a/b/c?post=foo").Query()
	assert.Equal(t, 1, len(par["post"]), "Na klar")
	assert.Equal(t, "foo", par["post"][0], "Na klar")

	purl := fmt.Sprintf("?post=%s&title=%s&source=%s", url.QueryEscape("http://example.com/foo?bar=baz#grr"), url.QueryEscape("A first post"), url.QueryEscape("me"))
	par = mustParseURL(purl).Query()
	assert.Equal(t, 1, len(par["post"]), "Na klar")
	assert.Equal(t, "http://example.com/foo?bar=baz#grr", par["post"][0], "Na klar")
	assert.Equal(t, "A first post", par["title"][0], "Na klar")
	assert.Equal(t, "me", par["source"][0], "Na klar")
}

func TestLfTimeFmt(t *testing.T) {
	t.Parallel()

	loc, err := time.LoadLocation("Europe/Berlin")
	assert.Nil(t, err, "aua")
	// loc = time.Local

	t0, err := time.ParseInLocation(fmtTimeLfTime, "20171106_223225", loc)
	assert.Nil(t, err, "aua")
	assert.Equal(t, "2017-11-06T22:32:25+01:00", t0.Format(time.RFC3339), "Na klar")
	assert.Equal(t, "20171106_223225", t0.Format(fmtTimeLfTime), "Na klar")
}

func TestParseLinkUrl(t *testing.T) {
	t.Parallel()

	// vorher schon geklärt:
	// - ist's die rechte Seite von :// eines bekannten Links
	// - ist's eine Id eines bekannten Links

	u, e := url.Parse("www.heise.de")
	assert.Nil(t, e, "aua")
	assert.Equal(t, "www.heise.de", u.String(), "aua")
	assert.False(t, u.IsAbs(), "aua")

	u, e = url.Parse("http://www.heise.de")
	assert.Equal(t, "http://www.heise.de", u.String(), "aua")
	assert.True(t, u.IsAbs(), "aua")

	u, e = url.Parse("https://www.heise.de")
	assert.Equal(t, "https://www.heise.de", u.String(), "aua")
	assert.True(t, u.IsAbs(), "aua")

	u, e = url.Parse("voo8Uo")
	assert.Equal(t, "voo8Uo", u.String(), "aua")
	assert.False(t, u.IsAbs(), "aua")

	assert.Nil(t, e, "aua")
	assert.NotNil(t, u, "aua")

	assert.Equal(t, "http://heise.de", parseLinkUrl("heise.de").String(), "aua")
	assert.Equal(t, "http://heise.de", parseLinkUrl("http://heise.de").String(), "aua")
	assert.Equal(t, "https://heise.de", parseLinkUrl("https://heise.de").String(), "aua")
	assert.Nil(t, parseLinkUrl("Eine Notiz"), "aua")
	assert.Nil(t, parseLinkUrl("genau www.heise.de"), "aua")
}

func TestToken(t *testing.T) {
	t.Parallel()

	src := []byte("foo\x00bar8901234567890")
	assert.Equal(t, 20, len(src), "sicher")
	hx := hex.EncodeToString(src)
	assert.Equal(t, 40, len(hx), "sicher")
	assert.Equal(t, "666f6f0062617238393031323334353637383930", hx, "Na klar")

	src = make([]byte, 20)
	_, err := io.ReadFull(rand.Reader, src)
	assert.Nil(t, err, "aua")
	assert.NotNil(t, hex.EncodeToString(src), "aua")
}

func TestSmallHash(t *testing.T) {
	t.Parallel()

	// assert.Equal(t, uint32(0xc2d07353), crc32.Checksum([]byte("20171108_231054"), crc32.MakeTable((0xC96C5795<<32)|0xD7870F42)), "hm")
	// assert.Equal(t, "wtBzUw", smallHash("20171108_231054"), "aha")
	// assert.Equal(t, "yZH23w", smallHash("20111006_131924"), "the original from https://github.com/sebsauvage/Shaarli/blob/master/index.php#L228")
	assert.Equal(t, "AfsQ8g", smallHash("20111006_131924"), "strange - that's what GO produces.")

	tt, _ := time.Parse(fmtTimeLfTime, "20171108_231054")
	assert.Equal(t, "_o4DWg", smallDateHash(tt), "aha")
}

func TestApi0LinkFormMap(t *testing.T) {
	t.Parallel()

	e := Entry{}
	assert.Equal(t, map[string]string{"lf_tags": "", "lf_linkdate": "00010101_000000", "lf_title": ""}, e.api0LinkFormMap(), "oha")

	e = Entry{
		Title: HumanText{Body: "My #Post"},
	}
	assert.Equal(t, map[string]string{"lf_linkdate": "00010101_000000", "lf_title": "My #Post", "lf_tags": ""}, e.api0LinkFormMap(), "oha")

	e = Entry{
		Title:      HumanText{Body: "My #Post"},
		Categories: []Category{Category{Term: "Post"}, Category{Term: "tag1"}},
	}
	assert.Equal(t, map[string]string{"lf_tags": "tag1", "lf_linkdate": "00010101_000000", "lf_title": "My #Post"}, e.api0LinkFormMap(), "oha")
	// assert.Equal(t, map[string]string{"lf_linkdate": "00010101_000000", "lf_title": "My #Post", "lf_tags": "tag1"}, e.api0LinkFormMap(), "oha")
}
