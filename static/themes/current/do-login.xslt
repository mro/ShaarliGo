<?xml version="1.0" encoding="UTF-8"?>
<!--
-->
<xsl:stylesheet
  xmlns="http://www.w3.org/1999/xhtml"
  xmlns:h="http://www.w3.org/1999/xhtml"
  xmlns:xsl="http://www.w3.org/1999/XSL/Transform"
  version="1.0">

  <xsl:output
    method="html"
    doctype-system="http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd"
    doctype-public="-//W3C//DTD XHTML 1.0 Strict//EN"/>

  <xsl:variable name="xml_base" select="/*/@xml:base"/>
  <xsl:variable name="xml_base_pub" select="concat($xml_base,'o')"/>
	<xsl:variable name="skin_base" select="concat($xml_base,'themes/current')"/>
  <xsl:variable name="cgi_base" select="concat($xml_base,'shaarli.cgi')"/>

  <xsl:template match="/">
    <xsl:apply-templates select="h:html"/>
  </xsl:template>

  <xsl:template match="h:html">
		<html xmlns="http://www.w3.org/1999/xhtml" class="logged-out">
      <xsl:apply-templates select="h:head"/>
      <xsl:apply-templates select="h:body"/>
    </html>
  </xsl:template>

  <xsl:template match="h:head">
    <head>
      <meta content="text/html; charset=utf-8" http-equiv="content-type"/>
      <!-- https://developer.apple.com/library/IOS/documentation/AppleApplications/Reference/SafariWebContent/UsingtheViewport/UsingtheViewport.html#//apple_ref/doc/uid/TP40006509-SW26 -->
      <!-- http://maddesigns.de/meta-viewport-1817.html -->
      <!-- meta name="viewport" content="width=device-width"/ -->
      <!-- http://www.quirksmode.org/blog/archives/2013/10/initialscale1_m.html -->
      <meta name="viewport" content="width=device-width,initial-scale=1.0"/>
      <!-- meta name="viewport" content="width=400"/ -->
      <link href="{$skin_base}/style.css" rel="stylesheet" type="text/css"/>

      <title>Login</title>
    </head>
  </xsl:template>

  <xsl:template match="h:body">
    <body>
      <div class="container">
        <noscript><p>JavaScript deactivated, almost fully functional, but <em>nicer</em> if on.</p></noscript>

        <xsl:apply-templates select="h:form"/>
      </div>
    </body>
  </xsl:template>

  <xsl:template match="h:form[@name='loginform']">
    <form method="{@method}" name="{@name}">
      <input name="token" type="hidden" value="{h:input[@name='token']/@value}"/>
      <input name="returnurl" type="hidden" value="{h:input[@name='returnurl']/@value}"/>
      <input tabindex="100" name="login" type="text" autofocus="autofocus" placeholder="Who's there?" value="{h:input[@name='login']/@value}"/>
      <input tabindex="200" name="password" type="password" placeholder="Do you know the password?"/>
      <button tabindex="300" type="submit">Go!</button>
    </form>
  </xsl:template>

</xsl:stylesheet>
