// Copyright 2022 Daniel Erat.
// All rights reserved.

package build

import (
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
)

// compressDir writes compressed versions of textual files within dir (including in subdirs).
func compressDir(dir string) error {
	return filepath.Walk(dir, func(p string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.Mode()&os.ModeType != 0 {
			return nil
		}
		switch filepath.Ext(p) {
		case ".css", ".htm", ".html", ".js", ".json", ".txt", ".xml":
			dp := p + ".gz"
			if err := compressFile(p, dp); err != nil {
				return err
			}
			// Give the .gz file the same mtime and atime as the original file so rsync can skip it.
			return os.Chtimes(dp, getAtime(fi), fi.ModTime())
		default:
			return nil
		}
	})
}

// compressFile writes a gzipped version of sp at dp.
// The modification time in the gzip header is unset.
func compressFile(sp, dp string) error {
	sf, err := os.Open(sp)
	if err != nil {
		return err
	}
	defer sf.Close()

	df, err := os.Create(dp)
	if err != nil {
		return err
	}
	defer df.Close()

	w := gzip.NewWriter(df)
	if _, err := io.Copy(w, sf); err != nil {
		return err
	}
	return w.Close()
}
