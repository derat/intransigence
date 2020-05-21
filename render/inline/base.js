function toggleNavbox(e) {
  var ul = document.getElementById('nav-list');
  var img = document.getElementById('nav-toggle-img');

  var collapsed = ul.className == 'collapsed-mobile';
  ul.className = collapsed ? '' : 'collapsed-mobile';
  img.className = collapsed ? '' : 'expand';
}

document.addEventListener('DOMContentLoaded', function() {
  document.getElementById('nav-logo').addEventListener('click', toggleNavbox, false);
  document.getElementById('nav-title').addEventListener('click', toggleNavbox, false);
}, false);
