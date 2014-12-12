(function(){

    $('<div id="hdict-bubble-container" class="hdict-bubble hide"><div class="word"></div><div class="trans"></div></div>').appendTo('body');

    var positionTooltip = function(event){
      var tPosX = event.pageX + 15;
      var tPosY = event.pageY + 15;
      $('#hdict-bubble-container').css({'position': 'absolute', 'top': tPosY, 'left': tPosX}).removeClass('hide');
    };

    var config = undefined;
    var loadConfig = function() {
      chrome.storage.sync.get({
        dblclickTrigger: true,
        selTrigger: false,
        triggerKey : "metaKey"
      }, function(items) {
        config = items;
        // if we're not in OSX
        if (window.navigator.platform.toLowerCase().lastIndexOf('mac') == -1 && config.triggerKey == "metaKey") {
          config.triggerKey = "ctrlKey";
        }
      });
    }

    loadConfig();
    chrome.storage.onChanged.addListener(function () {
      loadConfig();
    })

    var shouldPopup = function(event) {
      var bubble = $('#hdict-bubble-container');
      if (bubble.hasClass('hide') == false) {
        return false;
      }

      var target = event.target;
      if (target && target.tagName) {
        var tagName = target.tagName.toLowerCase();
        if (tagName == 'input' || tagName == 'textarea') {
          return false;
        }
      }

      if (document.designMode && document.designMode.toLowerCase() == 'on') {
        return false;
      }

      for (; target ; target = target.parentNode) {
        if (target.isContentEditable) {
          return false;
        }
      }

      if (config.triggerKey != "none" &&  event[config.triggerKey] == false) {
        return false;
      }

      return true;
    }

    var isMouseInPopup = function(e) {
      var target = e.target;
      for (; target; target = target.parentNode) {
        var id = $(target).attr('id');
        if (id == "hdict-bubble-container") {
          return true;
        }
      }
      return false;
    }

    var hideBubble = function() {
      $('#hdict-bubble-container .word').html('');
      $('#hdict-bubble-container .trans').html('')
      $('#hdict-bubble-container').addClass('hide');
    }

    var queryAndPopup = function(event, word) {
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
          if (response.basic.phonetic !== undefined) {
            title += ' <span class="phonetic">/' + response.basic.phonetic + '/</phonetic>';
          }
        }
        if (explain.length > 0) {
          $('#hdict-bubble-container .word').html(title);
          $('#hdict-bubble-container .trans').html(explain);
          positionTooltip(event);
        }
      });
    }

    $('body').dblclick(function(e) {
      if (!shouldPopup(e)) {
        return ;
      }
      if (!config.dblclickTrigger) {
        return ;
      }
      var selection = window.getSelection() || document.getSelection() || document.selection.createRange();
      var word = $.trim(selection.toString());
      if (word != '') {
        queryAndPopup(e, word);
      }
    });

    $(document).mouseup(function(e) {
      if (isMouseInPopup(e)) {
        return ;
      }
      var bubble = $('#hdict-bubble-container');
      if (bubble.hasClass('hide') == false) {
        hideBubble();
        return ;
      }
      if (!shouldPopup(e)) {
        return ;
      }
      if (!config.selTrigger) {
        return ;
      }
      var selection = window.getSelection() || document.getSelection() || document.selection.createRange();
      var word = $.trim(selection.toString());
      if(word != '') {
        queryAndPopup(e, word);
      }
    });

})();
