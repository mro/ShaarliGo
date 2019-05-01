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
	"os"
	"path"
	"regexp"
	"sort"
	"time"

	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSliceIndices(t *testing.T) {
	t.Parallel()
	s := []string{"a", "b", "c"}
	p := s[0:2]
	assert.Equal(t, 2, len(p), "Oha")
	assert.Equal(t, "a", p[0], "Oha")
	assert.Equal(t, "b", p[1], "Oha")

	p = s[2:]
	assert.Equal(t, 1, len(p), "Oha")
	assert.Equal(t, "c", p[0], "Oha")
}

func TestUriBasics(t *testing.T) {
	t.Parallel()
	u1 := mustParseURL("b")
	assert.Equal(t, "a/b", path.Join("a", u1.Path), "Oha")
	assert.Equal(t, "/b", mustParseURL("a").ResolveReference(u1).String(), "Oha")
	assert.Equal(t, "/b", mustParseURL(".").ResolveReference(u1).String(), "Oha")
	assert.Equal(t, "https://mro.name/b", mustParseURL("https://mro.name").ResolveReference(u1).String(), "Oha")
	assert.Equal(t, "https://mro.name/b", mustParseURL("https://mro.name/").ResolveReference(u1).String(), "Oha")

	assert.Equal(t, "https://mro.name/b", mustParseURL("https://mro.name/sub").ResolveReference(u1).String(), "Oha")
	assert.Equal(t, "https://mro.name/sub/b", mustParseURL("https://mro.name/sub/").ResolveReference(u1).String(), "Oha")

	urn := mustParseURL("urn:uuid:60a76c80-d399-11d9-b93C-0003939e0af6")
	assert.Equal(t, "urn:uuid:60a76c80-d399-11d9-b93C-0003939e0af6", urn.String(), "Oha")
	assert.Equal(t, "urn", urn.Scheme, "Oha")
	assert.Equal(t, "uuid:60a76c80-d399-11d9-b93C-0003939e0af6", urn.Opaque, "Oha")
	assert.Equal(t, "", urn.Host, "Oha")
	assert.Equal(t, "", urn.Path, "Oha")
	assert.Equal(t, "", urn.RawQuery, "Oha")
	assert.Equal(t, "", urn.Fragment, "Oha")

	assert.Equal(t, "../../..", regexp.MustCompile("[^/]+").ReplaceAllString("a/b/c", ".."), "Oha")
}

func TestComputeLastPage(t *testing.T) {
	t.Parallel()
	assert.Equal(t, 1, computePageCount(0, 100), "Oha")
	assert.Equal(t, 1, computePageCount(1, 100), "Oha")
	assert.Equal(t, 1, computePageCount(100, 100), "Oha")
	assert.Equal(t, 2, computePageCount(101, 100), "Oha")
}

func TestEntryFeedFilters(t *testing.T) {
	t.Parallel()
	itm := &Entry{
		Id:         "id_0",
		Published:  iso8601(mustParseRFC3339("2010-12-31T00:11:22Z")),
		Categories: []Category{{Term: "🐳"}},
	}

	keys := uriSliceSorted(itm.FeedFilters(nil))
	assert.Equal(t, []string{
		uriPubDays + "2010-12-31" + "/",
		uriPubPosts,
		uriPubPosts + "id_0" + "/",
		uriPubTags,
		uriPubTags + "🐳" + "/",
	}, keys, "Oha")
}

func TestPathJoin(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "a/b", path.Join("a", "b", ""), "Oha")
}

func TestAppendPageNumber(t *testing.T) {
	t.Parallel()
	s := "abc/"
	assert.Equal(t, "/", s[len(s)-1:], "Oha")
	assert.Equal(t, uriPub+"/"+uriPosts+"-"+"0"+"/", appendPageNumber(uriPubPosts, 0, 1+1), "Oha")
	assert.Equal(t, uriPubPosts, appendPageNumber(uriPubPosts, 1, 1+1), "Oha")
}

