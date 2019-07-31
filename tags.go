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
func isEmojiRune(ru rune) bool {
	r := int(ru)
	return false ||
		(0x2b50 <= r && r <= 0x2b50) || // star
		(0x1F600 <= r && r <= 0x1F64F) || // Emoticons
		(0x1F300 <= r && r <= 0x1F5FF) || // Misc Symbols and Pictographs
		(0x1F680 <= r && r <= 0x1F6FF) || // Transport and Map
		(0x1F1E6 <= r && r <= 0x1F1FF) || // Regional country flags
		(0x2600 <= r && r <= 0x26FF) || // Misc symbols
		(0x2700 <= r && r <= 0x27BF) || // Dingbats
		(0xFE00 <= r && r <= 0xFE0F) || // Variation Selectors
		(0x1F900 <= r && r <= 0x1F9FF) || // Supplemental Symbols and Pictographs
		(0x1f018 <= r && r <= 0x1f270) || // Various asian characters
		(0xfe00 <= r && r <= 0xfe0f) || // Variation selector
		(0x238c <= r && r <= 0x2454) || // Misc items
		(0x20d0 <= r && r <= 0x20ff) // Combining Diacritical Marks for Symbols
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

func tagsNormalise(ds string, ex string, ta []string, known map[string]string) (description string, extended string, tags []string) {
	tadi := make(map[string]string, len(ta))
	for _, tag := range ta {
		k := fold(tag)
		if "" == k {
			continue
		}
		if _, ok := tadi[k]; ok {
			continue
		}
		if v, ok := known[k]; ok {
			tadi[k] = v
			continue
		}
		tadi[k] = tag
	}

	midi := make(map[string]string, len(tadi))
	for k, v := range tadi {
		midi[k] = v
	}
	for _, tag := range append(tagsFromString(ds), tagsFromString(ex)...) {
		k := fold(tag)
		if _, ok := tadi[k]; ok {
			delete(midi, k)
			continue
		}
		if v, ok := known[k]; ok {
			tadi[k] = v
			continue
		}
		tadi[k] = tag
	}

	miss := make([]string, 0, len(midi))
	for _, v := range midi {
		miss = append(miss, v)
	}
	sort.Strings(miss)

	tags = make([]string, 0, len(tadi))
	for _, v := range tadi {
		tags = append(tags, v)
	}
	sort.Strings(tags)

	description = strings.TrimSpace(ds)
	extended = strings.TrimSpace(strings.Join(append([]string{strings.TrimSpace(ex)}, miss...), " #"))
	return
}
