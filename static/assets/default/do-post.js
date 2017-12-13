
document.addEventListener('DOMContentLoaded', function(event) {
  console.log(event.type);
  var tit = new Awesomplete('input[data-multiple]', {
    minChars: 3,
    maxItems: 15,
    filter: function(text, input) { return Awesomplete.FILTER_CONTAINS(text, input.match(/\S*$/)[0]); /* match */ },
    item: function(text, input) { return Awesomplete.ITEM(text, input.match(/\S*$/)[0]); /* highlight */ },
    replace: function(text) {
      var before = this.input.value.match(/^.+\s+|/)[0]; // ends with a whitespace
      this.input.value = before + text + " ";
    }
  });

  var txt = new Awesomplete('textarea[data-multiple]', {
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
