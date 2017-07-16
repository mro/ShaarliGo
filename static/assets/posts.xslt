<?xml version="1.0" encoding="UTF-8"?>
<!--
  AtomicShaarli, microblogging detox
  Copyright (C) 2017-2017  Marcus Rohrmoser, http://purl.mro.name/AtomicShaarli

  This program is free software: you can redistribute it and/or modify
  it under the terms of the GNU General Public License as published by
  the Free Software Foundation, either version 3 of the License, or
  (at your option) any later version.

  This program is distributed in the hope that it will be useful,
  but WITHOUT ANY WARRANTY; without even the implied warranty of
  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
  GNU General Public License for more details.

  You should have received a copy of the GNU General Public License
  along with this program.  If not, see <http://www.gnu.org/licenses/>.

  http://www.w3.org/TR/xslt/
-->
<xsl:stylesheet
  xmlns="http://www.w3.org/1999/xhtml"
  xmlns:a="http://www.w3.org/2005/Atom"
  xmlns:media="http://search.yahoo.com/mrss/"
  xmlns:georss="http://www.georss.org/georss"
  xmlns:xsl="http://www.w3.org/1999/XSL/Transform"
  exclude-result-prefixes="a media georss"
  version="1.0">

  <!-- replace linefeeds with <br> tags -->
  <xsl:template name="linefeed2br">
    <xsl:param name="string" select="''"/>
    <xsl:param name="pattern" select="'&#10;'"/>
    <xsl:choose>
      <xsl:when test="contains($string, $pattern)">
        <xsl:value-of select="substring-before($string, $pattern)"/><br class="br"/><xsl:comment>Why do we see 2 br on Safari and output/@method=html here? http://purl.mro.name/safari-xslt-br-bug</xsl:comment>
        <xsl:call-template name="linefeed2br">
          <xsl:with-param name="string" select="substring-after($string, $pattern)"/>
          <xsl:with-param name="pattern" select="$pattern"/>
        </xsl:call-template>
      </xsl:when>
      <xsl:otherwise>
        <xsl:value-of select="$string"/>
      </xsl:otherwise>
    </xsl:choose>
  </xsl:template>

  <xsl:template name="human_time">
    <xsl:param name="time">-</xsl:param>
    <xsl:value-of select="substring($time, 9, 2)"/><xsl:text>. </xsl:text>
    <xsl:variable name="month" select="substring($time, 6, 2)"/>
    <xsl:choose>
      <xsl:when test="'01' = $month">Jan</xsl:when>
      <xsl:when test="'02' = $month">Feb</xsl:when>
      <xsl:when test="'03' = $month">Mär</xsl:when>
      <xsl:when test="'04' = $month">Apr</xsl:when>
      <xsl:when test="'05' = $month">Mai</xsl:when>
      <xsl:when test="'06' = $month">Jun</xsl:when>
      <xsl:when test="'07' = $month">Jul</xsl:when>
      <xsl:when test="'08' = $month">Aug</xsl:when>
      <xsl:when test="'09' = $month">Sep</xsl:when>
      <xsl:when test="'10' = $month">Okt</xsl:when>
      <xsl:when test="'11' = $month">Nov</xsl:when>
      <xsl:when test="'12' = $month">Dez</xsl:when>
      <xsl:otherwise>?</xsl:otherwise>
    </xsl:choose><xsl:text> </xsl:text>
    <xsl:value-of select="substring($time, 1, 4)"/><xsl:text> </xsl:text>
    <xsl:value-of select="substring($time, 12, 5)"/><!-- xsl:text> Uhr</xsl:text -->
  </xsl:template>

  <xsl:template name="degrees">
    <xsl:param name="num" select="0"/>
    <xsl:choose>
      <xsl:when test="$num &lt; 0">-<xsl:call-template name="degrees"><xsl:with-param name="num" select="-$num"/></xsl:call-template></xsl:when>
      <xsl:when test="$num &gt;= 0">
        <xsl:variable name="deg" select="floor($num)"/>
        <xsl:variable name="min" select="floor(($num * 60) mod 60)"/>
        <xsl:variable name="sec" select="format-number((($num * 36000) mod 600) div 10, '0.0')"/>
        <xsl:value-of select="$deg"/>° <!--
        --><xsl:value-of select="$min"/>' <!--
        --><xsl:value-of select="$sec"/>"
      </xsl:when>
      <xsl:otherwise>?</xsl:otherwise>
    </xsl:choose>
  </xsl:template>

  <xsl:output
    method="html"
    doctype-system="http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd"
    doctype-public="-//W3C//DTD XHTML 1.0 Strict//EN"/>

  <!-- http://stackoverflow.com/a/16328207 -->
  <xsl:key name="CategorY" match="a:entry/a:category" use="@term" />

  <xsl:template match="a:feed">
    <html xmlns="http://www.w3.org/1999/xhtml">
      <head>
        <meta content="text/html; charset=utf-8" http-equiv="content-type"/>
        <!-- https://developer.apple.com/library/IOS/documentation/AppleApplications/Reference/SafariWebContent/UsingtheViewport/UsingtheViewport.html#//apple_ref/doc/uid/TP40006509-SW26 -->
        <!-- http://maddesigns.de/meta-viewport-1817.html -->
        <!-- meta name="viewport" content="width=device-width"/ -->
        <!-- http://www.quirksmode.org/blog/archives/2013/10/initialscale1_m.html -->
        <meta name="viewport" content="width=device-width,initial-scale=1.0"/>
        <!-- meta name="viewport" content="width=400"/ -->
        <link href="../../assets/bootstrap.css" rel="stylesheet" type="text/css"/>
        <link href="../../assets/bootstrap-theme.css" rel="stylesheet" type="text/css"/>

        <link href="." rel="alternate" type="application/atom+xml"/>
        <link href="." rel="self" type="application/xhtml+xml"/>

        <style type="text/css">
