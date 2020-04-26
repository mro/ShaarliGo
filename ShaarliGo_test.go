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
	"bufio"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/yhat/scrape"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"testing"
)

const dirTmp = "go-test~" // volatile cwd while testing

// https://stackoverflow.com/a/42310257
func prepTeardown(t *testing.T) func() {
	// t.Log("sub test [")
	assert.Nil(t, os.RemoveAll(dirTmp), "aha")
	assert.Nil(t, os.MkdirAll(dirTmp, 0700), "aha")
	cwd, _ := os.Getwd()

	os.Chdir(dirTmp)
	return func() {
		// t.Log("] sub test")
		os.Chdir(cwd)
		assert.Nil(t, os.RemoveAll(dirTmp), "aha")
	}
}

func TestQueryParse(t *testing.T) {
	t.Parallel()
	u := mustParseURL("http://example.com/a/shaarli.cgi?do=login&foo=bar&do=auch")

	assert.Equal(t, "http://example.com/a/shaarli.cgi?do=login&foo=bar&do=auch", u.String(), "ach")
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

// non-Ascii paths and Cookies...
func TestUrlParseµ(t *testing.T) {
	t.Parallel()
	u := mustParseURL("http://example.com/µ/")
	assert.Equal(t, "/µ/", u.Path, "omg")
	assert.Equal(t, "/%C2%B5/", u.EscapedPath(), "omg")
}

func TestGetConfigRaw(t *testing.T) {
	defer prepTeardown(t)()

	pi := "/config/"
	os.Setenv("PATH_INFO", pi)
	ts := httptest.NewServer(handleMux(&sync.WaitGroup{}))
	defer ts.Close()
	c := http.Client{Timeout: time.Second}

	r, _ := c.Get(ts.URL + pi)
	assert.Equal(t, http.StatusOK, r.StatusCode, "aha")

	b, _ := ioutil.ReadAll(r.Body)
	assert.Equal(t, xml.Header+`<?xml-stylesheet type='text/xsl' href='../../themes/current/config.xslt'?>
<!--
  The html you see here is for compatibility with vanilla shaarli.

  The main reason is backward compatibility for e.g. http://app.mro.name/ShaarliOS and
  https://github.com/dimtion/Shaarlier as tested via
  https://code.mro.name/mro/Shaarli-API-test
-->
<html xmlns="http://www.w3.org/1999/xhtml">
  <head/>
  <body>
    <form method="post" name="configform" id="configform">
      <input type="text" name="setlogin" value=""/>
      <input type="password" name="setpassword" />
      <input type="text" name="title" value=""/>
      <input type="submit" name="Save" value="Save config" />
    </form>
  </body>
</html>
`, string(b), "aha")

	fi, _ := os.Stat(fileFeedStorage)
	assert.Equal(t, int64(1025), fi.Size(), "uhu")

	_, err := os.Stat(filepath.Join("tpl", "tools.html"))
	assert.Equal(t, true, os.IsNotExist(err), "oje")
}

func TestGetConfigScraped(t *testing.T) {
	defer prepTeardown(t)()

	pi := "/config/"
	os.Setenv("PATH_INFO", pi)
	ts := httptest.NewServer(handleMux(&sync.WaitGroup{}))
	defer ts.Close()
	c := http.Client{Timeout: time.Second}

	r, _ := c.Get(ts.URL + pi)
	assert.Equal(t, http.StatusOK, r.StatusCode, "aha")
	assert.Equal(t, "200 OK", r.Status, "aha")
	fo, _ := formValuesFromReader(r.Body, "configform")
	assert.Equal(t, url.Values(url.Values{
		"setlogin":    []string{""},
		"setpassword": []string{""},
		"title":       []string{""},
		"Save":        []string{"Save config"},
	}), fo, "aha")
}

func TestHttpServer(t *testing.T) {
	defer prepTeardown(t)()

	var u *url.URL
	// Using this server doesn't result in absolute request urls as is the case running as CGI.
	// And Atom needs absolute urls.
	// Shaarligo relies on them, however.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%s", u.String())
	}))
	defer ts.Close()
	u, _ = url.Parse(ts.URL)

	c := http.Client{Timeout: time.Second}
	re, _ := c.Get(ts.URL + "/uhu")
	b, _ := ioutil.ReadAll(re.Body)
	assert.Equal(t, ts.URL, string(b), "aha")
}

