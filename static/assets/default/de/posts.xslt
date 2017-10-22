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
    <html xmlns="http://www.w3.org/1999/xhtml" class="loggedout">
      <head>
        <meta content="text/html; charset=utf-8" http-equiv="content-type"/>
        <!-- https://developer.apple.com/library/IOS/documentation/AppleApplications/Reference/SafariWebContent/UsingtheViewport/UsingtheViewport.html#//apple_ref/doc/uid/TP40006509-SW26 -->
        <!-- http://maddesigns.de/meta-viewport-1817.html -->
        <!-- meta name="viewport" content="width=device-width"/ -->
        <!-- http://www.quirksmode.org/blog/archives/2013/10/initialscale1_m.html -->
        <meta name="viewport" content="width=device-width,initial-scale=1.0"/>
        <!-- meta name="viewport" content="width=400"/ -->
        <link href="{@xml:base}assets/default/bootstrap.css" rel="stylesheet" type="text/css"/>
        <link href="{@xml:base}assets/default/bootstrap-theme.css" rel="stylesheet" type="text/css"/>

        <link href="." rel="alternate" type="application/atom+xml"/>
        <link href="." rel="self" type="application/xhtml+xml"/>

        <style type="text/css">
.container {
}
#links_commands {
  margin: 2ex 0;
}
.table {
  width: 100%;
  max-width: 100%;
}
li {
  background-color: #F8F8F8;
  margin: 1em 0;
}
.loggedout #post {
  display: none;
}
.loggedout .link_edit {
  display: none;
}
.loggedout #link_logout, .loggedout #link_tools {
  display: none;
}
.loggedin #link_login {
  display: none;
}
p.categories {
  margin: 1ex 0;
}
.categories a {
  padding: 0.5ex;
  background: linear-gradient(#F2F2F2, #ffffff);
  box-shadow: 0 0 2px rgba(0, 0, 0, 0.5);
  border-radius: 3px;
}

#demo {
  display: none;
}

/* This is a workaround for Browsers that insert additional br tags.
 * See http://purl.mro.name/safari-xslt-br-bug */
br { display:none; }
br.br { display:inline; }
        </style>
        <title><xsl:value-of select="a:title"/></title>
      </head>
      <body>
        <div class="container">
          <noscript><p>JavaScript ist aus, es geht zwar (fast) alles auch ohne, aber mit ist's <em>schöner</em>.</p></noscript>

          <table id="links_commands" class="toolbar table table-bordered table-striped table-inverse">
            <tbody>
              <tr>
                <td class="text-left"><a href="{@xml:base}pub/posts"><xsl:value-of select="a:title"/></a></td>
                <td class="text-center"><a href="../tags/">⛅ # Tags</a></td>
                <td class="text-center"><a href="../imgs/">🎨 Bilder</a></td>
                <td class="text-center"><a id="link_daily" href="../days/">📅 Tage</a></td>
                <td class="text-center"><a id="link_tools" href="{@xml:base}atom.cgi?do=tools">🔨 Tools</a></td>
                <td class="text-right">
                  <a id="link_login" href="{@xml:base}atom.cgi?do=login">Anmelden 🌺</a>
                  <a id="link_logout" href="{@xml:base}atom.cgi?do=logout">Abmelden 🍃</a>
                </td>
              </tr>
            </tbody>
          </table>

          <xsl:call-template name="prev-next"/>

          <!-- <h1><xsl:value-of select="a:title"/></h1> -->

          <xsl:if test="a:subtitle">
            <h2><xsl:value-of select="a:subtitle"/></h2>
          </xsl:if>

          <form id="post" class="form-horizontal" action="{@xml:base}atom.cgi?do=addlink" method="GET">
            <div class="form-group">
              <input type="text" class="form-control" id="post" name="post" placeholder="Was gibt's Neues? (Notiz oder URL)"/>
            </div>
            <div class="form-group" style="display:none">
              <input type="file" class="file pull-right" id="input-1" placeholder="Bild"/>
            </div>
            <div class="form-group text-right" style="display:none">
              <div class="form-check">
                <label class="form-check-label">
                  <input class="form-check-input" type="checkbox"/> Privat?
                </label>
              </div>
            </div>
            <div class="form-group text-right">
              <button type="submit" class="btn btn-primary">Shaaaare!</button>
            </div>
          </form>

          <ol id="entries" class="list-unstyled">
            <xsl:apply-templates select="a:entry"/>
          </ol>

          <xsl:call-template name="prev-next"/>

          <script src="{@xml:base}assets/default/script.js" type="text/javascript"></script>

          <hr style="clear:left;"/>
          <p id="footer">
            <a title="Validate my Atom 1.0 feed" href="https://validator.w3.org/feed/check.cgi?url={@xml:base}{a:link[@rel='self']/@href}">
              <img alt="Valid Atom 1.0" src="{@xml:base}assets/default/valid-atom.png" style="border:0;width:88px;height:31px"/>
            </a>
            <!-- <xsl:text> </xsl:text>
            <a href="https://validator.w3.org/check?uri=referer">
              <img alt="Valid XHTML 1.0 Strict" src="{@xml:base}assets/default/valid-xhtml10-blue-v.svg" style="border:0;width:88px;height:31px"/>
            </a>
            <a href="https://jigsaw.w3.org/css-validator/check/referer?profile=css3&amp;usermedium=screen&amp;warning=2&amp;vextwarning=false&amp;lang=de">
              <img alt="CSS ist valide!" src="{@xml:base}assets/default/valid-css-blue-v.svg" style="border:0;width:88px;height:31px"/>
            </a>
            -->
          </p>
          <p id="demo">
