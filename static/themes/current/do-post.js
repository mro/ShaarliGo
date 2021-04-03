
const xml_base_pub = document.documentElement.getAttribute("data-xml-base-pub");

// Firefox 56+ doesn't fire that one in xslt situation: document.addEventListener("DOMContentLoaded", function(event) { console.log("DOM fully loaded and parsed"); });
let tit;
document.onreadystatechange = function () {
  if(tit !== undefined)
    return;
  // console.log('setup awesomeplete');
  tit = new Awesomplete('input[data-multiple]', {
    minChars: 3,
    maxItems: 15,
    filter:  function(text, input) { const m = input.match(/#(\S*)$/); return m !== null && Awesomplete.FILTER_CONTAINS(text, m[1]); /* match */ },
    item:    function(text, input) { const m = input.match(/#(\S*)$/); return m !== null && Awesomplete.ITEM(text, m[1]); /* highlight */ },
    replace: function(text) { const inp = this.input; inp.value = inp.value.replace(/#[^#]+$/, text) + " "; },
  });

  const txt = new Awesomplete('textarea[data-multiple]', {
    minChars: 3,
    maxItems: 15,
    filter:  function(text, input) { const m = input.match(/#(\S*)$/); return m !== null && Awesomplete.FILTER_CONTAINS(text, m[1]); /* match */ },
    item:    function(text, input) { const m = input.match(/#(\S*)$/); return m !== null && Awesomplete.ITEM(text, m[1]); /* highlight */ },
    replace: function(text) { const inp = this.input; inp.value = inp.value.replace(/#[^#]+$/, text) + " "; },
  });

  const xhr = new XMLHttpRequest()
  xhr.onreadystatechange = function() {
    if (xhr.readyState > 3 && xhr.status == 200) {
      txt.list = tit.list = JSON.parse(xhr.response);
    }
  };
  xhr.open('GET', xml_base_pub + '/t/index.json');
  xhr.setRequestHeader('X-Requested-With', 'XMLHttpRequest');
  xhr.send();
};

const secret = 'popup is ready';

function fun0(event) {
  window.removeEventListener(event.type, fun0);
  console.info(event.data);
  function l(k){console.info(event.data.get(k));}
  l('token');
  l('post');
  l('source');
  l('title');
  l('scrape');
  l('tags');
  l('description');
  l('image');
}

window.addEventListener('message', fun0, false);
window.opener.postMessage(secret, '*');
