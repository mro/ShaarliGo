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
	"encoding/xml"
	"fmt"
	"net/url"
	"path"
	"regexp"
	"sort"

	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSliceIndices(t *testing.T) {
	s := []string{"a", "b", "c"}
	p := s[0:2]
	assert.Equal(t, 2, len(p), "Oha")
	assert.Equal(t, "a", p[0], "Oha")
	assert.Equal(t, "b", p[1], "Oha")

	p = s[2:len(s)]
	assert.Equal(t, 1, len(p), "Oha")
	assert.Equal(t, "c", p[0], "Oha")
}

func TestUriBasics(t *testing.T) {
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
	assert.Equal(t, 0, computeLastPage(0, 100), "Oha")
	assert.Equal(t, 0, computeLastPage(1, 100), "Oha")
	assert.Equal(t, 0, computeLastPage(100, 100), "Oha")
	assert.Equal(t, 1, computeLastPage(101, 100), "Oha")
}

func TestFeedUrlsForEntry(t *testing.T) {
	itm := &Entry{
		Id:         "id_0",
		Published:  iso8601{mustParseRFC3339("2010-12-31T00:11:22Z")},
		Categories: []Category{Category{Term: "🐳"}},
	}

	uris := feedUrlsForEntry(itm)

	assert.Equal(t, 2+1, len(uris), "Oha")
	assert.Equal(t, "pub/posts/", uris[0], "Oha")
	assert.Equal(t, "pub/days/2010-12-31/", uris[1], "Oha")
	assert.Equal(t, "pub/tags/🐳/", uris[2], "Oha")
}

func TestPathJoin(t *testing.T) {
	assert.Equal(t, "a/b", path.Join("a", "b", ""), "Oha")
}

func TestAppendPageNumber(t *testing.T) {
	s := "abc/"
	assert.Equal(t, "/", s[len(s)-1:], "Oha")
	assert.Equal(t, "pub/posts/", appendPageNumber("pub/posts/", 0), "Oha")
	assert.Equal(t, "pub/posts-1/", appendPageNumber("pub/posts/", 1), "Oha")
}

type buff struct {
	b []byte
}

func (buff *buff) Write(p []byte) (n int, err error) {
	buff.b = append(buff.b, p...)
	return len(p), nil
}

func (buff *buff) Close() error {
	return nil
}

func TestWriteFeedsEmpty0(t *testing.T) {
	assert.Equal(t, "feed/@xml:base must be set to an absolute URL with a trailing slash", new(Feed).writeFeeds(2, nil).Error(), "aha")
}

type saveFeedWriter struct {
	feeds   map[string]Feed
	entries map[string]Entry
	bufs    map[string]buff
}

func (sfw saveFeedWriter) Write(feedOrEntry interface{}, self *url.URL, xsltFileName string) error {
	uri := self.String()

	fep, ok := feedOrEntry.(*Feed)
	if ok {
		sfw.feeds[uri] = *fep
	}
	enp, ok := feedOrEntry.(*Entry)
	if ok {
		sfw.entries[uri] = *enp
	}

	w := new(buff)

	pathPrefix := rexPath.ReplaceAllString(uri, "..")
	xslt := path.Join(pathPrefix, dirAssets, "default/de", xsltFileName)

	enc := xml.NewEncoder(w)
	enc.Indent("", "  ")
	if err := xmlEncodeWithXslt(feedOrEntry, xslt, enc); err == nil {
		enc.Flush()
		w.Close()
		sfw.bufs[uri] = *w
	}

	return nil
}

