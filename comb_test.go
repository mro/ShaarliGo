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
	"os"
	"time"

	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseCombHeiseHttp(t *testing.T) {
	t.SkipNow()
	ent, err := entryFromURL(mustParseURL("https://www.heise.de/security/meldung/Notfall-Patch-fuer-Windows-Co-Kritische-Sicherheitsluecke-im-Virenscanner-von-Microsoft-3913800.html"), 10*time.Second)
	assert.Nil(t, err, "ouch")
	assert.Equal(t, "de", ent.XmlLang, "ouch")
}

func TestParseCombHeise(t *testing.T) {
	t.Parallel()
	f, err := os.Open("testdata/comb-heise-3913800.html")
	assert.NotNil(t, f, "ouch")
	assert.Nil(t, err, "ouch")

	ent, err := entryFromReader(f, mustParseURL("https://www.heise.de/security/meldung/Notfall-Patch-fuer-Windows-Co-Kritische-Sicherheitsluecke-im-Virenscanner-von-Microsoft-3913800.html"))

	assert.Nil(t, err, "ouch")
	assert.Equal(t, "de", ent.XmlLang, "ouch")
	assert.Equal(t, "Notfall-Patch für Windows & Co.: Kritische Sicherheitslücke im Virenscanner von Microsoft", ent.Title.Body, "ouch")
	//assert.Equal(t, "", ent.Content.Body, "ouch")
	assert.Equal(t, 5, len(ent.Categories), "ouch")
	assert.Equal(t, "Malware_Protection_Engine", ent.Categories[0].Term, "ouch")
	assert.Equal(t, "Windows_Defender", ent.Categories[4].Term, "ouch")
	assert.Equal(t, "Die Malware Protection Engine von Microsoft weist eine Schwachstelle auf, über die Angreifer Schadcode auf Computer schieben könnten. Die Engine kommt unter anderem bei Windows Defender zum Einsatz.", ent.Summary.Body, "ouch")
	assert.Equal(t, 1, len(ent.Authors), "ouch")
	assert.Equal(t, "Dennis Schirrmacher", ent.Authors[0].Name, "ouch")
	assert.Equal(t, "2017-12-08T09:57:00+01:00", ent.Published.Format(time.RFC3339), "ouch")
	assert.Equal(t, "https://www.heise.de/imgs/18/2/3/3/3/6/5/2/ms-27f2d5b32536ec59.png", ent.MediaThumbnail.Url, "ouch")
}

func TestParseCombSpon(t *testing.T) {
	t.Parallel()
	f, err := os.Open("testdata/comb-spon-1182013.html")
	assert.NotNil(t, f, "ouch")
	assert.Nil(t, err, "ouch")

	ent, err := entryFromReader(f, mustParseURL("http://www.spiegel.de/netzwelt/web/bitcoin-blockchain-hashgraph-die-blase-die-bleibt-kolumne-a-1182013.html"))

	assert.Nil(t, err, "ouch")
	assert.Equal(t, "de", ent.XmlLang, "ouch")
	assert.Equal(t, "Bitcoin, Blockchain, Hashgraph: Die Blase, die bleibt - SPIEGEL ONLINE - Netzwelt", ent.Title.Body, "ouch")
	//assert.Equal(t, "", ent.Content.Body, "ouch")
	assert.Equal(t, 0, len(ent.Categories), "ouch")
	assert.Equal(t, "Bitcoin ist nur deshalb so viel wert, weil so viele Menschen daran glauben, dass Bitcoin so viel wert ist. Die Krypto-Währung ist nichts anderes als das jüngste, digitale Gesicht des Kapitalismus.", ent.Summary.Body, "ouch")
	assert.Equal(t, 1, len(ent.Authors), "ouch")
	assert.Equal(t, "SPIEGEL ONLINE, Hamburg, Germany", ent.Authors[0].Name, "ouch")
	assert.Equal(t, "2017-12-06T16:13:00+01:00", ent.Published.Format(time.RFC3339), "ouch")
	assert.Equal(t, "http://cdn3.spiegel.de/images/image-1222932-galleryV9-yqyy-1222932.jpg", ent.MediaThumbnail.Url, "ouch")
}
