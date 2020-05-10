// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package build

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
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
		if err != nil {
			return err
		}
		if fi.Mode()&os.ModeType != 0 {
			return nil
		}
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

// copyFileTimes examines each file in src. If a file with the same contents exists
// at the same path relative to dst, the source atime and mtime are copied over.
func copyFileTimes(src, dst string) error {
	return filepath.Walk(src, func(sp string, si os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if si.Mode()&os.ModeDir != 0 {
			return nil
		}

		// If the dest file doesn't exist or has a different size or is a directory, skip it.
		dp := filepath.Join(dst, sp[len(src)+1:])
		di, err := os.Stat(dp)
		if os.IsNotExist(err) {
			return nil
		} else if err != nil {
			return err
		} else if si.Size() != di.Size() || di.Mode()&os.ModeDir != 0 {
			return nil
		}

		// Also skip the dest file if it doesn't have the same contents as the source file.
		if match, err := filesMatch(sp, dp); err != nil {
			return err
		} else if !match {
			return nil
		}

		return os.Chtimes(dp, getAtime(si), si.ModTime())
	})
}

// filesMatch reads the files at paths a and b and returns true if they have the same contents.
func filesMatch(a, b string) (bool, error) {
	ah, err := hashFile(a)
	if err != nil {
		return false, err
	}
	bh, err := hashFile(b)
	if err != nil {
		return false, err
	}
	return bytes.Equal(ah, bh), nil
}

// hashFile reads the file at p and returns a SHA256 hash of its contents.
func hashFile(p string) ([]byte, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}
