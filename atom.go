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
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
	// "golang.org/x/tools/blog/atom"

	"github.com/yhat/scrape"
	"golang.org/x/net/html"
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
	if read, err := os.Open(file); nil == read || nil != err {
		return Feed{}, err
	} else {
		defer read.Close()
		return FeedFromReader(read)
	}
}

func FeedFromReader(file io.Reader) (Feed, error) {
	un(trace("FeedFromReader"))
	ret := Feed{}
	err := xml.NewDecoder(file).Decode(&ret)
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
	if parse, err := time.Parse(time.RFC3339, v); err != nil {
		return err
	} else {
		*c = iso8601{parse}
		return nil
	}
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
	Lat, Lon float32
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

func xmlEncodeWithXslt(e interface{}, hrefXslt string, enc *xml.Encoder) error {
	var err error
	// preamble
	if err = enc.EncodeToken(xml.ProcInst{"xml", []byte(`version="1.0" encoding="UTF-8"`)}); err == nil {
		if err = enc.EncodeToken(xml.CharData("\n")); err == nil {
			if err = enc.EncodeToken(xml.ProcInst{"xml-stylesheet", []byte("type='text/xsl' href='" + hrefXslt + "'")}); err == nil {
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

func (feed *Feed) Append(e *Entry) (*Entry, error) {
	if err := e.Validate(); err != nil {
		return nil, err
	}
	// todo: pre-check uniqueness of Id
	feed.Entries = append(feed.Entries, e)
	return e, nil
}

// sort.Interface

type ByPublishedDesc []*Entry

func (a ByPublishedDesc) Len() int           { return len(a) }
func (a ByPublishedDesc) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByPublishedDesc) Less(i, j int) bool { return !a[i].Published.Time.Before(a[j].Published.Time) }

type ByUpdatedDesc []*Entry

func (a ByUpdatedDesc) Len() int           { return len(a) }
func (a ByUpdatedDesc) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByUpdatedDesc) Less(i, j int) bool { return !a[i].Updated.Time.Before(a[j].Updated.Time) }

// custom interface

func (feed *Feed) findEntry(doesMatch func(*Entry) bool) (int, *Entry) {
	defer un(trace(strings.Join([]string{"Feed.findEntry(f(*Entry))"}, "")))
	if nil != doesMatch {
		for idx, entry := range feed.Entries {
			if doesMatch(entry) {
				return idx, entry
			}
		}
	}
	return -1, nil
}

func (feed *Feed) findEntryById(id string) (int, *Entry) {
	defer un(trace(strings.Join([]string{"Feed.findEntryById('", id, "')"}, "")))
	if "" != id {
		return feed.findEntry(func(entry *Entry) bool { return id == entry.Id })
	}
	return feed.findEntry(nil)
}

func (feed *Feed) deleteEntry(id string) *Entry {
	if i, entry := feed.findEntryById(id); i >= 0 {
		a := feed.Entries
		// https://github.com/golang/go/wiki/SliceTricks
		copy(a[i:], a[i+1:])
		a[len(a)-1] = nil // or the zero value of T
		feed.Entries = a[:len(a)-1]
		return entry
	}
	return nil
}

func (feed Feed) Save(dst string) error {
	defer un(trace("Feed.Save"))
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

// Validate for storage
func (entry *Entry) Validate() error {
	if "" == entry.Id {
		return errors.New("Entry may not have empty Id.")
	}
	if 1 < len(entry.Links) {
		return fmt.Errorf("Entry may not have more than one link. Entry.Id='%s'", entry.Id)
	}
	if 1 == len(entry.Links) {
		if "" == entry.Links[0].Href {
			return fmt.Errorf("Entry may not have empty link. Entry.Id='%s'", entry.Id)
		}
		url := mustParseURL(entry.Links[0].Href)
		if !url.IsAbs() {
			return fmt.Errorf("Entry must have absolute Link. Entry.Id='%s'", entry.Id)
		}
		if "" == url.Host {
			return fmt.Errorf("Entry must have Link with non-empty host. Entry.Id='%s'", entry.Id)
		}
	}
	return nil
}

func (entry Entry) CategoriesMerged() []Category {
	a := entry.Title.Categories()
	b := entry.Content.Categories()
	ret := make([]Category, 0, len(a)+len(b)+len(entry.Categories))
	ret = append(ret, a...)
	ret = append(ret, b...)
	ret = append(ret, entry.Categories...)
	sort.Slice(ret, func(i, j int) bool { return strings.Compare(ret[i].Term, ret[j].Term) < 0 })
	return uniqCategory(ret)
}

func uniqCategory(data []Category) []Category {
	ret := make([]Category, 0, len(data))
	for i, e := range data {
		if "" == e.Term {
			continue
		}
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

// save string allocations but todo handle atEOF?
func tagsFromString1(str string) []string {
	scanner := bufio.NewScanner(strings.NewReader(str))
	split := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		advance, token, err = bufio.ScanWords(data, atEOF)
		if token != nil {
			if byte('#') == token[0] {
				token = token[1:] // de-prefix token
			} else {
				token = nil // drop token
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

func tagsFromString(str string) []string {
	scanner := bufio.NewScanner(strings.NewReader(str))
	scanner.Split(bufio.ScanWords)

	ret := make([]string, 0, 10)
	for scanner.Scan() {
		token := scanner.Text()
		if len(token) > 0 && byte('#') == token[0] {
			term := strings.TrimRightFunc(token[1:], unicode.IsPunct)
			if len(term) > 0 {
				ret = append(ret, term)
			}
		}
	}
	return ret
}

const iWillBeALineFeedMarker = "+,zX@D4X#%`lGdX-vWU?/==v"

func cleanLegacyContent(txt string) string {
	src := strings.Replace(txt, "<br />", iWillBeALineFeedMarker, -1)
	if node, err := html.Parse(strings.NewReader(src)); err == nil {
		str := strings.Replace(scrape.Text(node), iWillBeALineFeedMarker, "", -1)
		return strings.Trim(str[:len(str)-len("( Permalink )")], " ")
	} else {
		return err.Error()
	}
}
