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
	"compress/gzip"
	"encoding/gob"
	"os"
	"time"

	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIriAdd(t *testing.T) {
	t.Parallel()

	assert.Equal(t, Iri("ab"), Iri("a")+Iri("b"), "uh")
}

func ta(tags ...string) map[string]struct{} {
	ret := make(map[string]struct{}, len(tags))
	for _, tag := range tags {
		ret[tag] = struct{}{}
	}
	return ret
}

func TestTagsFromString(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "", isTag(""), "aha")
	assert.Equal(t, "ha", isTag("#ha"), "aha")
	assert.Equal(t, "🐳", isTag("🐳"), "aha")
	assert.Equal(t, "", isTag("foo#nein"), "aha")

	assert.Equal(t, "><(((°>", isTag("#><(((°>"), "aha")
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

func TestEntryCategoriesMerged(t *testing.T) {
	t.Parallel()
	e := Entry{
		Title:      HumanText{Body: "#ha, foo#nein"},
		Content:    &HumanText{Body: "B #so, bar"},
		Categories: []Category{{Term: "so"}, {Term: "da"}},
	}
	assert.NotNil(t, e, "genau")
	assert.Equal(t, []Category{{Term: "da"}, {Term: "ha"}, {Term: "so"}}, e.CategoriesMerged(), "genau")
}

var b64Tob24Tests = []struct {
	n        string
	expected string
}{
	{"FxW3Ow", "77tk489"},
	{"Zt3wHA", "4ezde68"},
	{"oHordQ", "c8x2syk"}, // http://sebsauvage.net/links/?oHordQ
	{"McLIuQ", "k9cr7c3"}, // http://sebsauvage.net/links/?McLIuQ
}

func TestFeedIdLowerFromIdMixed(t *testing.T) {
	// https://dave.cheney.net/2013/06/09/writing-table-driven-tests-in-go
	for _, tt := range b64Tob24Tests {
		actual, _ := base64ToBase24x7(tt.n)
		if actual != tt.expected {
			t.Errorf("base64ToBase24x7(%s): expected %s, actual %s", tt.n, tt.expected, actual)
		}
	}
}

func TestFeedNewEntry(t *testing.T) {
	f := Feed{}
	ent := f.newEntry(time.Time{})
	assert.True(t, ent.Published.IsZero(), "oha")
	assert.Equal(t, Id("dzcz8k2"), ent.Id, "soso")
}

func TestFeedFromFileName_Atom(t *testing.T) {
	t.Parallel()
	feed, err := FeedFromFileName("testdata/links.atom")
	assert.Nil(t, err, "soso")
	assert.Equal(t, "🔗 mro", feed.Title.Body, "soso")
	assert.Equal(t, "2017-02-09T22:44:52+01:00", feed.Updated.Format(time.RFC3339), "soso")
	assert.Equal(t, 0, len(feed.Links), "soso")
	assert.Equal(t, "m", feed.Authors[0].Name, "soso")
	assert.Equal(t, Iri(""), feed.Authors[0].Uri, "soso")
	assert.Equal(t, Id(""), feed.Id, "soso")

	assert.Equal(t, TextType("html"), feed.Entries[0].Content.Type, "soso")
	txt := `&quot;… Ein Vertreter der Bundesanwaltschaft (BAW) erklärte vor dem U-Ausschuss des Bundestages, Marschners Akte sei selbst für die BAW gesperrt. …&quot;<br />
<br />
Ach was, wen gibt's denn dann da noch so?<br>(<a href="https://links.mro.name/?aTh_gA">Permalink</a>)`
	assert.Equal(t, txt, feed.Entries[0].Content.Body, "soso")
	assert.Equal(t, `"… Ein Vertreter der Bundesanwaltschaft (BAW) erklärte vor dem U-Ausschuss des Bundestages, Marschners Akte sei selbst für die BAW gesperrt. …"

Ach was, wen gibt's denn dann da noch so?`, cleanLegacyContent(txt), "soso")

	assert.Equal(t, 3618, len(feed.Entries), "soso")
}

func _TestFeedFromFileName_AtomLarge(t *testing.T) {
	if testing.Short() {
		t.Skip("long running")
	}
	t.Parallel()
	feed, err := FeedFromFileName("testdata/sebsauvage.atom")
	assert.Nil(t, err, "soso")
	assert.Equal(t, 21900, len(feed.Entries), "soso")
}

func TestFeedFromFileName_PhotosAtom(t *testing.T) {
	t.Parallel()
	feed, err := FeedFromFileName("testdata/photos.atom")
	assert.Nil(t, err, "soso")
	assert.Equal(t, "Demo Album", feed.Title.Body, "soso")
	assert.Equal(t, "2016-11-27T12:32:57+01:00", feed.Updated.Format(time.RFC3339), "soso")
	assert.Equal(t, 2, len(feed.Links), "soso")
	assert.Equal(t, relSelf, feed.Links[0].Rel, "soso")
	assert.Equal(t, "https://lager.mro.name/galleries/demo/", feed.Links[0].Href, "soso")
	assert.Equal(t, relAlternate, feed.Links[1].Rel, "soso")
	assert.Equal(t, "https://lager.mro.name/galleries/demo/", feed.Links[1].Href, "soso")
	assert.Equal(t, "Marcus Rohrmoser", feed.Authors[0].Name, "soso")
	assert.Equal(t, Iri("http://mro.name/me"), feed.Authors[0].Uri, "soso")
	assert.Equal(t, Id("https://lager.mro.name/galleries/demo/"), feed.Id, "soso")

	assert.Equal(t, 127, len(feed.Entries), "soso")

	assert.Equal(t, Iri("https://lager.mro.name/galleries/demo/200p/fbb6669a533054da3747fb71790dc515bbf76da2.jpeg"), feed.Entries[0].MediaThumbnail.Url, "soso")
	assert.Equal(t, Latitude(48.047504), feed.Entries[0].GeoRssPoint.Lat, "soso")
	assert.Equal(t, Longitude(10.871933), feed.Entries[0].GeoRssPoint.Lon, "soso")
}

