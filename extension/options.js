(function(){

function save_options() {
  var dblclickTrigger = document.getElementById('dblclick-trigger').checked;
  var selTrigger = document.getElementById('select-trigger').checked;

  var obj = document.getElementById('trigger-key');
  var triggerKey = obj.options[obj.selectedIndex].value

  chrome.storage.sync.set({
    dblclickTrigger: dblclickTrigger,
    selTrigger: selTrigger,
    triggerKey: triggerKey
  }, function() {
  });
}

function restore_options() {
  chrome.storage.sync.get({
    dblclickTrigger: true,
    selTrigger: false,
    triggerKey : "metaKey"
  }, function(items) {
    document.getElementById('dblclick-trigger').checked = items.dblclickTrigger;
    document.getElementById('select-trigger').checked = items.selTrigger;
    document.getElementById('trigger-key').value = items.triggerKey;
  });
}

restore_options();

$(document).ready(function(){

  $('#dblclick-trigger').click(function() {
    save_options();
  });

  $('#select-trigger').click(function() {
    save_options();
  });

  $('#trigger-key').change(function() {
    save_options();
  })

})

})();
