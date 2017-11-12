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

  <xsl:variable name="xml_base_pub">pub</xsl:variable>

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
      <link href="{$xml_base_pub}/../assets/default/bootstrap.css" rel="stylesheet" type="text/css"/>
      <link href="{$xml_base_pub}/../assets/default/bootstrap-theme.css" rel="stylesheet" type="text/css"/>

      <style type="text/css">
.hidden-logged-in { display:initial; }
.logged-in .hidden-logged-in { display:none; }
.visible-logged-in { display:none; }
.logged-in .visible-logged-in { display:initial; }

.hidden-logged-out { display:initial; }
.logged-out .hidden-logged-out { display:none; }
.visible-logged-out { display:none; }
.logged-out .visible-logged-out { display:initial; }

.container {
}

#links_commands {
margin: 2ex 0;
}
.table {
width: 100%;
max-width: 100%;
}
form {
margin: 1.0ex 0;
}

#links_commands td {
min-width: 40px;
}
      </style>
      <title>Shaaare!</title>
    </head>
  </xsl:template>

  <xsl:template name="links_commands">
    <table id="links_commands" class="toolbar table table-bordered table-striped table-inverse" aria-label="Befehle">
      <tbody>
        <tr>
          <td class="text-left">
            <a href="{$xml_base_pub}/posts/">
              <xsl:value-of select="/h:html/h:head/h:title"/>
<!--              <xsl:choose>
                <xsl:when test="a:link[@rel = 'up']/@title">
                  <xsl:value-of select="a:link[@rel = 'up']/@title"/>
                </xsl:when>
                <xsl:otherwise>
                  <xsl:value-of select="a:title"/>
                </xsl:otherwise>
              </xsl:choose>
-->           </a>
          </td>
          <td class="text-right"><a href="{$xml_base_pub}/tags/">⛅ <span class="hidden-xs"># Tags</span></a></td>
          <td class="text-right"><a href="{$xml_base_pub}/days/">📅 <span class="hidden-xs">Tage</span></a></td>
          <td class="text-right"><a href="{$xml_base_pub}/imgs/">🎨 <span class="hidden-xs">Bilder</span></a></td>
          <td class="text-right hidden-logged-out"><a href="{$xml_base_pub}/../shaarligo.cgi?do=tools">🔨 <span class="hidden-xs">Tools</span></a></td>
          <td class="text-right">
            <a id="link_login" href="{$xml_base_pub}/../shaarligo.cgi?do=login" class="visible-logged-out"><span class="hidden-xs">Anmelden</span> 🌺 </a>
            <a id="link_logout" href="{$xml_base_pub}/../shaarligo.cgi?do=logout" class="hidden-logged-out"><span class="hidden-xs">Abmelden</span> 🍃 </a>
          </td>
        </tr>
      </tbody>
    </table>
  </xsl:template>

  <xsl:template match="h:body">
    <body>
      <div class="container">
        <noscript><p>JavaScript ist aus, es geht zwar (fast) alles auch ohne, aber mit ist's <em>schöner</em>.</p></noscript>

        <xsl:apply-templates select="h:form"/>
      </div>
    </body>
  </xsl:template>

  <xsl:template match="h:form[@name='linkform']">
    <form method="{@method}" name="{@name}" class="form-horizontal">
      <input name="token" type="hidden" value="{h:input[@name='token']/@value}"/>
      <input name="returnurl" type="hidden" value="{h:input[@name='returnurl']/@value}"/>
      <input name="lf_linkdate" type="hidden" value="{h:input[@name='lf_linkdate']/@value}" class="form-control"/>
      <div class="input-group">
        <div class="col-sm-12">
          <input name="lf_url" type="text" placeholder="https://..." value="{h:input[@name='lf_url']/@value}" class="form-control"/>
        </div>
      </div>
      <div class="input-group">
        <div class="col-sm-12">
          <input autofocus="autofocus" name="lf_title" type="text" placeholder="Ein Titel, gerne mit #Schlagwort" value="{h:input[@name='lf_title']/@value}" class="form-control"/>
        </div>
      </div>
      <div class="input-group">
        <div class="col-sm-12">
          <textarea name="lf_description" placeholder="Lorem #ipsum…" rows="4" cols="25" class="form-control"><xsl:value-of select="h:textarea[@name='lf_description']"/></textarea>
        </div>
      </div>
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
      <div class="input-group">
        <div class="col-sm-12">
          <span class="input-group-btn">
            <input name="save_edit" type="submit" value="Save" class="btn btn-primary"/>
          </span>
          <span class="input-group-btn">
            <input name="cancel_edit" type="submit" value="Cancel" class="btn btn-primary"/>
          </span>
        </div>
      </div>
    </form>
  </xsl:template>

</xsl:stylesheet>