func TestWriteFeedsEmpty0(t *testing.T) {
	assert.Equal(t, "feed/@xml:base must be set to an absolute URL with a trailing slash but not ''", Server{}.PublishFeedsForModifiedEntries(Feed{}, nil).Error(), "aha")
}

func TestWriteFeedsAddOneAndOneAndRemoveFirst(t *testing.T) {
	t.Parallel()
	feed := &Feed{XmlBase: Iri("http://example.com/")}
	{
		entry := &Entry{
			Id:         "id_0",
			Published:  iso8601(mustParseRFC3339("2010-12-31T00:11:22Z")),
			Categories: []Category{{Term: "🐳"}},
		}

		feed.Append(entry)

		complete := feed.CompleteFeedsForModifiedEntries([]*Entry{entry})

		assert.Equal(t, 5, len(complete), "ja")

		assert.Equal(t, Id(uriPubDays+"2010-12-31"+"/"), complete[0].Id, "ja")
		assert.Equal(t, 1, len(complete[0].Entries), "ja")

		assert.Equal(t, Id(uriPubPosts), complete[1].Id, "ja")
		assert.Equal(t, 1, len(complete[1].Entries), "ja")

		assert.Equal(t, Id(uriPubPosts+"id_0/"), complete[2].Id, "ja")
		assert.Equal(t, 1, len(complete[2].Entries), "ja")

		assert.Equal(t, Id(uriPubTags), complete[3].Id, "ja")
		assert.Equal(t, 0, len(complete[3].Entries), "ja")

		assert.Equal(t, Id(uriPubTags+"🐳/"), complete[4].Id, "ja")
		assert.Equal(t, 1, len(complete[4].Entries), "ja")
	}
	{
		entry := &Entry{
			Id:         "id_1",
			Published:  iso8601(mustParseRFC3339("2010-12-30T00:11:22Z")),
			Categories: []Category{{Term: "foo"}},
		}

		feed.Append(entry)

		complete := feed.CompleteFeedsForModifiedEntries([]*Entry{entry})

		assert.Equal(t, 5, len(complete), "ja")

		assert.Equal(t, Id(uriPubDays+"2010-12-30"+"/"), complete[0].Id, "ja")
		assert.Equal(t, 1, len(complete[0].Entries), "ja")

		assert.Equal(t, Id(uriPubPosts), complete[1].Id, "ja")
		assert.Equal(t, 2, len(complete[1].Entries), "ja")

		assert.Equal(t, Id(uriPubPosts+"id_1"+"/"), complete[2].Id, "ja")
		assert.Equal(t, 1, len(complete[2].Entries), "ja")

		assert.Equal(t, Id(uriPubTags), complete[3].Id, "ja")
		assert.Equal(t, 0, len(complete[3].Entries), "ja")

		assert.Equal(t, Id(uriPubTags+"foo"+"/"), complete[4].Id, "ja")
		assert.Equal(t, 1, len(complete[4].Entries), "ja")
	}
	{
		e0 := *feed.Entries[0]
		feed.deleteEntryById("id_0")

		complete := feed.CompleteFeedsForModifiedEntries([]*Entry{&e0})

		assert.Equal(t, 5, len(complete), "ja")

		assert.Equal(t, Id(uriPubDays+"2010-12-31"+"/"), complete[0].Id, "ja")
		assert.Equal(t, 0, len(complete[0].Entries), "ja")

		assert.Equal(t, Id(uriPubPosts), complete[1].Id, "ja")
		assert.Equal(t, 1, len(complete[1].Entries), "ja")

		assert.Equal(t, Id(uriPubPosts+"id_0"+"/"), complete[2].Id, "ja")
		assert.Equal(t, 0, len(complete[2].Entries), "ja")

		assert.Equal(t, Id(uriPubTags), complete[3].Id, "ja")
		assert.Equal(t, 0, len(complete[3].Entries), "ja")

		assert.Equal(t, Id(uriPubTags+"🐳"+"/"), complete[4].Id, "ja")
		assert.Equal(t, 0, len(complete[4].Entries), "ja")
	}
}

