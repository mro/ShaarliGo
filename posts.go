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

func newPostsHandler() http.Handler {
	return postsHandler{}
}

type postsHandler struct{}

func (postsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Server", myselfNamespace)
	w.Header().Set("Content-Type", "application/xhtml+xml; charset=utf-8")
	w.Header().Set("Handler", "postsHandler")
	// w.WriteHeader(http.StatusOK)

	enc := xml.NewEncoder(w)
	enc.Indent("", "  ")
	enc.EncodeToken(xml.ProcInst{"xml", []byte(`version="1.0" encoding="UTF-8"`)})
	enc.EncodeToken(xml.CharData("\n"))
	enc.EncodeToken(xml.ProcInst{"xml-stylesheet", []byte("type='text/xsl' href='../assets/setup.xslt'")})
	enc.EncodeToken(xml.CharData("\n"))
	enc.EncodeToken(xml.Comment(" Am Anfang war das Licht, dann kam bald das Atom! "))
	enc.EncodeToken(xml.CharData("\n"))
	enc.EncodeToken(xml.Comment(fmt.Sprintf(" Method: %s ", r.Method)))
	enc.EncodeToken(xml.CharData("\n"))
	enc.EncodeToken(xml.Comment(fmt.Sprintf(" TLS: %s ", r.TLS)))
	enc.EncodeToken(xml.CharData("\n"))
	n := xml.Name{Local: "setup"}
	enc.EncodeToken(xml.StartElement{Name: n})
	if false {
		writeElementKeyValue(enc, "form", "url", r.FormValue("url"))
	} else {
		err := r.ParseForm()
		if err != nil {
			enc.EncodeToken(xml.Comment(err.Error()))
		} else {
			for k, v := range r.PostForm {
				writeElementKeyValue(enc, "form", k, strings.Join(v, ""))
			}
		}
	}
	enc.EncodeToken(xml.EndElement{Name: n})
	enc.EncodeToken(xml.CharData("\n"))
	enc.Flush()
}
