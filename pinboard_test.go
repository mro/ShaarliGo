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

func TestClampString(t *testing.T) {
	t.Parallel()

	ti := "💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫"
	mx := "u💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫💫�"

	assert.Equal(t, 256, len(mx), "Na klar")
	assert.Equal(t, 1024, len(ti), "Na klar")

	ti = limit(255, "…", ti)
	assert.Equal(t, 255, len(ti), "Na klar")
}

func TestPinboardPostsAdd(t *testing.T) {
	t.Parallel()

	base, _ := url.Parse("https://api.pinboard.in/v1?auth_token=fee%3AABCDE445566")

	// dt := mustParseRFC3339("1990-12-31T02:02:02+01:00")
	en := Entry{
		Id:    Id("foo"),
		Links: []Link{Link{Href: "https://pinboard.in/api#posts_add"}},
		Title: HumanText{Body: "#Hello, #world!"},
		// Content: &HumanText{Body: "1st at l.mro.name/o/p/xyz"},
		Categories: []Category{
			Category{Term: "Uhu"},
			Category{Term: "🦉"},
		},
	}
	url, err := pinboardPostsAdd(*base, en, "=> "+string(en.Id))
	assert.Equal(t, nil, err, "Na klar")
	assert.Equal(t, "https://api.pinboard.in/v1/posts/add?auth_token=fee%3AABCDE445566&description=%23Hello%2C+%23world%21&dt=0001-01-01T00%3A00%3A00Z&extended=%3D%3E+foo&tags=Uhu+%F0%9F%A6%89&url=https%3A%2F%2Fpinboard.in%2Fapi%23posts_add", url.String(), "Na klar")
}
