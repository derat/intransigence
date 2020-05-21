// Code generated by gen_filemap.go. DO NOT EDIT.

package render

var stdInline = map[string]string{
	"amp-boilerplate-noscript.css": "body{-webkit-animation:none;-moz-animation:none;-ms-animation:none;animation:none}",
	"amp-boilerplate.css":          "body{-webkit-animation:-amp-start 8s steps(1,end) 0s 1 normal both;-moz-animation:-amp-start 8s steps(1,end) 0s 1 normal both;-ms-animation:-amp-start 8s steps(1,end) 0s 1 normal both;animation:-amp-start 8s steps(1,end) 0s 1 normal both}@-webkit-keyframes -amp-start{from{visibility:hidden}to{visibility:visible}}@-moz-keyframes -amp-start{from{visibility:hidden}to{visibility:visible}}@-ms-keyframes -amp-start{from{visibility:hidden}to{visibility:visible}}@-o-keyframes -amp-start{from{visibility:hidden}to{visibility:visible}}@keyframes -amp-start{from{visibility:hidden}to{visibility:visible}}",
	"base.js.min":                  "function toggleNavbox(c){var b=document.getElementById(\"nav-list\");var a=document.getElementById(\"nav-toggle-img\");var d=b.className==\"collapsed-mobile\";b.className=d?\"\":\"collapsed-mobile\";a.className=d?\"\":\"expand\"}document.addEventListener(\"DOMContentLoaded\",function(){document.getElementById(\"nav-logo\").addEventListener(\"click\",toggleNavbox,false);document.getElementById(\"nav-title\").addEventListener(\"click\",toggleNavbox,false)},false);",
	"graph-iframe.css":             "body{margin:0;overflow:hidden}svg.graph{background-color:white;display:inline-block;height:100%;position:absolute;width:100%}circle.line{fill:white;stroke:steelblue;stroke-width:1.5px}path.line{fill:none;stroke:steelblue;stroke-width:1.5px}rect.note{fill:#f5f5f5;shape-rendering:crispEdges;stroke:#eee;stroke-width:1px}text.title{font-family:Verdana,Helvetica,Arial,sans-serif;font-size:12px}.label rect{fill:#fffbe0;shape-rendering:crispEdges;stroke:#d2cfb9;stroke-width:1px;z-index:1}.label text{font-family:Helvetica,Arial,sans-serif;font-size:11px;z-index:2}.rule line{shape-rendering:crispEdges;stroke:#eee}.rule text{font-family:Helvetica,Arial,sans-serif;font-size:10px}\n",
	"graph-iframe.js.min":          "var d=null;function appendGraph(k,m,z,E,M,L,A){var i=A?A[0]:d3.min(E,function(O){return O.value});var q=A?A[1]:d3.max(E,function(O){return O.value});var c=d3.min(E,function(O){return O.time});var n=d3.max(E,function(O){return O.time});var f={HALF_HOUR:1,HOUR:2,YEAR:3};var D;if(n-c<=3*3600){D=f.HALF_HOUR}else{if(n-c<=24*3600){D=f.HOUR}else{D=f.YEAR}}function B(P,O){var Q=new Date(P*1000);switch(D){case f.HALF_HOUR:case f.HOUR:return d3.format(\"02f\")(Q.getUTCHours())+\":\"+d3.format(\"02f\")(Q.getUTCMinutes());case f.YEAR:return O?Q.getUTCFullYear()+\"\":Q.getUTCFullYear()+\"-\"+d3.format(\"02f\")(Q.getUTCMonth()+1)+\"-\"+d3.format(\"02f\")(Q.getUTCDate())}}var x=20;var r=15,y=20;var N=20,G=5;var u=5,t=3,j=15,h=20;var g=d3.select(k).append(\"svg:svg\").data([E]).attr(\"preserveAspectRatio\",\"xMinYMin meet\").attr(\"viewBox\",\"0 0 \"+m[0]+\" \"+m[1]).attr(\"class\",\"graph\");var b=m[0]-2*x-y,e=m[1]-2*x-r-N,s=d3.scale.linear().domain([c,n]).range([0,b]),F=d3.scale.linear().domain([i,q]).range([e,0]);var J=g.append(\"svg:g\").attr(\"transform\",\"translate(\"+(x+y)+\",\"+(x+N)+\")\");J.append(\"svg:text\").attr(\"class\",\"title\").attr(\"x\",0.5*b-y).attr(\"y\",-(N-G)).attr(\"text-anchor\",\"middle\").text(z);var K=J.selectAll(\"rect.note\").data(M).enter().append(\"svg:rect\").attr(\"class\",\"note\").attr(\"x\",function(O){return s(O.time)-3}).attr(\"y\",0).attr(\"width\",6).attr(\"height\",e);K.on(\"mouseover\",function(P,O){d3.select(this).transition().duration(150).style(\"fill\",\"#eee\").style(\"stroke\",\"#ddd\");d3.select(p[0][O]).transition().duration(150).style(\"opacity\",1)});K.on(\"mouseout\",function(P,O){d3.select(this).transition().duration(300).style(\"fill\",\"#f5f5f5\").style(\"stroke\",\"#eee\");d3.select(p[0][O]).transition().duration(150).style(\"opacity\",0)});s.ticks=function(R){var O=new Date(c*1000);var S=new Date(n*1000);var Q=new Date(c*1000);var T=null;switch(D){case f.HALF_HOUR:case f.HOUR:Q.setUTCMinutes(0);Q.setUTCSeconds(0);T=(D==f.HALF_HOUR)?function(U){U.setUTCMinutes(U.getUTCMinutes()+30)}:function(U){U.setUTCHours(U.getUTCHours()+1)};break;case f.YEAR:Q.setUTCMonth(0);Q.setUTCDate(1);Q.setUTCHours(0);Q.setUTCMinutes(0);Q.setUTCSeconds(0);T=function(U){U.setUTCFullYear(U.getUTCFullYear()+1)};break}var P=[];for(;Q<S;T(Q)){if(Q>=O){P.push(Q.getTime()/1000)}}return P};var H=J.selectAll(\"g.xrule\").data(s.ticks(10)).enter().append(\"svg:g\").attr(\"class\",\"rule\");H.append(\"svg:line\").attr(\"x1\",s).attr(\"x2\",s).attr(\"y1\",0).attr(\"y2\",e-1);H.append(\"svg:text\").attr(\"x\",s).attr(\"y\",e+15).attr(\"dy\",\".71em\").attr(\"text-anchor\",\"middle\").text(function(O){return B(O,true)});var w=J.selectAll(\"g.yrule\").data(F.ticks(10)).enter().append(\"svg:g\").attr(\"class\",\"rule\");w.append(\"svg:line\").attr(\"y1\",F).attr(\"y2\",F).attr(\"x1\",0).attr(\"x2\",b+1);w.append(\"svg:text\").attr(\"y\",F).attr(\"x\",-10).attr(\"dy\",\".35em\").attr(\"text-anchor\",\"end\").text(F.tickFormat(10));J.append(\"svg:path\").attr(\"class\",\"line\").attr(\"pointer-events\",\"none\").attr(\"d\",d3.svg.line().x(function(O){return s(O.time)}).y(function(O){return F(O.value)}));var o=J.selectAll(\"circle.line\").data(E).enter().append(\"svg:circle\").attr(\"class\",\"line\").attr(\"cx\",function(O){return s(O.time)}).attr(\"cy\",function(O){return F(O.value)}).attr(\"r\",3.5);o.on(\"mouseover\",function(P,O){d3.select(this).transition().duration(150).style(\"fill\",\"steelblue\");d3.select(l[0][O]).transition().duration(150).style(\"opacity\",1)});o.on(\"mouseout\",function(P,O){d3.select(this).transition().duration(300).style(\"fill\",\"white\");d3.select(l[0][O]).transition().duration(150).style(\"opacity\",0)});var p=J.selectAll(\"g.noteLabel\").data(M).enter().append(\"svg:g\").attr(\"class\",\"noteLabel label\").attr(\"pointer-events\",\"none\").attr(\"opacity\",0);var I=p.append(\"svg:rect\");var C=p.append(\"svg:text\").attr(\"text-anchor\",\"middle\").text(function(O){return B(O.time,false)+\": \"+O.text}).attr(\"x\",function(O){return Math.max(0.5*this.getBBox().width,Math.min(b-0.5*this.getBBox().width,s(O.time)))}).attr(\"y\",h);I.data(C[0]).attr(\"x\",function(O){return O.getBBox().x-u}).attr(\"y\",function(O){return O.getBBox().y-t}).attr(\"width\",function(O){return O.getBBox().width+2*u}).attr(\"height\",function(O){return O.getBBox().height+2*t});var l=J.selectAll(\"g.dataLabel\").data(E).enter().append(\"svg:g\").attr(\"class\",\"dataLabel label\").attr(\"pointer-events\",\"none\").attr(\"opacity\",0);var a=l.append(\"svg:rect\");var v=l.append(\"svg:text\").attr(\"text-anchor\",\"middle\").text(function(O){return B(O.time,false)+\": \"+O.value+(L?\" \"+L:\"\")}).attr(\"x\",function(O){return Math.max(0.5*this.getBBox().width,Math.min(b-0.5*this.getBBox().width,s(O.time)))}).attr(\"y\",function(O){return F(O.value)-j});a.data(v[0]).attr(\"x\",function(O){return O.getBBox().x-u}).attr(\"y\",function(O){return O.getBBox().y-t}).attr(\"width\",function(O){return O.getBBox().width+2*u}).attr(\"height\",function(O){return O.getBBox().height+2*t})}function initPage(){var a=window.location.search.substring(1);d=dataSets[a];if(!d){throw'Data not found for \"'+a+\"'\"}appendGraph(\"#graph-node\",[window.innerWidth,window.innerHeight],d.title,d.points,d.notes,d.units,d.range)}document.addEventListener(\"DOMContentLoaded\",initPage,false);",
	"map-iframe.css":               "body{background-size:100% 100%;margin:0;overflow:hidden}#map-div{display:inline-block;height:100%;position:absolute;visibility:hidden;width:100%}#map-div.loaded{visibility:visible}a.location{color:#555;cursor:pointer;font-family:Arial,Helvetica,sans-serif;text-decoration:underline}\n",
	"map-iframe.js.min":            "var pageUrl=null;var mapDiv=null;var map=null;var infoWindow=null;function initializeMap(){pageUrl=document.referrer.split(\"#\",1)[0];var b={mapTypeId:google.maps.MapTypeId.ROADMAP,scrollwheel:false};mapDiv=document.getElementById(\"map-div\");map=new google.maps.Map(mapDiv,b);infoWindow=new google.maps.InfoWindow();google.maps.event.addListenerOnce(map,\"idle\",function(){mapDiv.className=\"loaded\"});var e=new google.maps.LatLngBounds();for(var c=0;c<points.length;c++){var f=points[c];f.latLong=new google.maps.LatLng(f.latLong[0],f.latLong[1]);e.extend(f.latLong);var d=String.fromCharCode(65+c);var a={position:f.latLong,title:f.name,icon:\"https://chart.googleapis.com/chart?chst=d_map_pin_letter&chld=\"+d+\"|8cf|000\",map:map};f.marker=new google.maps.Marker(a);google.maps.event.addListener(f.marker,\"click\",selectPoint.bind(null,f.id,false))}map.fitBounds(e)}function selectPoint(h,c){var b=null;for(var e=0;e<points.length;++e){if(points[e].id==h){b=points[e];break}}if(!b){console.log(\"Unable to find point with ID \"+h);return}var d=document.createElement(\"a\");var g=function(){window.top.location=pageUrl+\"#\"+h};d.appendChild(document.createTextNode(b.name));d.className=\"location\";d.addEventListener(\"click\",g,false);infoWindow.setContent(d);infoWindow.open(map,b.marker);if(c){map.setCenter(b.latLong);mapDiv.scrollIntoView(true)}}google.maps.event.addDomListener(window,\"load\",initializeMap);window.addEventListener(\"message\",function(a){selectPoint(a.data.id,true)},false);",
	"map.js.min":                   "document.addEventListener(\"DOMContentLoaded\",function(){var d=document.getElementById(\"map\");var e=document.getElementsByClassName(\"map-link\");for(var c=0;c<e.length;c++){var b=e[c];var h=b.parentElement.parentElement.id;var g=d.contentWindow.postMessage.bind(d.contentWindow,{id:h},\"*\",[]);b.addEventListener(\"click\",g,false)}},false);"}
