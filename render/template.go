// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package render

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"path/filepath"
)

// Generate an std_templates.go file that defines a map[string]string named stdTemplates.
//go:generate sh -c "go run gen/gen_filemap.go stdTemplates templates/*.tmpl | gofmt -s >std_templates.go"

// templater caches and executes HTML templates.
type templater struct {
	dir         string
	commonFuncs template.FuncMap
	tmpls       map[string]*template.Template
}

// newTemplater returns a templater that will load templates from the supplied directory.
// TODO: Delete dir arg if it's unneeded.
func newTemplater(dir string, commonFuncs template.FuncMap) *templater {
	return &templater{dir, commonFuncs, make(map[string]*template.Template)}
}

// load loads and caches a template consisting of the supplied files and functions.
// files are loaded first from stdTemplates before falling back to actual files in t.dir.
func (t *templater) load(files []string, funcs template.FuncMap) (*template.Template, error) {
	if len(files) == 0 {
		return nil, errors.New("no files supplied")
	}

	name := files[0]
	if tmpl, ok := t.tmpls[name]; ok {
		return tmpl, nil
	}

	fm := template.FuncMap{}
	for n, f := range t.commonFuncs {
		fm[n] = f
	}
	for n, f := range funcs {
		fm[n] = f
	}

	tmpl := template.New(name).Funcs(fm)

	var paths []string
	for _, fn := range files {
		if s, ok := stdTemplates[fn]; ok {
			if _, err := tmpl.Parse(s); err != nil {
				return nil, fmt.Errorf("failed parsing %v: %v", fn, err)
			}
		} else {
			paths = append(paths, filepath.Join(t.dir, fn))
		}
	}
	if len(paths) > 0 {
		if _, err := tmpl.ParseFiles(paths...); err != nil {
			return nil, err
		}
	}

	t.tmpls[name] = tmpl
	return tmpl, nil
}

// run runs a template consisting of the named files using the supplied data and functions,
// plus any functions passed to newTemplater.
func (t *templater) run(w io.Writer, files []string, data interface{}, funcs template.FuncMap) error {
	tmpl, err := t.load(files, funcs)
	if err != nil {
		return err
	}
	return tmpl.Execute(w, data)
}

// runNamed is similar to run but runs the named template.
// See Template.ExecuteTemplate() vs. Template.Execute().
func (t *templater) runNamed(w io.Writer, files []string, name string, data interface{}, funcs template.FuncMap) error {
	tmpl, err := t.load(files, funcs)
	if err != nil {
		return err
	}
	return tmpl.ExecuteTemplate(w, name, data)
}
