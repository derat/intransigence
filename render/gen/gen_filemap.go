// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s [var] [file]...\n", os.Args[0])
	}

	data := struct {
		Var   string            // name for generated variable
		Hash  string            // hash of file names and contents
		Files map[string]string // keys are base filenames, values are data
	}{
		Var:   os.Args[1],
		Files: make(map[string]string),
	}

	// Keep a running hash of null-terminated file names and base64-encoded contents.
	hash := sha256.New()

	paths := append([]string{}, os.Args[2:]...)
	sort.Strings(paths)
	for _, p := range paths {
		b, err := ioutil.ReadFile(p)
		if err != nil {
			log.Fatal("Failed reading file: ", err)
		}
		name := filepath.Base(p)
		data.Files[name] = string(b)

		hash.Write([]byte(name))
		hash.Write([]byte{0})
		hash.Write([]byte(base64.StdEncoding.EncodeToString(b)))
		hash.Write([]byte{0})
	}

	data.Hash = hex.EncodeToString(hash.Sum(nil))

	if err := template.Must(template.New("").Parse(strings.TrimLeft(`
// Code generated by gen_filemap.go from {{.Hash}}. DO NOT EDIT.

package render

var {{.Var}} = map[string]string{
{{- range $name, $val := .Files}}
	{{printf "%q" $name}}: {{printf "%q" $val}},
{{- end -}}
}
`, "\n"))).Execute(os.Stdout, &data); err != nil {
		log.Fatal("Failed executing template: ", err)
	}
}
