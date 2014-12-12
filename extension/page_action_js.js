(function(){

  var queryAndShow = function(word) {
    chrome.runtime.sendMessage({
      command : 'query',
      word: word
    }, function(response) {
      var explain = '';
      var title = response.query;
      if (response.basic !== undefined && response.basic.explains !== undefined) {
        $.each(response.basic.explains, function(i, ele) {
          explain += '<div class="basic">' + ele + '</div>';
        })
      }
      if (explain.length > 0) {
        $('#trans').html(explain);
      } else {
        $('#trans').html("'" + word + "'" + " no such word");
      }
    });
  }

  $(document).ready(function(){
    var lookup = function() {
      var word = $.trim($('#query-field').val());
      if (word != '') {
        queryAndShow(word);
      } else {
        $('#trans').html('');
      }
    }
    $('#trans').html('');
    $('#query-field').focus();
    $('#button').click(lookup)
    $('#query-field').keyup(function(e){
      if (e.keyCode == 13) {
        lookup();
      }
    });
  });
})();
