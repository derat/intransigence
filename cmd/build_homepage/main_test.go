// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"syscall"
	"testing"
	"time"

	"github.com/otiai10/copy"
)

const repoPath = "../.." // path from package to the main repo dir

// newTestSiteDir creates a new temporary directory and copies test data
// and inlined files and templates into it.
func newTestSiteDir() (string, error) {
	dir, err := ioutil.TempDir("", "homepage_test.")
	if err != nil {
		return "", err
	}
	done := false
	defer func() {
		if !done {
			os.RemoveAll(dir)
		}
	}()

	if err := copy.Copy("testdata", dir); err != nil {
		return "", fmt.Errorf("failed copying test data: %v", err)
	}
	for _, subdir := range []string{"inline", "templates"} {
		src := filepath.Join(repoPath, subdir)
		dst := filepath.Join(dir, subdir)
		if err := copy.Copy(src, dst); err != nil {
			return "", fmt.Errorf("failed copying %v to %v: %v", src, dst, err)
		}
	}

	done = true // disarm cleanup
	return dir, nil
}

// checkPageContents reads the page at p and checks that its contents are
// matched by all the regular expressions in pats and by none of the ones in negPats.
func checkPageContents(t *testing.T, p string, pats, negPats []string) {
	b, err := ioutil.ReadFile(p)
	if err != nil {
		t.Errorf("Failed reading page: %v", err)
		return
	}
	for _, pat := range pats {
		re := regexp.MustCompile(pat)
		if !re.Match(b) {
			t.Errorf("Page %v not matched by %q", filepath.Base(p), pat)
		}
	}
	for _, pat := range negPats {
		re := regexp.MustCompile(pat)
		if re.Match(b) {
			t.Errorf("Page %v unexpectedly matched by %q", filepath.Base(p), pat)
		}
	}
}

// getFileTimes returns p's mtime and atime.
func getFileTimes(p string) (mtime, atime time.Time, err error) {
	fi, err := os.Stat(p)
	if err != nil {
		return mtime, atime, err
	}
	stat := fi.Sys().(*syscall.Stat_t)
	return fi.ModTime(), time.Unix(int64(stat.Atim.Sec), int64(stat.Atim.Nsec)), nil
}

// getFileContents reads and returns p's contents.
// If p ends in ".gz", the contents are uncompressed before being returned.
func getFileContents(p string) ([]byte, error) {
	b, err := ioutil.ReadFile(p)
	if err != nil {
		return nil, err
	}
	if filepath.Ext(p) == ".gz" {
		r, err := gzip.NewReader(bytes.NewReader(b))
		if err != nil {
			return nil, err
		}
		defer r.Close()
		b, err = ioutil.ReadAll(r)
	}
	return b, err
}

// checkTypes describes checks that should be performed by checkFile.
type checkTypes int

const (
	checkTimes    checkTypes = 1 << iota // compare mtime and atime
	checkContents                        // compare file contents
)

// checkFile compares p against ref.
func checkFile(t *testing.T, p, ref string, ct checkTypes) {
	if ct&checkTimes != 0 {
		if pm, pa, err := getFileTimes(p); err != nil {
			t.Errorf("Couldn't stat out file: %v", err)
		} else if rm, ra, err := getFileTimes(ref); err != nil {
			t.Errorf("Couldn't stat ref file: %v", err)
		} else {
			if !pm.Equal(rm) {
				t.Errorf("%v mtime (%v) doesn't match %v (%v)", p, pm, ref, rm)
			}
			if !pa.Equal(ra) {
				t.Errorf("%v atime (%v) doesn't match %v (%v)", p, pa, ref, ra)
			}
		}
	}
	if ct&checkContents != 0 {
		if pb, err := getFileContents(p); err != nil {
			t.Errorf("Couldn't read out file: %v", err)
		} else if rb, err := getFileContents(ref); err != nil {
			t.Errorf("Couldn't read ref file: %v", err)
		} else if !bytes.Equal(pb, rb) {
			t.Errorf("%v contents don't match %v (%q vs %q)", p, ref, pb, rb)
		}
	}
}