📝 ❌ 🔐 🔓 🌸 🐳  alt ok: ⛅ 🎑 📜 📄 🔧 🔨 🎨 📰 ⚛ ⚛ ⚛ ⚛ ⚛
          </p>
        </div>
      </body>
    </html>
  </xsl:template>

  <xsl:template name="prev-next">
    <xsl:if test="a:link[@rel='first'] or a:link[@rel='last']">
    <table class="table prev-next">
      <tbody>
        <tr>
          <xsl:if test="a:link[@rel='first']">
            <td class="text-left"  ><a href="{a:link[@rel='first']/@href}">Seite 1</a></td>
          </xsl:if>
          <xsl:if test="a:link[@rel='previous']">
            <td class="text-center"><a href="{a:link[@rel='previous']/@href}"><xsl:value-of select="a:link[@rel='previous']/@title"/></a></td>
          </xsl:if>
          <td class="text-center"><a href="{a:link[@rel='self']/@href}">Seite <xsl:value-of select="a:link[@rel='self']/@title"/></a></td>
          <xsl:if test="a:link[@rel='next']">
            <td class="text-center"><a href="{a:link[@rel='next']/@href}"><xsl:value-of select="a:link[@rel='next']/@title"/></a></td>
          </xsl:if>
          <xsl:if test="a:link[@rel='last']">
            <td class="text-right" ><a href="{a:link[@rel='last']/@href}">Seite <xsl:value-of select="a:link[@rel='last']/@title"/></a></td>
          </xsl:if>
        </tr>
      </tbody>
    </table>
    </xsl:if>
  </xsl:template>

  <xsl:template match="a:entry">
    <xsl:variable name="xml_base" select="/*/@xml:base"/>
    <xsl:variable name="link" select="a:link[@rel='alternate']/@href"/>
    <li id="{substring-after(a:link[@rel='self']/@href, '/posts/')}" class="clearfix">
      <p class="small text-right">
        <xsl:if test="media:thumbnail/@url">
          <a href="{$link}">
            <img alt="Vorschaubild" class="thumbnail img-responsive pull-right" src="{media:thumbnail/@url}"/>
          </a>
        </xsl:if>

        <xsl:variable name="entry_updated" select="a:updated"/>
        <xsl:variable name="entry_updated_human"><xsl:call-template name="human_time"><xsl:with-param name="time" select="$entry_updated"/></xsl:call-template></xsl:variable>

        <span class="link_edit" title="Bearbeiten">
          <a href="{$xml_base}{a:link[@rel='edit']/@href}">🔨</a><xsl:text> </xsl:text>
        </span>
        <a class="time" title="Einzelansicht" href="{$xml_base}{a:link[@rel='self']/@href}">¶ <xsl:value-of select="$entry_updated_human"/></a>
        <xsl:if test="$link">
          <xsl:text> ~ </xsl:text>
          <a title="Archiv" href="https://web.archive.org/web/{$link}">@archive.org</a>
        </xsl:if>
      </p>
      <h4>
        <xsl:choose>
          <xsl:when test="$link">
            <a href="{$link}" title="Original"><xsl:value-of select="a:title"/> 🚀</a>
          </xsl:when>
          <xsl:otherwise>
            <xsl:value-of select="a:title"/>
          </xsl:otherwise>
        </xsl:choose>
      </h4>
      <xsl:if test="a:summary">
        <h5>
          <xsl:call-template name="linefeed2br">
            <xsl:with-param name="string" select="a:summary"/>
          </xsl:call-template>
        </h5>
      </xsl:if>
      <div>
        <div class="renderhtml">
          <!-- html content won't work that easy (out-of-the-firebox): https://bugzilla.mozilla.org/show_bug.cgi?id=98168#c140 -->
          <!-- workaround via jquery: http://stackoverflow.com/a/9714567 -->

          <!-- Überbleibsel vom Shaarli Atom Feed raus: -->
          <!-- xsl:value-of select="substring-before(a:content[not(@src)], '&lt;br&gt;(&lt;a href=&quot;https://links.mro.name/?')" disable-output-escaping="yes" / -->
          <xsl:call-template name="linefeed2br">
            <xsl:with-param name="string" select="a:content"/>
          </xsl:call-template>
        </div>
        <p class="categories" title="Schlagworte">
          <xsl:for-each select="a:category">
            <xsl:sort select="@term"/>
            <a href="{@scheme}/{@term}">#<xsl:value-of select="@term"/></a><xsl:text> </xsl:text>
          </xsl:for-each>
        </p>
      </div>
    </li>
  </xsl:template>

</xsl:stylesheet>
