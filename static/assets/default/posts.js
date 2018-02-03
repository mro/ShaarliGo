
const xml_base_pub = document.documentElement.getAttribute("data-xml-base-pub");
{
  // check if we're logged-in (AJAX or Cookie?).
  const xhr = new XMLHttpRequest();
  xhr.onreadystatechange = function(data0) {
    if (xhr.readyState > 3) {
      // console.log('xhr.status = ' + xhr.status);
      document.documentElement.classList.add(xhr.status === 200 ? 'logged-in' : 'logged-out');
      // store the result locally and use as initial value for later calls?
    }
  }
  xhr.open('GET', xml_base_pub + '/../shaarligo.cgi/session/');
  xhr.setRequestHeader('X-Requested-With', 'XMLHttpRequest');
  xhr.send(null);
}

// onload="document.getElementById('q').removeAttribute('autofocus');document.getElementById('post').setAttribute('autofocus', 'autofocus');"
// onload="document.form_post.post.focus();"

// Firefox 56+ doesn't fire that one in xslt situation: document.addEventListener("DOMContentLoaded", function(event) { console.log("DOM fully loaded and parsed"); });
let addlink;
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
    if (xhr.readyState > 3 && xhr.status == 200)
      addlink.list = JSON.parse(xhr.response);
  };
  xhr.open('GET', xml_base_pub + '/tags/index.json');
  xhr.setRequestHeader('X-Requested-With', 'XMLHttpRequest');
  xhr.send();

  // list tags with font-size in relation to frequency
  // https://github.com/sebsauvage/Shaarli/blob/master/index.php#L1254
  let fontMin = 8;
  let fontMax = 48;
  const tags = document.getElementById('tags').getElementsByClassName('tag');
  const counts = new Array(tags.length);
  for (let i = tags.length - 1; i >= 0; i--) {
    const elm = tags[i].getElementsByClassName('count')[0];
    counts[i] = 1 * elm.innerText;
  }
  const countMaxLog = Math.log(Math.max.apply(Math, counts)); // https://johnresig.com/blog/fast-javascript-maxmin/
  const factor = 1.0 / countMaxLog * (fontMax - fontMin);
  for (let i = tags.length - 1; i >= 0; i--) {
    // https://stackoverflow.com/a/3717340
    const size = Math.log(counts[i]) * factor + fontMin;
    tags[i].style.fontSize = size + 'pt';
  }
};
