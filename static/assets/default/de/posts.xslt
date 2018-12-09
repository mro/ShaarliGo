<?xml version="1.0" encoding="UTF-8"?>
<!--
  ShaarliGo, microblogging detox
  Copyright (C) 2017-2018  Marcus Rohrmoser, http://purl.mro.name/ShaarliGo

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

  https://www.w3.org/TR/xslt-10/
-->
<xsl:stylesheet
  xmlns="http://www.w3.org/1999/xhtml"
  xmlns:a="http://www.w3.org/2005/Atom"
  xmlns:opensearch="http://a9.com/-/spec/opensearch/1.1/"
  xmlns:media="http://search.yahoo.com/mrss/"
  xmlns:georss="http://www.georss.org/georss"
  xmlns:sg="http://purl.mro.name/ShaarliGo/"
  xmlns:xsl="http://www.w3.org/1999/XSL/Transform"
  exclude-result-prefixes="a opensearch media georss sg"
  xmlns:math="http://exslt.org/math"
  extension-element-prefixes="math"
  version="1.0">

  <!-- xsl:variable name="redirector">https://anonym.to/?</xsl:variable --> <!-- mask the HTTP_REFERER -->
  <xsl:variable name="redirector"></xsl:variable>
  <xsl:variable name="archive">https://web.archive.org/web/</xsl:variable>

  <!-- replace linefeeds with <br> tags -->
  <xsl:template name="linefeed2br">
    <xsl:param name="string" select="''"/>
    <xsl:param name="pattern" select="'&#10;'"/>
    <xsl:choose>
      <xsl:when test="contains($string, $pattern)">
        <xsl:value-of select="substring-before($string, $pattern)"/><br class="br"/><xsl:comment> Why do we see 2 br on Safari and output/@method=html here? http://purl.mro.name/safari-xslt-br-bug </xsl:comment>
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

  <xsl:variable name="xml_base" select="/*/@xml:base"/>
  <xsl:variable name="xml_base_pub" select="concat($xml_base,'o')"/>
  <xsl:variable name="skin_base" select="concat($xml_base,'assets/default')"/>
  <xsl:variable name="cgi_base" select="concat($xml_base,'shaarligo.cgi')"/>

  <xsl:template match="/">
    <!--
      Do not set a class="logged-out" initially, but do via early JavaScript.
      If JavaScript is off, we need mixture between logged-in and -out.
    -->
    <html xmlns="http://www.w3.org/1999/xhtml" data-xml-base-pub="{$xml_base_pub}" style="background-color:#3d2400">
      <xsl:call-template name="head"/>

      <body>
        <xsl:apply-templates select="a:feed|a:entry" mode="root"/>
      </body>
    </html>
  </xsl:template>

  <xsl:template name="head">
    <head>
      <meta content="text/html; charset=utf-8" http-equiv="content-type"/>
      <!-- https://developer.apple.com/library/IOS/documentation/AppleApplications/Reference/SafariWebContent/UsingtheViewport/UsingtheViewport.html#//apple_ref/doc/uid/TP40006509-SW26 -->
      <!-- http://maddesigns.de/meta-viewport-1817.html -->
      <!-- meta name="viewport" content="width=device-width"/ -->
      <!-- http://www.quirksmode.org/blog/archives/2013/10/initialscale1_m.html -->
      <meta name="viewport" content="width=device-width,initial-scale=1.0"/>
      <!-- meta name="viewport" content="width=400"/ -->
      <link rel="icon" data-emoji="🌺" type="image/png"/>
      <link href="{$skin_base}/combined.css" rel="stylesheet" type="text/css"/>
      <script src="{$skin_base}/awesomplete.js"><!-- async="true" fails --></script>
      <script src="{$skin_base}/posts.js"><!-- async="true" fails --></script>

      <link href="." rel="alternate" type="application/atom+xml"/>
      <link href="." rel="self" type="application/xhtml+xml"/>

      <style type="text/css">
