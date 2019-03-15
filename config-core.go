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
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

var configFileName string

func init() {
	configFileName = filepath.Join(dirApp, "config.yaml")
}

type RegexpReplaceAllString struct {
	Regexp           string `yaml:"regexp"`
	ReplaceAllString string `yaml:"replace_all_string"`
}

type Config struct {
	Title             string                   `yaml:"title"` // title for cheap non-feed screens (login, tools, etc)
	Uid               string                   `yaml:"uid"`   // for login
	PwdBcrypt         string                   `yaml:"pwd_bcrypt"`
	CookieStoreSecret string                   `yaml:"cookie_secret"`
	TimeZone          string                   `yaml:"timezone"`
	Skin              string                   `yaml:"skin"`
	LinksPerPage      int                      `yaml:"links_per_page"` // https://github.com/sebsauvage/Shaarli/blob/master/index.php#L18
	BanAfter          int                      `yaml:"ban_after"`      // https://github.com/sebsauvage/Shaarli/blob/master/index.php#L20
	BanSeconds        int                      `yaml:"ban_seconds"`    // https://github.com/sebsauvage/Shaarli/blob/master/index.php#L21
	UrlCleaner        []RegexpReplaceAllString `yaml:"url_cleaner"`
	// Redirector     string                   `yaml:"redirector"` // actually a prefix to href - Hardcoded in xslt
}

func LoadConfig() (Config, error) {
	// seed the cookie store secret
	buf := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, buf); err != nil {
		return Config{}, err
	}
	ret := Config{
		CookieStoreSecret: base64.StdEncoding.EncodeToString(buf),
		TimeZone:          "Europe/Paris",
		Skin:              "default/de",
		LinksPerPage:      100,   // https://github.com/sebsauvage/Shaarli/blob/master/index.php#L18
		BanAfter:          4,     // https://github.com/sebsauvage/Shaarli/blob/master/index.php#L20
		BanSeconds:        14400, // https://github.com/sebsauvage/Shaarli/blob/master/index.php#L21
		UrlCleaner: []RegexpReplaceAllString{
			{Regexp: "[\\?&]utm_source=.*$", ReplaceAllString: ""}, // We remove the annoying parameters added by FeedBurner and GoogleFeedProxy (?utm_source=...)
			{Regexp: "#xtor=RSS-.*$", ReplaceAllString: ""},
			{Regexp: "^(?i)(?:https?://)?(?:(?:www|m)\\.)?heise\\.de/.*?(-\\d+)(?:\\.html)?(?:[\\?#].*)?$", ReplaceAllString: "https://heise.de/${1}"},
			{Regexp: "^(?i)(?:https?://)?(?:(?:www|m)\\.)?spiegel\\.de/.*?-(\\d+)(?:\\.html.*)?", ReplaceAllString: "https://spiegel.de/article.do?id=${1}"},
			{Regexp: "^(?i)(?:https?://)?(?:(?:www|m)\\.)?sueddeutsche\\.de/.*?-(\\d+\\.\\d+)(?:\\.html.*)?$", ReplaceAllString: "https://sz.de/${1}"},
			{Regexp: "^(?i)(?:https?://)?(?:(?:www|m)\\.)?youtube.com/watch\\?v=([^&]+)(?:.*&(t=[^&]+))?(?:.*)$", ReplaceAllString: "https://youtu.be/${1}?${2}"},
		},
		// Redirector: "http://anonym.to/?",
	}

	if read, err := ioutil.ReadFile(configFileName); err == nil {
		err = yaml.Unmarshal(read, &ret)
		if ret.LinksPerPage < 1 {
			ret.LinksPerPage = 1
		}
		return ret, err
	} else if os.IsNotExist(err) {
		return ret, nil
	} else {
		return ret, err
	}
}

func (cfg Config) Save() error {
	if out, err := yaml.Marshal(cfg); err == nil {
		tmpFileName := fmt.Sprintf("%s~%d", configFileName, os.Getpid())
		if err = os.MkdirAll(filepath.Dir(configFileName), 0770); err == nil {
			if err = ioutil.WriteFile(tmpFileName, out, 0660); err == nil {
				err = os.Rename(tmpFileName, configFileName)
			}
		}
		return err
	} else {
		return err
	}
}

func (cfg Config) IsConfigured() bool {
	return cfg.PwdBcrypt != ""
}
