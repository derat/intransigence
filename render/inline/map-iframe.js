let pageUrl = null;
let mapDiv = null;
let map = null;
let infoWindow = null;

function initializeMap() {
  // AMP effectively doesn't let us use allow-same-origin (see
  // https://github.com/ampproject/amphtml/blob/master/spec/amp-iframe-origin-policy.md),
  // which prevents us from just updating window.top.location.hash in
  // selectPoint(). Get the base page URL from document.referrer so we can use
  // it to construct a URL with the correct fragment and assign that directly to
  // window.top.location, which _is_ allowed.
  //
  // TODO: This doesn't work quite right. When a page is loaded from a Google
  // results page, it looks like we get a URL like
  // https://www-example-org.cdn.ampproject.org/v/s/www.example.org/page.amp.html
  // here, but the outer page seems to actually be
  // https://www.google.com/amp/s/www.example.org/page.amp.html. Per
  // https://developers.googleblog.com/2017/02/whats-in-amp-url.html, this
  // sounds like it's weirdness relating to the prerendering. The upshot is that
  // clicking on a location link triggers a navigation to the ampproject.org
  // URL. I'm not sure how to fix this, since I don't want to hardcode a
  // www.google.com/amp URL here.
  pageUrl = document.referrer.split('#', 1)[0];

  const mapOptions = {
    mapTypeId: google.maps.MapTypeId.ROADMAP,
    // Disable scrollwheel zooming; it's too easy to trigger while scrolling the
    // page up or down.
    scrollwheel: false,
    mapTypeControl: true,
    mapTypeControlOptions: {
      style: google.maps.MapTypeControlStyle.DROPDOWN_MENU,
      position: google.maps.ControlPosition.LEFT_TOP,
    },
  };
  mapDiv = document.getElementById('map-div');
  map = new google.maps.Map(mapDiv, mapOptions);
  infoWindow = new google.maps.InfoWindow();

  // Only show the map once it's fully loaded.
  google.maps.event.addListenerOnce(map, 'idle', () => {
    mapDiv.className = 'loaded';
  });

  const bounds = new google.maps.LatLngBounds();
  for (let i = 0; i < points.length; i++) {
    const p = points[i];
    p.latLong = new google.maps.LatLng(p.latLong[0], p.latLong[1]);
    bounds.extend(p.latLong);

    const letter = String.fromCharCode(65 + i);
    const markerOptions = {
      position: p.latLong,
      title: p.name,
      icon: `https://chart.googleapis.com/chart?chst=d_map_pin_letter&chld=${letter}|8cf|000`,
      map,
    };
    p.marker = new google.maps.Marker(markerOptions);
    google.maps.event.addListener(
      p.marker,
      'click',
      selectPoint.bind(null, p.id, false)
    );
  }

  map.fitBounds(bounds);
}

function selectPoint(id, center) {
  const point = points.find((p) => p.id == id);
  if (!point) {
    console.log('Unable to find point with ID ' + id);
    return;
  }

  const a = document.createElement('a');
  a.appendChild(document.createTextNode(point.name));
  a.className = 'location';
  a.addEventListener('click', () => (window.top.location = `${pageUrl}#${id}`));
  infoWindow.setContent(a);
  infoWindow.open(map, point.marker);

  if (center) {
    map.setCenter(point.latLong);
    mapDiv.scrollIntoView(true);
  }
}

window.addEventListener('load', () => initializeMap());
window.addEventListener('message', (e) => selectPoint(e.data.id, true));