.hidden-logged-in { display:initial; }
.logged-in .hidden-logged-in { display:none; }
.visible-logged-in { display:none; }
.logged-in .visible-logged-in { display:initial; }

.hidden-logged-out { display:initial; }
.logged-out .hidden-logged-out { display:none; }
.visible-logged-out { display:none; }
.logged-out .visible-logged-out { display:initial; }

body {
  background: none;
}
.container {
  background-color: black;
}

#links_commands {
  margin: 2ex 0;
}
.table {
  width: 100%;
  max-width: 100%;
}
table.prev-next td {
  width: 10%;
  padding: 2ex 0;
}
li {
  margin: 2ex -1ex;
  padding: 1ex;
}
p {
  overflow-wrap: break-word;
  word-break: break-word;
  hyphens: auto;
}
form {
  margin: 1.0ex 0;
}
form button { min-width: 14ex; }

#links_commands td {
  min-width: 40px;
}

img.img-thumbnail {
  max-width: 120px;
  max-height: 120px;
  height: auto;
}

#demo {
  display: none;
}

/* This is a workaround for Browsers that insert additional br tags.
 * See http://purl.mro.name/safari-xslt-br-bug */
.rendered.type-text br { display:none; }
.rendered.type-text br.br { display:inline; }

/* I'm surprised, that I need to fiddle: */
.awesomplete > ul { top: 5ex; z-index: 3; }
div.awesomplete { display: block; }

table.prev-next a {
  padding: 0.75ex;
}
      </style>
      <title><xsl:value-of select="a:*/a:title"/></title>
    </head>
  </xsl:template>

  <xsl:template match="a:feed" mode="root">
    <div class="container">
      <noscript><p>JavaScript ist aus, es geht zwar (fast) alles auch ohne, aber mit ist's <em>schöner</em>.</p></noscript>

      <xsl:call-template name="links_commands"/>

      <xsl:call-template name="prev-next"/>

      <!-- <h1><xsl:value-of select="a:title"/></h1> -->

      <xsl:if test="a:subtitle">
        <h2><xsl:value-of select="a:subtitle"/></h2>
      </xsl:if>

      <p id="tags" class="categories">

        <xsl:variable name="countMax">
          <!-- https://stackoverflow.com/a/17966412 -->
          <xsl:for-each select="a:category">
            <xsl:sort select="@label" data-type="number" order="descending"/>
            <xsl:if test="position() = 1"><xsl:value-of select="@label"/></xsl:if>
          </xsl:for-each>
        </xsl:variable>

        <xsl:variable name="labelsDesc">
          <xsl:for-each select="a:category">
            <xsl:sort select="@label" order="descending"/>
            <xsl:value-of select="@label"/>
          </xsl:for-each>
        </xsl:variable>
        <xsl:for-each select="a:category">
          <xsl:sort select="@term" order="ascending"/>
          <!-- not log, just linear, similar to https://github.com/sebsauvage/Shaarli/blob/master/index.php#L1254 -->
          <xsl:variable name="size" select="8 + 40 * @label div $countMax"/>
          <a style="font-size:{$size}pt" href="{$cgi_base}/search/?q=%23{@term}+" class="tag"><span class="label"><xsl:value-of select="@term"/></span><span style="font-size:8pt">&#160;(<span class="count"><xsl:value-of select="@label"/></span>)</span></a><xsl:text>, </xsl:text>
        </xsl:for-each>
      </p>

      <ol id="entries" class="list-unstyled">
        <xsl:apply-templates select="a:entry"/>
      </ol>

      <xsl:call-template name="prev-next"/>

      <xsl:call-template name="footer"/>

      <p id="demo">
