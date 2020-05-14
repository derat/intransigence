// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package render

import (
	"errors"
	"html/template"
	"io"
	"path/filepath"
)

// templater caches and executes HTML templates.
type templater struct {
	dir         string
	commonFuncs template.FuncMap
	tmpls       map[string]*template.Template
}

// newTemplater returns a templater that will load templates from the supplied directory.
func newTemplater(dir string, commonFuncs template.FuncMap) *templater {
	return &templater{dir, commonFuncs, make(map[string]*template.Template)}
}

// load loads and caches a template consisting of the supplied files and functions.
func (t *templater) load(files []string, funcs template.FuncMap) (*template.Template, error) {
	if len(files) == 0 {
		return nil, errors.New("no files supplied")
	}

	name := files[0]
	if tmpl, ok := t.tmpls[name]; ok {
		return tmpl, nil
	}

	var paths []string
	for _, fn := range files {
		paths = append(paths, filepath.Join(t.dir, fn))
	}

	fm := template.FuncMap{}
	for n, f := range t.commonFuncs {
		fm[n] = f
	}
	for n, f := range funcs {
		fm[n] = f
	}

	tmpl, err := template.New(name).Funcs(fm).ParseFiles(paths...)
	if err != nil {
		return nil, err
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
