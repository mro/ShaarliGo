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
	"time"

	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIsBanned(t *testing.T) {
	mgr := GetManager()
	assert.NotNil(t, mgr, "soso")

	var t0 time.Time
	banned, err := mgr.IsBanned(nil, t0)
	assert.Nil(t, err, "soso")
	assert.True(t, banned, "soso")

	addr := ""
	banned, err = mgr.isBanned(&addr, t0)
	assert.Nil(t, err, "soso")
	assert.False(t, banned, "soso")

	addr = "127.0.0.1"
	banned, err = mgr.isBanned(&addr, t0)
	assert.Nil(t, err, "soso")
	assert.False(t, banned, "soso")
}

func TestAfter(t *testing.T) {
	// var t0 *time.Time = nil
	// assert.True(t, time.Now().After(*t0), "soso")
}
