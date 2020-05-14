// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package build

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/otiai10/copy"
	"github.com/pmezard/go-difflib/difflib"
)

const repoPath = ".." // path from package to the main repo dir

func TestBuild_Full(t *testing.T) {
	dir, err := newTestSiteDir()
	if err != nil {
		t.Fatal("Failed creating site dir:", err)
	}
	if err := Build(context.Background(), dir, "", PrettyPrint|Validate); err != nil {
		os.RemoveAll(dir)
		t.Fatal("Build failed:", err)
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
		`Page created in 1995.`,        // from page_info
		`Last modified Dec. 31, 1999.`, // from page_info
	}, nil)
	checkPageContents(t, filepath.Join(out, "index.amp.html"), []string{
		"^<!DOCTYPE html>\n<html amp lang=\"en\">\n",
		`<link rel="canonical" href="https://www.example.org/">`,
	}, nil)

	// Check that custom features work.
	checkPageContents(t, filepath.Join(out, "features.html"), []string{
		`(?s)<div class="title">\s*Features\s*</div>`, // level-1 heading
		`(?s)<title>\s*Features\s*</title>`,           // hide_title_suffix
		// TODO: Also test prefix/suffix.
		`<figure class="desktop-right mobile-center custom-class">\s*` + // "image" code block
			`<a href="img_link.html">\s*` +
			`<img src="images/test.png" width="300" height="200" alt="Alt text">\s*` +
			`</a>\s*` +
			`<figcaption>\s*Image caption\s*</figcaption>\s*` +
			`</figure>`,
		`<div class="clear"></div>`,                                                        // "clear" code block
		`<img class="inline" src="images/test.png" width="48" height="24" alt="Alt text">`, // <image>
		`<code class="url">https://code.example.org/</code>`,                               // <code-url>
		`<span class="small">small text</span>`,                                            // <text-size small>
		`<span class="real-small">tiny text</span>`,                                        // <text-size tiny>
		`(?s)<p>\s*only for non-AMP\s*</p>`,                                                // </only-nonamp>
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

	compareFiles(t, filepath.Join(out, "images/test.png"), filepath.Join(dir, "static/images/test.png"), contentsEqual)
	compareFiles(t, filepath.Join(out, "other/test.css"), filepath.Join(dir, "static/other/test.css"), contentsEqual)
	compareFiles(t, filepath.Join(out, "more/file.txt"), filepath.Join(dir, "extra/file.txt"), contentsEqual)

	// Textual files should be gzipped, and the .gz file should have the same mtime as the original file.
	compareFiles(t, filepath.Join(out, "index.html.gz"), filepath.Join(out, "index.html"), contentsEqual|mtimeEqual)
	compareFiles(t, filepath.Join(out, "features.html.gz"), filepath.Join(out, "features.html"), contentsEqual|mtimeEqual)
	compareFiles(t, filepath.Join(out, "other/test.css.gz"), filepath.Join(out, "other/test.css"), contentsEqual|mtimeEqual)
	checkFileNotExist(t, filepath.Join(out, "images/test.png.gz")) // image files shouldn't be compressed

	// The generated sitemap should list all pages.
	checkFileContents(t, filepath.Join(out, sitemapFile), strings.TrimLeft(`
<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url>
    <loc>https://www.example.org/index.html</loc>
    <changefreq>weekly</changefreq>
  </url>
  <url>
    <loc>https://www.example.org/features.html</loc>
    <changefreq>weekly</changefreq>
  </url>
</urlset>
`, "\n"))

	if t.Failed() {
		fmt.Println("Output is in", out)
	} else {
		os.RemoveAll(dir)
	}
}

func TestBuild_Rebuild(t *testing.T) {
	// Generate the site within the site dir.
	dir, err := newTestSiteDir()
	if err != nil {
		t.Fatal("Failed creating site dir:", err)
	}
	if err := Build(context.Background(), dir, "", PrettyPrint); err != nil {
		os.RemoveAll(dir)
		t.Fatal("Build failed:", err)
	}

	// Add some content to the index page and regenerate the site.
	const newContent = "Here's a newly-added sentence."
	if err := appendToFile(filepath.Join(dir, "pages/index.md"), "\n"+newContent+"\n"); err != nil {
		os.RemoveAll(dir)
		t.Fatal("Failed appending content:", err)
	}
	if err := Build(context.Background(), dir, "", PrettyPrint); err != nil {
		os.RemoveAll(dir)
		t.Fatal("Build failed on rebuild:", err)
	}

	out := filepath.Join(dir, outSubdir)
	oldOut := filepath.Join(dir, oldOutSubdir)

	// Unchanged files' and directories' mtimes should be copied over from the first build.
	compareFiles(t, filepath.Join(out, "features.html"), filepath.Join(oldOut, "features.html"), contentsEqual|mtimeEqual)
	compareFiles(t, filepath.Join(out, "features.html.gz"), filepath.Join(oldOut, "features.html.gz"), contentsEqual|mtimeEqual)
	compareFiles(t, filepath.Join(out, "images"), filepath.Join(oldOut, "images"), mtimeEqual)
	compareFiles(t, filepath.Join(out, "images/test.png"), filepath.Join(oldOut, "images/test.png"), contentsEqual|mtimeEqual)
	compareFiles(t, filepath.Join(out, "more"), filepath.Join(oldOut, "more"), mtimeEqual)
	compareFiles(t, filepath.Join(out, "more/file.txt"), filepath.Join(oldOut, "more/file.txt"), contentsEqual|mtimeEqual)
	compareFiles(t, out, oldOut, mtimeEqual)

	// The new index file should have the newly-added content and an updated mtime.
	checkPageContents(t, filepath.Join(out, "index.html"), []string{regexp.QuoteMeta(newContent)}, nil)
	compareFiles(t, filepath.Join(out, "index.html"), filepath.Join(oldOut, "index.html"), mtimeAfter)
	compareFiles(t, filepath.Join(out, "index.html"), filepath.Join(out, "index.html.gz"), contentsEqual|mtimeEqual)

	if t.Failed() {
		fmt.Println("Output is in", out)
	} else {
		os.RemoveAll(dir)
	}
}

// newTestSiteDir creates a new temporary directory and copies test data
// and inlined files and templates into it.
func newTestSiteDir() (string, error) {
	dir, err := ioutil.TempDir("", "build_test.")
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
	for _, subdir := range []string{"inline", "static/resources", "templates"} {
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
func getFileTimes(p string) (atime, mtime time.Time, err error) {
	fi, err := os.Stat(p)
	if err != nil {
		return mtime, atime, err
	}
	return getAtime(fi), fi.ModTime(), nil
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

// checkFileContents checks that the contents of the file at path p exactly match want.
// If the file doesn't match, a failure containing a diff is reported.
func checkFileContents(t *testing.T, p, want string) {
	b, err := getFileContents(p)
	if err != nil {
		t.Error("Failed to get file contents:", err)
		return
	}
	got := string(b)
	if got != want {
		diff, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
			A:        difflib.SplitLines(want),
			FromFile: "want",
			B:        difflib.SplitLines(got),
			ToFile:   "got",
		})
		t.Errorf("%v doesn't have expected contents:\n%s", p, diff)
	}
}

// compareTypes describes checks that should be performed by compareFiles.
type compareTypes int

const (
	mtimeEqual    compareTypes = 1 << iota // mtimes are equal
	mtimeAfter                             // p's mtime is later than ref
	contentsEqual                          // file contents are equal
)

// compareFiles compares p against ref.
func compareFiles(t *testing.T, p, ref string, ct compareTypes) {
	if ct&(mtimeEqual|mtimeAfter) != 0 {
		// Don't compare atime, since it gets updated when the file is read and isn't used by rsync.
		if _, pm, err := getFileTimes(p); err != nil {
			t.Errorf("Couldn't stat out file: %v", err)
		} else if _, rm, err := getFileTimes(ref); err != nil {
			t.Errorf("Couldn't stat ref file: %v", err)
		} else if ct&mtimeEqual != 0 && !pm.Equal(rm) {
			t.Errorf("%v mtime (%v) doesn't match %v (%v)", p, pm, ref, rm)
		} else if ct&mtimeAfter != 0 && !pm.After(rm) {
			t.Errorf("%v mtime (%v) not after %v (%v)", p, pm, ref, rm)
		}
	}
	if ct&contentsEqual != 0 {
		if pb, err := getFileContents(p); err != nil {
			t.Errorf("Couldn't read out file: %v", err)
		} else if rb, err := getFileContents(ref); err != nil {
			t.Errorf("Couldn't read ref file: %v", err)
		} else if !bytes.Equal(pb, rb) {
			t.Errorf("%v contents don't match %v (%q vs %q)", p, ref, pb, rb)
		}
	}
}

// checkFileNotExist checks that the file at path p doesn't exist.
func checkFileNotExist(t *testing.T, p string) {
	if _, err := os.Stat(p); os.IsNotExist(err) {
		return
	} else if err != nil {
		t.Error("Couldn't stat file: ", err)
	} else {
		t.Error(p, " unexpectedly exists")
	}
}

// appendToFile appends data to the file at path p.
func appendToFile(p, data string) error {
	f, err := os.OpenFile(p, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	_, err = f.Write([]byte(data))
	if err != nil {
		f.Close()
		return err
	}
	return f.Close()
}