func TestBuildSite(t *testing.T) {
	dir, err := newTestSiteDir()
	if err != nil {
		t.Fatal("Failed creating site dir: ", err)
	}
	if err := buildSite(context.Background(), dir, "", true /* pretty */, true /* validate */); err != nil {
		os.RemoveAll(dir)
		t.Fatal("Failed building site: ", err)
	}

	out := filepath.Join(dir, outSubdir)

	// Perform high-level checks on the index pages.
	checkPageContents(t, filepath.Join(out, "index.html"), []string{
		"^<!DOCTYPE html>\n<html lang=\"en\">\n",
		`<meta charset="utf-8">`,
		`<link rel="amphtml" href="https://www.example.org/index.amp.html">`,
		`"datePublished":\s*"1995-01-01"`,              // struct data from page_info
		`(?s)<title>\s*Index - example.org\s*</title>`, // suffix added
		`Hello`, // from heading
		`This is the index page\.`,
		`Back to top`,
		`Last modified Dec. 31, 1999.`, // from page_info
	}, nil)
	checkPageContents(t, filepath.Join(out, "index.amp.html"), []string{
		"^<!DOCTYPE html>\n<html amp lang=\"en\">\n",
		`<link rel="canonical" href="https://www.example.org/">`,
	}, nil)

	// Check that custom features work.
	checkPageContents(t, filepath.Join(out, "features.html"), []string{
		`(?s)<div class="boxtitle">\s*Features\s*</div>`, // level-1 heading
		`(?s)<title>\s*Features\s*</title>`,              // hide_title_suffix
		// "image" code blocks
		`<figure class="imagebox desktop-right mobile-center custom-class">\s*` +
			`<a href="img_link.html">\s*` +
			`<img src="block_img.png" width="300" height="200" alt="Alt text">\s*` +
			`</a>\s*` +
			`<figcaption>\s*Image caption\s*</figcaption>\s*` +
			`</figure>`,
		// <img-inline>
		`<img class="inline" src="inline_img.jpg" width="48" height="24" alt="Alt text">`,
		`<code class="url">https://code.example.org/</code>`, // <code-url>
		`<span class="clear"></span>`,                        // <clear-floats>
		`<span class="small">small text</span>`,              // <text-size small>
		`<span class="real-small">tiny text</span>`,          // <text-size tiny>
		`(?s)<p>\s*only for non-AMP\s*</p>`,                  // </only-nonamp>
		// !force_amp and !force_nonamp link flags
		`<a href="https://www.google.com/">absolute link</a>`,
		`<a href="bare.html#frag">bare link</a>`,
		`<a href="files/file.txt">text link</a>`,
		`<a href="forced_amp.amp.html#frag">forced AMP link</a>`,
		`<a href="forced_nonamp.html#frag">forced non-AMP link</a>`,
	}, []string{
		`only for amp`,  // <only-amp>
		`Back to top`,   // hide_back_to_top
		`Last modified`, // no modified in page_info
	})
	checkPageContents(t, filepath.Join(out, "features.amp.html"), []string{
		`(?)<p>\s*only for AMP\s*</p>`, // <only-amp>
		// !force_amp and !force_nonamp link flags
		`<a href="https://www.google.com/">absolute link</a>`,
		`<a href="bare.amp.html#frag">bare link</a>`,
		`<a href="https://www.example.org/files/file.txt">text link</a>`,
		`(?s)<a href="forced_amp.amp.html#frag">forced\s+AMP\s+link</a>`,
		`(?s)<a href="forced_nonamp.html#frag">forced\s+non-AMP\s+link</a>`,
	}, []string{
		`only for non-AMP`,
	})

	// TODO: Also pass checkTimes for these after mtimes and atimes are preserved.
	checkFile(t, filepath.Join(out, "images/test.png"), filepath.Join(dir, "static/images/test.png"), checkContents)
	checkFile(t, filepath.Join(out, "other/test.css"), filepath.Join(dir, "static/other/test.css"), checkContents)
	checkFile(t, filepath.Join(out, "more/file.txt"), filepath.Join(dir, "extra/file.txt"), checkContents)

	checkFile(t, filepath.Join(out, "other/test.css.gz"), filepath.Join(out, "other/test.css"), checkTimes&checkContents)
	// Image files shouldn't be compressed.
	p := filepath.Join(out, "images/test.png.gz")
	if _, err := os.Stat(p); err == nil {
		t.Errorf("%v unexpectedly exists", p)
	} else if !os.IsNotExist(err) {
		t.Errorf("Unable to check %v: %v", p, err)
	}

	if t.Failed() {
		fmt.Println("Output is in", out)
	} else {
		os.RemoveAll(dir)
	}
}
