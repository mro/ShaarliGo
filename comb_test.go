//
// Copyright (C) 2017-2021 Marcus Rohrmoser, http://purl.mro.name/ShaarliGo
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
	assert.Equal(t, Lang("de"), ent.XmlLang, "ouch")
	assert.Equal(t, "Notfall-Patch für Windows & Co.: Kritische Sicherheitslücke im Virenscanner von Microsoft", ent.Title.Body, "ouch")
	//assert.Equal(t, "", ent.Content.Body, "ouch")
	assert.Equal(t, 5, len(ent.Categories), "ouch")
	assert.Equal(t, "Malware_Protection_Engine", ent.Categories[0].Term, "ouch")
	assert.Equal(t, "Windows_Defender", ent.Categories[4].Term, "ouch")
	assert.Equal(t, "Die Malware Protection Engine von Microsoft weist eine Schwachstelle auf, über die Angreifer Schadcode auf Computer schieben könnten. Die Engine kommt unter anderem bei Windows Defender zum Einsatz.", ent.Summary.Body, "ouch")
	assert.Equal(t, 1, len(ent.Authors), "ouch")
	assert.Equal(t, "Dennis Schirrmacher", ent.Authors[0].Name, "ouch")
	assert.Equal(t, "2017-12-08T09:57:00+01:00", ent.Published.Format(time.RFC3339), "ouch")
	assert.Equal(t, Iri("https://www.heise.de/imgs/18/2/3/3/3/6/5/2/ms-27f2d5b32536ec59.png"), ent.MediaThumbnail.Url, "ouch")
}

func TestParseCombSpon(t *testing.T) {
	t.Parallel()
	f, err := os.Open("testdata/comb-spon-1182013.html")
	assert.NotNil(t, f, "ouch")
	assert.Nil(t, err, "ouch")

	ent, err := entryFromReader(f, mustParseURL("http://www.spiegel.de/netzwelt/web/bitcoin-blockchain-hashgraph-die-blase-die-bleibt-kolumne-a-1182013.html"))

	assert.Nil(t, err, "ouch")
	assert.Equal(t, Lang("de"), ent.XmlLang, "ouch")
	assert.Equal(t, "Bitcoin, Blockchain, Hashgraph: Die Blase, die bleibt - SPIEGEL ONLINE - Netzwelt", ent.Title.Body, "ouch")
	//assert.Equal(t, "", ent.Content.Body, "ouch")
	assert.Equal(t, 0, len(ent.Categories), "ouch")
	assert.Equal(t, "Bitcoin ist nur deshalb so viel wert, weil so viele Menschen daran glauben, dass Bitcoin so viel wert ist. Die Krypto-Währung ist nichts anderes als das jüngste, digitale Gesicht des Kapitalismus.", ent.Summary.Body, "ouch")
	assert.Equal(t, 1, len(ent.Authors), "ouch")
	assert.Equal(t, "SPIEGEL ONLINE, Hamburg, Germany", ent.Authors[0].Name, "ouch")
	assert.Equal(t, "2017-12-06T16:13:00+01:00", ent.Published.Format(time.RFC3339), "ouch")
	assert.Equal(t, Iri("http://cdn3.spiegel.de/images/image-1222932-galleryV9-yqyy-1222932.jpg"), ent.MediaThumbnail.Url, "ouch")
}

func TestParseCombSpon2019(t *testing.T) {
	t.Parallel()
	f, err := os.Open("testdata/comb-spon-1276146.html")
	assert.NotNil(t, f, "ouch")
	assert.Nil(t, err, "ouch")

	ent, err := entryFromReader(f, mustParseURL("https://www.spiegel.de/kultur/musik/joao-gilberto-bossa-nova-legende-stirbt-mit-88-jahren-a-1276146.html"))

	assert.Nil(t, err, "ouch")
	assert.Equal(t, Lang("de"), ent.XmlLang, "ouch")
	assert.Equal(t, "João Gilberto: Bossa-Nova-Legende stirbt mit 88 Jahren - SPIEGEL ONLINE", ent.Title.Body, "ouch")
	assert.Nil(t, ent.Content, "ouch")
	assert.Equal(t, 0, len(ent.Categories), "ouch")
	assert.Equal(t, "Er war einer der einflussreichsten Vertreter lateinamerikanischer Musik. \"The Girl from Ipanema\" - komponiert von Antônio Carlos Jobim und interpretiert von seiner früheren Ehefrau Astrud - wurde zum Welthit. Jetzt ist der Gitarrist João Gilberto gestorben.", ent.Summary.Body, "ouch")
	assert.Equal(t, 1, len(ent.Authors), "ouch")
	assert.Equal(t, "SPIEGEL ONLINE, Hamburg, Germany", ent.Authors[0].Name, "ouch")
	assert.Equal(t, "2019-07-06T22:10:00+02:00", ent.Published.Format(time.RFC3339), "ouch")
	//	assert.Equal(t, Iri("https://cdn1.spiegel.de/images/image-1446841-860_poster_16x9-dxqw-1446841.jpg"), ent.MediaThumbnail.Url, "ouch")

	data := ent.api0LinkFormMap()
	assert.Equal(t, ent.Title.Body, data["lf_title"], "ouch")
}
