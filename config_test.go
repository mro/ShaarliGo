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
	"os"
	"syscall"

	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIsConfigured(t *testing.T) {
	cfg := Config{
		AuthorName: "A",
		Title:      "B",
		PwdBcrypt:  "C",
	}
	assert.True(t, cfg.IsConfigured(), "soso")
	cfg.AuthorName = ""
	assert.False(t, cfg.IsConfigured(), "soso")
}

func TestConfigFromFileName(t *testing.T) {
	config, err := configFromFileName("testdata/config.yaml")

	assert.Nil(t, err, "soso")
	assert.Equal(t, "🔗 mro", config.Title, "soso")
	assert.Equal(t, "MyUserName", config.AuthorName, "soso")
	assert.Equal(t, "***", config.PwdBcrypt, "soso")
}

func TestNonexConfigFromFileName(t *testing.T) {
	_, err := configFromFileName("testdata/config.nonex")

	assert.NotNil(t, err, "soso")
	pathErr, ok := err.(*os.PathError)
	assert.True(t, ok, "soso")
	assert.Equal(t, "open", pathErr.Op, "soso")
	assert.Equal(t, syscall.Errno(2), pathErr.Err, "soso")
}

func TestSaveFail(t *testing.T) {
	cfg := Config{}
	err := cfg.saveToFileName("testdata/foo/config.yaml")

	assert.NotNil(t, err, "soso")
	pathErr, ok := err.(*os.PathError)
	assert.True(t, ok, "soso")
	assert.Equal(t, "open", pathErr.Op, "soso")
	assert.Equal(t, syscall.Errno(2), pathErr.Err, "soso")
}

func TestSaveOk(t *testing.T) {
	cfg := Config{
		AuthorName: "A",
		Title:      "B",
		PwdBcrypt:  "C",
	}
	defer os.RemoveAll(CONFIG_FILE_PATH)
	err := cfg.Save()
	assert.Nil(t, err, "soso")
}
