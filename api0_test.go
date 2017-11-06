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
	"crypto/rand"
	"encoding/hex"
	"fmt"
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
