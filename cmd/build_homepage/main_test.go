// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"testing"

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

func TestBuildSite(t *testing.T) {
	dir, err := newTestSiteDir()
	if err != nil {
		t.Fatal("Failed creating site dir: ", err)
	}
	if err := buildSite(dir, "", true /* pprint */); err != nil {
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
		"^<!DOCTYPE html>\n<html amp>\n",
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

	if t.Failed() {
		fmt.Println("Output is in", out)
	} else {
		os.RemoveAll(dir)
	}
}
