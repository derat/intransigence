// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package render

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"sort"
	"testing"
)

var hashRegexp = regexp.MustCompile(` from (.+)\. DO NOT EDIT\.`)

// checkGenerated returns an error if the hash in file doesn't match the globbed files.
// This needs to match the hashing logic in gen/gen_filemap.go.
func checkGenerated(file string, globs []string) error {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	ms := hashRegexp.FindSubmatch(b)
	if len(ms) != 2 {
		return errors.New("failed to find hash in generated file")
	}
	fileHash := string(ms[1])

	var paths []string
	for _, glob := range globs {
		ps, err := filepath.Glob(glob)
		if err != nil {
			return err
		} else if len(ps) == 0 {
			return fmt.Errorf("no files matched by %q", glob)
		}
		paths = append(paths, ps...)
	}

	hash := sha256.New()
	sort.Strings(paths)
	for _, p := range paths {
		b, err := ioutil.ReadFile(p)
		if err != nil {
			return err
		}
		hash.Write([]byte(filepath.Base(p)))
		hash.Write([]byte{0})
		hash.Write([]byte(base64.StdEncoding.EncodeToString(b)))
		hash.Write([]byte{0})
	}
	globHash := hex.EncodeToString(hash.Sum(nil))

	if fileHash != globHash {
		return fmt.Errorf("hash in %v (%v) doesn't match source hash (%v)", file, fileHash, globHash)
	}
	return nil
}

func TestStdInline(t *testing.T) {
	// TODO: Figure out some way to check that gen_css.sh has been run.
	if err := checkGenerated("std_inline.go", []string{"inline/*.css", "inline/*.js"}); err != nil {
		t.Error("Checking generated inline files failed:", err)
	}
}

func TestStdTemplates(t *testing.T) {
	if err := checkGenerated("std_templates.go", []string{"templates/*.tmpl"}); err != nil {
		t.Error("Checking generated templates failed:", err)
	}
}
