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

  <!-- tags -->
  <xsl:template name="tags_with_hash">
    <xsl:param name="string" select="''"/>
    <xsl:param name="pattern" select="' '"/>
    <xsl:if test="$string != ''">
      <xsl:text> #</xsl:text>
    </xsl:if>
    <xsl:choose>
      <xsl:when test="contains($string, $pattern)">
        <xsl:value-of select="substring-before($string, $pattern)"/>
        <xsl:call-template name="tags_with_hash">
          <xsl:with-param name="string" select="substring-after($string, $pattern)"/>
          <xsl:with-param name="pattern" select="$pattern"/>
        </xsl:call-template>
      </xsl:when>
      <xsl:otherwise>
        <xsl:value-of select="$string"/>
      </xsl:otherwise>
    </xsl:choose>
  </xsl:template>

  <xsl:variable name="xml_base" select="/*/@xml:base"/>
  <xsl:variable name="xml_base_pub" select="concat($xml_base,'o')"/>
  <xsl:variable name="skin_base" select="concat($xml_base,'assets/default')"/>
  <xsl:variable name="cgi_base" select="concat($xml_base,'shaarligo.cgi')"/>

  <xsl:template match="/">
    <xsl:apply-templates select="h:html"/>
  </xsl:template>

  <xsl:template match="h:html">
    <html xmlns="http://www.w3.org/1999/xhtml" class="logged-out" data-xml-base-pub="{$xml_base_pub}">
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
      <link href="{$skin_base}/awesomplete.css" rel="stylesheet" type="text/css"/>
      <script src="{$skin_base}/awesomplete.js"><!-- async="true" fails --></script>
      <link href="{$skin_base}/style.css" rel="stylesheet" type="text/css"/>
      <script src="{$skin_base}/do-post.js"></script>

      <title>Shaaare!</title>
    </head>
  </xsl:template>

  <xsl:template name="links_commands">
  </xsl:template>

  <xsl:template match="h:body">
    <body>
      <div class="container">
        <noscript><p>JavaScript ist aus, es geht zwar (fast) alles auch ohne, aber mit ist's <em>schöner</em>.</p></noscript>

        <xsl:copy-of select="h:ul"/>
        <xsl:apply-templates select="h:form"/>
      </div>
    </body>
  </xsl:template>

  <xsl:template match="h:form[@name='linkform']">
    <form method="{@method}" name="{@name}">
      <xsl:copy-of select=".//h:input[@type='hidden']"/>
      <input name="lf_url" type="text" placeholder="https://..." value="{h:input[@name='lf_url']/@value}"/>
      <input autofocus="autofocus" name="lf_title" type="text" placeholder="Ein Titel, gerne mit #Schlagwort" value="{h:input[@name='lf_title']/@value}" class="awesomplete" data-multiple="true"/>
      <textarea name="lf_description" placeholder="Lorem #ipsum…" rows="14" cols="25" data-multiple="true">
        <xsl:value-of select="h:textarea[@name='lf_description']"/>
        <xsl:call-template name="tags_with_hash">
          <xsl:with-param name="string" select="h:input[@name='lf_tags']/@value"/>
        </xsl:call-template>
      </textarea>
  <!-- div class="input-group">
    <div class="col-sm-12">
      <input name="lf_tags" type="text" placeholder="Schlagwort NochEinSchlagwort" data-multiple="data-multiple" value="{h:input[@name='lf_tags']/@value}" class="form-control"/>
    </div>
  </div -->
  <!-- div class="input-group">
    <div class="col-sm-12">
      <input name="lf_private" type="checkbox" value="{h:input[@name='lf_private']/@value}" class="form-control"/>
    </div>
  </div -->
      <div style="display:flex; justify-content:space-between;">
        <button name="save_edit" type="submit" value="Save">Speichern</button>
        <button name="cancel_edit" type="submit" value="Cancel">Abbrechen</button>
        <button name="delete_edit" type="submit" value="Delete">Löschen</button>
      </div>
    </form>
  </xsl:template>

</xsl:stylesheet>
