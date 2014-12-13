(function(){
  var queryUrl = "https://fanyi.youdao.com/openapi.do?keyfrom=c-dict&key=1416039712&type=data&doctype=json&version=1.1&q=";

  var uploadUrl = "https://huangdx.net/c4pt0r"
  var token = "d9c207cb2ae7b88aeba47f8c824f1b23"

  var postQueryRecord = function(rec) {
    $.ajax
    ({
      type: "POST",
      url: "https://huangdx.net/c4pt0r",
      dataType: 'json',
      async: false,
      headers: {
        "Authorization": "Token " + token
      },
      data: JSON.stringify(rec),
    });
  }

  chrome.runtime.onMessage.addListener(
    function(request, sender, sendResponse) {
      if (request.command == "query") {
        $.get(queryUrl + encodeURIComponent(request.word), function(e) {
          sendResponse(e);
          if (e.basic !== undefined) {
            var trans = "";
            $.each(e.basic.explains, function(i, ele){
              trans += ele + "\n"
            })
            postQueryRecord({word:e.query, explain: trans, phonetic: e.basic.phonetic});
          }
        });
      }
      return true;
    }
  );
})();
