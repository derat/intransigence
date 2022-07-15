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
    styles: getStyles(),
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

  // Show the map after the tiles have fully loaded, but also watch for the
  // 'idle' event (which often fires earlier) as a fallback for slow
  // connections.
  google.maps.event.addListenerOnce(map, 'tilesloaded', () => {
    mapDiv.classList.add('loaded');
  });
  google.maps.event.addListenerOnce(map, 'idle', () => {
    window.setTimeout(() => mapDiv.classList.add('loaded'), 5000);
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
      icon: `https://chart.googleapis.com/chart?chst=d_map_pin_letter&chld=${letter}|fc783a|33180c`,
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
  updateStyle();
}

function selectPoint(id, center) {
  if (!map) {
    console.log('Map not initialized');
    return;
  }

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

// Returns the 'styles' value for google.maps.MapOptions.
function getStyles() {
  // Just use the default light style if the dark theme isn't being used.
  if (!document.body.classList.contains('dark')) return undefined;

  // Generated using https://mapstyle.withgoogle.com/
  return [
    {
      elementType: 'geometry',
      stylers: [{ color: '#242f3e' }],
    },
    {
      elementType: 'labels.text.fill',
      stylers: [{ color: '#746855' }],
    },
    {
      elementType: 'labels.text.stroke',
      stylers: [{ color: '#242f3e' }],
    },
    {
      featureType: 'administrative.locality',
      elementType: 'labels.text.fill',
      stylers: [{ color: '#d59563' }],
    },
    {
      featureType: 'poi',
      elementType: 'labels.text.fill',
      stylers: [{ color: '#d59563' }],
    },
    {
      featureType: 'poi.park',
      elementType: 'geometry',
      stylers: [{ color: '#263c3f' }],
    },
    {
      featureType: 'poi.park',
      elementType: 'labels.text.fill',
      stylers: [{ color: '#6b9a76' }],
    },
    {
      featureType: 'road',
      elementType: 'geometry',
      stylers: [{ color: '#38414e' }],
    },
    {
      featureType: 'road',
      elementType: 'geometry.stroke',
      stylers: [{ color: '#212a37' }],
    },
    {
      featureType: 'road',
      elementType: 'labels.text.fill',
      stylers: [{ color: '#9ca5b3' }],
    },
    {
      featureType: 'road.highway',
      elementType: 'geometry',
      stylers: [{ color: '#746855' }],
    },
    {
      featureType: 'road.highway',
      elementType: 'geometry.stroke',
      stylers: [{ color: '#1f2835' }],
    },
    {
      featureType: 'road.highway',
      elementType: 'labels.text.fill',
      stylers: [{ color: '#f3d19c' }],
    },
    {
      featureType: 'transit',
      elementType: 'geometry',
      stylers: [{ color: '#2f3948' }],
    },
    {
      featureType: 'transit.station',
      elementType: 'labels.text.fill',
      stylers: [{ color: '#d59563' }],
    },
    {
      featureType: 'water',
      elementType: 'geometry',
      stylers: [{ color: '#17263c' }],
    },
    {
      featureType: 'water',
      elementType: 'labels.text.fill',
      stylers: [{ color: '#515c6d' }],
    },
    {
      featureType: 'water',
      elementType: 'labels.text.stroke',
      stylers: [{ color: '#17263c' }],
    },
    // Deemphasize POI and road icons since they compete with our markers
    // otherwise. The styler ominously warns, "The effect of the following
    // stylers will change whenever Google updates the base map style.
    // Use with caution."
    {
      featureType: 'poi',
      elementType: 'labels.icon',
      stylers: [{ saturation: -50 }, { lightness: -30 }],
    },
    {
      featureType: 'road',
      elementType: 'labels.icon',
      stylers: [{ saturation: -50 }, { lightness: -30 }],
    },
  ];
}

function updateStyle() {
  // Handle dark/light mode using code defined in dark.js.
  applyTheme();
  map.setOptions({ styles: getStyles() });
}

window.addEventListener('DOMContentLoaded', () => {
  darkQuery.addEventListener('change', () => updateStyle());
  window.addEventListener('storage', () => updateStyle());
  initializeMap();
});

window.addEventListener('message', (e) => selectPoint(e.data.id, true));
