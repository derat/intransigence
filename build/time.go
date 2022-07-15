// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package build

import (
	"os"
	"path/filepath"
	"syscall"
	"time"
)

// getAtime returns the atime from fi.
func getAtime(fi os.FileInfo) time.Time {
	// See https://stackoverflow.com/a/20877193.
	stat := fi.Sys().(*syscall.Stat_t)
	return time.Unix(int64(stat.Atim.Sec), int64(stat.Atim.Nsec))
}

// maxTime returns the max of the supplied times.
func maxTime(t0 time.Time, ts ...time.Time) time.Time {
	max := t0
	for _, t := range ts {
		if t.After(max) {
			max = t
		}
	}
	return max
}

// copyTimes copies the first path's atime and mtime to the second path.
func copyTimes(from, to string) error {
	fi, err := os.Stat(from)
	if err != nil {
		return err
	}
	return os.Chtimes(to, getAtime(fi), fi.ModTime())
}

// updateDirTimes recursively updates dir and its subdirectories
// to have the max atime and mtime of their contents.
func updateDirTimes(dir string) (atime, mtime time.Time, err error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return atime, mtime, err
	}
	for _, e := range entries {
		fi, err := e.Info()
		if err != nil {
			return atime, mtime, err
		}
		var at, mt time.Time
		if fi.IsDir() {
			if at, mt, err = updateDirTimes(filepath.Join(dir, fi.Name())); err != nil {
				return atime, mtime, err
			}
		} else {
			at, mt = getAtime(fi), fi.ModTime()
		}
		atime = maxTime(atime, at)
		mtime = maxTime(mtime, mt)
	}
	if !atime.IsZero() && !mtime.IsZero() {
		err = os.Chtimes(dir, atime, mtime)
	}
	return atime, mtime, err

}
