//
// Copyright (C) 2017-2017 Marcus Rohrmoser, http://purl.mro.name/ShaarliGo
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
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
	// "golang.org/x/tools/blog/atom"
)

const lengthyAtomPreambleComment string = `
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
`

const atomNamespace = "http://www.w3.org/2005/Atom"

func FeedFromFileName(file string) (Feed, error) {
	read, err := os.Open(file)
	if read == nil {
		return Feed{}, err
	}
	defer read.Close()
	// read, err := bufio.NewReader(file)
	return FeedFromReader(read)
}

func FeedFromReader(file io.Reader) (Feed, error) {
	ret := Feed{}
	err := xml.NewDecoder(file).Decode(&ret)
	if nil != err {
		return Feed{}, err
	}
	return ret, err
}

// http://atomenabled.org/developers/syndication/
//
// see also https://godoc.org/golang.org/x/tools/blog/atom#Feed
type Feed struct {
	XMLName      xml.Name   `xml:"http://www.w3.org/2005/Atom feed"`
	XmlBase      string     `xml:"xml:base,attr,omitempty"`
	XmlLang      string     `xml:"xml:lang,attr,omitempty"`
	Title        HumanText  `xml:"title"`
	Subtitle     *HumanText `xml:"subtitle,omitempty"`
	Id           string     `xml:"id"`
	Updated      iso8601    `xml:"updated"`
	Generator    *Generator `xml:"generator,omitempty"`
	Icon         string     `xml:"icon,omitempty"`
	Logo         string     `xml:"logo,omitempty"`
	Links        []Link     `xml:"link"`
	Categories   []Category `xml:"category"`
	Authors      []Person   `xml:"author"`
	Contributors []Person   `xml:"contributor"`
	Rights       *HumanText `xml:"rights,omitempty"`
	Entries      []*Entry   `xml:"entry"`
}

type Generator struct {
	Uri     string `xml:"uri,attr"`
	Version string `xml:"version,attr,omitempty"`
	Body    string `xml:",chardata"`
}

// http://stackoverflow.com/a/25015260
type iso8601 struct{ time.Time }

func (c *iso8601) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var v string
	d.DecodeElement(&v, &start)
	parse, err := time.Parse(time.RFC3339, v)
	if err != nil {
		return err
	}
	*c = iso8601{parse}
	return nil
}

// see also https://godoc.org/golang.org/x/tools/blog/atom#Link
type Link struct {
	Href     string `xml:"href,attr"`
	Rel      string `xml:"rel,attr,omitempty"`
	Type     string `xml:"type,attr,omitempty"`
	HrefLang string `xml:"hreflang,attr,omitempty"`
	Title    string `xml:"title,attr,omitempty"`
	Length   int64  `xml:"length,attr,omitempty"`
}

// see also https://godoc.org/golang.org/x/tools/blog/atom#Person
type Person struct {
	Name  string `xml:"name"`
	Email string `xml:"email,omitempty"`
	Uri   string `xml:"uri,omitempty"`
}

// see also https://godoc.org/golang.org/x/tools/blog/atom#Entry
type Entry struct {
	XMLName      xml.Name   `xml:"http://www.w3.org/2005/Atom entry,omitempty"`
	XmlBase      string     `xml:"xml:base,attr,omitempty"`
	XmlLang      string     `xml:"xml:lang,attr,omitempty"`
	Title        HumanText  `xml:"title"`
	Summary      *HumanText `xml:"summary,omitempty"`
	Id           string     `xml:"id"`
	Updated      iso8601    `xml:"updated"`
	Published    iso8601    `xml:"published,omitempty"`
	Links        []Link     `xml:"link"`
	Categories   []Category `xml:"category"`
	Authors      []Person   `xml:"author"`
	Contributors []Person   `xml:"contributor"`
	Content      *HumanText `xml:"content"`
	// Vorsicht! beim Schreiben (Marshal/Encode) fuchst's noch: https://github.com/golang/go/issues/9519#issuecomment-252196382
	MediaThumbnail *MediaThumbnail `xml:"http://search.yahoo.com/mrss/ thumbnail,omitempty"`
	GeoRssPoint    *GeoRssPoint    `xml:"http://www.georss.org/georss point,omitempty"`
}

type HumanText struct {
	XmlLang string `xml:"xml:lang,attr,omitempty"`
	Body    string `xml:",chardata"`
	Type    string `xml:"type,attr,omitempty"`
	Src     string `xml:"src,attr,omitempty"`
}

type Category struct {
	Term   string `xml:"term,attr"`
	Scheme string `xml:"scheme,attr,omitempty"`
	Label  string `xml:"label,attr,omitempty"`
}

type MediaThumbnail struct {
	Url string `xml:"url,attr"`
}

type GeoRssPoint struct {
	Lat float32
	Lon float32
}

func (v GeoRssPoint) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	e.EncodeElement(fmt.Sprintf("%f %f", v.Lat, v.Lon), start)
	return nil
}

