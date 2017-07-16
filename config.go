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
	"io/ioutil"
	"os"
	"syscall"

	"gopkg.in/yaml.v2"
)

const (
	CONFIG_FILE_PATH string = "app"
	CONFIG_FILE_NAME string = CONFIG_FILE_PATH + "/config.yaml"
)

func LoadConfig() (*Config, error) {
	ret, err := configFromFileName(CONFIG_FILE_NAME)
	if err != nil {
		pathErr, ok := err.(*os.PathError)
		if ok && pathErr.Op == "open" && syscall.Errno(2) == pathErr.Err {
			// recover from "no such file or directory" and return empty config
			ret = Config{}
		} else {
			return nil, err
		}
	}
	return &ret, nil
}

type Config struct {
	AuthorName string `yaml:"author_name"`
	Title      string `yaml:"title"`
	PwdBcrypt  string `yaml:"pwd_bcrypt"`
}

func (cfg *Config) Save() error {
	err := cfg.saveToFileName(CONFIG_FILE_NAME)
	if err != nil {
		pathErr, ok := err.(*os.PathError)
		if ok && "open" == pathErr.Op && syscall.Errno(2) == pathErr.Err {
			// recover from "no such file or directory" and create dir
			err = os.Mkdir(CONFIG_FILE_PATH, os.FileMode(0770))
			if err != nil {
				return err
			}
			return cfg.saveToFileName(CONFIG_FILE_NAME)
		} else {
			return err
		}
	}
	return nil
}

func (cfg *Config) saveToFileName(fileName string) error {
	tmpFileName := fileName + "~"
	out, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(tmpFileName, out, os.FileMode(0660))
	defer os.Remove(tmpFileName)
	if err != nil {
		return err
	}
	return os.Rename(tmpFileName, fileName)
}

func configFromFileName(file string) (Config, error) {
	ret := Config{}
	read, err := ioutil.ReadFile(file)
	if read == nil {
		return ret, err
	}
	err = yaml.Unmarshal(read, &ret)
	return ret, err
}

func (m *Config) IsConfigured() bool {
	return m.AuthorName != ""
}
