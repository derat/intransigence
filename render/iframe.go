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

const (
	d3URL      = baseURL + "resources/d3.v3.min.js"
	mapsAPIKey = "AIzaSyBvPAIQSEuw7hTR6EOy-v9IhJv0otYtsP0"
)

func Iframe(js []byte, dir string) ([]byte, error) {
	var data struct {
		Type        string      `json:"type"`        // "graph" or "map"
		Data        interface{} `json:"data"`        // type-specific data
		Placeholder string      `json:"placeholder"` // placeholder image URL (just for maps)
	}
	if err := json.Unmarshal(js, &data); err != nil {
		return nil, err
	}

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
			ScriptURLs: []string{d3URL},
			InlineScripts: []template.JS{
				template.JS(readInline("graph-iframe.js.min")),
				template.JS("var dataSets = " + string(jsonData) + ";"),
			},
			InlineStyle: template.CSS(readInline("graph-iframe.css.min")),
		}

		csp := cspHasher{}
		csp.add(cspDefault, cspNone)
		for _, u := range td.ScriptURLs {
			csp.add(cspScript, cspSource(u))
		}
		for _, s := range td.InlineScripts {
			csp.hash(cspScript, string(s))
		}
		csp.hash(cspStyle, string(td.InlineStyle))
		td.CSPMeta = template.HTML(csp.tag())

		tp := filepath.Join(dir, templateDir, "graph_page.tmpl")
		tmpl := template.Must(template.ParseFiles(tp))
		if err := tmpl.Execute(&b, td); err != nil {
			return nil, err
		}
	case "map":
		// TODO: Replicate check_static_file placeholder check.
		var td = struct {
			ScriptURLs    []string
			InlineScripts []template.JS
			InlineStyle   template.CSS
		}{
			ScriptURLs: []string{"https://maps.googleapis.com/maps/api/js?key=" + mapsAPIKey},
			InlineScripts: []template.JS{
				template.JS("var points = " + string(jsonData) + ";"),
				template.JS(readInline("map-iframe.js.min")),
			},
			InlineStyle: template.CSS(readInline("map-iframe.css.min") +
				"body{background-image:url('" + data.Placeholder + "')}"),
		}

		// Don't use CSP here; Maps API's gonna do whatever it wants.
		tp := filepath.Join(dir, templateDir, "map_page.tmpl")
		tmpl := template.Must(template.ParseFiles(tp))
		if err := tmpl.Execute(&b, td); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unknown type %q", data.Type)
	}
	return b.Bytes(), nil
}
