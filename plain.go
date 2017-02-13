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

func (h plainHandler) ServeHTTP(wri http.ResponseWriter, req *http.Request) {
	wri.Header().Set("Server", "http://purl.mro.name/AtomicShaarli")
	wri.Header().Set("Content-Type", "text/plain; charset=utf-8")
	wri.Header().Set("Status", " 402 Heute nur für Stammgäste") // https://www.safaribooksonline.com/library/view/apache-cookbook/0596001916/ch09s02.html
	wri.Header().Set("My-Foo-Header", "My Bar Value")
	wri.WriteHeader(401)

	fmt.Fprintf(wri, "HTTP HEADER\n")
	for k, v := range req.Header {
		fmt.Fprintf(wri, "  %s: %s\n", k, v)
	}
	fmt.Fprintf(wri, "Method: %s\n", req.Method)
	fmt.Fprintf(wri, "URL: %s\n", req.URL)
	fmt.Fprintf(wri, "ContentLength: %d\n", req.ContentLength)
	fmt.Fprintf(wri, "Host: %s\n", req.Host)
	fmt.Fprintf(wri, "RemoteAddr: %s\n", req.RemoteAddr)
	fmt.Fprintf(wri, "RequestURI: %s\n", req.RequestURI)
	fmt.Fprintf(wri, "Referer: %s\n", req.Referer())
	// buf.WriteString(fmt.Sprintf("Referrer: %s\n", req.BasicAuth()))
	// fmt.Printf("Location: %s\n", "https://links.mro.name")
	fmt.Fprintf(wri, "\n")

	// wri.Write(h.buf.Bytes())
}
