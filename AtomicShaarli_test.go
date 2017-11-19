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
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
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
	// t.Log("sub test")
	assert.Nil(t, os.RemoveAll(dirTmp), "aha")
	assert.Nil(t, os.MkdirAll(dirTmp, 0700), "aha")
	cwd, _ := os.Getwd()
	os.Chdir(dirTmp)
	return func(t *testing.T) {
		// t.Log("sub test")
		os.Chdir(cwd)
		// assert.Nil(t, os.RemoveAll(dirTmp), "aha")
	}
}

func TestQueryParse(t *testing.T) {
	t.Parallel()
	u := mustParseURL("http://example.com/a/shaarligo.cgi?do=login&foo=bar&do=auch")

	assert.Equal(t, "http://example.com/a/shaarligo.cgi?do=login&foo=bar&do=auch", u.String(), "ach")
	assert.Equal(t, "do=login&foo=bar&do=auch", u.RawQuery, "ach")
	v := u.Query()

	assert.Equal(t, 2, len(v["do"]), "omg")
	assert.Equal(t, "login", v["do"][0], "omg")

	{
		parts := strings.Split("", "/")
		assert.Equal(t, 1, len(parts), "ja, genau")
		assert.Equal(t, "", parts[0], "ja, genau")
	}
	{
		parts := strings.Split("/config", "/")
		assert.Equal(t, 2, len(parts), "ja, genau")
		assert.Equal(t, "", parts[0], "ja, genau")
		assert.Equal(t, "config", parts[1], "ja, genau")
	}
}

