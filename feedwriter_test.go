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
	"io"
	"time"

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

func mustParseRFC3339(str string) time.Time {
	ret, err := time.Parse(time.RFC3339, str)
	if err != nil {
		panic(err)
	}
	return ret
}

func TestComputeLastPage(t *testing.T) {
	assert.Equal(t, 0, computeLastPage(0, 100), "Oha")
	assert.Equal(t, 0, computeLastPage(1, 100), "Oha")
	assert.Equal(t, 0, computeLastPage(100, 100), "Oha")
	assert.Equal(t, 1, computeLastPage(101, 100), "Oha")
}

func TestFeedUrlsForEntry(t *testing.T) {
	itm := &Entry{
		Published:  &iso8601{mustParseRFC3339("2010-12-31T00:11:22Z")},
		Categories: []Category{Category{Term: "🐳"}},
	}

	uris := feedUrlsForEntry(itm)

	assert.Equal(t, 3, len(uris), "Oha")
	assert.Equal(t, "pub/posts", uris[0], "Oha")
	assert.Equal(t, "pub/2010-12-31", uris[1], "Oha")
	assert.Equal(t, "pub/tags/🐳", uris[2], "Oha")
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
	bufs := make(map[string]*buff)
	// callback
	fctWriteCloser := func(uri string, page int) (io.WriteCloser, error) {
		bf := new(buff)
		bufs[appendPageNumber(uri, page)] = bf
		return bf, nil
	}

	feed := new(Feed)
	err := feed.writeFeeds(2, fctWriteCloser)
	assert.Equal(t, 0, len(bufs), "soso")
	assert.Nil(t, err, "aha")
}

func TestWriteFeedsEmpty1(t *testing.T) {
	bufs := make(map[string]*buff)
	// callback
	fctWriteCloser := func(uri string, page int) (io.WriteCloser, error) {
		bf := new(buff)
		bufs[appendPageNumber(uri, page)] = bf
		return bf, nil
	}

	feed := &Feed{Entries: []*Entry{new(Entry)}}
	err := feed.writeFeeds(2, fctWriteCloser)

	assert.Nil(t, err, "soso")
	assert.Equal(t, 2, len(bufs), "soso")
	assert.NotNil(t, bufs["pub/0001-01-01"], "aha")
	assert.Equal(t, `<?xml version="1.0" encoding="UTF-8"?>
<?xml-stylesheet type='text/xsl' href='../../assets/atom2html.xslt'?>
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
<feed xmlns="http://www.w3.org/2005/Atom">
  <title></title>
  <id>/pub/posts</id>
  <updated>0001-01-01T00:00:00Z</updated>
  <link href="/pub/posts" rel="self"></link>
  <entry>
    <title></title>
    <id>/pub/posts/</id>
    <updated>0001-01-01T00:00:00Z</updated>
  </entry>
</feed>
`, string(bufs["pub/posts"].b), "soso")
}

func TestWriteFeedsUnpaged(t *testing.T) {
	bufs := make(map[string]*buff)
	// callback
	fctWriteCloser := func(uri string, page int) (io.WriteCloser, error) {
		bf := new(buff)
		bufs[appendPageNumber(uri, page)] = bf
		return bf, nil
	}

	feed := &Feed{
		Id:    "http://example.com",
		Title: HumanText{Body: "Hello, Atom!"},
		Entries: []*Entry{&Entry{
			Id:      "e0",
			Title:   HumanText{Body: "Hello, Entry!"},
			Updated: iso8601{mustParseRFC3339("1990-12-31T01:02:03+01:00")},
		}},
	}
	err := feed.writeFeeds(2, fctWriteCloser)

	assert.Nil(t, err, "soso")
	assert.Equal(t, 2, len(bufs), "soso")
	assert.NotNil(t, bufs["pub/1990-12-31"], "aha")
	assert.Equal(t, `<?xml version="1.0" encoding="UTF-8"?>
<?xml-stylesheet type='text/xsl' href='../../assets/atom2html.xslt'?>
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
<feed xmlns="http://www.w3.org/2005/Atom">
  <title>Hello, Atom!</title>
  <id>http://example.com/pub/posts</id>
  <updated>1990-12-31T01:02:03+01:00</updated>
  <link href="http://example.com/pub/posts" rel="self"></link>
  <entry>
    <title>Hello, Entry!</title>
    <id>http://example.com/pub/posts/e0</id>
    <updated>1990-12-31T01:02:03+01:00</updated>
  </entry>
</feed>
`, string(bufs["pub/posts"].b), "soso")
}

func TestWriteFeedsPaged(t *testing.T) {
	bufs := make(map[string]*buff)
	// callback
	fctWriteCloser := func(uri string, page int) (io.WriteCloser, error) {
		bf := new(buff)
		// fmt.Printf("'%s'\n", appendPageNumber(uri, page))
		bufs[appendPageNumber(uri, page)] = bf
		return bf, nil
	}

	feed := &Feed{
		Id:    "http://example.com",
		Title: HumanText{Body: "Hello, Atom!"},
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
	err := feed.writeFeeds(2, fctWriteCloser)

	assert.Nil(t, err, "soso")
	assert.Equal(t, 4, len(bufs), "soso")
	assert.NotNil(t, bufs["pub/1990-12-30"], "aha")
	assert.NotNil(t, bufs["pub/1990-12-31"], "aha")
	assert.Equal(t, `<?xml version="1.0" encoding="UTF-8"?>
<?xml-stylesheet type='text/xsl' href='../../assets/atom2html.xslt'?>
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
<feed xmlns="http://www.w3.org/2005/Atom">
  <title>Hello, Atom!</title>
  <id>http://example.com/pub/posts</id>
  <updated>1990-12-31T02:02:02+01:00</updated>
  <link href="http://example.com/pub/posts" rel="self"></link>
  <link href="http://example.com/pub/posts" rel="first"></link>
  <link href="http://example.com/pub/posts-1" rel="next"></link>
  <link href="http://example.com/pub/posts-1" rel="last"></link>
  <entry>
    <title>Hello, Entry 2!</title>
    <id>http://example.com/pub/posts/e2</id>
    <updated>1990-12-31T02:02:02+01:00</updated>
  </entry>
  <entry>
    <title>Hello, Entry 1!</title>
    <id>http://example.com/pub/posts/e1</id>
    <updated>1990-12-31T01:01:01+01:00</updated>
  </entry>
</feed>
`, string(bufs["pub/posts"].b), "soso")

	assert.Equal(t, `<?xml version="1.0" encoding="UTF-8"?>
<?xml-stylesheet type='text/xsl' href='../../assets/atom2html.xslt'?>
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
<feed xmlns="http://www.w3.org/2005/Atom">
  <title>Hello, Atom!</title>
  <id>http://example.com/pub/posts</id>
  <updated>1990-12-31T02:02:02+01:00</updated>
  <link href="http://example.com/pub/posts-1" rel="self"></link>
  <link href="http://example.com/pub/posts" rel="first"></link>
  <link href="http://example.com/pub/posts" rel="previous"></link>
  <link href="http://example.com/pub/posts-1" rel="last"></link>
  <entry>
    <title>Hello, Entry 0!</title>
    <id>http://example.com/pub/posts/e0</id>
    <updated>1990-12-30T00:00:00+01:00</updated>
  </entry>
</feed>
`, string(bufs["pub/posts-1"].b), "soso")
}
