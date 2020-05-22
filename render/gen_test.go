// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package render

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// checkGenerated returns an error if file's modification time precedes that of any top-level file in dir.
func checkGenerated(file, dir string) error {
	fi, err := os.Stat(file)
	if err != nil {
		return err
	}

	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()

	dfis, err := d.Readdir(0)
	if err != nil {
		return err
	}
	for _, dfi := range dfis {
		if fi.ModTime().Before(dfi.ModTime()) {
			return fmt.Errorf("%v is older than %v", file, filepath.Join(dir, dfi.Name()))
		}
	}
	return nil
}

func TestStdInline(t *testing.T) {
	if err := checkGenerated("std_inline.go", "inline"); err != nil {
		t.Error("Checking generated inline files failed: ", err)
	}
}

func TestStdTemplates(t *testing.T) {
	if err := checkGenerated("std_templates.go", "templates"); err != nil {
		t.Error("Checking generated templates failed: ", err)
	}
}
