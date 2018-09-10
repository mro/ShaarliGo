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

  <xsl:variable name="xml_base_pub">../../=</xsl:variable>
  <xsl:variable name="skin_base" select="concat($xml_base_pub,'/../assets/default')"/>
  <xsl:variable name="cgi_base" select="concat($xml_base_pub,'/../shaarligo.cgi')"/>

  <xsl:template match="/">
    <xsl:apply-templates select="h:html"/>
  </xsl:template>

  <xsl:template match="h:html">
    <html xmlns="http://www.w3.org/1999/xhtml" style="background-color:#3d2400">
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
      <link rel="icon" data-emoji="🌺" type="image/png"/>
      <link href="{$skin_base}/bootstrap.css" rel="stylesheet" type="text/css"/>
      <link href="{$skin_base}/bootstrap-theme.css" rel="stylesheet" type="text/css"/>

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
  padding-bottom: 1ex;
}

#links_commands {
  margin: 2ex 0;
}
.table {
  max-width: 100%;
  width: 100%;
}
form {
  margin: 1.0ex 0;
}

#links_commands td {
  min-width: 40px;
}

br.br { display:none; }
      </style>
      <title>Tools</title>
    </head>
  </xsl:template>

  <xsl:template name="links_commands">
    <table id="links_commands" class="toolbar table table-bordered table-striped table-inverse" aria-label="Befehle">
      <tbody>
        <tr>
          <td class="text-left">
            <a tabindex="10" href="{$xml_base_pub}/p/">
              <xsl:value-of select="/h:html/h:head/h:title"/>
            </a>
          </td>
          <td tabindex="20" class="text-right"><a href="{$xml_base_pub}/t/">⛅ <span class="hidden-xs"># Tags</span></a></td>
          <td tabindex="30" class="text-right"><a href="{$xml_base_pub}/d/">📅 <span class="hidden-xs">Tage</span></a></td>
          <td tabindex="40" class="text-right"><a href="{$xml_base_pub}/i/">🎨 <span class="hidden-xs">Bilder</span></a></td>
          <td class="text-right"><!-- I'd prefer a class="text-right hidden-logged-out" but just don't get it right -->
            <a class="hidden-logged-out" href="{$cgi_base}/tools/">🔨 <span class="hidden-xs">Tools</span></a>
          </td>
          <td class="text-right">
            <a tabindex="50" id="link_login" href="{$cgi_base}?do=login" class="visible-logged-out"><span class="hidden-xs">Anmelden</span> 🌺 </a>
            <a tabindex="51" id="link_logout" href="{$cgi_base}?do=logout" class="hidden-logged-out"><span class="hidden-xs">Abmelden</span> 🐾 </a>
          </td>
        </tr>
      </tbody>
    </table>
  </xsl:template>

  <xsl:template match="h:body">
    <body>
<!--
   onload="document.getElementById('q').removeAttribute('autofocus');document.getElementById('post').setAttribute('autofocus', 'autofocus');"
   onload="document.form_post.post.focus();"
-->
      <script>
const xml_base_pub = '<xsl:value-of select="$xml_base_pub"/>';
const skin_base    = '<xsl:value-of select="$skin_base"/>';
const cgi_base     = '<xsl:value-of select="$cgi_base"/>';
// <![CDATA[
// check if we're logged-in (AJAX or Cookie?).
const xhr = new XMLHttpRequest();
xhr.onreadystatechange = function(data0) {
  if (xhr.readyState == 4) {
    console.log('xhr.status = ' + xhr.status);
    document.documentElement.classList.add(xhr.status === 200 ? 'logged-in' : 'logged-out');
    // store the result locally and use as initial value for later calls.
  }
}
xhr.timeout = 1000;
xhr.open('GET', cgi_base + '/session/');
xhr.send(null);

// https://koddsson.com/posts/emoji-favicon/
const favicon = document.querySelector("link[rel=icon]");
if (favicon) {
  const emoji = favicon.getAttribute("data-emoji");
  if (emoji) {
    const canvas = document.createElement("canvas");
    canvas.height = 64;
    canvas.width = 64;

    const ctx = canvas.getContext("2d");
    ctx.font = "64px serif";
    ctx.fillText(emoji, 0, 56);

    favicon.href = canvas.toDataURL();
  }
}
// ]]>
      </script>
      <div class="container">
        <noscript><p>JavaScript ist aus, es geht zwar (fast) alles auch ohne, aber mit ist's <em>schöner</em>.</p></noscript>

        <xsl:call-template name="links_commands"/>

        <ol>
          <xsl:apply-templates select="h:ol/h:li"/>
        </ol>

        <hr style="clear:left;"/>

        <p><a href="http://purl.mro.name/ShaarliGo"><img style="background-color:#10b210" src="{$skin_base}/qrcode.png" alt="QR Code"/></a></p>

      </div>
    </body>
  </xsl:template>

  <xsl:template match="h:li[@id='bookmarklet']">
    <li>
      <a class="btn btn-default btn-lg active" onclick="{h:a/@onclick}" href="{h:a/@href}">
        <xsl:value-of select="h:a"/>
      </a>
      <xsl:copy-of select="h:span"/>
    </li>
  </xsl:template>

  <xsl:template match="h:li">
    <xsl:copy-of select="."/>
  </xsl:template>

</xsl:stylesheet>