func keys4map(mymap map[string]buff) []string {
	keys := make([]string, len(mymap))
	i := 0
	for k := range mymap {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	return keys
}

func TestWriteFeedsEmpty1(t *testing.T) {
	feed := &Feed{
		XmlBase: mustParseURL("http://example.com/").String(),
		Entries: []*Entry{&Entry{Id: "abcd"}}, // a single, but almost empty, entry
	}

	sfw := saveFeedWriter{feeds: make(map[string]Feed), entries: make(map[string]Entry), bufs: make(map[string]buff)}
	err := feed.writeFeeds(2, sfw)
	assert.Nil(t, err, "soso")
	assert.Equal(t, []string{"pub/days/0001-01-01/", "pub/posts/", "pub/posts/abcd/"}, keys4map(sfw.bufs), "soso")

	assert.Equal(t, 1472, len(sfw.bufs["pub/days/0001-01-01/"].b), "aha")
	assert.Equal(t, `<?xml version="1.0" encoding="UTF-8"?>
<?xml-stylesheet type='text/xsl' href='../../assets/default/de/posts.xslt'?>
<!--
  https://developer.mozilla.org/en/docs/XSL_Transformations_in_Mozilla_FAQ#Why_isn.27t_my_stylesheet_applied.3F

  Note that Firefox will override your XSLT stylesheet if your XML is
  detected as an RSS or Atom feed. A known workaround is to add a
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
  <link href="pub/posts/" rel="self" title="1"></link>
  <entry xmlns="http://www.w3.org/2005/Atom" xml:base="http://example.com/">
    <title></title>
    <id>http://example.com/pub/posts/abcd/</id>
    <updated>0001-01-01T00:00:00Z</updated>
    <published>0001-01-01T00:00:00Z</published>
    <link href="pub/posts/abcd/" rel="self"></link>
    <link href="shaarligo.cgi?post=pub/posts/abcd/" rel="edit"></link>
    <link href="../" rel="up"></link>
  </entry>
</feed>
`, string(sfw.bufs["pub/posts/"].b), "soso")
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
		"pub/days/1990-12-31/",
		"pub/posts/",
		"pub/posts/e0/",
		"pub/tags/aha/",
	}, keys4map(sfw.bufs), "soso")

	assert.Equal(t, 1661, len(sfw.bufs["pub/days/1990-12-31/"].b), "aha")
	assert.Equal(t, `<?xml version="1.0" encoding="UTF-8"?>
<?xml-stylesheet type='text/xsl' href='../../assets/default/de/posts.xslt'?>
<!--
  https://developer.mozilla.org/en/docs/XSL_Transformations_in_Mozilla_FAQ#Why_isn.27t_my_stylesheet_applied.3F

  Note that Firefox will override your XSLT stylesheet if your XML is
  detected as an RSS or Atom feed. A known workaround is to add a
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
  <link href="pub/posts/" rel="self" title="1"></link>
  <category term="aha" label="1"></category>
  <entry xmlns="http://www.w3.org/2005/Atom" xml:base="http://example.com/">
    <title>Hello, Entry!</title>
    <id>http://example.com/pub/posts/e0/</id>
    <updated>1990-12-31T01:02:03+01:00</updated>
    <published>0001-01-01T00:00:00Z</published>
    <link href="pub/posts/e0/" rel="self"></link>
    <link href="shaarligo.cgi?post=pub/posts/e0/" rel="edit"></link>
    <link href="../" rel="up" title="Hello, Atom!"></link>
    <category term="aha" scheme="http://example.com/pub/tags/"></category>
  </entry>
</feed>
`, string(sfw.bufs["pub/posts/"].b), "soso")

	assert.Equal(t, `<?xml version="1.0" encoding="UTF-8"?>
<?xml-stylesheet type='text/xsl' href='../../../assets/default/de/posts.xslt'?>
<!--
  https://developer.mozilla.org/en/docs/XSL_Transformations_in_Mozilla_FAQ#Why_isn.27t_my_stylesheet_applied.3F

  Note that Firefox will override your XSLT stylesheet if your XML is
  detected as an RSS or Atom feed. A known workaround is to add a
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
  <id>http://example.com/pub/posts/e0/</id>
  <updated>1990-12-31T01:02:03+01:00</updated>
  <published>0001-01-01T00:00:00Z</published>
  <link href="pub/posts/e0/" rel="self"></link>
  <link href="shaarligo.cgi?post=pub/posts/e0/" rel="edit"></link>
  <link href="../" rel="up" title="Hello, Atom!"></link>
  <category term="aha" scheme="http://example.com/pub/tags/"></category>
</entry>
`, string(sfw.bufs["pub/posts/e0/"].b), "soso")

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
		"pub/days/1990-12-30/",
		"pub/days/1990-12-31/",
		"pub/posts-1/",
		"pub/posts/",
		"pub/posts/e0/",
		"pub/posts/e1/",
		"pub/posts/e2/",
	}, keys4map(sfw.bufs), "soso")

	assert.Equal(t, `<?xml version="1.0" encoding="UTF-8"?>
<?xml-stylesheet type='text/xsl' href='../../assets/default/de/posts.xslt'?>
<!--
  https://developer.mozilla.org/en/docs/XSL_Transformations_in_Mozilla_FAQ#Why_isn.27t_my_stylesheet_applied.3F

  Note that Firefox will override your XSLT stylesheet if your XML is
  detected as an RSS or Atom feed. A known workaround is to add a
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
  <link href="pub/posts/" rel="self" title="1"></link>
  <link href="pub/posts/" rel="first" title="1"></link>
  <link href="pub/posts-1/" rel="next" title="2"></link>
  <link href="pub/posts-1/" rel="last" title="2"></link>
  <entry xmlns="http://www.w3.org/2005/Atom" xml:base="http://example.com/">
    <title>Hello, Entry 0!</title>
    <id>http://example.com/pub/posts/e0/</id>
    <updated>1990-12-30T00:00:00+01:00</updated>
    <published>0001-01-01T00:00:00Z</published>
    <link href="pub/posts/e0/" rel="self"></link>
    <link href="shaarligo.cgi?post=pub/posts/e0/" rel="edit"></link>
    <link href="../" rel="up" title="Hello, Atom!"></link>
  </entry>
  <entry xmlns="http://www.w3.org/2005/Atom" xml:base="http://example.com/">
    <title>Hello, Entry 1!</title>
    <id>http://example.com/pub/posts/e1/</id>
    <updated>1990-12-31T01:01:01+01:00</updated>
    <published>0001-01-01T00:00:00Z</published>
    <link href="pub/posts/e1/" rel="self"></link>
    <link href="shaarligo.cgi?post=pub/posts/e1/" rel="edit"></link>
    <link href="../" rel="up" title="Hello, Atom!"></link>
  </entry>
</feed>
`,
		string(sfw.bufs["pub/posts/"].b), "page 1")

	assert.Equal(t, `<?xml version="1.0" encoding="UTF-8"?>
<?xml-stylesheet type='text/xsl' href='../../assets/default/de/posts.xslt'?>
<!--
  https://developer.mozilla.org/en/docs/XSL_Transformations_in_Mozilla_FAQ#Why_isn.27t_my_stylesheet_applied.3F

  Note that Firefox will override your XSLT stylesheet if your XML is
  detected as an RSS or Atom feed. A known workaround is to add a
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
  <link href="pub/posts/" rel="first" title="1"></link>
  <link href="pub/posts/" rel="previous" title="1"></link>
  <link href="pub/posts-1/" rel="last" title="2"></link>
  <entry xmlns="http://www.w3.org/2005/Atom" xml:base="http://example.com/">
    <title>Hello, Entry 2!</title>
    <id>http://example.com/pub/posts/e2/</id>
    <updated>1990-12-31T02:02:02+01:00</updated>
    <published>0001-01-01T00:00:00Z</published>
    <link href="pub/posts/e2/" rel="self"></link>
    <link href="shaarligo.cgi?post=pub/posts/e2/" rel="edit"></link>
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
  detected as an RSS or Atom feed. A known workaround is to add a
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
  <id>http://example.com/pub/posts/e0/</id>
  <updated>1990-12-30T00:00:00+01:00</updated>
  <published>0001-01-01T00:00:00Z</published>
  <link href="pub/posts/e0/" rel="self"></link>
  <link href="shaarligo.cgi?post=pub/posts/e0/" rel="edit"></link>
  <link href="../" rel="up" title="Hello, Atom!"></link>
</entry>
`,
		string(sfw.bufs["pub/posts/e0/"].b), "page 2")
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

func ExampleFeed_WriteFeeds() {
	fmt.Println("hello")
	// Output: hello
}
