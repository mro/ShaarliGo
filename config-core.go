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
	Title             string                   `yaml:"title"`
	Uid               string                   `yaml:"uid"`
	PwdBcrypt         string                   `yaml:"pwd_bcrypt"`
	CookieStoreSecret string                   `yaml:"cookie_secret"`
	TimeZone          string                   `yaml:"timezone"`
	Skin              string                   `yaml:"skin"`           // maybe should not be configurable, but rather be a filesystem softlink
	LinksPerPage      int                      `yaml:"links_per_page"` // https://github.com/sebsauvage/Shaarli/blob/master/index.php#L18
	BanAfter          int                      `yaml:"ban_after"`      // https://github.com/sebsauvage/Shaarli/blob/master/index.php#L20
	BanSeconds        int                      `yaml:"ban_seconds"`    // https://github.com/sebsauvage/Shaarli/blob/master/index.php#L21
	UrlCleaner        []RegexpReplaceAllString `yaml:"url_cleaner"`
	// Redirector     string                   `yaml:"redirector"` // actually a prefix to href - Hardcoded in xslt
}

func LoadConfig() (Config, error) {
	if read, err := ioutil.ReadFile(configFileName); err != nil {
		return Config{}, err
	} else {
		ret := Config{}
		if err = yaml.Unmarshal(read, &ret); err != nil {
			return Config{}, err
		}
		if ret.CookieStoreSecret == "" {
			buf := make([]byte, 32)
			if _, err := io.ReadFull(rand.Reader, buf); err != nil {
				return Config{}, err
			}
			ret.CookieStoreSecret = base64.StdEncoding.EncodeToString(buf)
		}
		ret.LinksPerPage = max(1, ret.LinksPerPage)
		ret.BanAfter = max(1, ret.BanAfter)
		ret.BanSeconds = max(1, ret.BanSeconds)
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
