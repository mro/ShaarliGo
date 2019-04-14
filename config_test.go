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
	"golang.org/x/crypto/bcrypt"

	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBcrypt(t *testing.T) {
	t.Parallel()
	pwd := "123456789012"
	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	assert.Nil(t, err, "soso")

	str := string(hash)

	err = bcrypt.CompareHashAndPassword([]byte(str), []byte(pwd))
	assert.Nil(t, err, "soso")

	err = bcrypt.CompareHashAndPassword([]byte(str), []byte("wrong"))
	assert.NotNil(t, err, "soso")
}

func TestLoadFeedSeed(t *testing.T) {
	t.Parallel()
	feed, err := FeedFromFileName("testdata/config-feed-seed.xml")
	assert.Nil(t, err, "soso")
	assert.Equal(t, "Seed entries for new feeds", feed.Title.Body, "soso")
	assert.Equal(t, 3, len(feed.Entries), "soso")
	assert.Equal(t, "Hello, #Atom!", feed.Entries[0].Title.Body, "soso")
	assert.Equal(t, "Was noch alles fehlt", feed.Entries[1].Title.Body, "soso")
	assert.Equal(t, "Shaarli — sebsauvage.net", feed.Entries[2].Title.Body, "soso")
}