.table {
  width: 100%;
  max-width: 100%;
}
li {
  background-color: #F8F8F8;
  margin: 1em 0;
}
        </style>
        <title><xsl:value-of select="a:title"/></title>
      </head>
      <body>
        <div class="container">
          <noscript><p>JavaScript ist aus, es geht zwar (fast) alles auch ohne, aber mit ist's <em>schöner</em>.</p></noscript>

          <p/>

          <table class="table table-bordered table-striped table-inverse">
            <tbody>
              <tr>
                <td class="text-left"><a href="../tags/">⛅ # Tag Cloud</a></td>
                <td class="text-center"><a href="../thumbnails/">🎨 Picture Wall</a></td>
                <td class="text-center"><a href="../days/now">📅 Daily</a></td>
                <td class="text-center"><a href="../tools/">🔨 Tools</a></td>
                <td class="text-right">
                  <span class="hidden"><a href="../login">Login</a> <a href="../../?do=login">🌺</a></span>
                  <span><a href="../logout">Logout</a> <a href="../../?do=logout">🌺</a></span>
                </td>
              </tr>
            </tbody>
          </table>

          <p/>

          <table class="table">
            <tbody>
              <tr>
                <td class="text-left"><a href="{a:link[@rel='first']/@href}" title="Erste Seite">|&lt; Erste</a></td>
                <td class="text-center"><a href="{a:link[@rel='previous']/@href}" title="Vorige Seite">&lt; Vorige</a></td>
                <td class="text-center"><xsl:value-of select="a:subtitle"/></td>
                <td class="text-center"><a href="{a:link[@rel='next']/@href}" title="Nächste Seite">Nächste &gt;</a></td>
                <td class="text-right"><a href="{a:link[@rel='last']/@href}" title="Letzte Seite">Letzte &gt;|</a></td>
              </tr>
            </tbody>
          </table>

          <p/>

          <h1><xsl:value-of select="a:title"/></h1>

          <h2>Feed Untertitel</h2>

          <xsl:choose>
          	<!-- http://getbootstrap.com/css/#forms-horizontal -->
            <xsl:when test="not(a:link[@rel='previous'])">
              <form class="form-horizontal">
                <div class="form-group">
                  <input type="text" class="form-control" id="email_id" name="email_name" placeholder="Was gibt's Neues? (Text oder URL)"/>
                </div>
                <div class="form-group">
                  <input type="file" class="file" id="input-1" placeholder="Bild"/>
                </div>
                <div class="form-group">
                  <div class="form-check">
                    <label class="form-check-label">
                      <input class="form-check-input" type="checkbox"/> Privat?
                    </label>
                  </div>
                </div>
                <div class="form-group">
                  <button type="submit" class="btn btn-primary">Shaaaare!</button>
                </div>
              </form>
            </xsl:when>
            <xsl:when test="a:link[@rel='previous'] and a:link[@rel='next']">
              <form class="form-horizontal">
                <div class="form-group">
                  <label for="email_id" class="control-label col-sm-1">URL</label>
                  <div class="col-sm-11">
                    <input type="text" class="form-control" id="email_id" name="email_name" placeholder="example.com/nice-blog-post/"/>
                  </div>
                </div>
                <div class="form-group">
                  <label for="email_id" class="control-label col-sm-1">Titel</label>
                  <div class="col-sm-11">
                    <input type="text" class="form-control" id="email_id" name="email_name" placeholder="Foo Bar"/>
                  </div>
                </div>
                <div class="form-group">
                  <label for="email_id" class="control-label col-sm-1">Text</label>
                  <div class="col-sm-11">
                    <textarea class="form-control" id="message_id" name="message" rows="3" placeholder="Lorem Ipsum"></textarea>
                  </div>
                </div>
                <div class="form-group">
                  <label for="email_id" class="control-label col-sm-1">Bild</label>
                  <div class="col-sm-11">
                    <input type="file" class="file" id="input-1" placeholder="Bild"/>
                  </div>
                </div>
                <div class="form-group">
                  <div class="col-sm-11 col-sm-offset-1">
                    <div class="form-check">
                      <label class="form-check-label">
                        <input class="form-check-input" type="checkbox"/> Privat?
                      </label>
                    </div>
                  </div>
                </div>
                <div class="form-group">
                  <div class="col-sm-11 col-sm-offset-1">
                    <button type="submit" class="btn btn-primary">Shaaaare!</button>
                  </div>
                </div>
              </form>
            </xsl:when>
          </xsl:choose>

          <ol id="entries" class="list-unstyled">
            <xsl:apply-templates select="a:entry"/>
          </ol>

          <table class="table">
            <tbody>
              <tr>
                <td class="text-left"><a href="{a:link[@rel='first']/@href}" title="Erste Seite">|&lt; Erste</a></td>
                <td class="text-center"><a href="{a:link[@rel='previous']/@href}" title="Vorige Seite">&lt; Vorige</a></td>
                <td class="text-center"><xsl:value-of select="a:subtitle"/></td>
                <td class="text-center"><a href="{a:link[@rel='next']/@href}" title="Nächste Seite">Nächste &gt;</a></td>
                <td class="text-right"><a href="{a:link[@rel='last']/@href}" title="Letzte Seite">Letzte &gt;|</a></td>
              </tr>
            </tbody>
          </table>

          <script src="../../../assets/script.js" type="text/javascript"></script>

          <hr style="clear:left;"/>
          <p id="footer">
            <a title="Validate my Atom 1.0 feed" href="https://validator.w3.org/feed/check.cgi?url={a:link[@rel='self']/@href}">
              <img alt="Valid Atom 1.0" src="../../../assets/valid-atom.png" style="border:0;width:88px;height:31px"/>
            </a><xsl:text> </xsl:text>
            <a href="https://validator.w3.org/check?uri=referer">
              <img alt="Valid XHTML 1.0 Strict" src="../../../assets/valid-xhtml10-blue-v.svg" style="border:0;width:88px;height:31px"/>
            </a>
            <a href="https://jigsaw.w3.org/css-validator/check/referer?profile=css3&amp;usermedium=screen&amp;warning=2&amp;vextwarning=false&amp;lang=de">
              <img alt="CSS ist valide!" src="../../../assets/valid-css-blue-v.svg" style="border:0;width:88px;height:31px"/>
            </a>
          </p>
        </div>
        📝 ❌ 🔐 🔓 🌸 🐳  alt ok: ⛅ 🎑 📜 📄 🔧 🔨 🎨 📰 ⚛ ⚛ ⚛ ⚛ ⚛
      </body>
    </html>
  </xsl:template>

  <xsl:template match="a:entry">
    <li id="{substring-after(a:id, '#')}" class="clearfix">
      <p class="small">
      	<img alt="Vorschaubild" class="img-responsive pull-right" src="https://links.mro.name/?do=genthumbnail&amp;hmac=d8f746960e34eb1ece5cb067a03363f71d419bb7bba2a7d146ee7be3026ad3c6&amp;url=https%3A%2F%2Fheise.de%2F-3619788"/>
      	
        <xsl:variable name="entry_updated" select="a:updated"/>
        <xsl:variable name="entry_updated_human"><xsl:call-template name="human_time"><xsl:with-param name="time" select="$entry_updated"/></xsl:call-template></xsl:variable>
        <a class="time" title="{$entry_updated}" href="{a:id}">🔗 <xsl:value-of select="$entry_updated_human"/></a>
        <xsl:text> ~ </xsl:text>
        <a title="Archiv" href="https://web.archive.org/web/{a:link[not(@rel)]/@href}">@archive.org</a>
      </p>
      <h4><a href="{a:link[not(@rel)]/@href}"><xsl:value-of select="a:title"/></a></h4>
      <h5><xsl:value-of select="a:summary"/></h5>
      <div>
        <div class="renderhtml">
          <!-- html content won't work that easy (out-of-the-firebox): https://bugzilla.mozilla.org/show_bug.cgi?id=98168#c140 -->
          <!-- workaround via jquery: http://stackoverflow.com/a/9714567 -->
          
          <!-- Überbleibsel vom Shaarli Atom Feed raus: -->
          <xsl:value-of select="substring-before(a:content[not(@src)], '&lt;br&gt;(&lt;a href=&quot;https://links.mro.name/?')" disable-output-escaping="yes" />
        </div>
        <p class="categories" title="Schlagworte">
          <xsl:for-each select="a:category">
          	<xsl:sort select="@term"/>
            <a href="../tags/{@term}/">#<xsl:value-of select="@term"/></a><xsl:text> </xsl:text>
          </xsl:for-each>
        </p>
      </div>
    </li>
  </xsl:template>

</xsl:stylesheet>