func _TestPostConfigG(t *testing.T) {
	defer prepTeardown(t)()

	cgi := "/sub/shaarli.cgi"
	pi := "/config/"
	os.Setenv("SCRIPT_NAME", cgi)
	os.Setenv("PATH_INFO", pi)

	ts := httptest.NewServer(handleMux(&sync.WaitGroup{}))
	defer ts.Close()

	c := http.Client{Timeout: time.Second}

	r, err := c.PostForm(ts.URL+cgi+pi, url.Values{
		"title":                      []string{"A"},
		"setlogin":                   []string{"B"},
		"setpassword":                []string{"123456789012"},
		"import_shaarli_url":         []string{""},
		"import_shaarli_setlogin":    []string{""},
		"import_shaarli_setpassword": []string{""},
	})

	assert.Equal(t, nil, err, "aha")
	assert.Equal(t, http.StatusFound, r.StatusCode, "aha")
	assert.Equal(t, "/sub/"+uriPubPosts, r.Header["Location"], "aha")

	body, _ := ioutil.ReadAll(r.Body)
	assert.Equal(t, "", string(body), "soso")

	cfg, err := ioutil.ReadFile(filepath.Join(dirApp, "config.yaml"))
	assert.Nil(t, err, "aha")
	assert.True(t, strings.HasPrefix(string(cfg), "title: A\nuid: B\n"), string(cfg))

	//	assert.Equal(t, 1, len(r.Header["Set-Cookie"]), "naja")

	// stat, _ := os.Stat(uriPub)
	// assert.Equal(t, 0755, int(stat.Mode()&os.ModePerm), "ach, wieso?")
}

