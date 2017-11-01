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
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestQueryParse(t *testing.T) {
	u := mustParseURL("http://example.com/a/atom.cgi?do=login&foo=bar&do=auch")

	assert.Equal(t, "http://example.com/a/atom.cgi?do=login&foo=bar&do=auch", u.String(), "ach")
	assert.Equal(t, "do=login&foo=bar&do=auch", u.RawQuery, "ach")
	v := u.Query()

	assert.Equal(t, 2, len(v["do"]), "omg")
	assert.Equal(t, "login", v["do"][0], "omg")
}