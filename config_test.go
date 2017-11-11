//
// Copyright (C) 2017-2017 Marcus Rohrmoser, http://purl.mro.name/GoShaarli
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
	pwd := "123456789012"
	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	assert.Nil(t, err, "soso")

	str := string(hash)

	err = bcrypt.CompareHashAndPassword([]byte(str), []byte(pwd))
	assert.Nil(t, err, "soso")

	err = bcrypt.CompareHashAndPassword([]byte(str), []byte("wrong"))
	assert.NotNil(t, err, "soso")
}

func TestXmlBaseFromRequestURL(t *testing.T) {
	assert.Equal(t, "http://example.com/", xmlBaseFromRequestURL(mustParseURL("http://example.com/goshaarli.cgi"), "/goshaarli.cgi").String(), "soso")
	assert.Equal(t, "http://example.com/b/", xmlBaseFromRequestURL(mustParseURL("http://example.com/b/goshaarli.cgi"), "/b/goshaarli.cgi").String(), "soso")
}

func TestFeedFromFileName__(t *testing.T) {
	assert.Nil(t, nil, "soso")
}
