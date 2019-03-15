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
	"net/url"

	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPinboardAddUrl(t *testing.T) {
	t.Parallel()

	base, _ := url.Parse("https://api.pinboard.in/v1")
	ur, _ := url.Parse("https://m.heise.de/")
	tags := []string{"a", "b", "c"}
	dt := mustParseRFC3339("1990-12-31T02:02:02+01:00")
	assert.Equal(t, "https://api.pinboard.in/v1/posts/add?auth_token=fee%3AABCDE445566&description=desc%26rip%3Dtion&dt=1990-12-31T02%3A02%3A02%2B01%3A00&extended=ex%3Fte%3Dded&tags=a+b+c&url=https%3A%2F%2Fm.heise.de%2F", pinboardAddUrl(base, "fee:ABCDE445566", ur, "desc&rip=tion", "ex?te=ded", tags, dt).String(), "Na klar")
}