func TestWriteFeedsPaged(t *testing.T) {
	t.Parallel()
	feed := &Feed{
		XmlBase: Iri("http://example.com/"),
		XmlLang: "deu",
		Id:      Id(mustParseURL("http://example.com").String()),
		Title:   HumanText{Body: "Hello, Atom!"},
		Entries: []*Entry{
			{
				Id:        "e2",
				Title:     HumanText{Body: "Hello, Entry 2!"},
				Published: iso8601(mustParseRFC3339("1990-12-31T02:02:02+01:00")),
			},
			{
				Id:        "e1",
				Title:     HumanText{Body: "Hello, Entry 1!"},
				Published: iso8601(mustParseRFC3339("1990-12-31T01:01:01+01:00")),
			},
			{
				Id:        "e0",
				Title:     HumanText{Body: "Hello, Entry 0!"},
				Published: iso8601(mustParseRFC3339("1990-12-30T00:00:00+01:00")),
			},
		},
	}

	sort.Sort(ByPublishedDesc(feed.Entries))

	complete := feed.CompleteFeedsForModifiedEntries(feed.Entries)
	assert.Equal(t, 7, len(complete), "uhu")

	pages, err := feed.PagedFeeds(complete, 1)
	assert.Nil(t, err, "uhu")
	assert.Equal(t, 10, len(pages), "uhu")

	i := 0
	assert.Equal(t, Id(uriPubDays+"1990-12-30/"), pages[i].Id, "ja")
	assert.Equal(t, uriPubDays+"1990-12-30/", LinkRelSelf(pages[i].Links).Href, "ja")
	assert.Equal(t, 1, len(pages[i].Entries), "ja")
	i++
	assert.Equal(t, Id(uriPubDays+"1990-12-31/"), pages[i].Id, "ja")
	assert.Equal(t, uriPubDays+"1990-12-31"+"-"+"0"+"/", LinkRelSelf(pages[i].Links).Href, "ja")
	assert.Equal(t, 1, len(pages[i].Entries), "ja")
	i++
	assert.Equal(t, Id(uriPubDays+"1990-12-31/"), pages[i].Id, "ja")
	assert.Equal(t, uriPubDays+"1990-12-31/", LinkRelSelf(pages[i].Links).Href, "ja")
	assert.Equal(t, 1, len(pages[i].Entries), "ja")
	i++
	assert.Equal(t, Id(uriPubPosts), pages[i].Id, "ja")
	assert.Equal(t, uriPub+"/"+uriPosts+"-"+"0"+"/", LinkRelSelf(pages[i].Links).Href, "ja")
	assert.Equal(t, 1, len(pages[i].Entries), "ja")
	i++
	assert.Equal(t, Id(uriPubPosts), pages[i].Id, "ja")
	assert.Equal(t, uriPub+"/"+uriPosts+"-"+"1"+"/", LinkRelSelf(pages[i].Links).Href, "ja")
	assert.Equal(t, 1, len(pages[i].Entries), "ja")
	i++
	assert.Equal(t, Id(uriPubPosts), pages[i].Id, "ja")
	assert.Equal(t, uriPubPosts, LinkRelSelf(pages[i].Links).Href, "ja")
	assert.Equal(t, 1, len(pages[i].Entries), "ja")
	i++
	assert.Equal(t, Id(uriPubPosts+"e0/"), pages[i].Id, "ja")
	assert.Equal(t, uriPubPosts+"e0/", LinkRelSelf(pages[i].Links).Href, "ja")
	assert.Equal(t, 1, len(pages[i].Entries), "ja")
	i++
	assert.Equal(t, Id(uriPubPosts+"e1/"), pages[i].Id, "ja")
	assert.Equal(t, uriPubPosts+"e1/", LinkRelSelf(pages[i].Links).Href, "ja")
	assert.Equal(t, 1, len(pages[i].Entries), "ja")
	i++
	assert.Equal(t, Id(uriPubPosts+"e2/"), pages[i].Id, "ja")
	assert.Equal(t, uriPubPosts+"e2/", LinkRelSelf(pages[i].Links).Href, "ja")
	assert.Equal(t, 1, len(pages[i].Entries), "ja")
	i++
	assert.Equal(t, Id(uriPubTags), pages[i].Id, "ja")
	assert.Equal(t, uriPubTags, LinkRelSelf(pages[i].Links).Href, "ja")
	assert.Equal(t, 0, len(pages[i].Entries), "ja")
	i++
	assert.Equal(t, len(pages), i, "ja")
}

