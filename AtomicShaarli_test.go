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
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/yhat/scrape"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/stretchr/testify/assert"
	"testing"
)

const dirTmp = "go-test~"

// https://stackoverflow.com/a/42310257
func setupTest(t *testing.T) func(t *testing.T) {
	t.Log("sub test")
	assert.Nil(t, os.RemoveAll(dirTmp), "aha")
	assert.Nil(t, os.MkdirAll(dirTmp, 0700), "aha")
	cwd, _ := os.Getwd()
	os.Chdir(dirTmp)
	return func(t *testing.T) {
		t.Log("sub test")
		os.Chdir(cwd)
		// assert.Nil(t, os.RemoveAll(dirTmp), "aha")
	}
}

func TestQueryParse(t *testing.T) {
	t.Parallel()
	u := mustParseURL("http://example.com/a/atom.cgi?do=login&foo=bar&do=auch")

	assert.Equal(t, "http://example.com/a/atom.cgi?do=login&foo=bar&do=auch", u.String(), "ach")
	assert.Equal(t, "do=login&foo=bar&do=auch", u.RawQuery, "ach")
	v := u.Query()

	assert.Equal(t, 2, len(v["do"]), "omg")
	assert.Equal(t, "login", v["do"][0], "omg")
}

func doHttp(method, path_info string) (*http.Response, error) {
	cgi := "atom.cgi"
	os.Setenv("SCRIPT_NAME", "/sub/"+cgi)
	os.Setenv("SERVER_PROTOCOL", "HTTP/1.1")
	os.Setenv("HTTP_HOST", "example.com")

	os.Setenv("REQUEST_METHOD", method)
	os.Setenv("PATH_INFO", path_info)

	fname := "stdout"
	old := os.Stdout
	temp, _ := os.Create(fname)
	os.Stdout = temp
	defer func() { temp.Close(); os.Stdout = old }()

	fmt.Print("HTTP/1.1 600 Overwrite me asap.\r\n")
	fmt.Print("Server: go-test\r\n")
	main()
	temp.Close()

	if f, err := os.Open(fname); err == nil {
		if ret, err := http.ReadResponse(bufio.NewReader(f), nil); err == nil {
			ret.Status = ret.Header["Status"][0]
			if i, err := strconv.Atoi(strings.SplitN(ret.Status, " ", 2)[0]); err == nil {
				delete(ret.Header, "Status")
				ret.StatusCode = i
				return ret, err
			} else {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

func doGet(path_info string) (*http.Response, error) {
	return doHttp("GET", path_info)
}

func TestGetConfigRaw(t *testing.T) {
	teardownTest := setupTest(t)
	defer teardownTest(t)

	r, err := doGet("/config")

	assert.Nil(t, err, "aha")
	assert.Equal(t, http.StatusOK, r.StatusCode, "aha")
	assert.Equal(t, "200 OK", r.Status, "aha")
	assert.Equal(t, "go-test", r.Header["Server"][0], "aha")
	assert.Nil(t, r.Header["Status"], "aha")
	body, err := ioutil.ReadAll(r.Body)
	assert.Nil(t, err, "aha")
	assert.Equal(t, `<?xml version='1.0' encoding='UTF-8'?>
<?xml-stylesheet type='text/xsl' href='../assets/default/de/config.xslt'?>
<!--
  The html you see here is for compatibilty with vanilla shaarli.

  The main reason is backward compatibility for e.g. https://github.com/mro/ShaarliOS and
  https://github.com/dimtion/Shaarlier as tested via
  https://github.com/mro/Shaarli-API-test
-->
<html xmlns="http://www.w3.org/1999/xhtml">
  <head/>
  <body>
    <form method="post" action="#" name="installform" id="installform">
      <input type="text" name="setlogin" value=""/>
      <input type="password" name="setpassword" />
      <input type="text" name="title" value=""/>
      <input type="submit" name="Save" value="Save config" />
    </form>
  </body>
</html>
`, string(body), "aha")
}

func TestGetConfigScraped(t *testing.T) {
	teardownTest := setupTest(t)
	defer teardownTest(t)

	r, err := doGet("/config")

	assert.Nil(t, err, "aha")
	assert.Equal(t, http.StatusOK, r.StatusCode, "aha")
	assert.Equal(t, "200 OK", r.Status, "aha")
	assert.Equal(t, "go-test", r.Header["Server"][0], "aha")
	assert.Nil(t, r.Header["Status"], "aha")

	root, err := html.Parse(r.Body)
	assert.Nil(t, err, "aha")
	assert.NotNil(t, root, "aha")

	all := scrape.FindAll(root, func(n *html.Node) bool { return atom.Input == n.DataAtom })
	assert.Equal(t, 4, len(all), "aha")
	assert.Equal(t, "text", scrape.Attr(all[0], "type"), "aha")
	assert.Equal(t, "password", scrape.Attr(all[1], "type"), "aha")
	assert.Equal(t, "text", scrape.Attr(all[2], "type"), "aha")
	assert.Equal(t, "submit", scrape.Attr(all[3], "type"), "aha")
}
