// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package render

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"path/filepath"
	"strings"
)

// IframeOutDir is the subdirectory under the output dir for generated iframe pages.
const IframeOutDir = "iframes"

// Iframe renders and returns the framed page described by the supplied YAML data.
func Iframe(si SiteInfo, yb []byte) ([]byte, error) {
	var data struct {
		// Map-specific data.
		MapPlaceholderLight string `yaml:"map_placeholder_light"` // placeholder image path (relative to iframe)
		MapPlaceholderDark  string `yaml:"map_placeholder_dark"`  // placeholder for dark theme
		MapPoints           []struct {
			Name    string     `json:"name" yaml:"name"`        // name displayed on label
			LatLong [2]float64 `json:"latLong" yaml:"lat_long"` // [latitude, longitude]
			ID      string     `json:"id" yaml:"id"`            // matches anchor ID on page
		} `yaml:"map_points"` // points of interest

		// Graph-specific data.
		Graphs map[string]struct {
			Title  string `json:"title" yaml:"title"` // title displayed in graph
			Points []struct {
				Time  int64   `json:"time" yaml:"time"` // seconds since Unix epoch
				Value float64 `json:"value" yaml:"value"`
			} `json:"points" yaml:"points"`
			Notes []struct {
				Time int64  `json:"time" yaml:"time"` // seconds since Unix epoch
				Text string `json:"text" yaml:"text"` // label for note
			} `json:"notes" yaml:"notes"`
			Range [2]float64 `json:"range" yaml:"range"` // [min, max]
			Units string     `json:"units" yaml:"units"` // units displayed on graph
		} `yaml:"graphs"` // keyed by ID from page
	}
	if err := unmarshalYAML(yb, &data); err != nil {
		return nil, err
	}

	tmpl := newTemplater(filepath.Join(si.TemplateDir()), nil)

	var b bytes.Buffer
	switch {
	case data.Graphs != nil:
		jsonData, err := json.MarshalIndent(data.Graphs, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal data to JSON: %v", err)
		}
		var td = struct {
			CSPMeta       template.HTML
			ScriptURLs    []string
			InlineScripts []template.JS
			InlineStyle   template.CSS
		}{
			ScriptURLs: []string{si.D3ScriptURL},
			InlineScripts: []template.JS{
				template.JS(getStdInline("dark.js")), // used by graph-iframe.js
				template.JS(getStdInline("graph-iframe.js")),
				template.JS("const dataSets = " + string(jsonData) + ";"),
			},
			InlineStyle: template.CSS(getStdInline("graph-iframe.css") + si.ReadInline("graph-iframe.css")),
		}

		csp := cspBuilder{}
		csp.add(cspDefault, cspNone)
		for _, u := range td.ScriptURLs {
			csp.add(cspScript, cspSource(u))
		}
		for _, s := range td.InlineScripts {
			csp.hash(cspScript, string(s))
		}
		csp.hash(cspStyle, string(td.InlineStyle))
		td.CSPMeta = template.HTML(csp.tag())

		if err := tmpl.run(&b, []string{"graph_page.tmpl"}, td, nil); err != nil {
			return nil, err
		}
	case data.MapPoints != nil:
		jsonData, err := json.MarshalIndent(data.MapPoints, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal data to JSON: %v", err)
		}

		// Generate background-image CSS property declarations for the placeholders.
		// The generated page will be in a subdir, so make sure that the placeholder
		// path takes that into account.
		lightImg, err := makeBackgroundImage(&si, data.MapPlaceholderLight, IframeOutDir)
		if err != nil {
			return nil, fmt.Errorf("map_placeholder_light: %v", err)
		}
		darkImg, err := makeBackgroundImage(&si, data.MapPlaceholderDark, IframeOutDir)
		if err != nil {
			return nil, fmt.Errorf("map_placeholder_dark: %v", err)
		}

		var td = struct {
			ScriptURLs    []string
			InlineScripts []template.JS
			BodyScript    template.JS
			InlineStyle   template.CSS
		}{
			ScriptURLs: []string{
				// The no-op callback parameter is needed to avoid a dumb "Loading the Google Maps
				// JavaScript API without a callback is not supported" error:
				// https://stackoverflow.com/a/75212692
				"https://maps.googleapis.com/maps/api/js?key=" + si.GoogleMapsAPIKey + "&callback=Function.prototype",
			},
			InlineScripts: []template.JS{
				template.JS("const points = " + string(jsonData) + ";"),
				template.JS(getStdInline("dark.js")), // used by map-iframe.js
				template.JS(getStdInline("map-iframe.js")),
			},
			BodyScript: template.JS(getStdInline("map-iframe-body.js")),
			InlineStyle: template.CSS(getStdInline("map-iframe.css") + si.ReadInline("map-iframe.css") +
				"body{" + strings.Join(lightImg, ";") + "}" +
				"body.dark{" + strings.Join(darkImg, ";") + "}",
			),
		}
		// Don't use CSP here; Maps API's gonna do whatever it wants.
		if err := tmpl.run(&b, []string{"map_page.tmpl"}, td, nil); err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("unknown iframe type")
	}
	return b.Bytes(), nil
}
