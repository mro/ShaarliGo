<?xml version="1.0" encoding="UTF-8"?>
<!--
  ShaarliGo, microblogging detox
  Copyright (C) 2017-2017  Marcus Rohrmoser, http://purl.mro.name/ShaarliGo

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
  xmlns:h="http://www.w3.org/1999/xhtml"
  xmlns:xsl="http://www.w3.org/1999/XSL/Transform"
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

  <xsl:template match="/h:html/h:body/h:form">
    <html xmlns="http://www.w3.org/1999/xhtml">
      <head>
        <meta content="text/html; charset=utf-8" http-equiv="content-type"/>
        <!-- https://developer.apple.com/library/IOS/documentation/AppleApplications/Reference/SafariWebContent/UsingtheViewport/UsingtheViewport.html#//apple_ref/doc/uid/TP40006509-SW26 -->
        <!-- http://maddesigns.de/meta-viewport-1817.html -->
        <!-- meta name="viewport" content="width=device-width"/ -->
        <!-- http://www.quirksmode.org/blog/archives/2013/10/initialscale1_m.html -->
        <meta name="viewport" content="width=device-width,initial-scale=1.0"/>
        <!-- meta name="viewport" content="width=400"/ -->
        <link href="../assets/default/bootstrap.css" rel="stylesheet" type="text/css"/>
        <link href="../assets/default/bootstrap-theme.css" rel="stylesheet" type="text/css"/>

        <style type="text/css">
.table {
  width: 100%;
  max-width: 100%;
}
li {
  background-color: #F8F8F8;
  margin: 1em 0;
}
div.if_hasdiv_pwd { display:none; }
.has_pwd div.if_has_pwd { display:block; }
        </style>
        <title><xsl:value-of select="title"/></title>
      </head>
      <body onload="document.{@name}.title.focus();">
        <div class="container">
          <p><img
            alt="Sehr schön, der Webserver ist gut eingestellt, ./app/ ist geschützt."
            title="Wenn hier ein Filmzitat sichtbar ist, liegt ./app/ ungeschützt im Netz."
            src="../app/i-must-be-403.svg"/></p>

          <xsl:if test="setlogin = ''">
            <p>Huch, das sieht ja alles recht frisch aus hier.</p>
          </xsl:if>

          <form id="{@id}" name="{@name}" class="form-horizontal" method="POST">
            <!-- https://www.tjvantoll.com/2012/08/05/html5-form-validation-showing-all-error-messages/ -->

            <div class="form-group">
              <label for="title" class="control-label col-sm-1">Titel</label>
              <div class="col-sm-11">
                <input tabindex="100" name="title" autofocus="autofocus" type="text" placeholder="My ShaarliGo 🐳" required="required" pattern="\S(.*\S)?" value="{h:input[@name='title']/@value}" class="form-control"/>
              </div>
            </div>
            <div class="form-group">
              <label for="setlogin" class="control-label col-sm-1">User</label>
              <div class="col-sm-11">
                <input tabindex="200" name="setlogin" type="text" placeholder="Benutzername dieses neuen ShaarliGo" required="required" pattern="\S(.*\S)?" _oninvalid="setCustomValidity('Das ist nicht Dein Ernst oder?')" value="{h:input[@name='setlogin']/@value}" class="form-control"/>
              </div>
            </div>

            <xsl:if test="h:input[@name='oldpassword']/@value != ''">
              <div class="if_has_pwd form-group">
                <label for="oldpassword" class="control-label col-sm-1">Pwd (bestehend)</label>
                <div class="col-sm-11">
                  <input tabindex="300" name="oldpassword" type="password" placeholder="Das bisherige Passwort" required="required" minlength="12" pattern="\S(.*\S)?" value="{h:input[@name='oldpassword']/@value}" class="form-control"/>
                </div>
              </div>
            </xsl:if>
            <div class="form-group">
              <label for="setpassword" class="control-label col-sm-1">Pwd</label>
              <div class="col-sm-11">
                <input tabindex="400" name="setpassword" type="password" placeholder="gute Passworte: xkcd.com/936" required="required" minlength="12" pattern="\S(.*\S)?" value="{h:input[@name='setpassword']/@value}" class="form-control"/>
              </div>
            </div>
            <div class="if_has_pwd form-group">
              <label for="confirmpassword" class="control-label col-sm-1">Pwd (Wiederholung)</label>
              <div class="col-sm-11">
                <input tabindex="500" name="confirmpassword" type="password" placeholder="dasselbe nochmal" required="required" minlength="12" pattern="\S(.*\S)?" class="form-control"/>
              </div>
            </div>
            <!-- evtl. Zeitzone, continent / city? -->

            <p>Mag man Material aus einem alten Shaarli übernehmen?</p>

            <div class="form-group">
              <label for="import_shaarli_url" class="control-label col-sm-1">alte Shaarli Adresse</label>
              <div class="col-sm-11">
                <input tabindex="600" name="import_shaarli_url" type="url" placeholder="example.com/shaarli" pattern="\S+" class="form-control"/>
              </div>
            </div>
            <div class="form-group">
              <label for="import_shaarli_setlogin" class="control-label col-sm-1">Benutzer</label>
              <div class="col-sm-11">
                <input tabindex="700" name="import_shaarli_setlogin" type="text" placeholder="Benutzername des alten Shaarli" class="form-control"/>
              </div>
            </div>
            <div class="form-group">
              <label for="import_shaarli_setpassword" class="control-label col-sm-1">Pwd</label>
              <div class="col-sm-11">
                <input tabindex="800" name="import_shaarli_password" type="password" placeholder="Passwort des alten Shaarli" class="form-control"/>
              </div>
            </div>

            <div class="form-group">
              <div class="col-sm-11 col-sm-offset-1">
                <button tabindex="900" type="submit" class="btn btn-primary">Loooooos!</button>
              </div>
            </div>
          </form>
        </div>
      </body>
    </html>
  </xsl:template>

</xsl:stylesheet>