func _TestFeedLargeToGob(t *testing.T) {
	if testing.Short() {
		t.Skip("long running")
	}
	t.Parallel()

	file, err := os.Create("testdata/sebsauvage.gob~")
	assert.Nil(t, err, "soso")
	defer file.Close()

	feed, err := FeedFromFileName("testdata/sebsauvage.atom")
	assert.Nil(t, err, "soso")
	err = gob.NewEncoder(file).Encode(feed)
	assert.Nil(t, err, "soso")
}

func _TestFeedLargeToAtomClean(t *testing.T) {
	if testing.Short() {
		t.Skip("long running")
	}
	t.Parallel()

	feed, err := FeedFromFileName("testdata/sebsauvage.atom")
	assert.Nil(t, err, "soso")

	for _, entry := range feed.Entries {
		entry.Content.Body = cleanLegacyContent(entry.Content.Body)
		entry.Content.Type = "text"
		if entry.Published.IsZero() {
			entry.Published = entry.Updated
		}
	}

	err = feed.SaveToFile("testdata/sebsauvage.atom~")
	assert.Nil(t, err, "soso")
}

func _TestFeedFromFileName_GobLarge(t *testing.T) {
	t.Parallel()

	file, err := os.Open("testdata/sebsauvage.gob")
	assert.Nil(t, err, "soso")
	defer file.Close()

	feed := Feed{}
	err = gob.NewDecoder(file).Decode(&feed)
	assert.Nil(t, err, "soso")
	assert.Equal(t, 21900, len(feed.Entries), "soso")
}

func _TestFeedFromFileName_Gob(t *testing.T) {
	t.Parallel()

	file, err := os.Open("testdata/links.gob")
	assert.Nil(t, err, "soso")
	defer file.Close()

	feed := Feed{}
	err = gob.NewDecoder(file).Decode(&feed)
	assert.Nil(t, err, "soso")

	assert.Nil(t, err, "soso")
	assert.Equal(t, "🔗 mro", feed.Title.Body, "soso")
	assert.Equal(t, "2017-02-09T22:44:52+01:00", feed.Updated.Format(time.RFC3339), "soso")
	assert.Equal(t, 2, len(feed.Links), "soso")
	assert.Equal(t, "self", feed.Links[0].Rel, "soso")
	assert.Equal(t, "https://links.mro.name/?do=atom&nb=all", feed.Links[0].Href, "soso")
	assert.Equal(t, "hub", feed.Links[1].Rel, "soso")
	assert.Equal(t, "http://blog.mro.name/?pushpress=hub", feed.Links[1].Href, "soso")
	assert.Equal(t, "https://links.mro.name/", feed.Authors[0].Name, "soso")
	assert.Equal(t, "https://links.mro.name/", feed.Authors[0].Uri, "soso")
	assert.Equal(t, Id("https://links.mro.name/"), feed.Id, "soso")

	assert.Equal(t, 3618, len(feed.Entries), "soso")
}

func _TestFeedFromFileName_GobGz(t *testing.T) {
	t.Parallel()
	file, err := os.Open("testdata/links.gob.gz")
	assert.Nil(t, err, "soso")
	defer file.Close()

	gunzip, err := gzip.NewReader(file)
	assert.Nil(t, err, "soso")
	defer gunzip.Close()

	feed := Feed{}
	err = gob.NewDecoder(gunzip).Decode(&feed)
	assert.Nil(t, err, "soso")

	assert.Nil(t, err, "soso")
	assert.Equal(t, "🔗 mro", feed.Title.Body, "soso")
	assert.Equal(t, "2017-02-09T22:44:52+01:00", feed.Updated.Format(time.RFC3339), "soso")
	assert.Equal(t, 2, len(feed.Links), "soso")
	assert.Equal(t, "self", feed.Links[0].Rel, "soso")
	assert.Equal(t, "https://links.mro.name/?do=atom&nb=all", feed.Links[0].Href, "soso")
	assert.Equal(t, "hub", feed.Links[1].Rel, "soso")
	assert.Equal(t, "http://blog.mro.name/?pushpress=hub", feed.Links[1].Href, "soso")
	assert.Equal(t, "https://links.mro.name/", feed.Authors[0].Name, "soso")
	assert.Equal(t, "https://links.mro.name/", feed.Authors[0].Uri, "soso")
	assert.Equal(t, Id("https://links.mro.name/"), feed.Id, "soso")

	assert.Equal(t, 3618, len(feed.Entries), "soso")
}
