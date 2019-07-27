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
	"encoding/base64"
	"encoding/binary"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
	// "golang.org/x/tools/blog/atom"

	"github.com/yhat/scrape"
	"golang.org/x/net/html"
)

const lengthyAtomPreambleComment string = `
  https://developer.mozilla.org/en/docs/XSL_Transformations_in_Mozilla_FAQ#Why_isn.27t_my_stylesheet_applied.3F

  Caution! Firefox ignores your XSLT stylesheet if your XML looks like a RSS or Atom feed. A typical workaround is to insert an XML comment at the beginning of your XML file to move the <fEEd or <rsS tag out of the first 512 bytes used by Firefox to guess whether it is a feed or not.

  See also the discussion at https://bugzilla.mozilla.org/show_bug.cgi?id=338621#c72.

  For best results, serve both atom feed and xslt as 'text/xml' or 'application/xml' without charset specified.
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
	ret := Feed{}
	err := xml.NewDecoder(file).Decode(&ret)
	return ret, err
}

type Iri string      // https://tools.ietf.org/html/rfc3987
type Id Iri          // we allow relative Ids (in persistent store)
type Lang string     // https://tools.ietf.org/html/rfc3066
type Relation string // https://www.iana.org/assignments/link-relations/link-relations.xhtml#link-relations-1
type MimeType string // https://tools.ietf.org/html/rfc2045#section-5.1
type TextType string // https://tools.ietf.org/html/rfc4287#section-4.1.3.1

// https://mro.github.io/atomenabled.org/
// https://tools.ietf.org/html/rfc4287#section-4.1.1
//
// see also https://godoc.org/golang.org/x/tools/blog/atom#Feed
type Feed struct {
	XMLName         xml.Name   `xml:"http://www.w3.org/2005/Atom feed"`
	XmlBase         Iri        `xml:"xml:base,attr,omitempty"`
	XmlLang         Lang       `xml:"xml:lang,attr,omitempty"`
	XmlNSShaarliGo  string     `xml:"xmlns:sg,attr,omitempty"`         // https://github.com/golang/go/issues/9519#issuecomment-252196382
	SearchTerms     string     `xml:"sg:searchTerms,attr,omitempty"`   // rather use http://www.opensearch.org/Specifications/OpenSearch/1.1#Example_of_OpenSearch_response_elements_in_Atom_1.0
	XmlNSOpenSearch string     `xml:"xmlns:opensearch,attr,omitempty"` // https://github.com/golang/go/issues/9519#issuecomment-252196382
	Query           string     `xml:"opensearch:Query,omitempty"`      // http://www.opensearch.org/Specifications/OpenSearch/1.1#Example_of_OpenSearch_response_elements_in_Atom_1.0
	Title           HumanText  `xml:"title"`
	Subtitle        *HumanText `xml:"subtitle,omitempty"`
	Id              Id         `xml:"id"`
	Updated         iso8601    `xml:"updated"`
	Generator       *Generator `xml:"generator,omitempty"`
	Icon            Iri        `xml:"icon,omitempty"`
	Logo            Iri        `xml:"logo,omitempty"`
	Links           []Link     `xml:"link"`
	Categories      []Category `xml:"category"`
	Authors         []Person   `xml:"author"`
	Contributors    []Person   `xml:"contributor"`
	Rights          *HumanText `xml:"rights,omitempty"`
	Entries         []*Entry   `xml:"entry"`
}

type Generator struct {
	Uri     Iri    `xml:"uri,attr"`
	Version string `xml:"version,attr,omitempty"`
	Body    string `xml:",chardata"`
}

// http://stackoverflow.com/a/25015260
type iso8601 time.Time

func (v iso8601) IsZero() bool             { return time.Time(v).IsZero() }
func (a iso8601) After(b iso8601) bool     { return time.Time(a).After(time.Time(b)) }
func (a iso8601) Before(b iso8601) bool    { return time.Time(a).Before(time.Time(b)) }
func (a iso8601) Format(fmt string) string { return time.Time(a).Format(fmt) }

func (v iso8601) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	e.EncodeElement(v.Format(time.RFC3339), start)
	return nil
}

func (c *iso8601) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var v string
	d.DecodeElement(&v, &start)
	if parse, err := time.Parse(time.RFC3339, v); err != nil {
		return err
	} else {
		*c = iso8601(parse)
		return nil
	}
}

// see also https://godoc.org/golang.org/x/tools/blog/atom#Link
type Link struct {
	Href     string   `xml:"href,attr"`
	Rel      Relation `xml:"rel,attr,omitempty"`
	Type     MimeType `xml:"type,attr,omitempty"`
	HrefLang Lang     `xml:"hreflang,attr,omitempty"`
	Title    string   `xml:"title,attr,omitempty"`
	Length   int64    `xml:"length,attr,omitempty"`
}

// see also https://godoc.org/golang.org/x/tools/blog/atom#Person
type Person struct {
	Name  string `xml:"name"`
	Email string `xml:"email,omitempty"`
	Uri   Iri    `xml:"uri,omitempty"`
}

// see also https://godoc.org/golang.org/x/tools/blog/atom#Entry
type Entry struct {
	XMLName      xml.Name   `xml:"http://www.w3.org/2005/Atom entry,omitempty"`
	XmlBase      Iri        `xml:"xml:base,attr,omitempty"`
	XmlLang      Lang       `xml:"xml:lang,attr,omitempty"`
	Title        HumanText  `xml:"title"`
	Summary      *HumanText `xml:"summary,omitempty"`
	Id           Id         `xml:"id"`
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
	XmlLang Lang     `xml:"xml:lang,attr,omitempty"`
	Body    string   `xml:",chardata"`
	Type    TextType `xml:"type,attr,omitempty"`
	Src     Iri      `xml:"src,attr,omitempty"`
}

type Category struct {
	Term   string `xml:"term,attr"`
	Scheme Iri    `xml:"scheme,attr,omitempty"`
	Label  string `xml:"label,attr,omitempty"`
}

type MediaThumbnail struct {
	Url Iri `xml:"url,attr"`
}

type Latitude float32
type Longitude float32

type GeoRssPoint struct {
	Lat Latitude
	Lon Longitude
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
	*c = GeoRssPoint{Lat: Latitude(lat), Lon: Longitude(lon)}
	return nil
}

func xmlEncodeWithXslt(e interface{}, hrefXslt string, enc *xml.Encoder) error {
	var err error
	// preamble
	if err = enc.EncodeToken(xml.ProcInst{Target: "xml", Inst: []byte(`version="1.0" encoding="UTF-8"`)}); err == nil {
		if err = enc.EncodeToken(xml.CharData("\n")); err == nil {
			if err = enc.EncodeToken(xml.ProcInst{Target: "xml-stylesheet", Inst: []byte("type='text/xsl' href='" + hrefXslt + "'")}); err == nil {
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
func (a ByPublishedDesc) Less(i, j int) bool { return !a[i].Published.Before(a[j].Published) }

type ByUpdatedDesc []*Entry

func (a ByUpdatedDesc) Len() int           { return len(a) }
func (a ByUpdatedDesc) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByUpdatedDesc) Less(i, j int) bool { return !a[i].Updated.Before(a[j].Updated) }

// custom interface

// sufficient for 32 bit.
func base64ToBase24x7(b64 string) (string, error) {
	if data, err := base64.RawURLEncoding.DecodeString(b64); err != nil {
		return "", err
	} else {
		// check len(data) ?
		ui32 := binary.LittleEndian.Uint32(data)
		base24 := fmt.Sprintf("%07s", strconv.FormatUint(uint64(ui32), 24))
		return strings.Map(mapBase24ToSuperCareful, base24), nil
	}
}

// Being "super-careful" https://code.mro.name/mro/ProgrammableWebSwartz2013/src/master/content/pages/2-building-for-users.md
//
// 0123456789abcdefghijklmn ->
// 23456789abcdefghkrstuxyz
func mapBase24ToSuperCareful(r rune) rune {
	digits := []rune("23456789abcdefghkrstuxyz")
	switch {
	case '0' <= r && r <= '9':
		return digits[:10][r-'0']
	case r >= 'a' && r <= 'n':
		return digits[10:][r-'a']
	}
	panic("ouch")
}

func newRandomId(t time.Time) Id {
	ui32 := uint32(t.Unix() & 0xFFFFFFFF) // unix time in seconds as uint32
	base24 := fmt.Sprintf("%07s", strconv.FormatUint(uint64(ui32), 24))
	return Id(strings.Map(mapBase24ToSuperCareful, base24))
}

func (feed Feed) newUniqueId(t time.Time) Id {
	id := newRandomId(t)
	for _, entry := range feed.Entries {
		if entry.Id == id {
			panic("id not unique")
		}
	}
	return id
}

func (feed Feed) newEntry(t time.Time) *Entry {
	defer un(trace("Feed.newEntry(t)"))
	return &Entry{
		Authors:   feed.Authors,
		Published: iso8601(t),
		Id:        feed.newUniqueId(t),
	}
}

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

func (feed *Feed) findEntryById(id Id) (int, *Entry) {
	defer un(trace(strings.Join([]string{"Feed.findEntryById('", string(id), "')"}, "")))
	if "" != id {
		return feed.findEntry(func(entry *Entry) bool { return id == entry.Id })
	}
	return feed.findEntry(nil)
}

func (feed *Feed) deleteEntryById(id Id) *Entry {
	if i, entry := feed.findEntryById(id); i < 0 {
		return nil
	} else {
		a := feed.Entries
		// https://github.com/golang/go/wiki/SliceTricks
		copy(a[i:], a[i+1:])
		// a[len(a)-1] = nil // or the zero value of T
		feed.Entries = a[:len(a)-1]
		feed.Updated = iso8601(time.Now())

		// don' try to be smart. When removing days feeds, we rely on correct Published date.
		// entry.Published = iso8601{time.Time{}}
		// entry.Updated = entry.Published

		return entry
	}
}

func (feed Feed) SaveToFile(dst string) error {
	defer un(trace("Feed.SaveToFile"))
	sort.Sort(ByPublishedDesc(feed.Entries))
	// remove deleted entries? Maybe Published date zero.

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

func AggregateCategories(entries []*Entry) []Category {
	// aggregate & count feed entry categories
	cats := make(map[string]int, 1*len(entries)) // raw len guess
	for _, ent := range entries {
		for _, cat := range ent.Categories {
			cats[cat.Term] += 1
		}
	}
	cs := make([]Category, 0, len(cats))
	for term, count := range cats {
		if term != "" && count != 0 {
			cs = append(cs, Category{Term: term, Label: strconv.Itoa(count)})
		}
	}
	sort.Slice(cs, func(i, j int) bool {
		return strings.Compare(cs[i].Term, cs[j].Term) < 0
	})
	return cs
}

func (ht HumanText) Categories() []Category {
	ret := make([]Category, 0, 10)
	for _, t := range tagsFromString(ht.Body) {
		ret = append(ret, Category{Term: t})
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
