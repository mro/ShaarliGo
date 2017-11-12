
# ShaarliGo

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
GET  shaarligo.cgi
GET  shaarligo.cgi/config
POST shaarligo.cgi/config 							token? session?
GET  pub/posts/Kk-eZA
GET  pub/tags/Design
GET  shaarligo.cgi?do=login
POST shaarligo.cgi?do=login
GET  shaarligo.cgi/logout
GET  shaarligo.cgi/login
POST shaarligo.cgi/login
GET  shaarligo.cgi/logout
GET  shaarligo.cgi?post=url&title=Foo&source=ShaarliGo
POST shaarligo.cgi?do=login&login=uid&password=pwd&token=xyz
POST shaarligo.cgi?post=url&title=Foo&source=ShaarliGo
```

### 0. Routes

GET

shaarligo.cgi
shaarligo.cgi/config
shaarligo.cgi/session
shaarligo.cgi?q=%23Design+%23URI+Foo+Bar
./pub/posts
./~me/posts                    	Merged. Check Basic/Digest Auth!!
./@me/posts/DK0BTg							allow other ids e.g. guid (base64) or sha1 (base64)
./pub/tags
./pub/tags/Design
./~me/2017-07-13
./~me/enclosures/foo.svg
./assets/default/style.css
./assets/default/de/config.xslt
./assets/default/de/posts.xslt

announced via link/@rel/@uri https://martinfowler.com/articles/richardsonMaturityModel.html#level3:

POST

shaarligo.cgi/config
shaarligo.cgi/session
shaarligo.cgi/session								(HTML form fallback)
shaarligo.cgi/~me/posts
shaarligo.cgi/@me/posts
shaarligo.cgi/posts
shaarligo.cgi/enclosures
shaarligo.cgi/posts/DK0BTg						(HTML form fallback)

PUT

shaarligo.cgi/config
shaarligo.cgi/@me/tags/Design
shaarligo.cgi/posts/DK0BTg
shaarligo.cgi/enclosures/foo.svg

DELETE

shaarligo.cgi/~me/tags/Design
shaarligo.cgi/posts/DK0BTg
shaarligo.cgi/enclosures/foo.svg

### Shaarli(OS|er) Compatibilty

see https://github.com/mro/ShaarliOS/blob/master/ios/ShaarliOS/API/ShaarliCmdUpdateEndpoint.m
and https://github.com/mro/Shaarli-API-test/blob/master/tests/test-post.sh

Login/Logout

GET    shaarligo.cgi?do=login
POST   shaarligo.cgi?do=login
GET  	 shaarligo.cgi?do=logout

Posting

GET  shaarligo.cgi?post=url&title=Foo&source=ShaarliGo -> ?do=login
POST shaarligo.cgi?do=login&login=uid&password=pwd&token=xyz -> .
POST shaarligo.cgi?post=url&title=Foo&source=ShaarliGo -> ../../@me/posts?#Kk-eZA

### 0.1 Storage

Settings

./app/config.yaml
./app/config.yaml~

Posts

./app/var/posts.gob.gz
./app/var/posts.gob.gz~
./app/var/posts.atom.gz
./app/var/posts.atom.gz~
./~me/enclosures/
./pub/enclosures/

Ban

./app/var/bans.gob
./app/var/bans.gob~

(penalty > threshold) => HTTP 403 oder 429

### 1. Setup

1. drop Go binary on server,
2. (Todo: setup Webserver rewrites),
3. point browser to base url,
4. unpack assets if not there,
6. lock down dirctory permissions,
4. redirect to ./config and prepare first post:
5. post `title`, `uid` and `pwd`,
6. store stuff
7. redirect to .
