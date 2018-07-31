
🌩 [Lightning Talk at the 34c3](https://events.ccc.de/congress/2017/wiki/Lightning:ShaarliGo:_self-hosted_microblogging) 🚀

[![Build Status](https://travis-ci.org/mro/ShaarliGo.svg?branch=master)](https://travis-ci.org/mro/ShaarliGo)

# ShaarliGo

self-hosted microblogging inspired by
http://sebsauvage.net/wiki/doku.php?id=php:shaarli. Destilled down to the bare
minimum, with easy hosting and security in mind. No PHP, no DB, no server-side
templating, JS optional.

## Design Goals

- [x] backwards compatible posting (https://git.mro.name/mro/Shaarli-API-test)
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
| Efficiency      |           |      |    ×   |            |
| Changeability   |           |  ×   |        |            |
| Portability     |     ×     |      |        |            |

## Dependencies

_tl;dr:_ a webserver that can execute [CGI](https://tools.ietf.org/html/rfc3875)s and serve files
from disc.

ShaarliGo is an old-school CGI binary executable, so it needs a webserver to drive it. Example
configurations come for [Apache](http://httpd.apache.org/) (see `static/.htaccess`) and
[Lighttpd](http://www.lighttpd.net/) (see `static/app/lighttpd.conf`).

As a self-contained, statically linked, [Go](https://golang.org/) executable, it has no software
dependencies and can run on a variety of platforms.

It needs write access to it's webserver's filesystem location to unpack the web assets and update
the content when posting.

Storage footprint is <25 [KiB](https://en.wikipedia.org/wiki/Kibibyte) per post.

When posting a page, it is once accessed via HTTP GET to suggest title, tags and a thumbnail image
URL.

## Install / Update

Linux amd64:

1. `$ curl -L http://purl.mro.name/shaarligo_cgi.gz | tee shaarligo_cgi.gz | gunzip > shaarligo.cgi && chmod a+x shaarligo.cgi`
2. visit in your browser: http://my.web.space/subdir/shaarligo.cgi

done!

Or build from source at http://purl.mro.name/ShaarliGo

See example `static/.htaccess` or `static/app/lighttpd.conf` how to set up webserver integration.

## Todos

1. private posts,
2. [PuSH/PubSubhubbub](https://github.com/pubsubhubbub/pubsubhubbub) / [WebSub](https://www.w3.org/TR/websub/),
3. import shaarlis (login?),
4. pwd reset (maybe deleting from `app/config.yaml` is acceptable),
5. images/enclosures,
7. comments,
8. trackback/pingback

### Shaarli(OS|er) Compatibilty

see https://git.mro.name/mro/ShaarliOS/src/master/ios/ShaarliOS/API/ShaarliCmdUpdateEndpoint.m
and https://git.mro.name/mro/Shaarli-API-test/src/master/tests/test-post.sh
