// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package render

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"path/filepath"
)

// IframeOutDir is the subdirectory under the output dir for generated iframe pages.
const IframeOutDir = "iframes"

// Iframe renders and returns the framed page described by the supplied JSON data.
func Iframe(si SiteInfo, jb []byte) ([]byte, error) {
	var data struct {
		Type        string      `json:"type"`        // "graph" or "map"
		Data        interface{} `json:"data"`        // type-specific data
		Placeholder string      `json:"placeholder"` // placeholder image path (just for maps)
	}
	dec := json.NewDecoder(bytes.NewReader(jb))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&data); err != nil {
		return nil, err
	}

	tmpl := newTemplater(filepath.Join(si.TemplateDir()), nil)

	jsonData, err := json.MarshalIndent(data.Data, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data to JSON: %v", err)
	}

	var b bytes.Buffer
	switch data.Type {
	case "graph":
		var td = struct {
			CSPMeta       template.HTML
			ScriptURLs    []string
			InlineScripts []template.JS
			InlineStyle   template.CSS
		}{
			ScriptURLs: []string{si.D3ScriptURL},
			InlineScripts: []template.JS{
				template.JS(getStdInline("graph-iframe.js.min")),
				template.JS("var dataSets = " + string(jsonData) + ";"),
			},
			InlineStyle: template.CSS(getStdInline("graph-iframe.css") + si.ReadInline("graph-iframe.css.min")),
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
	case "map":
		// The generated page will be in a subdir, so make sure that the placeholder path takes
		// that into account.
		if err := si.CheckStatic(filepath.Join(IframeOutDir, data.Placeholder)); err != nil {
			return nil, err
		}
		var td = struct {
			ScriptURLs    []string
			InlineScripts []template.JS
			InlineStyle   template.CSS
		}{
			ScriptURLs: []string{"https://maps.googleapis.com/maps/api/js?key=" + si.GoogleMapsAPIKey},
			InlineScripts: []template.JS{
				template.JS("var points = " + string(jsonData) + ";"),
				template.JS(getStdInline("map-iframe.js.min")),
			},
			InlineStyle: template.CSS(getStdInline("map-iframe.css") + si.ReadInline("map-iframe.css.min") +
				"body{background-image:url('" + data.Placeholder + "')}"),
		}
		// Don't use CSP here; Maps API's gonna do whatever it wants.
		if err := tmpl.run(&b, []string{"map_page.tmpl"}, td, nil); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unknown type %q", data.Type)
	}
	return b.Bytes(), nil
}
