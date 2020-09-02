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
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestYamlDeser(t *testing.T) {
	t.Parallel()

	cfg, err := loadConfi([]byte("posse:\n- pinboard: foo\n- mastodon: bar\n  limit: 200"))
	assert.Equal(t, nil, err, "Na klar")
	assert.Equal(t, 2, len(cfg.Posse), "Na klar")
	{
		pi := (cfg.Posse[0]).(Pinboard)
		assert.Equal(t, "foo", pi.Endpoint, "Na klar")
	}
	{
		ma := (cfg.Posse[1]).(Mastodon)
		assert.Equal(t, "bar", ma.Endpoint, "Na klar")
		assert.Equal(t, "200", ma.Limit, "Na klar")
	}
}
