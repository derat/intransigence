// Code generated by gen_filemap.go from b323886b09ee2727e75d722dd40ac08a7583f023ca906fd1983bbc87b7a66128. DO NOT EDIT.

package render

var stdInline = map[string]string{
	"amp-boilerplate-noscript.css": "body{-webkit-animation:none;-moz-animation:none;-ms-animation:none;animation:none}",
	"amp-boilerplate.css":          "body{-webkit-animation:-amp-start 8s steps(1,end) 0s 1 normal both;-moz-animation:-amp-start 8s steps(1,end) 0s 1 normal both;-ms-animation:-amp-start 8s steps(1,end) 0s 1 normal both;animation:-amp-start 8s steps(1,end) 0s 1 normal both}@-webkit-keyframes -amp-start{from{visibility:hidden}to{visibility:visible}}@-moz-keyframes -amp-start{from{visibility:hidden}to{visibility:visible}}@-ms-keyframes -amp-start{from{visibility:hidden}to{visibility:visible}}@-o-keyframes -amp-start{from{visibility:hidden}to{visibility:visible}}@keyframes -amp-start{from{visibility:hidden}to{visibility:visible}}",
	"amp.css":                      "amp-img.thumb{filter:blur(12px)}main .box>.body .mapbox amp-img[placeholder]{max-width:100%}\n",
	"base-body.js":                 "applyTheme(); // defined in dark.js\n",
	"base.css":                     "body{color-scheme:light}body.dark{color-scheme:dark}iframe{color-scheme:normal}header .dark{cursor:pointer}main .box{display:block}main .box>.body:after{clear:both;content:'';display:block}main .box>.body>*:first-child,main .box>.body>*:first-child>h2:first-child,main .box>.body>*:first-child>h3:first-child{margin-top:0}main .box>.body>*:last-child{margin-bottom:0}main .box>.body figure.left{float:left}main .box>.body figure.right{float:right}main .box>.body figure.center{margin-left:auto;margin-right:auto}main .box>.body figure *{max-width:100%}main .box>.body figure img{border:0;display:block;height:auto}main .box>.body img.inline,main .box>.body amp-img.inline{vertical-align:middle}main .box>.body img.pixelated,main .box>.body amp-img.pixelated{image-rendering:pixelated}main .box>.body img.inline{display:inline}main .box>.body pre{max-width:100%;white-space:pre-wrap;word-wrap:break-word}main .box>.body table{border-collapse:collapse}main .box>.body .clear{clear:both}main .box>.body .small{font-size:90%}main .box>.body .real-small{font-size:80%}main .box>.body .no-select{user-select:none}\n",
	"base.js":                      "document.addEventListener('DOMContentLoaded', () => {\n  const nav = document.querySelector('.sitenav');\n  const navBody = nav.querySelector('.box > .body');\n  const navList = navBody.querySelector('ul');\n  const navPadding = 32; // >= navBody's non-collapsed padding\n\n  // Toggle the navbox when the logo or anything in its title are clicked.\n  const toggleNav = () => {\n    // Animating height is a mess: https://stackoverflow.com/questions/3508605\n    // When collapsing, set max-height to the actual height first so the\n    // animation begins immediately. When expanding, set it to list's height\n    // (plus extra for padding) so the animation takes roughly the right time.\n    if (!nav.classList.contains('collapsed-mobile')) {\n      navBody.style.maxHeight = navBody.clientHeight + 'px';\n      window.setTimeout(() => (navBody.style.maxHeight = ''));\n    } else {\n      navBody.style.maxHeight = navList.clientHeight + navPadding + 'px';\n    }\n    nav.classList.toggle('collapsed-mobile');\n  };\n  document.querySelector('header .logo').addEventListener('click', toggleNav);\n  document\n    .querySelector('.sitenav .box .title')\n    .addEventListener('click', toggleNav);\n\n  // At the end of a transition, tell the body to use its natural height in case\n  // the window is later resized.\n  navBody.addEventListener('transitionend', () => {\n    navBody.style.maxHeight = '';\n  });\n\n  // |darkQuery| and applyTheme() are defined in dark.js.\n  // Toggle the theme when the dark-mode icon is clicked.\n  // The initial state is set in base-body.js: we can't do this in the top level\n  // of this file since document.body isn't available, and we also don't want to\n  // do it in DOMContentLoaded since we'll get a flash of the light theme then.\n  document\n    .querySelector('header .dark')\n    .addEventListener('click', () => applyTheme(true));\n\n  // We may also need to update the theme if prefers-color-scheme changes.\n  darkQuery.addEventListener('change', () => applyTheme());\n});\n",
	"dark.js":                      "const darkQuery = window.matchMedia('(prefers-color-scheme: dark)');\n\n// Adds or remove the 'dark' class from document.body per localStorage and\n// prefers-color-scheme. If |toggle| is truthy, toggles the current value and\n// saves the updated value to localStorage.\nfunction applyTheme(toggle) {\n  // AMP iframes can't use allow-same-origin since they might be served from the\n  // cache. Check document.domain to determine if we're sandboxed, which\n  // prevents us from accessing localStorage: https://stackoverflow.com/a/34073811\n  //\n  // Just give up and use the light theme in this case, since we won't be able\n  // to tell if the user toggles the theme, and using the dark theme in an\n  // iframe while the rest of the page is using the light theme looks weird.\n  if (!document.domain) return;\n\n  const hasStorage = typeof Storage !== 'undefined';\n  let dark = false;\n  if (toggle) {\n    dark = !document.body.classList.contains('dark');\n    if (hasStorage) localStorage.setItem('theme', dark ? 'dark' : 'light');\n  } else {\n    const saved = hasStorage ? localStorage.getItem('theme') : null;\n    dark = saved !== null ? saved === 'dark' : darkQuery.matches;\n  }\n  dark\n    ? document.body.classList.add('dark')\n    : document.body.classList.remove('dark');\n}\n",
	"desktop.css":                  ".mobile-only{display:none}.sitenav .toggle{display:none}main .box>.body>figure.desktop-left{float:left}main .box>.body>figure.desktop-right{float:right}main .box>.body>figure.desktop-left:first-child+p,main .box>.body>figure.desktop-right:first-child+p{margin-top:0}\n",
	"graph-iframe.css":             "body{color-scheme:light;margin:0;overflow:hidden}body.dark{color-scheme:dark}svg.graph{background-color:white;display:inline-block;height:100%;position:absolute;width:100%}circle.line{fill:white;stroke:steelblue;stroke-width:1.5px}circle.line:hover{fill:steelblue}path.line{fill:none;stroke:steelblue;stroke-width:1.5px}rect.note{fill:#f5f5f5;shape-rendering:crispEdges;stroke:#eee;stroke-width:1px}rect.note:hover{fill:#eee;stroke:#ddd}text.title{font-family:Verdana, Helvetica, Arial, sans-serif;font-size:12px}.label rect{fill:#fffbe0;shape-rendering:crispEdges;stroke:#d2cfb9;stroke-width:1px;z-index:1}.label text{font-family:Helvetica, Arial, sans-serif;font-size:11px;z-index:2}.rule line{pointer-events:none;shape-rendering:crispEdges;stroke:#eee}.rule text{font-family:Helvetica, Arial, sans-serif;font-size:10px}body.dark svg.graph{background-color:#333}body.dark circle.line{fill:#333}body.dark circle.line:hover{fill:steelblue}body.dark rect.note{fill:#383838;stroke:#444}body.dark rect.note:hover{fill:#444;stroke:#555}body.dark text{fill:#ccc}body.dark .label rect{fill:#444;stroke:#555}body.dark .rule line{stroke:#444}\n",
	"graph-iframe.js":              "var d = null;\n\nfunction appendGraph(selector, size, title, timeseries, noteData, units, valueRange) {\n  var minValue = valueRange ? valueRange[0] : d3.min(timeseries, function(d) { return d.value; });\n  var maxValue = valueRange ? valueRange[1] : d3.max(timeseries, function(d) { return d.value; });\n  var minTime = d3.min(timeseries, function(d) { return d.time; });\n  var maxTime = d3.max(timeseries, function(d) { return d.time; });\n\n  var tickUnitsEnum = {\n    \"HALF_HOUR\": 1,\n    \"HOUR\": 2,\n    \"YEAR\": 3\n  };\n\n  var tickUnits;\n  if (maxTime - minTime <= 3 * 3600) {\n    tickUnits = tickUnitsEnum.HALF_HOUR;\n  } else if (maxTime - minTime <= 24 * 3600) {\n    tickUnits = tickUnitsEnum.HOUR;\n  } else {\n    tickUnits = tickUnitsEnum.YEAR;\n  }\n\n  // Given a time as seconds since the epoch, return a String representing the time in UTC in appropriate units.\n  function formatTime(time, forTicks) {\n    var d = new Date(time * 1000);\n    switch (tickUnits) {\n      case tickUnitsEnum.HALF_HOUR:\n      case tickUnitsEnum.HOUR:\n        return d3.format(\"02f\")(d.getUTCHours()) + \":\" + d3.format(\"02f\")(d.getUTCMinutes());\n      case tickUnitsEnum.YEAR:\n        return forTicks ?\n            d.getUTCFullYear() + '' :\n            d.getUTCFullYear() + \"-\" + d3.format(\"02f\")(d.getUTCMonth() + 1) + \"-\" + d3.format(\"02f\")(d.getUTCDate());\n    }\n  }\n\n  var edgePadding = 20;\n  var xAxisSpace = 15, yAxisSpace = 20;\n  var titleSpace = 20, titleOffset = 5;\n  var labelPaddingX = 5, labelPaddingY = 3, dataLabelSpacing = 15, noteLabelSpacing = 20;\n\n  var svg = d3.select(selector)\n      .append(\"svg:svg\")\n      .data([timeseries])\n      // From https://stackoverflow.com/questions/16265123/resize-svg-when-window-is-resized-in-d3-js.\n      .attr(\"preserveAspectRatio\", \"xMinYMin meet\")\n      .attr(\"viewBox\", \"0 0 \" + size[0] + \" \" + size[1])\n      .attr(\"class\", \"graph\");\n\n  var width = size[0] - 2 * edgePadding - yAxisSpace,\n      height = size[1] - 2 * edgePadding - xAxisSpace - titleSpace,\n      xScale = d3.scale.linear().domain([minTime, maxTime]).range([0, width]),\n      yScale = d3.scale.linear().domain([minValue, maxValue]).range([height, 0]);\n\n  var vis = svg.append(\"svg:g\")\n      .attr(\"transform\", \"translate(\" + (edgePadding + yAxisSpace) + \",\" + (edgePadding + titleSpace) + \")\");\n\n  // Title.\n  vis.append(\"svg:text\")\n      .attr(\"class\", \"title\")\n      .attr(\"x\", 0.5 * width - yAxisSpace)\n      .attr(\"y\", - (titleSpace - titleOffset))\n      .attr(\"text-anchor\", \"middle\")\n      .text(title);\n\n  // Notes.\n  var notes = vis.selectAll(\"rect.note\")\n      .data(noteData)\n    .enter().append(\"svg:rect\")\n      .attr(\"class\", \"note\")\n      .attr(\"x\", function(d) { return xScale(d.time) - 3; })\n      .attr(\"y\", 0)\n      .attr(\"width\", 6)\n      .attr(\"height\", height);\n  notes.on(\"mouseover\", function(d, i) {\n    d3.select(noteLabels[0][i]).transition().duration(150).style(\"opacity\", 1);\n  });\n  notes.on(\"mouseout\", function(d, i) {\n    d3.select(noteLabels[0][i]).transition().duration(150).style(\"opacity\", 0);\n  });\n\n  // X ticks.\n  xScale.ticks = function(count) {\n    var startDate = new Date(minTime * 1000);\n    var endDate = new Date(maxTime * 1000);\n    var tickDate = new Date(minTime * 1000)\n    var advanceFunc = null;\n\n    switch (tickUnits) {\n      case tickUnitsEnum.HALF_HOUR:\n      case tickUnitsEnum.HOUR:\n        tickDate.setUTCMinutes(0);\n        tickDate.setUTCSeconds(0);\n        advanceFunc = (tickUnits == tickUnitsEnum.HALF_HOUR) ?\n            function(d) { d.setUTCMinutes(d.getUTCMinutes() + 30); } :\n            function(d) { d.setUTCHours(d.getUTCHours() + 1); };\n        break;\n      case tickUnitsEnum.YEAR:\n        // Firefox 3.6 doesn't seem willing to parse a UTC string.\n        tickDate.setUTCMonth(0);  // <-- whoever did this is a jerk\n        tickDate.setUTCDate(1);\n        tickDate.setUTCHours(0);\n        tickDate.setUTCMinutes(0);\n        tickDate.setUTCSeconds(0);\n        advanceFunc = function(d) { d.setUTCFullYear(d.getUTCFullYear() + 1); };\n        break;\n    }\n\n    var values = [];\n    for (; tickDate < endDate; advanceFunc(tickDate)) {\n      if (tickDate >= startDate) {\n        values.push(tickDate.getTime() / 1000);\n      }\n    }\n    return values;\n  }\n\n  var xRules = vis.selectAll(\"g.xrule\")\n      .data(xScale.ticks(10))\n    .enter().append(\"svg:g\")\n      .attr(\"class\", \"rule\");\n\n  xRules.append(\"svg:line\")\n      .attr(\"x1\", xScale)\n      .attr(\"x2\", xScale)\n      .attr(\"y1\", 0)\n      .attr(\"y2\", height - 1);\n\n  xRules.append(\"svg:text\")\n      .attr(\"x\", xScale)\n      .attr(\"y\", height + 15)\n      .attr(\"dy\", \".71em\")\n      .attr(\"text-anchor\", \"middle\")\n      .text(function(d) { return formatTime(d, true); });\n\n  // Y ticks.\n  var yRules = vis.selectAll(\"g.yrule\")\n      .data(yScale.ticks(10))\n    .enter().append(\"svg:g\")\n      .attr(\"class\", \"rule\");\n\n  yRules.append(\"svg:line\")\n      .attr(\"y1\", yScale)\n      .attr(\"y2\", yScale)\n      .attr(\"x1\", 0)\n      .attr(\"x2\", width + 1);\n\n  yRules.append(\"svg:text\")\n      .attr(\"y\", yScale)\n      .attr(\"x\", -10)\n      .attr(\"dy\", \".35em\")\n      .attr(\"text-anchor\", \"end\")\n      .text(yScale.tickFormat(10));\n\n  // Line.\n  vis.append(\"svg:path\")\n      .attr(\"class\", \"line\")\n      .attr(\"pointer-events\", \"none\")\n      .attr(\"d\", d3.svg.line()\n        .x(function(d) { return xScale(d.time); })\n        .y(function(d) { return yScale(d.value); }));\n\n  // Circles.\n  var circles = vis.selectAll(\"circle.line\")\n      .data(timeseries)\n    .enter().append(\"svg:circle\")\n      .attr(\"class\", \"line\")\n      .attr(\"cx\", function(d) { return xScale(d.time); })\n      .attr(\"cy\", function(d) { return yScale(d.value); })\n      .attr(\"r\", 3.5);\n  circles.on(\"mouseover\", function(d, i) {\n    d3.select(dataLabels[0][i]).transition().duration(150).style(\"opacity\", 1);\n  });\n  circles.on(\"mouseout\", function(d, i) {\n    d3.select(dataLabels[0][i]).transition().duration(150).style(\"opacity\", 0);\n  });\n\n  // Note labels.\n  var noteLabels = vis.selectAll(\"g.noteLabel\")\n      .data(noteData)\n    .enter().append(\"svg:g\")\n      .attr(\"class\", \"noteLabel label\")\n      .attr(\"pointer-events\", \"none\")\n      .attr(\"opacity\", 0);\n  var noteLabelBoxes = noteLabels.append(\"svg:rect\");\n  var noteLabelText = noteLabels.append(\"svg:text\")\n      .attr(\"text-anchor\", \"middle\")\n      .text(function(d) { return formatTime(d.time, false) + \": \" + d.text; })\n      .attr(\"x\", function(d) { return Math.max(0.5 * this.getBBox().width, Math.min(width - 0.5 * this.getBBox().width, xScale(d.time))); })\n      .attr(\"y\", noteLabelSpacing);\n  noteLabelBoxes.data(noteLabelText[0])\n      .attr(\"x\", function(d) { return d.getBBox().x - labelPaddingX; })\n      .attr(\"y\", function(d) { return d.getBBox().y - labelPaddingY; })\n      .attr(\"width\", function(d) { return d.getBBox().width + 2 * labelPaddingX; })\n      .attr(\"height\", function(d) { return d.getBBox().height + 2 * labelPaddingY; });\n\n  // Data labels.\n  var dataLabels = vis.selectAll(\"g.dataLabel\")\n      .data(timeseries)\n    .enter().append(\"svg:g\")\n      .attr(\"class\", \"dataLabel label\")\n      .attr(\"pointer-events\", \"none\")\n      .attr(\"opacity\", 0);\n  var dataLabelBoxes = dataLabels.append(\"svg:rect\");\n  var dataLabelText = dataLabels.append(\"svg:text\")\n      .attr(\"text-anchor\", \"middle\")\n      .text(function(d) { return formatTime(d.time, false) + \": \" + d.value + (units ? ' ' + units : ''); })\n      .attr(\"x\", function(d) { return Math.max(0.5 * this.getBBox().width, Math.min(width - 0.5 * this.getBBox().width, xScale(d.time))); })\n      .attr(\"y\", function(d) { return yScale(d.value) - dataLabelSpacing });\n  dataLabelBoxes.data(dataLabelText[0])\n      .attr(\"x\", function(d) { return d.getBBox().x - labelPaddingX; })\n      .attr(\"y\", function(d) { return d.getBBox().y - labelPaddingY; })\n      .attr(\"width\", function(d) { return d.getBBox().width + 2 * labelPaddingX; })\n      .attr(\"height\", function(d) { return d.getBBox().height + 2 * labelPaddingY; });\n}\n\n\ndocument.addEventListener('DOMContentLoaded', () => {\n  // Get the data for the requested graph.\n  // |dataSets| is an object of objects with the following properties:\n  // title:  string\n  // points: array of { time: epoch_time, value: num } objects\n  // notes:  array of { time: epoch_time, text: string } objects\n  // range:  [min, max]\n  // units:  string\n  var name = window.location.search.substring(1);\n  d = dataSets[name];\n  if (!d) {\n    throw 'Data not found for \"' + name + \"'\";;\n  }\n  appendGraph('#graph-node', [window.innerWidth, window.innerHeight],\n              d.title, d.points, d.notes, d.units, d.range);\n\n  // Handle dark/light mode using code defined in dark.js.\n  applyTheme();\n  darkQuery.addEventListener('change', () => applyTheme());\n  window.addEventListener('storage', () => applyTheme());\n});\n",
	"graph.css":                    "main .box>.body .graph{background-color:transparent;overflow:hidden;padding:0}\n",
	"map-iframe-body.js":           "applyTheme(); // defined in dark.js\n",
	"map-iframe.css":               "body{background-size:100% 100%;color-scheme:light;margin:0;overflow:hidden}body.dark{color-scheme:dark}body.dark .gm-style-mtc,body.dark .gm-fullscreen-control,body.dark .gm-bundled-control{filter:brightness(0.7)}.loading{position:absolute}#map-div{display:inline-block;height:100%;position:absolute;visibility:hidden;width:100%}#map-div.loaded{visibility:visible}a.location{color:#555;cursor:pointer;font-family:Arial, Helvetica, sans-serif;text-decoration:underline}.gm-style-iw button:focus{outline:0}.gm-style-mtc *{font-size:16px !important}.gm-style-mtc button{padding:7px 18px 6px 12px !important}.gm-style-mtc button img{margin-top:0 !important}\n",
	"map-iframe.js":                "let pageUrl = null;\nlet mapDiv = null;\nlet map = null;\nlet infoWindow = null;\n\nfunction initializeMap() {\n  // AMP effectively doesn't let us use allow-same-origin (see\n  // https://github.com/ampproject/amphtml/blob/master/spec/amp-iframe-origin-policy.md),\n  // which prevents us from just updating window.top.location.hash in\n  // selectPoint(). Get the base page URL from document.referrer so we can use\n  // it to construct a URL with the correct fragment and assign that directly to\n  // window.top.location, which _is_ allowed.\n  //\n  // TODO: This doesn't work quite right. When a page is loaded from a Google\n  // results page, it looks like we get a URL like\n  // https://www-example-org.cdn.ampproject.org/v/s/www.example.org/page.amp.html\n  // here, but the outer page seems to actually be\n  // https://www.google.com/amp/s/www.example.org/page.amp.html. Per\n  // https://developers.googleblog.com/2017/02/whats-in-amp-url.html, this\n  // sounds like it's weirdness relating to the prerendering. The upshot is that\n  // clicking on a location link triggers a navigation to the ampproject.org\n  // URL. I'm not sure how to fix this, since I don't want to hardcode a\n  // www.google.com/amp URL here.\n  pageUrl = document.referrer.split('#', 1)[0];\n\n  const mapOptions = {\n    mapTypeId: google.maps.MapTypeId.ROADMAP,\n    styles: getStyles(),\n    // Disable scrollwheel zooming; it's too easy to trigger while scrolling the\n    // page up or down.\n    scrollwheel: false,\n    // Make controls less huge.\n    controlSize: 32,\n    mapTypeControl: true,\n    mapTypeControlOptions: {\n      style: google.maps.MapTypeControlStyle.DROPDOWN_MENU,\n      position: google.maps.ControlPosition.LEFT_TOP,\n    },\n  };\n  mapDiv = document.getElementById('map-div');\n  map = new google.maps.Map(mapDiv, mapOptions);\n  infoWindow = new google.maps.InfoWindow();\n\n  // Show the map after the tiles have fully loaded, but also watch for the\n  // 'idle' event (which often fires earlier) as a fallback for slow\n  // connections.\n  google.maps.event.addListenerOnce(map, 'tilesloaded', () => {\n    mapDiv.classList.add('loaded');\n  });\n  google.maps.event.addListenerOnce(map, 'idle', () => {\n    window.setTimeout(() => mapDiv.classList.add('loaded'), 5000);\n  });\n\n  const bounds = new google.maps.LatLngBounds();\n  for (let i = 0; i < points.length; i++) {\n    const p = points[i];\n    p.latLong = new google.maps.LatLng(p.latLong[0], p.latLong[1]);\n    bounds.extend(p.latLong);\n\n    const letter = String.fromCharCode(65 + i);\n    const markerOptions = {\n      position: p.latLong,\n      title: p.name,\n      icon: `https://chart.googleapis.com/chart?chst=d_map_pin_letter&chld=${letter}|fc783a|33180c`,\n      map,\n    };\n    p.marker = new google.maps.Marker(markerOptions);\n    google.maps.event.addListener(\n      p.marker,\n      'click',\n      selectPoint.bind(null, p.id, false)\n    );\n  }\n\n  map.fitBounds(bounds);\n  updateStyle();\n}\n\nfunction selectPoint(id, center) {\n  if (!map) {\n    console.log('Map not initialized');\n    return;\n  }\n\n  const point = points.find((p) => p.id == id);\n  if (!point) {\n    console.log('Unable to find point with ID ' + id);\n    return;\n  }\n\n  const a = document.createElement('a');\n  a.appendChild(document.createTextNode(point.name));\n  a.className = 'location';\n  a.addEventListener('click', () => (window.top.location = `${pageUrl}#${id}`));\n  infoWindow.setContent(a);\n  infoWindow.open(map, point.marker);\n\n  if (center) {\n    map.setCenter(point.latLong);\n    mapDiv.scrollIntoView(true);\n  }\n}\n\n// Returns the 'styles' value for google.maps.MapOptions.\nfunction getStyles() {\n  // Just use the default light style if the dark theme isn't being used.\n  if (!document.body.classList.contains('dark')) return undefined;\n\n  // Generated using https://mapstyle.withgoogle.com/\n  return [\n    {\n      elementType: 'geometry',\n      stylers: [{ color: '#242f3e' }],\n    },\n    {\n      elementType: 'labels.text.fill',\n      stylers: [{ color: '#746855' }],\n    },\n    {\n      elementType: 'labels.text.stroke',\n      stylers: [{ color: '#242f3e' }],\n    },\n    {\n      featureType: 'administrative.locality',\n      elementType: 'labels.text.fill',\n      stylers: [{ color: '#d59563' }],\n    },\n    {\n      featureType: 'poi',\n      elementType: 'labels.text.fill',\n      stylers: [{ color: '#d59563' }],\n    },\n    {\n      featureType: 'poi.park',\n      elementType: 'geometry',\n      stylers: [{ color: '#263c3f' }],\n    },\n    {\n      featureType: 'poi.park',\n      elementType: 'labels.text.fill',\n      stylers: [{ color: '#6b9a76' }],\n    },\n    {\n      featureType: 'road',\n      elementType: 'geometry',\n      stylers: [{ color: '#38414e' }],\n    },\n    {\n      featureType: 'road',\n      elementType: 'geometry.stroke',\n      stylers: [{ color: '#212a37' }],\n    },\n    {\n      featureType: 'road',\n      elementType: 'labels.text.fill',\n      stylers: [{ color: '#9ca5b3' }],\n    },\n    {\n      featureType: 'road.highway',\n      elementType: 'geometry',\n      stylers: [{ color: '#746855' }],\n    },\n    {\n      featureType: 'road.highway',\n      elementType: 'geometry.stroke',\n      stylers: [{ color: '#1f2835' }],\n    },\n    {\n      featureType: 'road.highway',\n      elementType: 'labels.text.fill',\n      stylers: [{ color: '#f3d19c' }],\n    },\n    {\n      featureType: 'transit',\n      elementType: 'geometry',\n      stylers: [{ color: '#2f3948' }],\n    },\n    {\n      featureType: 'transit.station',\n      elementType: 'labels.text.fill',\n      stylers: [{ color: '#d59563' }],\n    },\n    {\n      featureType: 'water',\n      elementType: 'geometry',\n      stylers: [{ color: '#17263c' }],\n    },\n    {\n      featureType: 'water',\n      elementType: 'labels.text.fill',\n      stylers: [{ color: '#515c6d' }],\n    },\n    {\n      featureType: 'water',\n      elementType: 'labels.text.stroke',\n      stylers: [{ color: '#17263c' }],\n    },\n    // Deemphasize POI and road icons since they compete with our markers\n    // otherwise. The styler ominously warns, \"The effect of the following\n    // stylers will change whenever Google updates the base map style.\n    // Use with caution.\"\n    {\n      featureType: 'poi',\n      elementType: 'labels.icon',\n      stylers: [{ saturation: -50 }, { lightness: -30 }],\n    },\n    {\n      featureType: 'road',\n      elementType: 'labels.icon',\n      stylers: [{ saturation: -50 }, { lightness: -30 }],\n    },\n  ];\n}\n\nfunction updateStyle() {\n  // Handle dark/light mode using code defined in dark.js.\n  applyTheme();\n  map.setOptions({ styles: getStyles() });\n}\n\nwindow.addEventListener('DOMContentLoaded', () => {\n  darkQuery.addEventListener('change', () => updateStyle());\n  window.addEventListener('storage', () => updateStyle());\n  initializeMap();\n});\n\nwindow.addEventListener('message', (e) => selectPoint(e.data.id, true));\n",
	"map.css":                      "main .box>.body .mapbox{height:0;position:relative}main .box>.body .mapbox iframe{background-size:100% 100%;border:none;height:100%;left:0;overflow:hidden;position:absolute;top:0;width:100%}\n",
	"map.js":                       "// Wire up links to post messages to the iframe to activate markers.\ndocument.addEventListener('DOMContentLoaded', () => {\n  const iframe = document.getElementById('map');\n  const anchors = document.getElementsByClassName('map-link');\n  for (let i = 0; i < anchors.length; i++) {\n    const a = anchors[i];\n    const id = a.parentElement.parentElement.id;\n    a.addEventListener('click', () =>\n      iframe.contentWindow.postMessage({ id }, '*', [])\n    );\n  }\n});\n",
	"mobile.css":                   ".desktop-only{display:none}header .toggle{cursor:pointer}header .box>.body{overflow:hidden}header .collapsed-mobile .toggle{transform:rotate(180deg)}header .collapsed-mobile .box>.body{max-height:0px}header .collapsed-mobile .box>.body>ul{opacity:0}main .box{width:100%}main .box>.body figure.mobile-center{margin-left:auto;margin-right:auto}\n",
	"nonamp.css":                   ".img-wrapper{display:inline-block;position:relative;vertical-align:bottom}.img-wrapper>svg{position:absolute}.img-wrapper>picture{position:relative}@media screen and (-ms-high-contrast: active),(-ms-high-contrast: none){.img-wrapper>svg{display:none}}\n"}