🍂 🍃 📝 ❌ 🔐 🔓 🌸 🐳  alt ok: ⛅ 🎑 📜 📄 🔧 🔨 🎨 📰 ⚛ ⚛ ⚛ ⚛ ⚛
      </p>
    </div>
  </xsl:template>

  <xsl:template name="links_commands">
    <table id="links_commands" class="toolbar table table-bordered table-striped table-inverse" aria-label="Befehle">
      <tbody>
        <tr>
          <td class="text-left">
            <a tabindex="10" href="{$xml_base_pub}/p/">
              <xsl:choose>
                <xsl:when test="a:link[@rel = 'up']/@title">
                  <xsl:value-of select="a:link[@rel = 'up']/@title"/>
                </xsl:when>
                <xsl:otherwise>
                  <xsl:value-of select="a:title"/>
                </xsl:otherwise>
              </xsl:choose>
            </a>
          </td>
          <td tabindex="20" class="text-right"><a href="{$xml_base_pub}/t/">⛅ <span class="hidden-xs"># Tags</span></a></td>
          <td tabindex="30" class="text-right"><a href="{$xml_base_pub}/d/">📅 <span class="hidden-xs">Tage</span></a></td>
          <td tabindex="40" class="text-right"><a href="{$xml_base_pub}/i/">🎨 <span class="hidden-xs">Bilder</span></a></td>
          <td class="text-right"><!-- I'd prefer a class="text-right hidden-logged-out" but just don't get it right -->
            <a class="hidden-logged-out" href="{$cgi_base}/tools/" rel="nofollow">🔨 <span class="hidden-xs">Tools</span></a>
          </td>
          <td class="text-right">
            <a tabindex="50" id="link_login" href="{$cgi_base}?do=login" class="visible-logged-out" rel="nofollow"><span class="hidden-xs">Anmelden</span> 🌺 </a>
            <a tabindex="51" id="link_logout" href="{$cgi_base}?do=logout" class="hidden-logged-out" rel="nofollow"><span class="hidden-xs">Abmelden</span> 🐾 </a>
          </td>
        </tr>
      </tbody>
    </table>

    <xsl:comment> https://stackoverflow.com/a/18520870 http://jsfiddle.net/66Ynx/ </xsl:comment>
    <form id="form_search" name="form_search" class="form-horizontal form-search" action="{$cgi_base}/search/">
      <div class="input-group">
        <input tabindex="100" name="q" value="{@sg:searchTerms}" autofocus="autofocus" type="text" placeholder="Suche Wort oder #Tag..." class="awesomplete form-control search-query" data-multiple="true"/>
        <span class="input-group-btn">
          <button tabindex="200" type="submit" class="btn btn-primary">Suche</button>
        </span>
      </div>
    </form>

    <form id="form_post" name="form_post" class="form-horizontal" action="{$cgi_base}">
      <div class="input-group">
        <input tabindex="300" name="post" type="text" placeholder="Was gibt's #Neues? (Notiz oder URL)" class="awesomplete form-control" data-multiple="true"/>
        <span class="input-group-btn">
          <button tabindex="400" type="submit" class="btn btn-primary">Shaaaare!</button>
        </span>
      </div>
    </form>
  </xsl:template>

  <xsl:template name="prev-next">
    <xsl:if test="a:link[@rel='first'] or a:link[@rel='last']">
      <table class="table prev-next">
        <tbody>
          <tr>
            <td class="text-left">
              <xsl:variable name="disabled"><xsl:if test="a:link[@rel='first']/@href = a:link[@rel='self']/@href">disabled</xsl:if></xsl:variable>
              <a href="{$xml_base}{a:link[@rel='first']/@href}" class="{$disabled} btn btn-primary btn-sm"><xsl:value-of select="a:link[@rel='first']/@title"/>&#160;&lt;&lt;</a>
            </td>
            <td class="text-center">
              <xsl:variable name="disabled"><xsl:if test="not(a:link[@rel='previous'])">disabled</xsl:if></xsl:variable>
              <a href="{$xml_base}{a:link[@rel='previous']/@href}" class="{$disabled} btn btn-primary btn-sm"><xsl:value-of select="a:link[@rel='previous']/@title"/>&#160;&lt;</a>
            </td>
            <td class="text-center">
              <span class="hidden-xs">Seite&#160;</span><xsl:value-of select="a:link[@rel='self']/@title"/>
            </td>
            <td class="text-center">
              <xsl:variable name="disabled"><xsl:if test="not(a:link[@rel='next'])">disabled</xsl:if></xsl:variable>
              <a href="{$xml_base}{a:link[@rel='next']/@href}" class="{$disabled} btn btn-primary btn-sm">&gt;&#160;<xsl:value-of select="a:link[@rel='next']/@title"/></a>
            </td>
            <td class="text-right">
              <xsl:variable name="disabled"><xsl:if test="a:link[@rel='last']/@href = a:link[@rel='self']/@href">disabled</xsl:if></xsl:variable>
              <a href="{$xml_base}{a:link[@rel='last']/@href}" class="{$disabled} btn btn-primary btn-sm">&gt;&gt;&#160;<xsl:value-of select="a:link[@rel='last']/@title"/></a>
            </td>
          </tr>
        </tbody>
      </table>
    </xsl:if>
  </xsl:template>

  <xsl:template name="footer">
    <hr style="clear:left;"/>
    <p id="footer">
      <a title="Abonnieren" href="{$xml_base}{a:link[@rel='self']/@href}">
        <img alt="Feed" src="{$skin_base}/feed-icon.svg" style="border:0;width:27px;height:27px"/>
      </a>
      <xsl:text> </xsl:text>
       <a title="Prüfen (Atom 1.0)" href="https://validator.w3.org/feed/check.cgi?url={$xml_base}{a:link[@rel='self']/@href}">
        <img alt="Prüfplakette (Atom 1.0)" src="{$skin_base}/valid-atom.svg" style="border:0;width:77px;height:27px"/>
      </a>
      <!-- <xsl:text> </xsl:text>
      <a href="https://validator.w3.org/check?uri=referer">
        <img alt="Valid XHTML 1.0 Strict" src="{$skin_base}/valid-xhtml10-blue-v.svg" style="border:0;width:88px;height:31px"/>
      </a>
      <a href="https://jigsaw.w3.org/css-validator/check/referer?profile=css3&amp;usermedium=screen&amp;warning=2&amp;vextwarning=false&amp;lang=de">
        <img alt="CSS ist valide!" src="{$skin_base}/valid-css-blue-v.svg" style="border:0;width:88px;height:31px"/>
      </a>
      -->
      <xsl:text> </xsl:text>
      <a href="{a:link[@rel='about']/@href}">
        Über<xsl:if test="string-length(a:link[@rel='about']/@href) &lt; 2"><span>&lt;= Link fehlt!</span></xsl:if>
      </a>
      <xsl:text> </xsl:text>
      <a href="{a:link[@rel='license']/@href}">
        <xsl:value-of select="a:link[@rel='license']/@title"/><xsl:if test="string-length(a:link[@rel='license']/@href) &lt; 2"><span>&lt;= Link fehlt!</span></xsl:if>
      </a>
      <xsl:text> </xsl:text>
      <a href="{a:link[@rel='terms-of-service']/@href}">
        Impressum<xsl:if test="string-length(a:link[@rel='terms-of-service']/@href) &lt; 2"><span>&lt;= Link fehlt!</span></xsl:if>
      </a>
      <xsl:text> </xsl:text>
      <a href="{a:link[@rel='privacy-policy']/@href}">
        Datenschutzerklärung<xsl:if test="string-length(a:link[@rel='terms-of-service']/@href) &lt; 2"><span>&lt;= Link fehlt!</span></xsl:if>
      </a>
      <xsl:text> </xsl:text>
      <a title="Generator" href="{a:generator/@uri}#{a:generator/@version}">
        <xsl:value-of select="a:generator"/>
        <xsl:text> </xsl:text>
        <img style="background-color:#10b210;width:27px;height:27px" src="{$skin_base}/qrcode.png" alt="QR Code (Generator URI)"/>
      </a>
    </p>
  </xsl:template>


  <xsl:template match="a:entry" mode="root">
    <div class="container">
      <noscript><p>JavaScript ist aus, es geht zwar (fast) alles auch ohne, aber mit ist's <em>schöner</em>.</p></noscript>

      <xsl:call-template name="links_commands"/>

      <ol id="entries" class="list-unstyled">
        <xsl:apply-templates select="."/>
      </ol>

      <xsl:call-template name="footer"/>
    </div>
  </xsl:template>

  <xsl:template match="a:entry">
    <xsl:variable name="link" select="a:link[not(@rel)]/@href"/>
    <xsl:variable name="self" select="a:link[@rel='self']/@href"/>
    <xsl:variable name="id_slash" select="substring-after($self, '/p/')"/>
    <xsl:variable name="id" select="substring-before($id_slash, '/')"/>
    <li id="{$id}" class="clearfix">
      <p class="small text-right">
        <xsl:if test="media:thumbnail/@url">
          <a href="{$redirector}{$link}" rel="noopener noreferrer" referrerpolicy="no-referrer">
            <!-- https://varvy.com/pagespeed/defer-images.html -->
            <img alt="Vorschaubild" class="img-thumbnail pull-right" data-src="{media:thumbnail/@url}" src="data:image/png;base64,R0lGODlhAQABAAD/ACwAAAAAAQABAAACADs=" />
          </a>
        </xsl:if>
      </p>
      <h4>
        <xsl:choose>
          <xsl:when test="$link">
            <a href="{$redirector}{$link}" title="Original" rel="noopener noreferrer" referrerpolicy="no-referrer"><xsl:value-of select="a:title"/></a>
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
      <div class="content">
        <!-- p class="small text-right"><a><xsl:value-of select="$link"/></a></p -->

        <!-- html content won't work that easy (out-of-the-firebox): https://bugzilla.mozilla.org/show_bug.cgi?id=98168#c140 -->
        <!-- workaround via jquery: http://stackoverflow.com/a/9714567 -->

        <!-- Überbleibsel vom Shaarli Atom Feed raus: -->
        <!-- xsl:value-of select="substring-before(a:content[not(@src)], '&lt;br&gt;(&lt;a href=&quot;https://links.mro.name/?')" disable-output-escaping="yes" / -->
        <xsl:apply-templates select="a:content"/>

        <p class="categories" title="Schlagworte">
          <xsl:for-each select="a:category">
            <xsl:sort select="@term"/>
            <a href="{@scheme}{@term}/" class="tag">#<xsl:value-of select="@term"/></a><xsl:text>, </xsl:text>
          </xsl:for-each>
        </p>
      </div>
      <p>
        <xsl:variable name="entry_updated" select="a:updated"/>
        <xsl:variable name="entry_updated_human"><xsl:call-template name="human_time"><xsl:with-param name="time" select="$entry_updated"/></xsl:call-template></xsl:variable>
        <xsl:variable name="entry_published" select="a:published"/>
        <xsl:variable name="entry_published_human"><xsl:call-template name="human_time"><xsl:with-param name="time" select="$entry_published"/></xsl:call-template></xsl:variable>

        <a class="time" title="zuletzt: {$entry_updated_human}" href="{$xml_base}{a:link[@rel='self']/@href}"><xsl:value-of select="$entry_published_human"/></a>
        <xsl:if test="$link">
          <xsl:text> ~ </xsl:text>
          <a title="Archiv" href="{$archive}{$link}" rel="noopener noreferrer" referrerpolicy="no-referrer">@archive.org</a>
        </xsl:if>
        <span class="hidden-logged-out" title="Bearbeiten">
          <xsl:text> ~ </xsl:text>
          <a href="{$xml_base}{a:link[@rel='edit']/@href}" rel="nofollow"><span class="hidden-xs">🔧</span>🔨</a><xsl:text> </xsl:text>
        </span>
      </p>
    </li>
  </xsl:template>

  <xsl:template match="a:content[not(@type) or @type = 'text']">
    <p class="rendered type-text">
      <xsl:call-template name="linefeed2br">
        <xsl:with-param name="string" select="."/>
      </xsl:call-template>
    </p>
  </xsl:template>

</xsl:stylesheet>
