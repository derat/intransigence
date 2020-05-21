var d = null;

function appendGraph(selector, size, title, timeseries, noteData, units, valueRange) {
  var minValue = valueRange ? valueRange[0] : d3.min(timeseries, function(d) { return d.value; });
  var maxValue = valueRange ? valueRange[1] : d3.max(timeseries, function(d) { return d.value; });
  var minTime = d3.min(timeseries, function(d) { return d.time; });
  var maxTime = d3.max(timeseries, function(d) { return d.time; });

  var tickUnitsEnum = {
    "HALF_HOUR": 1,
    "HOUR": 2,
    "YEAR": 3
  };

  var tickUnits;
  if (maxTime - minTime <= 3 * 3600) {
    tickUnits = tickUnitsEnum.HALF_HOUR;
  } else if (maxTime - minTime <= 24 * 3600) {
    tickUnits = tickUnitsEnum.HOUR;
  } else {
    tickUnits = tickUnitsEnum.YEAR;
  }

  // Given a time as seconds since the epoch, return a String representing the time in UTC in appropriate units.
  function formatTime(time, forTicks) {
    var d = new Date(time * 1000);
    switch (tickUnits) {
      case tickUnitsEnum.HALF_HOUR:
      case tickUnitsEnum.HOUR:
        return d3.format("02f")(d.getUTCHours()) + ":" + d3.format("02f")(d.getUTCMinutes());
      case tickUnitsEnum.YEAR:
        return forTicks ?
            d.getUTCFullYear() + '' :
            d.getUTCFullYear() + "-" + d3.format("02f")(d.getUTCMonth() + 1) + "-" + d3.format("02f")(d.getUTCDate());
    }
  }

  var edgePadding = 20;
  var xAxisSpace = 15, yAxisSpace = 20;
  var titleSpace = 20, titleOffset = 5;
  var labelPaddingX = 5, labelPaddingY = 3, dataLabelSpacing = 15, noteLabelSpacing = 20;

  var svg = d3.select(selector)
      .append("svg:svg")
      .data([timeseries])
      // From https://stackoverflow.com/questions/16265123/resize-svg-when-window-is-resized-in-d3-js.
      .attr("preserveAspectRatio", "xMinYMin meet")
      .attr("viewBox", "0 0 " + size[0] + " " + size[1])
      .attr("class", "graph");

  var width = size[0] - 2 * edgePadding - yAxisSpace,
      height = size[1] - 2 * edgePadding - xAxisSpace - titleSpace,
      xScale = d3.scale.linear().domain([minTime, maxTime]).range([0, width]),
      yScale = d3.scale.linear().domain([minValue, maxValue]).range([height, 0]);

  var vis = svg.append("svg:g")
      .attr("transform", "translate(" + (edgePadding + yAxisSpace) + "," + (edgePadding + titleSpace) + ")");

  // Title.
  vis.append("svg:text")
      .attr("class", "title")
      .attr("x", 0.5 * width - yAxisSpace)
      .attr("y", - (titleSpace - titleOffset))
      .attr("text-anchor", "middle")
      .text(title);

  // Notes.
  var notes = vis.selectAll("rect.note")
      .data(noteData)
    .enter().append("svg:rect")
      .attr("class", "note")
      .attr("x", function(d) { return xScale(d.time) - 3; })
      .attr("y", 0)
      .attr("width", 6)
      .attr("height", height);
  notes.on("mouseover", function(d, i) {
    d3.select(this).transition().duration(150).style("fill", "#eee").style("stroke", "#ddd");
    d3.select(noteLabels[0][i]).transition().duration(150).style("opacity", 1);
  });
  notes.on("mouseout", function(d, i) {
    d3.select(this).transition().duration(300).style("fill", "#f5f5f5").style("stroke", "#eee");
    d3.select(noteLabels[0][i]).transition().duration(150).style("opacity", 0);
  });

  // X ticks.
  xScale.ticks = function(count) {
    var startDate = new Date(minTime * 1000);
    var endDate = new Date(maxTime * 1000);
    var tickDate = new Date(minTime * 1000)
    var advanceFunc = null;

    switch (tickUnits) {
      case tickUnitsEnum.HALF_HOUR:
      case tickUnitsEnum.HOUR:
        tickDate.setUTCMinutes(0);
        tickDate.setUTCSeconds(0);
        advanceFunc = (tickUnits == tickUnitsEnum.HALF_HOUR) ?
            function(d) { d.setUTCMinutes(d.getUTCMinutes() + 30); } :
            function(d) { d.setUTCHours(d.getUTCHours() + 1); };
        break;
      case tickUnitsEnum.YEAR:
        // Firefox 3.6 doesn't seem willing to parse a UTC string.
        tickDate.setUTCMonth(0);  // <-- whoever did this is a jerk
        tickDate.setUTCDate(1);
        tickDate.setUTCHours(0);
        tickDate.setUTCMinutes(0);
        tickDate.setUTCSeconds(0);
        advanceFunc = function(d) { d.setUTCFullYear(d.getUTCFullYear() + 1); };
        break;
    }

    var values = [];
    for (; tickDate < endDate; advanceFunc(tickDate)) {
      if (tickDate >= startDate) {
        values.push(tickDate.getTime() / 1000);
      }
    }
    return values;
  }

  var xRules = vis.selectAll("g.xrule")
      .data(xScale.ticks(10))
    .enter().append("svg:g")
      .attr("class", "rule");

  xRules.append("svg:line")
      .attr("x1", xScale)
      .attr("x2", xScale)
      .attr("y1", 0)
      .attr("y2", height - 1);

  xRules.append("svg:text")
      .attr("x", xScale)
      .attr("y", height + 15)
      .attr("dy", ".71em")
      .attr("text-anchor", "middle")
      .text(function(d) { return formatTime(d, true); });

  // Y ticks.
  var yRules = vis.selectAll("g.yrule")
      .data(yScale.ticks(10))
    .enter().append("svg:g")
      .attr("class", "rule");

  yRules.append("svg:line")
      .attr("y1", yScale)
      .attr("y2", yScale)
      .attr("x1", 0)
      .attr("x2", width + 1);

  yRules.append("svg:text")
      .attr("y", yScale)
      .attr("x", -10)
      .attr("dy", ".35em")
      .attr("text-anchor", "end")
      .text(yScale.tickFormat(10));

  // Line.
  vis.append("svg:path")
      .attr("class", "line")
      .attr("pointer-events", "none")
      .attr("d", d3.svg.line()
        .x(function(d) { return xScale(d.time); })
        .y(function(d) { return yScale(d.value); }));

  // Circles.
  var circles = vis.selectAll("circle.line")
      .data(timeseries)
    .enter().append("svg:circle")
      .attr("class", "line")
      .attr("cx", function(d) { return xScale(d.time); })
      .attr("cy", function(d) { return yScale(d.value); })
      .attr("r", 3.5);
  circles.on("mouseover", function(d, i) {
    d3.select(this).transition().duration(150).style("fill", "steelblue");
    d3.select(dataLabels[0][i]).transition().duration(150).style("opacity", 1);
  });
  circles.on("mouseout", function(d, i) {
    d3.select(this).transition().duration(300).style("fill", "white");
    d3.select(dataLabels[0][i]).transition().duration(150).style("opacity", 0);
  });

  // Note labels.
  var noteLabels = vis.selectAll("g.noteLabel")
      .data(noteData)
    .enter().append("svg:g")
      .attr("class", "noteLabel label")
      .attr("pointer-events", "none")
      .attr("opacity", 0);
  var noteLabelBoxes = noteLabels.append("svg:rect");
  var noteLabelText = noteLabels.append("svg:text")
      .attr("text-anchor", "middle")
      .text(function(d) { return formatTime(d.time, false) + ": " + d.text; })
      .attr("x", function(d) { return Math.max(0.5 * this.getBBox().width, Math.min(width - 0.5 * this.getBBox().width, xScale(d.time))); })
      .attr("y", noteLabelSpacing);
  noteLabelBoxes.data(noteLabelText[0])
      .attr("x", function(d) { return d.getBBox().x - labelPaddingX; })
      .attr("y", function(d) { return d.getBBox().y - labelPaddingY; })
      .attr("width", function(d) { return d.getBBox().width + 2 * labelPaddingX; })
      .attr("height", function(d) { return d.getBBox().height + 2 * labelPaddingY; });

  // Data labels.
  var dataLabels = vis.selectAll("g.dataLabel")
      .data(timeseries)
    .enter().append("svg:g")
      .attr("class", "dataLabel label")
      .attr("pointer-events", "none")
      .attr("opacity", 0);
  var dataLabelBoxes = dataLabels.append("svg:rect");
  var dataLabelText = dataLabels.append("svg:text")
      .attr("text-anchor", "middle")
      .text(function(d) { return formatTime(d.time, false) + ": " + d.value + (units ? ' ' + units : ''); })
      .attr("x", function(d) { return Math.max(0.5 * this.getBBox().width, Math.min(width - 0.5 * this.getBBox().width, xScale(d.time))); })
      .attr("y", function(d) { return yScale(d.value) - dataLabelSpacing });
  dataLabelBoxes.data(dataLabelText[0])
      .attr("x", function(d) { return d.getBBox().x - labelPaddingX; })
      .attr("y", function(d) { return d.getBBox().y - labelPaddingY; })
      .attr("width", function(d) { return d.getBBox().width + 2 * labelPaddingX; })
      .attr("height", function(d) { return d.getBBox().height + 2 * labelPaddingY; });
}

// |dataSets| is an object of objects with the following properties:
// title:  string
// points: array of { time: epoch_time, value: num } objects
// notes:  array of { time: epoch_time, text: string } objects
// range:  [min, max]
// units:  string
function initPage() {
  // Get the data for the requested graph.
  var name = window.location.search.substring(1);
  d = dataSets[name];
  if (!d) {
    throw 'Data not found for "' + name + "'";;
  }
  appendGraph('#graph-node', [window.innerWidth, window.innerHeight],
              d.title, d.points, d.notes, d.units, d.range);
}

document.addEventListener('DOMContentLoaded', initPage, false);