func TestPagedFeeds(t *testing.T) {
	t.Parallel()
	feed, err := FeedFromFileName("testdata/feedwriter.TestPagedFeeds.pub.atom")
	assert.Nil(t, err, "ja")
	assert.Equal(t, 5, len(feed.Entries), "ja")
	feed.XmlBase = "http://foo.eu/s/"

	feeds := feed.CompleteFeedsForModifiedEntries([]*Entry{feed.Entries[0]})
	assert.Equal(t, 4, len(feeds), "ja")
	assert.Equal(t, Id(uriPubDays+"2018-01-22/"), feeds[0].Id, "ja")
	assert.Equal(t, Id(uriPubPosts), feeds[1].Id, "ja")
	assert.Equal(t, Id(uriPubPosts+"XsuMcA/"), feeds[2].Id, "ja")
	assert.Equal(t, Id(uriPubTags), feeds[3].Id, "ja")

	// test low level
	assert.Equal(t, uriPubPosts, LinkRelSelf(feeds[1].Pages(100)[0].Links).Href, "ja")

	pages := make([]Feed, 0, 2*len(feeds))
	for _, comp := range feeds {
		pages = append(pages, comp.Pages(100)...)
	}
	assert.Equal(t, 4, len(pages), "ja")
	assert.Equal(t, Id(uriPubPosts), pages[1].Id, "ja")
	assert.Equal(t, uriPubPosts, LinkRelSelf(pages[1].Links).Href, "ja")

	// high level
	pages, err = feed.PagedFeeds(feeds, 100)
	assert.Nil(t, err, "ja")
	assert.Equal(t, 4, len(pages), "ja")
	assert.Equal(t, Id(uriPubDays+"2018-01-22/"), pages[0].Id, "ja")
	assert.Equal(t, Id(uriPubPosts), pages[1].Id, "ja")
	assert.Equal(t, Id(uriPubPosts+"XsuMcA/"), pages[2].Id, "ja")
	assert.Equal(t, Id(uriPubTags), pages[3].Id, "ja")

	assert.Equal(t, uriPubDays+"2018-01-22/", LinkRelSelf(pages[0].Links).Href, "ja")
	assert.Equal(t, uriPubPosts, LinkRelSelf(pages[1].Links).Href, "ja")
	assert.Equal(t, uriPubPosts+"XsuMcA/", LinkRelSelf(pages[2].Links).Href, "ja")
	assert.Equal(t, uriPubTags, LinkRelSelf(pages[3].Links).Href, "ja")

	pages = feeds[1].Pages(100)
	assert.Equal(t, Id(uriPubPosts), pages[0].Id, "ja")
	assert.Equal(t, uriPubPosts, LinkRelSelf(pages[0].Links).Href, "ja")
}

func BenchmarkFileStat(b *testing.B) {
	t0 := time.Time{}
	for i := 0; i < b.N; i++ {
		if fi, err := os.Stat("."); (fi != nil && fi.ModTime().Before(t0)) || os.IsNotExist(err) {
			// older or missing: (re-)create
		}
	}
}

