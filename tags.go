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
	"sort"
	"strings"
	"unicode"

	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var emojiRunes map[rune]struct{}

func init() {
	emojiRunes = make(map[rune]struct{}, len(emojiCodeMap))
	for _, v := range emojiCodeMap {
		r := []rune(v)[0]
		emojiRunes[r] = struct{}{}
	}
	emojiCodeMap = nil
}

// https://stackoverflow.com/a/39425959
// http://cldr-build.unicode.org/UnicodeJsps/list-unicodeset.jsp?a=%5B%3Aemoji%3A%5D&g=emoji
func isEmojiRune(ru rune) bool {
	return false ||
		('\u20d0' <= ru && ru <= '\u20ff') || // Combining Diacritical Marks for Symbols
		('\u2328' == ru) || // keyboard
		('\u238c' <= ru && ru <= '\u2454') || // Misc items
		('\u2600' <= ru && ru <= '\u26FF') || // Misc symbols
		('\u2700' <= ru && ru <= '\u27BF') || // Dingbats
		('\u2b50' == ru) || // star
		('\uFE00' <= ru && ru <= '\uFE0F') || // Variation Selectors
		('\U0001f018' <= ru && ru <= '\U0001f270') || // Various asian characters
		('\U0001F1E6' <= ru && ru <= '\U0001F1FF') || // Regional country flags
		('\U0001F300' <= ru && ru <= '\U0001F5FF') || // Misc Symbols and Pictographs
		('\U0001F600' <= ru && ru <= '\U0001F64F') || // Emoticons
		('\U0001F680' <= ru && ru <= '\U0001F6FF') || // Transport and Map
		('\U0001F900' <= ru && ru <= '\U0001F9FF') // Supplemental Symbols and Pictographs
}

const tpf = '#'

func myPunct(r rune) bool {
	switch r {
	case '@', '§', '†', tpf:
		return false
	default:
		return unicode.IsPunct(r)
	}
}

func isTag(tag string) string {
	for _, c := range tag {
		if tpf == c {
			tag = tag[1:]
			break
		}
		if isEmojiRune(c) {
			break
		}
		return ""
	}
	return strings.TrimFunc(tag, myPunct)
}

func tagsFromString(str string) []string {
	scanner := bufio.NewScanner(strings.NewReader(str))
	scanner.Split(bufio.ScanWords)

	ret := make([]string, 0, 10)
	tmp := make(map[string]struct{}, 10)
	tmp[""] = struct{}{}
	for scanner.Scan() {
		tag := isTag(scanner.Text())
		if _, ok := tmp[tag]; ok {
			continue
		}
		ret = append(ret, tag)
		tmp[tag] = struct{}{}
	}
	return ret
}

// https://stackoverflow.com/a/26722698
func fold(str string) string {
	tr := transform.Chain(norm.NFD, transform.RemoveFunc(func(r rune) bool {
		return unicode.Is(unicode.Mn, r) // Mn: nonspacing marks
	}), norm.NFC)
	// todo: chain lowercase + trim
	if result, _, err := transform.String(tr, str); err != nil {
		panic(err)
	} else {
		return strings.TrimSpace(strings.ToLower(result))
	}
}

func tagsVisitor(tags ...string) func(func(string)) {
	return func(callback func(string)) {
		for _, tag := range tags {
			callback(tag)
		}
	}
}

func tagsNormalise(ds string, ex string, tavi func(func(string)), knovi func(func(string))) (description string, extended string, tags []string) {
	knodi := make(map[string]string, 1000)
	knovi(func(tag string) { knodi[fold(tag)] = tag })

	tags = make([]string, 0, 20)
	// 1. iterate text tags
	tadi := make(map[string]string, 20)
	tadi[""] = ""

	add := func(tag string) string {
		k := fold(tag)
		if _, ok := tadi[k]; ok {
			return ""
		}
		if v, ok := knodi[k]; ok {
			tag = v
		}
		tadi[k] = tag
		tags = append(tags, tag) // updating the reference correctly?
		return tag
	}

	for _, tag := range append(tagsFromString(ds), tagsFromString(ex)...) {
		add(tag)
	}

	// 2. visit all previous tags and add missing ones to tadi and extended
	tavi(func(tag string) {
		if t := add(tag); t == "" {
			return
		}
		ex += " #" + tag // todo: skip superfluous # before emojis
	})

	description = strings.TrimSpace(ds)
	extended = strings.TrimSpace(ex)
	sort.Strings(tags)
	return
}
