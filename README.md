
🌩 [Lightning Talk at the 34c3](https://events.ccc.de/congress/2017/wiki/Lightning:ShaarliGo:_self-hosted_microblogging) 🚀

[![Build Status](https://travis-ci.org/mro/ShaarliGo.svg?branch=master)](https://travis-ci.org/mro/ShaarliGo)

# ShaarliGo

🌺 Self-reliant publishing for laypeople like your loved ones and mine. Have a say
and not be subjected to any T&Cs, just local law. All without setup headaches, but
truly self-sustained and enduringly independent:

## Install / Update

1. Rent any web space from EUR 2 monthly with a domain-name as your enduring
   digital property (e.g. https://variomedia.de/hosting),
2. download https://mro.name/Linux-x86_64/shaarligo.cgi and
3. copy it to the webspace, see e.g.
   https://www.variomedia.de/faq/Wie-uebertrage-ich-meine-Seite-auf-den-Server/article/177,
4. set the file permissions (chmod) to read-only+execute for all (numeric 555), see
   e.g. https://wiki.filezilla-project.org/Other_Features#Chmod,
5. visit http://example.com/shaarligo.cgi and off you go!

That's if the webserver is Apache (Linux, 64 bit, modules cgi and rewrite) as
common with shared hosting.

For lighttpd see `static/app/lighttpd.conf`. Nginx lacks CGI support (srsly?).

Or build from source at http://mro.name/ShaarliGo

## Responsible Disclosure

In case you are reluctant to file a [public
issue](https://mro.name/ShaarliGo/issues), feel free to email
[security@mro.name](mailto:security@mro.name?subject=ShaarliGo)
([🔏key](https://mro.name/.well-known/openpgpkey/hu/t5s8ztdbon8yzntexy6oz5y48etqsnbb?security)).

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
- pinboard: https://api.pinboard.in/v1?auth_token=johndoe:XOG6EJIYMIZZ
  prefix:
```

It's ok to leave `prefix` empty. Each pinboard post gets a backlink as an
additional footer line. If `prefix` is set, the footer line is `prefix` + `id`.

### Mastodon

at first manually obtain an access token (example server here is
https://social.tchncs.de/):

1. create an application in https://social.tchncs.de/settings/applications
2. give it permission `write:statuses`
3. note the access token and enter it below.

Then enter the server endpoint plus `/api/v1/` and access token into
`app/config.yaml` like so:

```yaml
posse:
- mastodon: https://social.tchncs.de/api/v1/
  token: …boph1koomie4eikaiG…
  prefix:
```

It's ok to leave `prefix` empty. Each mastodon post gets a backlink as an
additional footer line. If `prefix` is set, the footer line is `prefix` + `id`.

Also, if you don't know the token but the endpoint accepts basic auth
([pleroma](https://pleroma.social/)), you may [add the credentials to the
url](https://developer.mozilla.org/en-US/docs/Web/HTTP/Authentication#Access_using_credentials_in_the_URL)
for the time being until I figure out how to get a token from pleroma or do proper
OAuth2.

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

## Todos

1. pinned posts,
1. private posts,
2. [PuSH/PubSubhubbub](https://github.com/pubsubhubbub/pubsubhubbub) / [WebSub](https://www.w3.org/TR/websub/),
3. import shaarlis (login?),
4. pwd reset (maybe deleting from `app/config.yaml` is acceptable),
5. images/enclosures,
7. comments,
8. trackback/pingback

## Credits

inspired by and compatible to http://sebsauvage.net/wiki/doku.php?id=php:shaarli.

