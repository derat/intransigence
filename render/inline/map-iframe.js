var pageUrl = null;
var mapDiv = null;
var map = null;
var infoWindow = null;

function initializeMap() {
  // AMP effectively doesn't let us use allow-same-origin (see
  // https://github.com/ampproject/amphtml/blob/master/spec/amp-iframe-origin-policy.md),
  // which prevents us from just updating window.top.location.hash in
  // selectPoint(). Get the base page URL from document.referrer so we can use
  // it to construct a URL with the correct fragment and assign that directly to
  // window.top.location, which _is_ allowed.
  pageUrl = document.referrer.split('#', 1)[0];

  var mapOptions = {
    mapTypeId: google.maps.MapTypeId.ROADMAP,
    // Disable scrollwheel zooming; it's too easy to trigger while scrolling the
    // page up or down.
    scrollwheel: false
  };
  mapDiv = document.getElementById('map-div');
  map = new google.maps.Map(mapDiv, mapOptions);
  infoWindow = new google.maps.InfoWindow();

  // Only show the map once it's fully loaded.
  google.maps.event.addListenerOnce(map, 'idle', function() {
    mapDiv.className = 'loaded';
  });

  var bounds = new google.maps.LatLngBounds();
  for (var i = 0; i < points.length; i++) {
    var p = points[i];
    p.latLong = new google.maps.LatLng(p.latLong[0], p.latLong[1]);
    bounds.extend(p.latLong);

    var letter = String.fromCharCode(65 + i);
    var markerOptions = {
      position: p.latLong,
      title: p.name,
      icon: 'https://chart.googleapis.com/chart?chst=d_map_pin_letter&chld=' + letter + '|8cf|000',
      map: map
    };
    p.marker = new google.maps.Marker(markerOptions);
    google.maps.event.addListener(p.marker, 'click', selectPoint.bind(null, p.id, false));
  }

  map.fitBounds(bounds);
}

function selectPoint(id, center) {
  var point = null;
  for (var i = 0; i < points.length; ++i) {
    if (points[i].id == id) {
      point = points[i];
      break;
    }
  }
  if (!point) {
    console.log('Unable to find point with ID ' + id);
    return;
  }

  var a = document.createElement('a');
  var f = function() { window.top.location = pageUrl + '#' + id };
  a.appendChild(document.createTextNode(point.name));
  a.className = 'location';
  a.addEventListener('click', f, false);
  infoWindow.setContent(a);
  infoWindow.open(map, point.marker);

  if (center) {
    map.setCenter(point.latLong);
    mapDiv.scrollIntoView(true);
  }
}

google.maps.event.addDomListener(window, 'load', initializeMap);
window.addEventListener('message', function(e) { selectPoint(e.data.id, true) }, false);
