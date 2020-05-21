// Wire up links to post messages to the iframe to activate markers.
document.addEventListener('DOMContentLoaded', function() {
  var iframe = document.getElementById('map');
  var anchors = document.getElementsByClassName('map-link');
  for (var i = 0; i < anchors.length; i++) {
    var a = anchors[i];
    var id = a.parentElement.parentElement.id;
    var f = iframe.contentWindow.postMessage.bind(
        iframe.contentWindow, {id: id}, '*', []);
    a.addEventListener('click', f, false);
  }
}, false);
