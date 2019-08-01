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

func ta(tags ...string) []string {
	return tags
}

func TestTagsFromString(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "", isTag(""), "aha")
	assert.Equal(t, "ha", isTag("#ha"), "aha")
	assert.Equal(t, "🐳", isTag("🐳"), "aha")
	assert.Equal(t, "⌨️", isTag("⌨️"), "aha")
	assert.Equal(t, "", isTag("foo#nein"), "aha")

	assert.Equal(t, "><(((°>", isTag("#><(((°>"), "aha")
	assert.Equal(t, "@DeMaiziere", isTag("#@DeMaiziere"), "aha")
	assert.Equal(t, "F#", isTag("#F#"), "aha")
	assert.Equal(t, "#F#", isTag("##F#"), "aha")

	assert.Equal(t, ta("ha"), tagsFromString("#ha, 1.2 foo#nein"), "aha")
	assert.Equal(t, ta("🐳"), tagsFromString("🐳, foo#nein"), "aha")
	assert.Equal(t, ta("§", "$", "†"), tagsFromString("#§, #$ #† foo#nein"), "aha")
	assert.Equal(t, ta("🐳"), tagsFromString("#🐳, foo#nein #"), "aha")
	assert.Equal(t, ta("ipsum", "opensource", "🐳"), tagsFromString(`Lorem #ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet.

		Duis autem vel eum iriure dolor in hendrerit in vulputate velit esse molestie consequat, vel illum dolore eu feugiat nulla facilisis at vero eros et accumsan et iusto odio dignissim qui blandit praesent luptatum zzril delenit augue duis dolore te feugait nulla facilisi. Lorem ipsum dolor sit amet, consectetuer adipiscing elit, sed diam nonummy nibh euismod tincidunt ut laoreet dolore magna aliquam erat volutpat.

		Ut wisi enim ad minim veniam, quis nostrud exerci tation ullamcorper suscipit lobortis nisl ut aliquip ex ea commodo consequat. Duis autem vel eum iriure dolor in hendrerit in vulputate velit esse molestie consequat, vel illum dolore eu feugiat nulla facilisis at vero eros et accumsan et iusto odio dignissim qui blandit praesent luptatum zzril delenit augue duis dolore te feugait nulla facilisi. #opensource #🐳`), "ja, genau")
	assert.Equal(t, ta("⭐️"), tagsFromString("a single ⭐️ is also a tag"), "aha")

}

func TestFold(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "hallo wyrld!", fold(" Hälló wÿrld! "), "2")
	assert.Equal(t, "demaiziere", fold(" DeMaizière \n"), "1")
	assert.Equal(t, "cegłowski", fold("\tCegłowski"), "-")
}

func TestTagsNormalise(t *testing.T) {
	t.Parallel()

	description, extended, tags := tagsNormalise("#A", "#B #C", []string{"a", "c", "D"}, map[string]string{"c": "c"})
	assert.Equal(t, "#A", description, "u1")
	assert.Equal(t, "#B #C #D", extended, "u2")
	assert.Equal(t, []string{"B", "D", "a", "c"}, tags, "u3")

	description, extended, tags = tagsNormalise("#foo #Foo #fOo #foö", "", []string{}, map[string]string{})
	assert.Equal(t, "#foo #Foo #fOo #foö", description, "u1")
	assert.Equal(t, "", extended, "u2")
	assert.Equal(t, []string{"foo"}, tags, "u3")

	description, extended, tags = tagsNormalise("a b c", "nix", []string{"", ""}, map[string]string{})
	assert.Equal(t, "a b c", description, "u1")
	assert.Equal(t, "nix", extended, "u2")
	assert.Equal(t, []string{}, tags, "u3")

	description, extended, tags = tagsNormalise("#atöm und so weitr", "", []string{"Atom"}, map[string]string{})
	assert.Equal(t, "", extended, "u2")
	assert.Equal(t, []string{"Atom"}, tags, "u3")

	description, extended, tags = tagsNormalise("🏊 #Traunstein: Neue Wasserrutsche im Schwimmbad kommt in Sicht", "…Lieferung und Montage der 🚦 Ampelanlage und der ⏱ Rutschzeitnahme…", []string{"🏊", "🚦", "⏱ ", "Traunstein"}, map[string]string{})
	assert.Equal(t, "🏊 #Traunstein: Neue Wasserrutsche im Schwimmbad kommt in Sicht", description, "u2")
	assert.Equal(t, "…Lieferung und Montage der 🚦 Ampelanlage und der ⏱ Rutschzeitnahme…", extended, "u2")
	assert.Equal(t, []string{"Traunstein", "⏱ ", "🏊", "🚦"}, tags, "u3")
}
