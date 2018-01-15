//
// Copyright (C) 2017-2018 Marcus Rohrmoser, http://purl.mro.name/ShaarliGo
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

func TestLoadParseUrl(t *testing.T) {
	t.Parallel()

	_, err := url.Parse("http://„Epic Zen Garden“")
	assert.NotNil(t, err, "my bookmarks  OY0RWg")
}
