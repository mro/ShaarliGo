
🌩 [Lightning Talk at the 34c3](https://events.ccc.de/congress/2017/wiki/Lightning:ShaarliGo:_self-hosted_microblogging) 🚀

[![Build Status](https://travis-ci.org/mro/ShaarliGo.svg?branch=master)](https://travis-ci.org/mro/ShaarliGo)

# ShaarliGo

self-hosted microblogging inspired by
http://sebsauvage.net/wiki/doku.php?id=php:shaarli. Destilled down to the bare
minimum, with easy hosting and security in mind. No PHP, no DB, no server-side
templating, JS optional.

## Design Goals

- [x] backwards compatible posting (https://code.mro.name/mro/Shaarli-API-test)
- [x] trivial installation and minimal hosting requirements (run on simple hosted webspace),
- [x] keep server lean, especially for readers,
- [ ] standards compliant ([Atom](https://tools.ietf.org/html/rfc4287),
  [Atompub](https://tools.ietf.org/html/rfc5023),
  [WebSub](https://www.w3.org/TR/websub/)),
- [ ] easy migration from existing shaarlis,
- [x] run ok without javascript,
- [x] visitor reading operates on static flat files only (no server code),
- [ ] secure against brute force login attacks,
- [x] easy translation & skinning,
- [x] leverage existing, widely deployed web tec ([CGI](https://tools.ietf.org/html/rfc3875), [XSLT](https://www.w3.org/TR/xslt-10/),
  [HTML](https://www.w3.org/TR/xhtml11/), [CSS](https://www.w3.org/TR/CSS/)),
- [ ] easy fail2ban integration / DOS mitigation,

| Quality         | very good | good | normal | irrelevant |
|-----------------|:---------:|:----:|:------:|:----------:|
| Functionality   |           |      |    ×   |            |
| Reliability     |           |  ×   |        |            |
| Usability       |     ×     |      |        |            |
| Efficiency      |           |  ×   |        |            |
| Changeability   |           |  ×   |        |            |
| Portability     |           |  ×   |        |            |

## Dependencies

_tl;dr:_ a webserver that can execute [CGI](https://tools.ietf.org/html/rfc3875)s and serve files
from disc.

ShaarliGo is an old-school CGI binary executable, so it needs a webserver to drive it.
Configurations come for [Apache](http://httpd.apache.org/) (automatic, see `static/.htaccess`) and
[Lighttpd](http://www.lighttpd.net/) (see `static/app/lighttpd.conf`).

As a self-contained, statically linked, [Go](https://golang.org/) executable, it has no runtime
dependencies and works on a variety of platforms.

ShaarliGo needs write access to the webroot filesystem to once unpack the web assets and when posting
update the content.

Storage footprint is <25 [KiB](https://en.wikipedia.org/wiki/Kibibyte) per post.

When posting a page, it is once accessed via HTTP GET to infer title, tags and a thumbnail image
URL.

## Install / Update

If the webserver is Apache (Linux, 64 bit, set up and running, modules cgi and
rewrite):

1. Download http://purl.mro.name/Linux-x86_64/shaarligo.cgi,
2. copy this single file to your webspace,
3. set it's file permissions (chmod) to numeric 555 (readonly + executable for all),
4. visit in your browser: http://my.web.space/subdir/shaarligo.cgi,

done! For lighttpd see `static/app/lighttpd.conf`.

Or build from source at http://mro.name/ShaarliGo

## POSSE

> POSSE is an abbreviation for Publish (on your) Own Site, Syndicate Elsewhere, a
> content publishing model that starts with posting content on your own domain
> first, then syndicating out copies to 3rd party services with permashortlinks
> back to the original on your site.

(says https://indieweb.org/POSSE)

You can POSSE to

### Pinboard.in

enter your Pinboard Auth Token from https://pinboard.in/settings/password at the
end of `app/config.yaml` like this

```yaml
posse:
- pinboard: https://api.pinboard.in/v1?auth_token=johndoe:XOG86E7JIYMI
  prefix:
```

It's ok to leave `prefix` empty. Each pinboard post gets a backlink as an
additional footer line. If `prefix` is set, the footer line is `prefix` + `id`.

## Todos

1. private posts,
2. [PuSH/PubSubhubbub](https://github.com/pubsubhubbub/pubsubhubbub) / [WebSub](https://www.w3.org/TR/websub/),
3. import shaarlis (login?),
4. pwd reset (maybe deleting from `app/config.yaml` is acceptable),
5. images/enclosures,
7. comments,
8. trackback/pingback

