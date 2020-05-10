// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package main

import (
	"compress/gzip"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
)

// minifyInline updates the minified versions of CSS or JS files within dir if needed.
// It seems like it'd be cleaner to write the minified files into a temporary dir instead
// of alongside the source files, but yui-compressor is slow (1 second for a 9 KB JS file!)
// so it seems better to skip unnecessary calls.
func minifyInline(dir string) error {
	// filepath.Glob doesn't appear to support globs like "*.{css,js}".
	ps, err := filepath.Glob(filepath.Join(dir, "*.css"))
	if err != nil {
		return err
	} else if js, err := filepath.Glob(filepath.Join(dir, "*.js")); err != nil {
		return err
	} else {
		ps = append(ps, js...)
	}

	for _, p := range ps {
		pi, err := os.Stat(p)
		if err != nil {
			return err
		}
		mp := p + minSuffix
		mi, err := os.Stat(mp)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		// Minify if the minimized version doesn't exist or has an mtime before the original file's.
		if err != nil || mi.ModTime().Before(pi.ModTime()) {
			if err := exec.Command("yui-compressor", "-o", mp, p).Run(); err != nil {
				return err
			}
		}
	}

	return nil
}

// compressDir writes compressed versions of textual files within dir (including in subdirs).
func compressDir(dir string) error {
	return filepath.Walk(dir, func(p string, fi os.FileInfo, err error) error {
		switch filepath.Ext(p) {
		case ".css", ".htm", ".html", "js", ".json", ".txt", ".xml":
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
	// TODO: Necessary to set w.Header with filename and mtime?
	if _, err := io.Copy(w, sf); err != nil {
		return err
	}
	return w.Close()
}

// getAtime returns the atime from fi.
func getAtime(fi os.FileInfo) time.Time {
	// See https://stackoverflow.com/a/20877193.
	stat := fi.Sys().(*syscall.Stat_t)
	return time.Unix(int64(stat.Atim.Sec), int64(stat.Atim.Nsec))
}
