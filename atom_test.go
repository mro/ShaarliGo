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
	"bufio"
	"compress/gzip"
	"encoding/gob"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/stretchr/testify/assert"
	"testing"
)

func TestURLEqual(t *testing.T) {
	t.Parallel()

	// todo!
}

func TestScanner(t *testing.T) {
	t.Parallel()

	scanner := bufio.NewScanner(strings.NewReader(`Lorem #ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. 

Duis autem vel eum iriure dolor in hendrerit in vulputate velit esse molestie consequat, vel illum dolore eu feugiat nulla facilisis at vero eros et accumsan et iusto odio dignissim qui blandit praesent luptatum zzril delenit augue duis dolore te feugait nulla facilisi. Lorem ipsum dolor sit amet, consectetuer adipiscing elit, sed diam nonummy nibh euismod tincidunt ut laoreet dolore magna aliquam erat volutpat. 

Ut wisi enim ad minim veniam, quis nostrud exerci tation ullamcorper suscipit lobortis nisl ut aliquip ex ea commodo consequat. Duis autem vel eum iriure dolor in hendrerit in vulputate velit esse molestie consequat, vel illum dolore eu feugiat nulla facilisis at vero eros et accumsan et iusto odio dignissim qui blandit praesent luptatum zzril delenit augue duis dolore te feugait nulla facilisi. #opensource #🐳`))
	scanner.Split(bufio.ScanWords)

	ret := make([]string, 0, 10)
	for scanner.Scan() {
		t := scanner.Text()
		ret = append(ret, strings.TrimRightFunc(t, unicode.IsPunct))
	}
	assert.Equal(t, 279, len(ret), "so")
	assert.Equal(t, "facilisi", ret[279-3], "so")
	assert.Equal(t, "#opensource", ret[279-2], "so")
	assert.Equal(t, "#🐳", ret[279-1], "so")
}

func TestTagsFromString(t *testing.T) {
	t.Parallel()

	assert.Equal(t, []string{"ha"}, tagsFromString("#ha, foo#nein"), "aha")
	assert.Equal(t, []string{"🐳"}, tagsFromString("#🐳, foo#nein"), "aha")
	assert.Equal(t, []string{"🐳"}, tagsFromString("#🐳, foo#nein #"), "aha")
	assert.Equal(t, []string{"ipsum", "opensource", "🐳"}, tagsFromString(`Lorem #ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. 

Duis autem vel eum iriure dolor in hendrerit in vulputate velit esse molestie consequat, vel illum dolore eu feugiat nulla facilisis at vero eros et accumsan et iusto odio dignissim qui blandit praesent luptatum zzril delenit augue duis dolore te feugait nulla facilisi. Lorem ipsum dolor sit amet, consectetuer adipiscing elit, sed diam nonummy nibh euismod tincidunt ut laoreet dolore magna aliquam erat volutpat. 

Ut wisi enim ad minim veniam, quis nostrud exerci tation ullamcorper suscipit lobortis nisl ut aliquip ex ea commodo consequat. Duis autem vel eum iriure dolor in hendrerit in vulputate velit esse molestie consequat, vel illum dolore eu feugiat nulla facilisis at vero eros et accumsan et iusto odio dignissim qui blandit praesent luptatum zzril delenit augue duis dolore te feugait nulla facilisi. #opensource #🐳`), "ja, genau")
}

func TestEntryCategoriesMerged(t *testing.T) {
	t.Parallel()
	e := Entry{
		Title:      HumanText{Body: "#ha, foo#nein"},
		Content:    &HumanText{Body: "B #so, bar"},
		Categories: []Category{Category{Term: "so"}, Category{Term: "da"}},
	}
	assert.NotNil(t, e, "genau")
	assert.Equal(t, []Category{Category{Term: "da"}, Category{Term: "ha"}, Category{Term: "so"}}, e.CategoriesMerged(), "genau")
}

func TestFeedFromFileName_Atom(t *testing.T) {
	t.Parallel()
	feed, err := FeedFromFileName("testdata/links.atom")
	assert.Nil(t, err, "soso")
	assert.Equal(t, "🔗 mro", feed.Title.Body, "soso")
	assert.Equal(t, "2017-02-09T22:44:52+01:00", feed.Updated.Format(time.RFC3339), "soso")
	assert.Equal(t, 0, len(feed.Links), "soso")
	assert.Equal(t, "m", feed.Authors[0].Name, "soso")
	assert.Equal(t, "", feed.Authors[0].Uri, "soso")
	assert.Equal(t, "", feed.Id, "soso")

	assert.Equal(t, "html", feed.Entries[0].Content.Type, "soso")
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
	assert.Equal(t, "self", feed.Links[0].Rel, "soso")
	assert.Equal(t, "https://lager.mro.name/galleries/demo/", feed.Links[0].Href, "soso")
	assert.Equal(t, "alternate", feed.Links[1].Rel, "soso")
	assert.Equal(t, "https://lager.mro.name/galleries/demo/", feed.Links[1].Href, "soso")
	assert.Equal(t, "Marcus Rohrmoser", feed.Authors[0].Name, "soso")
	assert.Equal(t, "http://mro.name/me", feed.Authors[0].Uri, "soso")
	assert.Equal(t, "https://lager.mro.name/galleries/demo/", feed.Id, "soso")

	assert.Equal(t, 127, len(feed.Entries), "soso")

	assert.Equal(t, "https://lager.mro.name/galleries/demo/200p/fbb6669a533054da3747fb71790dc515bbf76da2.jpeg", feed.Entries[0].MediaThumbnail.Url, "soso")
	assert.Equal(t, float32(48.047504), feed.Entries[0].GeoRssPoint.Lat, "soso")
	assert.Equal(t, float32(10.871933), feed.Entries[0].GeoRssPoint.Lon, "soso")
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
	assert.Equal(t, "https://links.mro.name/", feed.Id, "soso")

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
	assert.Equal(t, "https://links.mro.name/", feed.Id, "soso")

	assert.Equal(t, 3618, len(feed.Entries), "soso")
}
