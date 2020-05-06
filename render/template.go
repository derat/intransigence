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

// run runs a template consisting of the named files using the supplied
// data and functions, plus any functions passed to newTemplater.
// The template is cached after it's loaded for the first time.
func (t *templater) run(w io.Writer, files []string, data interface{}, funcs template.FuncMap) error {
	if len(files) == 0 {
		return errors.New("no files supplied")
	}
	name := files[0]
	tmpl, ok := t.tmpls[name]
	if !ok {
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

		var err error
		if tmpl, err = template.New(name).Funcs(fm).ParseFiles(paths...); err != nil {
			return err
		}
		t.tmpls[name] = tmpl
	}
	return tmpl.Execute(w, data)
}
