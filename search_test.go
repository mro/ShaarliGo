//
// Copyright (C) 2018-2018 Marcus Rohrmoser, http://purl.mro.name/ShaarliGo
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
	"golang.org/x/text/language"
	"golang.org/x/text/search"

	"github.com/stretchr/testify/assert"
	"testing"
)

func entry(title, content string) *Entry {
	return &Entry{
		Title:   HumanText{Body: title},
		Content: &HumanText{Body: content},
	}
}

func TestRankEntryTerms(t *testing.T) {
	t.Parallel()
	lang := language.Make("DE-LU")
	assert.Equal(t, "de-LU", lang.String(), "aha")

	matcher := search.New(language.German, search.IgnoreDiacritics, search.IgnoreCase)

	assert.Equal(t, 0, rankEntryTerms(nil, nil, nil), "soso")
	assert.Equal(t, 0, rankEntryTerms(&Entry{}, nil, nil), "soso")
	assert.Equal(t, 2, rankEntryTerms(&Entry{Title: HumanText{Body: "my foo bar"}}, []string{"foo"}, matcher), "soso")

	assert.Equal(t, 2, rankEntryTerms(entry("my foo bar", ""), []string{"fòO"}, matcher), "soso")
}

//