/*
func TestWriteFeedsEmpty1(t *testing.T) {
	feed := &Feed{
		XmlBase: mustParseURL("http://example.com/").String(),
		Entries: []*Entry{&Entry{Id: "abcd"}}, // a single, but almost empty, entry
	}

	sfw := saveFeedWriter{feeds: make(map[string]Feed), entries: make(map[string]Entry), bufs: make(map[string]buff)}
	err := feed.writeFeeds(2, sfw)
	assert.Nil(t, err, "soso")
	assert.Equal(t, []string{uriPubDays+"0001-01-01/", uriPubPosts, uriPubPosts+"abcd/"}, keys4map(sfw.bufs), "soso")

	assert.Equal(t, 1472, len(sfw.bufs[uriPubDays+"0001-01-01/"].b), "aha")
	assert.Equal(t, `<?xml version="1.0" encoding="UTF-8"?>
<?xml-stylesheet type='text/xsl' href='../../assets/default/de/posts.xslt'?>
<!--
  https://developer.mozilla.org/en/docs/XSL_Transformations_in_Mozilla_FAQ#Why_isn.27t_my_stylesheet_applied.3F

  Note that Firefox will override your XSLT stylesheet if your XML is
  detected as a RSS or Atom feed. A known workaround is to add a
  sufficiently long XML comment to the beginning of your XML file in
  order to 'push' the <.feed> or <.rss> tag out of the first 512 bytes,
  which is analyzed by Firefox to determine if it's a feed or not. See
  the discussion on bug
  https://bugzilla.mozilla.org/show_bug.cgi?id=338621#c72 for more
  information.

  For best results serve both atom feed and xslt as 'text/xml' or
  'application/xml' without charset specified.
-->
<feed xmlns="http://www.w3.org/2005/Atom" xml:base="http://example.com/">
  <title></title>
  <id></id>
  <updated>0001-01-01T00:00:00Z</updated>
  <link href=uriPubPosts rel="self" title="1"></link>
  <entry xmlns="http://www.w3.org/2005/Atom" xml:base="http://example.com/">
    <title></title>
    <id>http://example.com/"+uriPubPosts+"abcd/</id>
    <updated>0001-01-01T00:00:00Z</updated>
    <published>0001-01-01T00:00:00Z</published>
    <link href=uriPubPosts+"abcd/" rel="self"></link>
    <link href="shaarligo.cgi?post="+uriPubPosts+"abcd/" rel="edit"></link>
    <link href="../" rel="up"></link>
  </entry>
</feed>
`, string(sfw.bufs[uriPubPosts].b), "soso")
}

func TestWriteFeedsUnpaged(t *testing.T) {
	feed := &Feed{
		XmlBase: mustParseURL("http://example.com/").String(),
		Id:      mustParseURL("http://example.com/").String(),
		Title:   HumanText{Body: "Hello, Atom!"},
		Entries: []*Entry{&Entry{
			Id:         "e0",
			Title:      HumanText{Body: "Hello, Entry!"},
			Updated:    iso8601{mustParseRFC3339("1990-12-31T01:02:03+01:00")},
			Categories: []Category{Category{Term: "aha"}},
		}},
	}
	sfw := saveFeedWriter{feeds: make(map[string]Feed), entries: make(map[string]Entry), bufs: make(map[string]buff)}
	err := feed.writeFeeds(2, sfw)
	assert.Nil(t, err, "soso")
	assert.Equal(t, []string{
		uriPubDays+"1990-12-31/",
		uriPubPosts,
		uriPubPosts+"e0/",
		uriPubTags+"aha/",
	}, keys4map(sfw.bufs), "soso")

	assert.Equal(t, 1661, len(sfw.bufs[uriPubDays+"1990-12-31/"].b), "aha")
	assert.Equal(t, `<?xml version="1.0" encoding="UTF-8"?>
<?xml-stylesheet type='text/xsl' href='../../assets/default/de/posts.xslt'?>
<!--
  https://developer.mozilla.org/en/docs/XSL_Transformations_in_Mozilla_FAQ#Why_isn.27t_my_stylesheet_applied.3F

  Note that Firefox will override your XSLT stylesheet if your XML is
  detected as a RSS or Atom feed. A known workaround is to add a
  sufficiently long XML comment to the beginning of your XML file in
  order to 'push' the <.feed> or <.rss> tag out of the first 512 bytes,
  which is analyzed by Firefox to determine if it's a feed or not. See
  the discussion on bug
  https://bugzilla.mozilla.org/show_bug.cgi?id=338621#c72 for more
  information.

  For best results serve both atom feed and xslt as 'text/xml' or
  'application/xml' without charset specified.
-->
<feed xmlns="http://www.w3.org/2005/Atom" xml:base="http://example.com/">
  <title>Hello, Atom!</title>
  <id>http://example.com/</id>
  <updated>1990-12-31T01:02:03+01:00</updated>
  <link href=uriPubPosts rel="self" title="1"></link>
  <category term="aha" label="1"></category>
  <entry xmlns="http://www.w3.org/2005/Atom" xml:base="http://example.com/">
    <title>Hello, Entry!</title>
    <id>http://example.com/"+uriPubPosts+"e0/</id>
    <updated>1990-12-31T01:02:03+01:00</updated>
    <published>0001-01-01T00:00:00Z</published>
    <link href=uriPubPosts+"e0/" rel="self"></link>
    <link href="shaarligo.cgi?post="+uriPubPosts+"e0/" rel="edit"></link>
    <link href="../" rel="up" title="Hello, Atom!"></link>
    <category term="aha" scheme="http://example.com/"+uriPubTags></category>
  </entry>
</feed>
`, string(sfw.bufs[uriPubPosts].b), "soso")

	assert.Equal(t, `<?xml version="1.0" encoding="UTF-8"?>
<?xml-stylesheet type='text/xsl' href='../../../assets/default/de/posts.xslt'?>
<!--
  https://developer.mozilla.org/en/docs/XSL_Transformations_in_Mozilla_FAQ#Why_isn.27t_my_stylesheet_applied.3F

  Note that Firefox will override your XSLT stylesheet if your XML is
  detected as a RSS or Atom feed. A known workaround is to add a
  sufficiently long XML comment to the beginning of your XML file in
  order to 'push' the <.feed> or <.rss> tag out of the first 512 bytes,
  which is analyzed by Firefox to determine if it's a feed or not. See
  the discussion on bug
  https://bugzilla.mozilla.org/show_bug.cgi?id=338621#c72 for more
  information.

  For best results serve both atom feed and xslt as 'text/xml' or
  'application/xml' without charset specified.
-->
<entry xmlns="http://www.w3.org/2005/Atom" xml:base="http://example.com/">
  <title>Hello, Entry!</title>
  <id>http://example.com/"+uriPubPosts+"e0/</id>
  <updated>1990-12-31T01:02:03+01:00</updated>
  <published>0001-01-01T00:00:00Z</published>
  <link href=uriPubPosts+"e0/" rel="self"></link>
  <link href="shaarligo.cgi?post="+uriPubPosts+"e0/" rel="edit"></link>
  <link href="../" rel="up" title="Hello, Atom!"></link>
  <category term="aha" scheme="http://example.com/"+uriPubTags></category>
</entry>
`, string(sfw.bufs[uriPubPosts+"e0/"].b), "soso")

}

func TestWriteFeedsPaged(t *testing.T) {
	feed := &Feed{
		XmlBase: mustParseURL("http://example.com/").String(),
		XmlLang: "deu",
		Id:      mustParseURL("http://example.com").String(),
		Title:   HumanText{Body: "Hello, Atom!"},
		Entries: []*Entry{
			&Entry{
				Id:      "e2",
				Title:   HumanText{Body: "Hello, Entry 2!"},
				Updated: iso8601{mustParseRFC3339("1990-12-31T02:02:02+01:00")},
			},
			&Entry{
				Id:      "e1",
				Title:   HumanText{Body: "Hello, Entry 1!"},
				Updated: iso8601{mustParseRFC3339("1990-12-31T01:01:01+01:00")},
			},
			&Entry{
				Id:      "e0",
				Title:   HumanText{Body: "Hello, Entry 0!"},
				Updated: iso8601{mustParseRFC3339("1990-12-30T00:00:00+01:00")},
			},
		},
	}

	sfw := saveFeedWriter{feeds: make(map[string]Feed), entries: make(map[string]Entry), bufs: make(map[string]buff)}
	err := feed.writeFeeds(2, sfw)
	assert.Nil(t, err, "soso")
	assert.Equal(t, []string{
		uriPubDays+"1990-12-30/",
		uriPubDays+"1990-12-31/",
		"pub/posts-1/",
		uriPubPosts,
		uriPubPosts+"e0/",
		uriPubPosts+"e1/",
		uriPubPosts+"e2/",
	}, keys4map(sfw.bufs), "soso")

	assert.Equal(t, `<?xml version="1.0" encoding="UTF-8"?>
<?xml-stylesheet type='text/xsl' href='../../assets/default/de/posts.xslt'?>
<!--
  https://developer.mozilla.org/en/docs/XSL_Transformations_in_Mozilla_FAQ#Why_isn.27t_my_stylesheet_applied.3F

  Note that Firefox will override your XSLT stylesheet if your XML is
  detected as a RSS or Atom feed. A known workaround is to add a
  sufficiently long XML comment to the beginning of your XML file in
  order to 'push' the <.feed> or <.rss> tag out of the first 512 bytes,
  which is analyzed by Firefox to determine if it's a feed or not. See
  the discussion on bug
  https://bugzilla.mozilla.org/show_bug.cgi?id=338621#c72 for more
  information.

  For best results serve both atom feed and xslt as 'text/xml' or
  'application/xml' without charset specified.
-->
<feed xmlns="http://www.w3.org/2005/Atom" xml:base="http://example.com/" xml:lang="deu">
  <title>Hello, Atom!</title>
  <id>http://example.com</id>
  <updated>1990-12-30T00:00:00+01:00</updated>
  <link href=uriPubPosts rel="self" title="1"></link>
  <link href=uriPubPosts rel="first" title="1"></link>
  <link href="pub/posts-1/" rel="next" title="2"></link>
  <link href="pub/posts-1/" rel="last" title="2"></link>
  <entry xmlns="http://www.w3.org/2005/Atom" xml:base="http://example.com/">
    <title>Hello, Entry 0!</title>
    <id>http://example.com/"+uriPubPosts+"e0/</id>
    <updated>1990-12-30T00:00:00+01:00</updated>
    <published>0001-01-01T00:00:00Z</published>
    <link href=uriPubPosts+"e0/" rel="self"></link>
    <link href="shaarligo.cgi?post="+uriPubPosts+"e0/" rel="edit"></link>
    <link href="../" rel="up" title="Hello, Atom!"></link>
  </entry>
  <entry xmlns="http://www.w3.org/2005/Atom" xml:base="http://example.com/">
    <title>Hello, Entry 1!</title>
    <id>http://example.com/"+uriPubPosts+"e1/</id>
    <updated>1990-12-31T01:01:01+01:00</updated>
    <published>0001-01-01T00:00:00Z</published>
    <link href=uriPubPosts+"e1/" rel="self"></link>
    <link href="shaarligo.cgi?post="+uriPubPosts+"e1/" rel="edit"></link>
    <link href="../" rel="up" title="Hello, Atom!"></link>
  </entry>
</feed>
`,
		string(sfw.bufs[uriPubPosts].b), "page 1")

	assert.Equal(t, `<?xml version="1.0" encoding="UTF-8"?>
<?xml-stylesheet type='text/xsl' href='../../assets/default/de/posts.xslt'?>
<!--
  https://developer.mozilla.org/en/docs/XSL_Transformations_in_Mozilla_FAQ#Why_isn.27t_my_stylesheet_applied.3F

  Note that Firefox will override your XSLT stylesheet if your XML is
  detected as a RSS or Atom feed. A known workaround is to add a
  sufficiently long XML comment to the beginning of your XML file in
  order to 'push' the <.feed> or <.rss> tag out of the first 512 bytes,
  which is analyzed by Firefox to determine if it's a feed or not. See
  the discussion on bug
  https://bugzilla.mozilla.org/show_bug.cgi?id=338621#c72 for more
  information.

  For best results serve both atom feed and xslt as 'text/xml' or
  'application/xml' without charset specified.
-->
<feed xmlns="http://www.w3.org/2005/Atom" xml:base="http://example.com/" xml:lang="deu">
  <title>Hello, Atom!</title>
  <id>http://example.com</id>
  <updated>1990-12-30T00:00:00+01:00</updated>
  <link href="pub/posts-1/" rel="self" title="2"></link>
  <link href=uriPubPosts rel="first" title="1"></link>
  <link href=uriPubPosts rel="previous" title="1"></link>
  <link href="pub/posts-1/" rel="last" title="2"></link>
  <entry xmlns="http://www.w3.org/2005/Atom" xml:base="http://example.com/">
    <title>Hello, Entry 2!</title>
    <id>http://example.com/"+uriPubPosts+"e2/</id>
    <updated>1990-12-31T02:02:02+01:00</updated>
    <published>0001-01-01T00:00:00Z</published>
    <link href=uriPubPosts+"e2/" rel="self"></link>
    <link href="shaarligo.cgi?post="+uriPubPosts+"e2/" rel="edit"></link>
    <link href="../" rel="up" title="Hello, Atom!"></link>
  </entry>
</feed>
`,
		string(sfw.bufs["pub/posts-1/"].b), "page 2")

	assert.Equal(t, `<?xml version="1.0" encoding="UTF-8"?>
<?xml-stylesheet type='text/xsl' href='../../../assets/default/de/posts.xslt'?>
<!--
  https://developer.mozilla.org/en/docs/XSL_Transformations_in_Mozilla_FAQ#Why_isn.27t_my_stylesheet_applied.3F

  Note that Firefox will override your XSLT stylesheet if your XML is
  detected as a RSS or Atom feed. A known workaround is to add a
  sufficiently long XML comment to the beginning of your XML file in
  order to 'push' the <.feed> or <.rss> tag out of the first 512 bytes,
  which is analyzed by Firefox to determine if it's a feed or not. See
  the discussion on bug
  https://bugzilla.mozilla.org/show_bug.cgi?id=338621#c72 for more
  information.

  For best results serve both atom feed and xslt as 'text/xml' or
  'application/xml' without charset specified.
-->
<entry xmlns="http://www.w3.org/2005/Atom" xml:base="http://example.com/">
  <title>Hello, Entry 0!</title>
  <id>http://example.com/"+uriPubPosts+"e0/</id>
  <updated>1990-12-30T00:00:00+01:00</updated>
  <published>0001-01-01T00:00:00Z</published>
  <link href=uriPubPosts+"e0/" rel="self"></link>
  <link href="shaarligo.cgi?post="+uriPubPosts+"e0/" rel="edit"></link>
  <link href="../" rel="up" title="Hello, Atom!"></link>
</entry>
`,
		string(sfw.bufs[uriPubPosts+"e0/"].b), "page 2")
}

func BenchmarkWriteFeedsPaged(b *testing.B) {
	for i := 0; i < b.N; i++ {
		feed := &Feed{
			XmlBase: mustParseURL("http://example.com/").String(),
			XmlLang: "deu",
			Id:      mustParseURL("http://example.com").String(),
			Title:   HumanText{Body: "Hello, Atom!"},
			Entries: []*Entry{
				&Entry{
					Id:      "e2",
					Title:   HumanText{Body: "Hello, Entry 2!"},
					Updated: iso8601{mustParseRFC3339("1990-12-31T02:02:02+01:00")},
				},
				&Entry{
					Id:      "e1",
					Title:   HumanText{Body: "Hello, Entry 1!"},
					Updated: iso8601{mustParseRFC3339("1990-12-31T01:01:01+01:00")},
				},
				&Entry{
					Id:      "e0",
					Title:   HumanText{Body: "Hello, Entry 0!"},
					Updated: iso8601{mustParseRFC3339("1990-12-30T00:00:00+01:00")},
				},
			},
		}

		sfw := saveFeedWriter{feeds: make(map[string]Feed), entries: make(map[string]Entry), bufs: make(map[string]buff)}
		feed.writeFeeds(2, sfw)
	}
}
*/
