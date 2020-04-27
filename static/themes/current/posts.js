
function toggle() {
	switch(document.cookie) {
	case 'dark':
			document.cookie = 'light';
			break;
	case 'light':
			document.cookie = '';
			break;
	default:
			document.cookie = 'dark';
			break;
	}
  console.log("uhu: "+document.cookie);
	const lst = document.documentElement.classList;
	lst.remove('dark');
	lst.remove('light');
	lst.add(document.cookie);
}

// list tags with font-size in relation to frequency
// https://github.com/sebsauvage/Shaarli/blob/master/index.php#L1254
function computeTagFontsize(tag0) {
  if (!tag0)
    return;
  const fontMin = 8;
  const fontMax = 32;
  var countMaxLog = Math.log(1);
  const tags = tag0.getElementsByTagName('a');
  const counts = new Array(tags.length);
  const log = {};
  for (var i = tags.length - 1; i >= 0; i--) {
    const lbl = tags[i].getElementsByClassName('label')[0].textContent;
    if ('2018-01-15T12:52' == lbl) {
      counts[i] = 1;
      continue;
    }
    const elm = tags[i].getElementsByClassName('count')[0];
    const txt = elm.textContent;
    var v = log[txt];
    if (!v)
      log[txt] = v = Math.log(parseInt(txt, 10));
    counts[i] = v;
    countMaxLog = Math.max(countMaxLog, counts[i]);
  }
  log.length = 0;
  const factor = 1.0 / countMaxLog * (fontMax - fontMin);
  requestAnimationFrame(function() { // http://wilsonpage.co.uk/preventing-layout-thrashing/
    for (var i = tags.length - 1; i >= 0; i--) {
      const k = counts[i];
      var v = log[k];
      if (!v)
        // https://stackoverflow.com/a/3717340
        log[k] = v = Math.ceil(k * factor + fontMin) + 'pt';
      tags[i].style.fontSize = v;
    }
  });
}

// https://varvy.com/pagespeed/defer-images.html
function loadDeferredImages(imgsDefer) {
  // console.log('loadDeferredImages: ' + imgsDefer.length);
  for (var i = imgsDefer.length - 1; i >= 0 ; i--) {
    const v = imgsDefer[i].getAttribute('data-src');
    if (!v)
      continue;
    imgsDefer[i].setAttribute('src', v);
  }
}

// make http and geo URIs (RFC 5870) clickable + microformat
function clickableTextLinks(elmsRendered) {
  // console.log('make http and geo URIs (RFC 5870) clickable + microformat');
  for (var i = elmsRendered.length - 1; i >= 0 ; i--) {
    const elm = elmsRendered[i];
    elm.innerHTML = elm.innerHTML.replace(/(https?:\/\/[^ \t\r\n"']+[^ ?\t\r\n"'.,;()])/gi, '<a rel="noreferrer" class="http" href="$1">$1</a>');
    // https://alanstorm.com/url_regex_explained/ \b(([\w-]+://?|www[.])[^\s()<>]+(?:\([\w\d]+\)|([^[:punct:]\s]|/)))
    // elm.innerHTML = elm.innerHTML.replace(/\b(([\w-]+:\/\/?|www[.])[^\s()<>]+(?:\([\w\d]+\)|([^[:punct:]\s]|\/)))/gi, '<a rel="noreferrer" class="http" href="$1">$1</a>');
    elm.innerHTML = elm.innerHTML.replace(/geo:(-?\d+.\d+),(-?\d+.\d+)(\?z=(\d+))?/gi, '<a class="geo" href="https://opentopomap.org/#marker=12/$1/$2" title="zoom=$4">geo:<span class="latitude">$1</span>,<span class="longitude">$2</span>$3</a>');
    elm.innerHTML = elm.innerHTML.replace(/(urn:ietf:rfc:(\d+)(#\S*[0-9a-z])?)/gi, '<a class="rfc" href="https://tools.ietf.org/html/rfc$2$3" title="RFC $2">$1</a>');
    elm.innerHTML = elm.innerHTML.replace(/(urn:isbn:([0-9-]+)(#\S*[0-9a-z])?)/gi, '<a class="isbn" href="https://de.wikipedia.org/wiki/Spezial:ISBN-Suche?isbn=$2" title="ISBN $2">$1</a>');
    elm.innerHTML = elm.innerHTML.replace(/(CVE-[0-9-]+-[0-9]+)/gi, '<a class="cve" href="https://cve.mitre.org/cgi-bin/cvename.cgi?name=$1">$1</a>');
  }
}

const xml_base_pub = document.documentElement.getAttribute("data-xml-base-pub");
{
  document.documentElement.classList.add('logged-out'); // do in js early on load
  // check if we're logged-in (AJAX or Cookie?).
  const xhr = new XMLHttpRequest();
  xhr.onreadystatechange = function(data0) {
    if (this.readyState === XMLHttpRequest.HEADERS_RECEIVED) {
      if (this.status === 200) {
        document.documentElement.classList.add('logged-in');
        document.documentElement.classList.remove('logged-out');
      }
      // store the result locally and use as initial value for later calls to avoid a logged-in flicker?
    }
  }
  xhr.timeout = 1000;
  xhr.open('GET', xml_base_pub + '/../shaarligo.cgi/session/');
  xhr.setRequestHeader('X-Requested-With', 'XMLHttpRequest');
  xhr.send();
}

// onload="document.getElementById('q').removeAttribute('autofocus');document.getElementById('post').setAttribute('autofocus', 'autofocus');"
// onload="document.form_post.post.focus();"

// Firefox 56+ doesn't fire that one in xslt situation: document.addEventListener("DOMContentLoaded", function(event) { console.log("DOM fully loaded and parsed"); });
var addlink;
document.onreadystatechange = function () {
  if(addlink !== undefined)
    return;
  // console.log('setup awesomeplete');
  // inspired by http://leaverou.github.io/awesomplete/#extensibility
  addlink = new Awesomplete('input[data-multiple]', {
    minChars: 3,
    maxItems: 15,
    filter:  function(text, input) { const m = input.match(/#(\S*)$/); return m !== null && Awesomplete.FILTER_CONTAINS(text, m[1]); /* match */ },
    item:    function(text, input) { const m = input.match(/#(\S*)$/); return m !== null && Awesomplete.ITEM(text, m[1]); /* highlight */ },
    replace: function(text) { const inp = this.input; inp.value = inp.value.replace(/#[^#]+$/, text) + " "; },
  });

  const xhr = new XMLHttpRequest()
  xhr.onreadystatechange = function() {
    if (this.readyState === XMLHttpRequest.DONE && this.status == 200)
      addlink.list = JSON.parse(this.response);
  };
  xhr.timeout = 1000;
  xhr.open('GET', xml_base_pub + '/t/index.json');
  xhr.setRequestHeader('X-Requested-With', 'XMLHttpRequest');
  xhr.send();

  document.getElementById('q').focus();
  computeTagFontsize(document.getElementById('tags'));
  loadDeferredImages(document.getElementsByTagName('img'));
  clickableTextLinks(document.getElementById('entries').getElementsByClassName('rendered'));

  // https://koddsson.com/posts/emoji-favicon/
  const favicon = document.querySelector("html > head > link[rel=icon]");
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
};