func doHttp(method, path_info string) (*http.Response, error) {
	cgi := "shaarli.cgi"
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

func TestPostConfig(t *testing.T) {
	defer prepTeardown(t)()

	r, err := doPost("/config/", []byte(`title=A&setlogin=B&setpassword=123456789012&import_shaarli_url=&import_shaarli_setlogin=&import_shaarli_setpassword=`))

	assert.Nil(t, err, "aha")
	assert.Equal(t, http.StatusFound, r.StatusCode, "aha")
	assert.Equal(t, "/sub/"+uriPubPosts, r.Header["Location"][0], "aha")

	body, err := ioutil.ReadAll(r.Body)
	assert.Nil(t, err, "aha")
	assert.Equal(t, 0, len(body), "soso")

	cfg, err := ioutil.ReadFile(filepath.Join(dirApp, "config.yaml"))
	assert.Nil(t, err, "aha")
	assert.True(t, strings.HasPrefix(string(cfg), "title: A\nuid: B\n"), string(cfg))

	assert.Equal(t, 1, len(r.Header["Set-Cookie"]), "naja")

	// stat, _ := os.Stat(uriPub)
	// assert.Equal(t, 0755, int(stat.Mode()&os.ModePerm), "ach, wieso?")
}

func TestGetLoginWithoutRedir(t *testing.T) {
	defer prepTeardown(t)()

	r, err := doPost("/config/", []byte(`title=A&setlogin=B&setpassword=123456789012&import_shaarli_url=&import_shaarli_setlogin=&import_shaarli_setpassword=`))
	assert.Nil(t, err, "aha")
	assert.Equal(t, http.StatusFound, r.StatusCode, "aha")
	assert.Equal(t, "/sub/"+uriPubPosts, r.Header["Location"][0], "aha")

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
	assert.Equal(t, "/sub/"+uriPubPosts, r.Header["Location"][0], "aha")
	cook := r.Header["Set-Cookie"][0]
	assert.True(t, strings.HasPrefix(cook, "ShaarliGo=MTU"), cook)
}

func TestGetLoginWithRedir(t *testing.T) {
	defer prepTeardown(t)()

	os.Unsetenv("COOKIE")
	r, err := doPost("/config/", []byte(`title=A&setlogin=B&setpassword=123456789012&import_shaarli_url=&import_shaarli_setlogin=&import_shaarli_setpassword=`))
	assert.Nil(t, err, "aha")
	assert.Equal(t, http.StatusFound, r.StatusCode, "aha")
	assert.Equal(t, "/sub/"+uriPubPosts, r.Header["Location"][0], "aha")

	returnurl := "/sub/" + uriPubPosts + "anyid/?foo=bar#baz"
	os.Setenv("QUERY_STRING", "do=login&returnurl="+url.QueryEscape(returnurl))
	r, err = doGet("")
	assert.Nil(t, err, "aha")
	assert.Equal(t, http.StatusOK, r.StatusCode, "aha")
	root, err := html.Parse(r.Body)
	assert.Nil(t, err, "aha")
	assert.NotNil(t, root, "aha")
	inputs := scrape.FindAll(root, func(n *html.Node) bool { return atom.Input == n.DataAtom })
	assert.Equal(t, 6, len(inputs), "aha")

	r, err = doPost("", []byte(`login=B&password=123456789012&token=foo&returnurl=/sub/`+uriPubPosts+`anyid/?foo=bar#baz`))
	assert.Equal(t, http.StatusFound, r.StatusCode, "aha")
	assert.Equal(t, returnurl, r.Header["Location"][0], "aha")
	cook := r.Header["Set-Cookie"][0]
	assert.True(t, strings.HasPrefix(cook, "ShaarliGo=MTU"), cook)
}

func _TestGetPostNew(t *testing.T) {
	defer prepTeardown(t)()

	r, err := doPost("/config", []byte(`title=A&setlogin=B&setpassword=123456789012&import_shaarli_url=&import_shaarli_setlogin=&import_shaarli_setpassword=`))
	assert.Nil(t, err, "aha")
	assert.Equal(t, http.StatusFound, r.StatusCode, "aha")
	assert.Equal(t, "/sub/"+uriPubPosts, r.Header["Location"][0], "aha")

	purl := fmt.Sprintf("?post=%s&title=%s&source=%s", url.QueryEscape("http://example.com/foo?bar=baz#grr"), url.QueryEscape("A first post"), url.QueryEscape("me"))
	os.Setenv("QUERY_STRING", purl)
	r, err = doGet("")
	assert.Nil(t, err, "aha")
	assert.Equal(t, http.StatusFound, r.StatusCode, "aha")
	assert.Equal(t, "/sub/shaarli.cgi?do=login", r.Header["Location"], "aha")

	r, err = doGet(fmt.Sprintf("?do=login&returnurl=/sub/shaarli.cgi%s", url.QueryEscape(purl)))
	assert.Nil(t, err, "aha")
	assert.Equal(t, http.StatusOK, r.StatusCode, "aha")
	cook := r.Header["Set-Cookie"][0]
	assert.True(t, strings.HasPrefix(cook, "ShaarliGo=MTU"), cook)
	os.Setenv("COOKIE", cook)
	root, err := html.Parse(r.Body)
	assert.Nil(t, err, "aha")
	assert.NotNil(t, root, "aha")
	assert.Equal(t, 4, len(scrape.FindAll(root, func(n *html.Node) bool { return atom.Input == n.DataAtom })), "aha")

	r, err = doPost(fmt.Sprintf("?do=login&returnurl=/sub/shaarli.cgi%s", url.QueryEscape(purl)), []byte(`login=B&password=123456789012`))
	os.Setenv("COOKIE", r.Header["Set-Cookie"][0])

	r, err = doGet(purl)
	assert.Equal(t, http.StatusOK, r.StatusCode, "aha")
	os.Setenv("COOKIE", r.Header["Set-Cookie"][0])
	root, err = html.Parse(r.Body)

	r, err = doPost(purl, nil)
	assert.Equal(t, http.StatusFound, r.StatusCode, "aha")
	assert.Equal(t, "/sub/"+uriPubPosts+"?#foo", r.Header["Location"], "aha")
}

func BenchmarkHello(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := fmt.Sprintf("hello")
		assert.NotNil(b, s, "aha")
	}
}

func fileIOPayload(idx int) {
	strFile := filepath.Join("testdata", strconv.Itoa(idx))
	if f, err := os.Create(strFile); err == nil {
		f.WriteString(strFile)
		f.Close()
		os.Remove(strFile)
	} else {
		panic(err)
	}
}

func BenchmarkFileCreateDeleteSequential(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fileIOPayload(i)
	}
}

func _BenchmarkFileCreateDeleteParallel(b *testing.B) {
	var wg sync.WaitGroup
	for i := 0; i < b.N; i++ {
		go func(ii int) {
			wg.Add(1)
			defer wg.Done()
			fileIOPayload(ii)
		}(i)
	}
	wg.Wait()
}

func BenchmarkFileCreateDeleteParallelChannel(b *testing.B) {
	var wg sync.WaitGroup

	worker := func(id int, jobs <-chan int) {
		for j := range jobs {
			func() {
				wg.Add(1)
				defer wg.Done()
				fileIOPayload(j)
			}()
		}
	}

	jobs := make(chan int, 10)
	for w := 0; w < 5; w++ {
		go worker(w, jobs)
	}
	for j := 0; j < b.N; j++ {
		jobs <- j
	}
	close(jobs)

	wg.Wait()
}
