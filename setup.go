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
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"
)

var _ = strings.Join

func newSetupHandler() http.Handler {
	return setupHandler{}
}

type setupHandler struct{}

func writeElementKeyValue(enc *xml.Encoder, element string, key string, value string) {
	a := []xml.Attr{xml.Attr{Name: xml.Name{Local: "name"}, Value: key}, xml.Attr{Name: xml.Name{Local: "content"}, Value: value}}
	n := xml.Name{Local: "meta"}
	enc.EncodeToken(xml.StartElement{Name: n, Attr: a})
	enc.EncodeToken(xml.EndElement{Name: n})
}

func (setupHandler) ServeHTTP(wri http.ResponseWriter, req *http.Request) {
	wri.Header().Set("Server", "http://purl.mro.name/AtomicShaarli")
	wri.Header().Set("Content-Type", "application/xhtml+xml; charset=utf-8")
	// wri.WriteHeader(http.StatusOK)

	enc := xml.NewEncoder(wri)
	enc.Indent("", "  ")
	enc.EncodeToken(xml.ProcInst{"xml", []byte(`version="1.0" encoding="UTF-8"`)})
	enc.EncodeToken(xml.CharData("\n"))
	enc.EncodeToken(xml.ProcInst{"xml-stylesheet", []byte("type='text/xsl' href='../assets/setup.xslt'")})
	enc.EncodeToken(xml.CharData("\n"))
	enc.EncodeToken(xml.Comment(" Am Anfang war das Licht, dann kam bald das Atom! "))
	enc.EncodeToken(xml.CharData("\n"))
	enc.EncodeToken(xml.Comment(fmt.Sprintf(" Method: %s ", req.Method)))
	enc.EncodeToken(xml.CharData("\n"))
	enc.EncodeToken(xml.Comment(fmt.Sprintf(" TLS: %s ", req.TLS)))
	enc.EncodeToken(xml.CharData("\n"))
	n := xml.Name{Local: "setup"}
	enc.EncodeToken(xml.StartElement{Name: n})
	if false {
		writeElementKeyValue(enc, "form", "url", req.FormValue("url"))
	} else {
		err := req.ParseForm()
		if err != nil {
			enc.EncodeToken(xml.Comment(err.Error()))
		} else {
			for k, v := range req.PostForm {
				writeElementKeyValue(enc, "form", k, strings.Join(v, ""))
			}
		}
	}
	enc.EncodeToken(xml.EndElement{Name: n})
	enc.EncodeToken(xml.CharData("\n"))
	enc.Flush()
}

func loadSetup() (interface{}, error) {
	return nil, nil //, errors.New("Not set up yet.")
}
