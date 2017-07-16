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
	"fmt"
	"net/http"
)

func newPlainHandler() http.Handler {
	return plainHandler{}
}

type plainHandler struct{}

func (h plainHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Server", myselfNamespace)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Status", " 402 Heute nur für Stammgäste") // https://www.safaribooksonline.com/library/view/apache-cookbook/0596001916/ch09s02.html
	w.Header().Set("My-Foo-Header", "My Bar Value")
	w.Header().Set("Handler", "plainHandler")
	w.WriteHeader(401)

	fmt.Fprintf(w, "HTTP HEADER\n")
	for k, v := range r.Header {
		fmt.Fprintf(w, "  %s: %s\n", k, v)
	}
	fmt.Fprintf(w, "Method: %s\n", r.Method)
	fmt.Fprintf(w, "URL: %s\n", r.URL)
	fmt.Fprintf(w, "ContentLength: %d\n", r.ContentLength)
	fmt.Fprintf(w, "Host: %s\n", r.Host)
	fmt.Fprintf(w, "RemoteAddr: %s\n", r.RemoteAddr)
	fmt.Fprintf(w, "RequestURI: %s\n", r.RequestURI)
	fmt.Fprintf(w, "Referer: %s\n", r.Referer())
	// buf.WriteString(fmt.Sprintf("Referrer: %s\n", r.BasicAuth()))
	// fmt.Printf("Location: %s\n", "https://links.mro.name")
	fmt.Fprintf(w, "\n")

	// w.Write(h.buf.Bytes())
	// w.Flush()
}
