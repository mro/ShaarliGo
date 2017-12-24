
🌩 [Lightning Talk at the 34c3](https://events.ccc.de/congress/2017/wiki/Lightning:ShaarliGo:_self-hosted_microblogging) 🚀

# ShaarliGo

self-hosted microblogging inspired by
http://sebsauvage.net/wiki/doku.php?id=php:shaarli. Destilled down to the bare
minimum, with easy hosting and security in mind. No PHP, no DB, no server-side
templating.

## Design Goals

* standards compliant ([Atom](https://tools.ietf.org/html/rfc4287), 
  [Atompub](https://tools.ietf.org/html/rfc5023), 
  [WebSub](https://www.w3.org/TR/websub/)),
* keep server lean, especially for readers,
* easy migration from existing shaarlis,
* backwards compatible posting (https://github.com/mro/Shaarli-API-test/)
* trivial installation and minimal hosting requirements (run on simple hosted webspace),
* run ok without javascript,
* visitor reading operates on static flat files only (no server code),
* secure against brute force login attacks,
* easy translation & skinning,
* leverage existing, widely deployed web tec ([XSLT](https://www.w3.org/TR/xslt-10/), HTML, CSS),
* easy fail2ban integration / DOS mitigation,

| Quality         | very good | good | normal | irrelevant |
|-----------------|:---------:|:----:|:------:|:----------:|
| Functionality   |           |      |    ×   |            |
| Reliability     |     ×     |      |        |            |
| Usability       |           |  ×   |        |            |
| Efficiency      |     ×     |      |        |            |
| Changeability   |           |  ×   |        |            |
| Portability     |           |      |        |     ×      |

## Install / Update

Linux amd64:

1. `$ curl -L http://purl.mro.name/shaarligo_cgi.gz | tee shaarligo_cgi.gz | gunzip > shaarligo.cgi && chmod a+x shaarligo.cgi`
2. visit in your browser: http://my.web.space/subdir/shaarligo.cgi

done!

Or build from source at http://purl.mro.name/ShaarliGo

## Todos

6. search,
1. private posts,
2. [PuSH/PubSubhubbub](https://github.com/pubsubhubbub/pubsubhubbub) / [WebSub](https://www.w3.org/TR/websub/),
3. import shaarlis (login?),
4. pwd reset,
5. images/enclosures,
7. comments,
8. trackback/pingback

### Shaarli(OS|er) Compatibilty

see https://github.com/mro/ShaarliOS/blob/master/ios/ShaarliOS/API/ShaarliCmdUpdateEndpoint.m
and https://github.com/mro/Shaarli-API-test/blob/master/tests/test-post.sh