func (c *GeoRssPoint) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var v string
	d.DecodeElement(&v, &start)
	res := strings.SplitN(v, " ", 2)
	if len(res) != 2 {
		return errors.New("Not a proper 'lat lon' pair.")
	}
	lat, err := strconv.ParseFloat(res[0], 32)
	if err != nil {
		return err
	}
	lon, err := strconv.ParseFloat(res[1], 32)
	if err != nil {
		return err
	}
	*c = GeoRssPoint{Lat: float32(lat), Lon: float32(lon)}
	return nil
}

func xmlEncodeWithXslt(e interface{}, xslt string, enc *xml.Encoder) error {
	var err error
	// preamble
	if err = enc.EncodeToken(xml.ProcInst{"xml", []byte(`version="1.0" encoding="UTF-8"`)}); err == nil {
		if err = enc.EncodeToken(xml.CharData("\n")); err == nil {
			if err = enc.EncodeToken(xml.ProcInst{"xml-stylesheet", []byte("type='text/xsl' href='" + xslt + "'")}); err == nil {
				if err = enc.EncodeToken(xml.CharData("\n")); err == nil {
					if err = enc.EncodeToken(xml.Comment(lengthyAtomPreambleComment)); err == nil {
						if err = enc.EncodeToken(xml.CharData("\n")); err == nil {
							if err = enc.Encode(e); err == nil {
								err = enc.EncodeToken(xml.CharData("\n"))
							}
						}
					}
				}
			}
		}
	}
	return err
}

func (feed *Feed) Append(e *Entry) *Feed {
	feed.Entries = append(feed.Entries, e)
	return feed
}

// sort.Interface

type ByPublishedDesc []*Entry

func (a ByPublishedDesc) Len() int           { return len(a) }
func (a ByPublishedDesc) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByPublishedDesc) Less(i, j int) bool { return !a[i].Published.Time.Before(a[j].Published.Time) }

// custom interface

func (feed *Feed) findOrCreateEntryForURL(url *url.URL, now time.Time, doAppend bool) *Entry {
	if url != nil {
		ur := url.String()
		for _, entry := range feed.Entries {
			for _, l := range entry.Links {
				if l.Rel == "" && l.Href == ur {
					return entry
				}
			}
		}
	}
	ret := &Entry{
		Published: iso8601{now},
	}
	if url != nil {
		ret.Links = []Link{Link{Href: url.String()}}
	}
	if doAppend {
		feed.Append(ret)
	}
	return ret
}

func (feed Feed) Save(dst string) error {
	defer un(trace("Feed.Save"))
	// sort.Sort(ByPublishedDesc(feed.Entries))
	sort.Slice(feed.Entries, func(i, j int) bool { return !feed.Entries[i].Published.Time.Before(feed.Entries[j].Published.Time) })
	tmp := dst + "~"
	var err error
	var w *os.File
	if w, err = os.Create(tmp); err == nil {
		enc := xml.NewEncoder(w)
		enc.Indent("", "  ")
		if err = enc.Encode(feed); err == nil {
			if err = enc.Flush(); err == nil {
				if err = w.Close(); err == nil {
					if err = os.Rename(dst, dst+".bak"); err == nil || os.IsNotExist(err) {
						if err = os.Rename(tmp, dst); err == nil {
							return nil
						}
					}
				}
			}
		}
	}
	return err
}

func (entry Entry) CategoriesMerged() []Category {
	a := entry.Title.Categories()
	b := entry.Content.Categories()
	c := entry.Categories
	ret := make([]Category, 0, len(a)+len(b)+len(c))
	ret = append(ret, a...)
	ret = append(ret, b...)
	ret = append(ret, c...)
	sort.Slice(ret, func(i, j int) bool { return strings.Compare(ret[i].Term, ret[j].Term) < 0 })
	return uniqCategory(ret)
}

func uniqCategory(data []Category) []Category {
	ret := make([]Category, 0, len(data))
	for i, e := range data {
		if i == 0 || e.Term != data[i-1].Term {
			ret = append(ret, e)
		}
	}
	return ret
}

func (ht HumanText) Categories() []Category {
	ret := make([]Category, 0, 10)
	for _, t := range tagsFromString(ht.Body) {
		ret = append(ret, Category{Term: t})
	}
	return ret
}

func tagsFromString(str string) []string {
	scanner := bufio.NewScanner(strings.NewReader(str))
	split := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		advance, token, err = bufio.ScanWords(data, atEOF)
		if err == nil && token != nil {
			if token[0] == byte('#') {
				token = token[1:]
			} else {
				token = nil
			}
		}
		return
	}
	scanner.Split(split)

	ret := make([]string, 0, 10)
	for scanner.Scan() {
		t := scanner.Text()
		ret = append(ret, strings.TrimRightFunc(t, unicode.IsPunct))
	}
	return ret
}
