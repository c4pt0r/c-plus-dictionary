(function(){
  var queryUrl = "https://fanyi.youdao.com/openapi.do?keyfrom=c-dict&key=1416039712&type=data&doctype=json&version=1.1&q=";

  var uploadUrl = "http://huangdx.net/c4pt0r"

  chrome.runtime.onMessage.addListener(
    function(request, sender, sendResponse) {
      if (request.command == "query") {
        $.get(queryUrl + encodeURIComponent(request.word), function(e) {
          sendResponse(e);
        });
      }
      return true;
    }
  );
})();
