
# AtomicShaarli

shaarli on diet. Built on Atom.

## Design Goals

* standards compliant API (Atompub/PuSH),
* seamless migration from existing shaarlis,
* backwards compatible posting (https://github.com/mro/Shaarli-API-test/)
* trivial installation and minimal hosting requirements (run on simple hosted webspace),
* run ok without javascript,
* reading operates on static flat files only (no server code),
* secure against brute force login attacks,
* easy translation & skinning,
* leverage existing, widely deployed web tec,
* easy fail2ban integration / DOS mitigation,

| Quality         | very good | good | normal | irrelevant |
|-----------------|:---------:|:----:|:------:|:----------:|
| Functionality   |           |      |    ×   |            |
| Reliability     |     ×     |      |        |            |
| Usability       |           |  ×   |        |            |
| Efficiency      |     ×     |      |        |            |
| Changeability   |           |  ×   |        |            |
| Portability     |           |      |        |     ×      |

## Todos

1. Setup,
2. login + -out,
3. create posts,
4. show posts,
5. store outside public space & lock down permisions,
6. paged feeds,
7. update/delete posts,
8. tags + bookmarklet client JS,
9. private posts,
10. PuSH
11. import shaarlis (login+atom),
12. pwd recovery,
13. images/enclosures,
14. further security lockdown ([HPKP](https://de.wikipedia.org/wiki/HTTP_Public_Key_Pinning)?), throttle search
15. comments,
16. trackback/pingback

```
GET  atom.cgi
GET  atom.cgi/settings
POST atom.cgi/settings 							token? session?
GET  pub/posts?#Kk-eZA
GET  pub/posts/Kk-eZA
GET  pub/tags/Design
GET  atom.cgi?do=login
POST atom.cgi?do=login
GET  atom.cgi/logout
GET  atom.cgi/login
POST atom.cgi/login
GET  atom.cgi/logout
GET  atom.cgi?post=url&title=Foo&source=AtomicShaarli
POST atom.cgi?do=login&login=uid&password=pwd&token=xyz
POST atom.cgi?post=url&title=Foo&source=AtomicShaarli
```

### 0. Routes

GET

atom.cgi
atom.cgi/config
atom.cgi/session
atom.cgi?q=%23Design+%23URI+Foo+Bar
./pub/posts
./~me/posts                    	Merged. Check Basic/Digest Auth!!
./@me/posts/DK0BTg							allow other ids e.g. guid (base64) or sha1 (base64)
./pub/tags
./pub/tags/Design
./~me/2017-07-13
./~me/enclosures/foo.svg
./assets/default/style.css
./assets/default/de/settings.xslt
./assets/default/de/posts.xslt

announced via link/@rel/@uri https://martinfowler.com/articles/richardsonMaturityModel.html#level3:

POST

atom.cgi/config
atom.cgi/session
atom.cgi/session?method=DELETE (HTML form fallback)
atom.cgi/~me/posts
atom.cgi/@me/posts
atom.cgi/pub/posts
atom.cgi/pub/enclosures
atom.cgi/pub/posts/DK0BTg?method=PUT (HTML form fallback)

PUT

atom.cgi/config
atom.cgi/@me/tags/Design
atom.cgi/pub/posts/DK0BTg
atom.cgi/pub/enclosures/foo.svg

DELETE

atom.cgi/~me/tags/Design
atom.cgi/pub/posts/DK0BTg
atom.cgi/pub/enclosures/foo.svg

### Shaarli(OS|er) Compatibilty

see https://github.com/mro/ShaarliOS/blob/master/ios/ShaarliOS/API/ShaarliCmdUpdateEndpoint.m
and https://github.com/mro/Shaarli-API-test/blob/master/tests/test-post.sh

Login/Logout

GET  atom.cgi?do=login
POST atom.cgi?do=login
GET  atom.cgi?do=logout

Posting

GET  atom.cgi?post=url&title=Foo&source=AtomicShaarli -> ?do=login
POST atom.cgi?do=login&login=uid&password=pwd&token=xyz -> .
POST atom.cgi?post=url&title=Foo&source=AtomicShaarli -> ../../@me/posts?#Kk-eZA

### 0.1 Storage

Settings

./app/config.yaml
./app/config.yaml~

Posts

./app/posts.gob
./app/posts.gob~
./app/posts.atom
./~me/enclosures/
./pub/enclosures/

Ban

./app/bans.gob
./app/bans.gob~

(penalty > threshold) => HTTP 403 oder 429

### 1. Setup

1. drop Go binary on server,
2. (Todo: setup Webserver rewrites),
3. point browser to base url,
4. unpack assets if not there,
6. lock down dirctory permissions,
4. redirect to ./settings and prepare first post:
5. post `title`, `uid` and `pwd`,
6. store stuff
7. redirect to .
