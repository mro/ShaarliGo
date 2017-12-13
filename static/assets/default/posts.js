
{
  var xml_base_pub = document.documentElement.getAttribute("data-xml-base-pub");
  // <![CDATA[
  // check if we're logged-in (AJAX or Cookie?).
  var xhr = new XMLHttpRequest();
  xhr.onreadystatechange = function(data0) {
    if (xhr.readyState == 4) {
      // console.log('xhr.status = ' + xhr.status);
      document.documentElement.classList.add(xhr.status === 200 ? 'logged-in' : 'logged-out');
      // store the result locally and use as initial value for later calls?
    }
  }
  xhr.open('GET', xml_base_pub + '/../shaarligo.cgi/session/');
  xhr.send(null);
}

// onload="document.getElementById('q').removeAttribute('autofocus');document.getElementById('post').setAttribute('autofocus', 'autofocus');"
// onload="document.form_post.post.focus();"

document.addEventListener('DOMContentLoaded', function(event) {
  console.log(event.type);
  // inspired by http://leaverou.github.io/awesomplete/#extensibility
  var addlink = new Awesomplete('input[data-multiple]', {
    minChars: 3,
    maxItems: 15,
    filter: function(text, input) { return Awesomplete.FILTER_CONTAINS(text, input.match(/\S*$/)[0]); /* match */ },
    item: function(text, input) { return Awesomplete.ITEM(text, input.match(/\S*$/)[0]); /* highlight */ },
    replace: function(text) {
      var before = this.input.value.match(/^.+\s+|/)[0]; // ends with a whitespace
      this.input.value = before + text + " ";
    }
  });
});
