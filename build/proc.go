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
	"reflect"
	"sort"
	"syscall"
	"time"
)

// minifyInline updates the minified versions of JS files within dir if needed.
// It seems like it'd be cleaner to write the minified files into a temporary dir instead
// of alongside the source files, but yui-compressor is slow (1 second for a 9 KB JS file!)
// so it seems better to skip unnecessary calls.
// TODO: Delete this? No customizable JS right now.
func minifyInline(dir string) error {
	ps, err := filepath.Glob(filepath.Join(dir, "*.js"))
	if err != nil {
		return err
	}

	defer clearStatus()
	for i, p := range ps {
		statusf("Minifying files: [%d/%d]", i, len(ps))
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
			if err := exec.Command("yui-compressor", "--type", "js", "-o", mp, p).Run(); err != nil {
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

// getAtime returns the atime from fi.
func getAtime(fi os.FileInfo) time.Time {
	// See https://stackoverflow.com/a/20877193.
	stat := fi.Sys().(*syscall.Stat_t)
	return time.Unix(int64(stat.Atim.Sec), int64(stat.Atim.Nsec))
}

// copyTimes examines each file in src. If a file with the same contents exists
// at the same path relative to dst, the source atime and mtime are copied over.
func copyTimes(src, dst string) error {
	return filepath.Walk(src, func(sp string, si os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		dp := filepath.Join(dst, sp[len(src):])
		di, err := os.Stat(dp)
		if os.IsNotExist(err) {
			return nil
		} else if err != nil {
			return err
		}

		if si.Mode() != di.Mode() || si.Size() != di.Size() {
			return nil
		}
		switch {
		case di.Mode()&os.ModeDir != 0: // directory
			if match, err := dirsMatch(sp, dp); err != nil {
				return err
			} else if !match {
				return nil
			}
		case di.Mode()&os.ModeType == 0: // regular file
			if match, err := filesMatch(sp, dp); err != nil {
				return err
			} else if !match {
				return nil
			}
		default:
			// TODO: Handle symlinks?
			return nil
		}

		// If we didn't bail out, update the times to match.
		return os.Chtimes(dp, getAtime(si), si.ModTime())
	})
}

// dirsMatch returns true if the directories at paths a and b have the same contents (i.e. filenames).
func dirsMatch(a, b string) (bool, error) {
	getNames := func(p string) ([]string, error) {
		f, err := os.Open(p)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		names, err := f.Readdirnames(0)
		if err != nil {
			return nil, err
		}
		sort.Strings(names)
		return names, nil
	}

	an, err := getNames(a)
	if err != nil {
		return false, err
	}
	bn, err := getNames(b)
	if err != nil {
		return false, err
	}
	return reflect.DeepEqual(an, bn), nil
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
