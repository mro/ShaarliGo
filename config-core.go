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
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type Config struct {
	AuthorName        string `yaml:"author_name"`
	Title             string `yaml:"title"`
	PwdBcrypt         string `yaml:"pwd_bcrypt"`
	CookieStoreSecret string `yaml:"cookie_secret"`
}

func LoadConfig() (Config, error) {
	ret := Config{}
	if read, err := ioutil.ReadFile(filepath.Join(dirApp, "config.yaml")); err == nil {
		err = yaml.Unmarshal(read, &ret)
		return ret, err
	} else if os.IsNotExist(err) {
		// seed the cookie store secret
		buf := make([]byte, 32)
		if _, err := io.ReadFull(rand.Reader, buf); err != nil {
			return ret, err
		}
		ret.CookieStoreSecret = base64.StdEncoding.EncodeToString(buf)

		return ret, nil
	} else {
		return ret, err
	}
}

func (cfg Config) Save() error {
	if out, err := yaml.Marshal(cfg); err == nil {
		fileName := filepath.Join(dirApp, "config.yaml")
		tmpFileName := filepath.Join(dirApp, fmt.Sprintf("%s~%d", "config.yaml", os.Getpid()))
		if err = ioutil.WriteFile(tmpFileName, out, os.FileMode(0660)); err == nil {
			err = os.Rename(tmpFileName, fileName)
		}
		return err
	} else {
		return err
	}
}

func (cfg Config) IsConfigured() bool {
	return cfg.PwdBcrypt != ""
}
