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
  xmlns:as="http://purl.mro.name/AtomicShaarli"
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

  <xsl:template match="as:setup">
    <html xmlns="http://www.w3.org/1999/xhtml">
      <head>
        <meta content="text/html; charset=utf-8" http-equiv="content-type"/>
        <!-- https://developer.apple.com/library/IOS/documentation/AppleApplications/Reference/SafariWebContent/UsingtheViewport/UsingtheViewport.html#//apple_ref/doc/uid/TP40006509-SW26 -->
        <!-- http://maddesigns.de/meta-viewport-1817.html -->
        <!-- meta name="viewport" content="width=device-width"/ -->
        <!-- http://www.quirksmode.org/blog/archives/2013/10/initialscale1_m.html -->
        <meta name="viewport" content="width=device-width,initial-scale=1.0"/>
        <!-- meta name="viewport" content="width=400"/ -->
        <link href="../assets/bootstrap.css" rel="stylesheet" type="text/css"/>
        <link href="../assets/bootstrap-theme.css" rel="stylesheet" type="text/css"/>

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
          <p><img alt="Sehr schön, der Webserver ist gut eingestellt." src="../app/i-must-be-403.svg"/></p>
          
          <xsl:if test="a:author/a:name = ''">
            <p>Huch, das sieht ja alles recht frisch aus hier.</p>
          </xsl:if>

          <form id="installform" name="installform" class="form-horizontal" method="POST" action="#">
            <!-- https://www.tjvantoll.com/2012/08/05/html5-form-validation-showing-all-error-messages/ -->

            <div class="form-group">
              <label for="title" class="control-label col-sm-1">Titel</label>
              <div class="col-sm-11">
                <input type="text" class="form-control" name="title" placeholder="My AtomicShaarli 🐳" required="required" pattern="\S(.*\S)?" value="{a:title}"/>
              </div>
            </div>
            <div class="form-group">
              <label for="author/name" class="control-label col-sm-1">User</label>
              <div class="col-sm-11">
                <input type="text" class="form-control" name="author/name" placeholder="Benutzername dieses neuen AtomicShaarli" required="required" pattern="\S(.*\S)?" _oninvalid="setCustomValidity('Das ist nicht Dein Ernst oder?')" value="{a:author/a:name}"/>
              </div>
            </div>
            <div class="form-group">
              <label for="password" class="control-label col-sm-1">Pwd</label>
              <div class="col-sm-11">
                <input type="password" class="form-control" name="password" placeholder="gute Passworte: xkcd.com/936" required="required" minlength="12" pattern="\S(.*\S)?"/>
              </div>
            </div>
            <!-- evtl. continent / city -->

            <p>Möchtest Du Daten aus einem früheren Shaarli übernehmen?</p>

            <div class="form-group">
              <label for="import_shaarli_url" class="control-label col-sm-1">URL</label>
              <div class="col-sm-11">
                <input type="text" class="form-control" name="import_shaarli_url" placeholder="example.com/shaarli" pattern="\S+"/>
              </div>
            </div>
            <div class="form-group">
              <label for="import_shaarli_setlogin" class="control-label col-sm-1">User</label>
              <div class="col-sm-11">
                <input type="text" class="form-control" name="import_shaarli_setlogin" placeholder="Benutzername des früheren Shaarli"/>
              </div>
            </div>
            <div class="form-group">
              <label for="import_shaarli_setpassword" class="control-label col-sm-1">Pwd</label>
              <div class="col-sm-11">
                <input type="password" class="form-control" name="import_shaarli_setpassword" placeholder="Passwort des früheren Shaarli"/>
              </div>
            </div>

            <div class="form-group">
              <div class="col-sm-11 col-sm-offset-1">
                <button type="submit" class="btn btn-primary">Loooooos!</button>
              </div>
            </div>
          </form>
        </div>
      </body>
    </html>
  </xsl:template>

</xsl:stylesheet>
