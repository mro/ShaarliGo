//
// Copyright (C) 2017-2021 Marcus Rohrmoser, http://purl.mro.name/ShaarliGo
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
	"gopkg.in/yaml.v2"
	"time"

	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSuffix(t *testing.T) {
	assert.Equal(t, "1.2.3.4", remoteAddressToKey("1.2.3.4:5"), "soso")
}

func TestIsRemoteAddrBanned(t *testing.T) {
	now := mustParseRFC3339("2017-11-01T00:00:00+01:00")

	{
		data, err := yaml.Marshal(now)
		assert.Nil(t, err, "soso")
		assert.Equal(t, "2017-11-01T00:00:00+01:00\n", string(data), "Oh je")
	}

	bp := BanPenalties{
		Penalties: map[string]Penalty{
			"92.194.87.209": {Badness: -100, End: mustParseRFC3339("2017-01-01T01:02:03+02:00")},
			"92.194.87.210": {Badness: -100, End: now.Add(10 * time.Minute)},
			"1.2.3.3":       {Badness: 2, End: now.Add(4 * time.Hour)},
			"1.2.3.4":       {Badness: 10, End: now.Add(10 * time.Minute)},
		},
	}

	{
		data, err := yaml.Marshal(bp)
		assert.Nil(t, err, "soso")
		assert.Equal(t, `penalties:
  1.2.3.3:
    badness: 2
    end: 2017-11-01T04:00:00+01:00
  1.2.3.4:
    badness: 10
    end: 2017-11-01T00:10:00+01:00
  92.194.87.209:
    badness: -100
    end: 2017-01-01T01:02:03+02:00
  92.194.87.210:
    badness: -100
    end: 2017-11-01T00:10:00+01:00
`, string(data), "ach!")
	}

	assert.NotNil(t, &bp, "soso")
	assert.Equal(t, -100, bp.Penalties["92.194.87.209"].Badness, "soso")

	assert.False(t, BanPenalties{}.isRemoteAddrBanned("nix", time.Time{}), "unknown shouldn't be banned from the start")
	assert.False(t, bp.isRemoteAddrBanned("nix", time.Time{}), "unknown shouldn't be banned from the start")
	assert.False(t, bp.isRemoteAddrBanned("1.2.3.3", time.Time{}), "should not be banned yet")
	assert.False(t, bp.isRemoteAddrBanned("1.2.3.3", now), "should not be banned yet")
	assert.True(t, bp.isRemoteAddrBanned("1.2.3.4", time.Time{}), "should be banned")
	assert.True(t, bp.isRemoteAddrBanned("1.2.3.4", now), "should be banned")
	assert.False(t, bp.isRemoteAddrBanned("1.2.3.4", now.Add(10*24*time.Hour)), "ban should be lifted meanwhile")
}