func doHttp(method, path_info string) (*http.Response, error) {
	cgi := "shaarligo.cgi"
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

func doPost(path_info string, body []byte) (*http.Response, error) {
	fname := "stdin"
	if err := ioutil.WriteFile(fname, body, 0600); err != nil {
		panic(err)
	}
	old := os.Stdin
	temp, err := os.Open(fname)
	if err != nil {
		panic(err)
	}
	os.Stdin = temp
	defer func() { temp.Close(); os.Stdin = old }()

	os.Setenv("CONTENT_LENGTH", fmt.Sprintf("%d", len(body)))
	os.Setenv("CONTENT_TYPE", "application/x-www-form-urlencoded")
	ret, err := doHttp("POST", path_info)

	return ret, err
}

func TestGetConfigRaw(t *testing.T) {
	teardownTest := setupTest(t)
	defer teardownTest(t)

	r, err := doGet("/config/")

	assert.Nil(t, err, "aha")
	assert.Equal(t, http.StatusOK, r.StatusCode, "aha")
	assert.Equal(t, "200 OK", r.Status, "aha")
	assert.Equal(t, "go-test", r.Header["Server"][0], "aha")
	assert.Nil(t, r.Header["Status"], "aha")
	body, err := ioutil.ReadAll(r.Body)
	assert.Nil(t, err, "aha")
	assert.Equal(t, `<?xml version='1.0' encoding='UTF-8'?>
<?xml-stylesheet type='text/xsl' href='../../assets/default/de/config.xslt'?>
<!--
  The html you see here is for compatibilty with vanilla shaarli.

  The main reason is backward compatibility for e.g. https://github.com/mro/ShaarliOS and
  https://github.com/dimtion/Shaarlier as tested via
  https://github.com/mro/Shaarli-API-test
-->
<html xmlns="http://www.w3.org/1999/xhtml">
  <head/>
  <body>
    <form method="post" name="installform" id="installform">
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

	r, err := doGet("/config/")

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
	assert.Equal(t, "setlogin", scrape.Attr(all[0], "name"), "aha")
	assert.Equal(t, "setpassword", scrape.Attr(all[1], "name"), "aha")
	assert.Equal(t, "title", scrape.Attr(all[2], "name"), "aha")
	assert.Equal(t, "Save", scrape.Attr(all[3], "name"), "aha")
}

func TestPostConfig(t *testing.T) {
	teardownTest := setupTest(t)
	defer teardownTest(t)

	r, err := doPost("/config/", []byte(`title=A&setlogin=B&setpassword=123456789012&import_shaarli_url=&import_shaarli_setlogin=&import_shaarli_setpassword=`))

	assert.Nil(t, err, "aha")
	assert.Equal(t, http.StatusFound, r.StatusCode, "aha")
	assert.Equal(t, "/sub/pub/posts/", r.Header["Location"][0], "aha")

	body, err := ioutil.ReadAll(r.Body)
	assert.Nil(t, err, "aha")
	assert.Equal(t, 0, len(body), "soso")

	cfg, err := ioutil.ReadFile(filepath.Join("app", "config.yaml"))
	assert.Nil(t, err, "aha")
	assert.True(t, strings.HasPrefix(string(cfg), "title: A\nauthor_name: B\n"), string(cfg))

	assert.Equal(t, 1, len(r.Header["Set-Cookie"]), "naja")

	stat, _ := os.Stat("pub")
	assert.Equal(t, 0755, int(stat.Mode()&os.ModePerm), "ach, wieso?")
}

func TestGetLoginWithoutRedir(t *testing.T) {
	teardownTest := setupTest(t)
	defer teardownTest(t)

	r, err := doPost("/config/", []byte(`title=A&setlogin=B&setpassword=123456789012&import_shaarli_url=&import_shaarli_setlogin=&import_shaarli_setpassword=`))
	assert.Nil(t, err, "aha")
	assert.Equal(t, http.StatusFound, r.StatusCode, "aha")
	assert.Equal(t, "/sub/pub/posts/", r.Header["Location"][0], "aha")

	os.Setenv("QUERY_STRING", "do=login")
	r, err = doGet("")
	assert.Nil(t, err, "aha")
	assert.Equal(t, http.StatusOK, r.StatusCode, "aha")
	root, err := html.Parse(r.Body)
	assert.Nil(t, err, "aha")
	assert.NotNil(t, root, "aha")
	inputs := scrape.FindAll(root, func(n *html.Node) bool { return atom.Input == n.DataAtom })
	assert.Equal(t, 6, len(inputs), "aha")

	r, err = doPost("", []byte(`login=B&password=123456789012&token=foo`))
	assert.Equal(t, http.StatusFound, r.StatusCode, "aha")
	assert.Equal(t, "/sub/pub/posts/", r.Header["Location"][0], "aha")
	cook := r.Header["Set-Cookie"][0]
	assert.True(t, strings.HasPrefix(cook, "ShaarliGo=MTU"), cook)
}

func TestGetLoginWithRedir(t *testing.T) {
	teardownTest := setupTest(t)
	defer teardownTest(t)

	os.Unsetenv("COOKIE")
	r, err := doPost("/config/", []byte(`title=A&setlogin=B&setpassword=123456789012&import_shaarli_url=&import_shaarli_setlogin=&import_shaarli_setpassword=`))
	assert.Nil(t, err, "aha")
	assert.Equal(t, http.StatusFound, r.StatusCode, "aha")
	assert.Equal(t, "/sub/pub/posts/", r.Header["Location"][0], "aha")

	returnurl := "/sub/pub/posts/anyid/?foo=bar#baz"
	os.Setenv("QUERY_STRING", "do=login&returnurl="+url.QueryEscape(returnurl))
	r, err = doGet("")
	assert.Nil(t, err, "aha")
	assert.Equal(t, http.StatusOK, r.StatusCode, "aha")
	root, err := html.Parse(r.Body)
	assert.Nil(t, err, "aha")
	assert.NotNil(t, root, "aha")
	inputs := scrape.FindAll(root, func(n *html.Node) bool { return atom.Input == n.DataAtom })
	assert.Equal(t, 6, len(inputs), "aha")

	r, err = doPost("", []byte(`login=B&password=123456789012&token=foo&returnurl=/sub/pub/posts/anyid/?foo=bar#baz`))
	assert.Equal(t, http.StatusFound, r.StatusCode, "aha")
	assert.Equal(t, returnurl, r.Header["Location"][0], "aha")
	cook := r.Header["Set-Cookie"][0]
	assert.True(t, strings.HasPrefix(cook, "ShaarliGo=MTU"), cook)
}

func _TestGetPostNew(t *testing.T) {
	teardownTest := setupTest(t)
	defer teardownTest(t)

	r, err := doPost("/config", []byte(`title=A&setlogin=B&setpassword=123456789012&import_shaarli_url=&import_shaarli_setlogin=&import_shaarli_setpassword=`))
	assert.Nil(t, err, "aha")
	assert.Equal(t, http.StatusFound, r.StatusCode, "aha")
	assert.Equal(t, "/sub/pub/posts/", r.Header["Location"][0], "aha")

	purl := fmt.Sprintf("?post=%s&title=%s&source=%s", url.QueryEscape("http://example.com/foo?bar=baz#grr"), url.QueryEscape("A first post"), url.QueryEscape("me"))
	os.Setenv("QUERY_STRING", purl)
	r, err = doGet("")
	assert.Nil(t, err, "aha")
	assert.Equal(t, http.StatusFound, r.StatusCode, "aha")
	assert.Equal(t, "/sub/shaarligo.cgi?do=login", r.Header["Location"], "aha")

	r, err = doGet(fmt.Sprintf("?do=login&returnurl=/sub/shaarligo.cgi%s", url.QueryEscape(purl)))
	assert.Nil(t, err, "aha")
	assert.Equal(t, http.StatusOK, r.StatusCode, "aha")
	cook := r.Header["Set-Cookie"][0]
	assert.True(t, strings.HasPrefix(cook, "ShaarliGo=MTU"), cook)
	os.Setenv("COOKIE", cook)
	root, err := html.Parse(r.Body)
	assert.Nil(t, err, "aha")
	assert.NotNil(t, root, "aha")
	assert.Equal(t, 4, len(scrape.FindAll(root, func(n *html.Node) bool { return atom.Input == n.DataAtom })), "aha")

	r, err = doPost(fmt.Sprintf("?do=login&returnurl=/sub/shaarligo.cgi%s", url.QueryEscape(purl)), []byte(`login=B&password=123456789012`))
	os.Setenv("COOKIE", r.Header["Set-Cookie"][0])

	r, err = doGet(purl)
	assert.Equal(t, http.StatusOK, r.StatusCode, "aha")
	os.Setenv("COOKIE", r.Header["Set-Cookie"][0])
	root, err = html.Parse(r.Body)

	r, err = doPost(purl, nil)
	assert.Equal(t, http.StatusFound, r.StatusCode, "aha")
	assert.Equal(t, "/sub/pub/posts/?#foo", r.Header["Location"], "aha")
}
