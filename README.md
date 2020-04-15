
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

1. `$ curl -LRo shaarligo.cgi.gz http://purl.mro.name/shaarligo-Linux-x86_64.cgi.gz  # uname -s; uname -m`
2. `$ gunzip shaarligo.cgi.gz`
3. `$ chmod a+x,a-w shaarligo.cgi`
4. visit in your browser: http://my.web.space/subdir/shaarligo.cgi

done (Apache)! For lighttpd see `static/app/lighttpd.conf`.

Or build from source at http://mro.name/ShaarliGo

## Todos

1. private posts,
2. [PuSH/PubSubhubbub](https://github.com/pubsubhubbub/pubsubhubbub) / [WebSub](https://www.w3.org/TR/websub/),
3. import shaarlis (login?),
4. pwd reset (maybe deleting from `app/config.yaml` is acceptable),
5. images/enclosures,
7. comments,
8. trackback/pingback

### Shaarli(OS|er) Compatibilty

see https://code.mro.name/mro/ShaarliOS/src/master/ios/ShaarliOS/API/ShaarliCmdUpdateEndpoint.m
and https://code.mro.name/mro/Shaarli-API-test/src/master/tests/test-post.sh
