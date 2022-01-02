// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package build

import (
	"bytes"
	"compress/gzip"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/derat/intransigence/render"

	"github.com/otiai10/copy"
	"github.com/pmezard/go-difflib/difflib"
)

var validateTest bool

func init() {
	// Make it easy to skip validation when iterating to try to fix failures.
	flag.BoolVar(&validateTest, "validate", true, "Validate generated pages")
}

func TestBuild_Full(t *testing.T) {
	dir, err := newTestSiteDir()
	if err != nil {
		t.Fatal("Failed creating site dir:", err)
	}
	flags := PrettyPrint
	if validateTest {
		flags |= Validate
	}
	if err := Build(context.Background(), dir, "", flags); err != nil {
		os.RemoveAll(dir)
		t.Fatal("Build failed:", err)
	}

	out := filepath.Join(dir, outSubdir)

	// Perform high-level checks on the index pages.
	checkPageContents(t, filepath.Join(out, "index.html"), []string{
		"^<!DOCTYPE html>\n<html lang=\"en\">\n",
		`<meta charset="utf-8">`,
		`<link rel="amphtml"\s+href="https://www.example.org/index.amp.html">`,
		`<link rel="alternate"\s+type="application/atom\+xml"\s+href="https://www\.example\.org/atom\.xml">`,
		`(?s)<title>\s*Welcome\s+to\s+an\s+example\s+site\s*</title>`, // hide_title_suffix
		`<link rel="icon" sizes="192x192" href="resources/favicon\.png">`,
		`<script type="application/ld\+json">` + // structured data
			`{"@context":"http://schema\.org","@type":"Article",` +
			`"mainEntityOfPage":"https://www\.example\.org/",` +
			`"headline":"Welcome to an example site",` +
			`"description":"Provides a high-level overview of features",` +
			`"dateModified":"2020-05-20","datePublished":"2020-05-19",` +
			`"author":{"@type":"Person","name":"Site Author","email":"user@example\.org"},` +
			`"publisher":{"@type":"Organization","name":"example\.org","url":"https://www\.example\.org/",` +
			`"logo":{"@type":"ImageObject","url":"https://www\.example\.org/resources/logo-460\.png","width":460,"height":460}},` +
			`"image":{"@type":"ImageObject","url":"https://www\.example\.org/scottish_fold/maru-800\.jpg","width":800,"height":500}}` +
			`</script>`,
		`<meta name="description" content="Provides a high-level overview of features">`, // desc
		`Welcome`, // from heading
		`This is the site's landing page\.`,
	}, []string{
		`class="collapsed-mobile"`, // navbox shouldn't be collapsed for index
		`Cheshire`,                 // omit_from_menu
		`cheshire.html`,            // omit_from_menu
		`Back to top`,              // hide_back_to_top
		`Last modified`,            // hide_dates
	})
	checkPageContents(t, filepath.Join(out, "index.amp.html"), []string{
		"^<!DOCTYPE html>\n<html amp lang=\"en\">\n",
		`<link rel="canonical"\s+href="https://www.example.org/">`,
		`<link rel="alternate"\s+type="application/atom\+xml"\s+href="https://www\.example\.org/atom\.xml">`,
	}, nil)

	// scottish_fold.html demonstrates a large number of custom features.
	checkPageContents(t, filepath.Join(out, "scottish_fold.html"), []string{
		`(?s)<title>\s*Scottish Fold\s+-\s+example.org\s*</title>`, // suffix added to title
		`<li><span\s+class="selected">Scottish\s+Fold</span>`,      // nav item selected
		`(?s)<div class="title">\s*Scottish\s+Fold\s*</div>`,       // box created by level-1 heading
		// "image" code block
		`<figure class="desktop-left mobile-center custom-class">\s*` +
			`<a href="scottish_fold/maru-800\.jpg">` +
			`<span class="img-wrapper">` +
			// placeholder
			`<svg width="400"\s+height="250">\s*` +
			`<filter\s+id="thumb-filter">\s*` +
			`<feGaussianBlur[^>]*></feGaussianBlur>\s*` +
			`<feComponentTransfer>\s*<feFuncA[^>]*></feFuncA>\s*</feComponentTransfer>\s*` +
			`</filter>\s*` +
			`<image\s+href="data:image/gif;base64,[^"]+"\s+width="400"\s+height="250"\s+` +
			`filter="url\(#thumb-filter\)"\s+preserveAspectRatio="none"></image>\s*` +
			`</svg>` +
			// real image
			`<picture>\s*` +
			`<source\s+type="image/webp"\s+sizes="400px"\s+` +
			`srcset="scottish_fold/maru-400\.webp 400w, scottish_fold/maru-800\.webp 800w">\s*` +
			`<img\s+src="scottish_fold/maru-400\.jpg"\s+sizes="400px"\s+` +
			`srcset="scottish_fold/maru-400\.jpg 400w, scottish_fold/maru-800\.jpg 800w"\s+` +
			`width="400"\s+height="250"\s+alt="Maru the cat sitting in a small cardboard box">\s*` +
			`</picture></span></a>\s*` +
			`<figcaption>\s*Maru\s*</figcaption>\s*` +
			`</figure>`,
		`<div class="clear"></div>`, // "clear" code block
		// inline <image>
		`<picture>\s*` +
			`<source\s+type="image/webp"\s+sizes="61px"\s+srcset="scottish_fold/nyan\.webp 61w">\s*` +
			`<img\s+class="inline"\s+src="scottish_fold/nyan\.gif"\s+sizes="61px"\s+` +
			`srcset="scottish_fold/nyan\.gif 61w"\s+width="61"\s+height="24"\s+alt="Nyan Cat">\s*` +
			`</picture>`,
		// "image" code block for WebP image
		`<figure>\s*` +
			`<span\s+class="img-wrapper">` +
			// placeholder
			`<svg width="400"\s+height="300">\s*` +
			`<image\s+href="data:image/gif;base64,[^"]+"\s+width="400"\s+height="300"\s+` +
			`filter="url\(#thumb-filter\)"\s+preserveAspectRatio="none"></image>\s*` +
			`</svg>` +
			// real image
			`<picture>\s*` +
			`<img\s+src="scottish_fold/christmas\.webp"\s+sizes="400px"\s+` +
			`srcset="scottish_fold/christmas.webp 400w"\s+` +
			`width="400"\s+height="300"\s+alt="Scottish Fold cat under a Christmas tree">\s*` +
			`</picture>` +
			`</span>\s*` +
			`</figure>`,
		`(?s)viewing\s+the\s+non-AMP\s+version`,                       // <only-nonamp>
		`<a href="scottish_fold\.amp\.html">AMP\s+version</a>`,        // !force_amp
		`<span class="small">makes\s+text\s+small</span>`,             // <text-size small>
		`<span class="real-small">makes\s+it\s+even\s+smaller</span>`, // <text-size tiny>
		`<code class="url">https://www.example.org/a/quite/long/url` + // <code-url>
			`/that/should/be/wrapped</code>`,
		`Text can also be <span class="no-select">marked as ` + // ‹...›
			`non-selectable</span> within a code block`,
		`<a href="#top">Back\s+to\s+top</a>`,
		`Page created in 2020\.`,
		`Last modified May 21, 2020\.`,
	}, []string{
		`class="collapsed-mobile"`,          // navbox shouldn't be collapsed due to children
		`Cheshire`,                          // omit_from_menu
		`cheshire.html`,                     // omit_from_menu
		`(?s)viewing\s+the\s+AMP\s+version`, // <only-amp>
	})

	// Check AMP-specific markup in the AMP version of the page.
	checkPageContents(t, filepath.Join(out, "scottish_fold.amp.html"), []string{
		// "image" code block
		`<figure class="desktop-left mobile-center custom-class">\s*` +
			`<a href="https://www\.example\.org/scottish_fold/maru-800\.jpg">` +
			// preferred
			`<amp-img\s+layout="responsive"\s+src="scottish_fold/maru-400\.webp"\s+sizes="400px"\s+` +
			`srcset="scottish_fold/maru-400\.webp 400w, scottish_fold/maru-800\.webp 800w"\s+` +
			`width="400"\s+height="250"\s+alt="Maru the cat sitting in a small cardboard box">\s*` +
			// fallback
			`<amp-img\s+fallback\s+layout="responsive"\s+src="scottish_fold/maru-400\.jpg"\s+sizes="400px"\s+` +
			`srcset="scottish_fold/maru-400\.jpg 400w, scottish_fold/maru-800\.jpg 800w"\s+` +
			`width="400"\s+height="250"\s+alt="Maru the cat sitting in a small cardboard box"></amp-img>\s*` +
			// placeholder
			`<amp-img\s+placeholder\s+layout="responsive"\s+class="thumb"\s+` +
			`src="data:image/gif;base64,[^"]+"\s+` +
			`width="400"\s+height="250"\s+alt="Maru the cat sitting in a small cardboard box"></amp-img>\s*` +
			`</amp-img>` +
			`</a>\s*` +
			`<figcaption>\s*Maru\s*</figcaption>\s*` +
			`</figure>`,
		// inline <image>
		`<amp-img\s+layout="fixed"\s+class="inline"\s+src="scottish_fold/nyan\.webp"\s+sizes="61px"\s+` +
			`srcset="scottish_fold/nyan\.webp 61w"\s+width="61"\s+height="24"\s+alt="Nyan Cat">\s*` +
			// fallback
			`<amp-img\s+fallback\s+layout="fixed"\s+class="inline"\s+src="scottish_fold/nyan\.gif"\s+sizes="61px"\s+` +
			`srcset="scottish_fold/nyan\.gif 61w"\s+width="61"\s+height="24"\s+alt="Nyan Cat"></amp-img>\s*` +
			// (too small for placeholder)
			`</amp-img>`,
		// "image" code block for WebP image
		`<figure>\s*` +
			`<amp-img\s+layout="responsive"\s+src="scottish_fold/christmas\.webp"\s+sizes="400px"\s+` +
			`srcset="scottish_fold/christmas\.webp 400w"\s+width="400"\s+height="300"\s+` +
			`alt="Scottish Fold cat under a Christmas tree">\s*` +
			// fallback
			`<amp-img\s+placeholder\s+layout="responsive"\s+class="thumb"\s+` +
			`src="data:image/gif;base64,[^"]+"\s+` +
			`width="400"\s+height="300"\s+alt="Scottish Fold cat under a Christmas tree">\s*</amp-img>\s*` +
			`</amp-img>\s*` +
			`</figure>`,
		`(?s)viewing\s+the\s+AMP\s+version`,                                             // <only-amp>
		`<a href="https://www\.example\.org/scottish_fold\.html">non-AMP\s+version</a>`, // !force_nonamp
	}, []string{
		`(?s)viewing\s+the\s+non-AMP\s+version`, // <only-nonamp>
	})

	checkPageContents(t, filepath.Join(out, "cheshire.html"), []string{}, []string{
		`Scottish fold`,      // omit_from_menu (don't expand parent)
		`scottish_fold.html`, // omit_from_menu (don't expand parent)
		`<li><span\s+class="selected">Cheshire</span>`, // omit_from_menu
	})

	// Static data should be copied into the output directory.
	compareFiles(t, filepath.Join(out, "static.html"), filepath.Join(dir, "static/static.html"), contentsEqual)
	compareFiles(t, filepath.Join(out, "favicon.ico"), filepath.Join(dir, "static/favicon.ico"), contentsEqual)
	compareFiles(t, filepath.Join(out, "scottish_fold/nyan.gif"), filepath.Join(dir, "static/scottish_fold/nyan.gif"), contentsEqual)
	compareFiles(t, filepath.Join(out, "other/extra.html"), filepath.Join(dir, "extra/extra.html"), contentsEqual)

	// Textual files should be gzipped, and the .gz file should have the same mtime as the original file.
	compareFiles(t, filepath.Join(out, "index.html.gz"), filepath.Join(out, "index.html"), contentsEqual|mtimeEqual)
	compareFiles(t, filepath.Join(out, "cats.html.gz"), filepath.Join(out, "cats.html"), contentsEqual|mtimeEqual)
	compareFiles(t, filepath.Join(out, "scottish_fold.html.gz"), filepath.Join(out, "scottish_fold.html"), contentsEqual|mtimeEqual)
	compareFiles(t, filepath.Join(out, "static.html.gz"), filepath.Join(out, "static.html"), contentsEqual|mtimeEqual)
	compareFiles(t, filepath.Join(out, "other/extra.html.gz"), filepath.Join(out, "other/extra.html"), contentsEqual|mtimeEqual)

	// Image files shouldn't be compressed.
	checkFileNotExist(t, filepath.Join(out, "favicon.ico.gz"))
	checkFileNotExist(t, filepath.Join(out, "scottish_fold/nyan.gif.gz"))

	// The generated sitemap should list all pages.
	checkFileContents(t, filepath.Join(out, sitemapFile), strings.TrimLeft(`
<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url>
    <loc>https://www.example.org/</loc>
    <changefreq>weekly</changefreq>
  </url>
  <url>
    <loc>https://www.example.org/cats.html</loc>
    <changefreq>weekly</changefreq>
  </url>
  <url>
    <loc>https://www.example.org/scottish_fold.html</loc>
    <changefreq>weekly</changefreq>
  </url>
  <url>
    <loc>https://www.example.org/cheshire.html</loc>
    <changefreq>weekly</changefreq>
  </url>
</urlset>
`, "\n"))

	checkFileContents(t, filepath.Join(out, render.FeedFile), strings.TrimLeft(`
<?xml version="1.0" encoding="UTF-8"?><feed xmlns="http://www.w3.org/2005/Atom">
  <title>example.org</title>
  <id>https://www.example.org/</id>
  <updated>2021-09-07T00:00:00Z</updated>
  <subtitle>New pages on example.org</subtitle>
  <link href="https://www.example.org/"></link>
  <author>
    <name>Site Author</name>
    <email>user@example.org</email>
  </author>
  <entry>
    <title>Cheshire Cat</title>
    <updated>2021-09-07T00:00:00Z</updated>
    <id>tag:www.example.org,2021-09-07:/cheshire.html</id>
    <link href="https://www.example.org/cheshire.html" rel="alternate"></link>
    <summary type="html">Example site</summary>
    <author>
      <name>Site Author</name>
      <email>user@example.org</email>
    </author>
  </entry>
  <entry>
    <title>Cats</title>
    <updated>2020-05-20T00:00:00Z</updated>
    <id>tag:www.example.org,2020-05-20:/cats.html</id>
    <link href="https://www.example.org/cats.html" rel="alternate"></link>
    <summary type="html">Example site</summary>
    <author>
      <name>Site Author</name>
      <email>user@example.org</email>
    </author>
  </entry>
  <entry>
    <title>Scottish Fold</title>
    <updated>2020-05-20T00:00:00Z</updated>
    <id>tag:www.example.org,2020-05-20:/scottish_fold.html</id>
    <link href="https://www.example.org/scottish_fold.html" rel="alternate"></link>
    <summary type="html">Example site</summary>
    <author>
      <name>Site Author</name>
      <email>user@example.org</email>
    </author>
  </entry>
</feed>`, "\n"))

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
	compareFiles(t, out, oldOut, mtimeEqual)
	compareFiles(t, filepath.Join(out, "cats.html"), filepath.Join(oldOut, "cats.html"), contentsEqual|mtimeEqual)
	compareFiles(t, filepath.Join(out, "cats.html.gz"), filepath.Join(oldOut, "cats.html.gz"), contentsEqual|mtimeEqual)
	compareFiles(t, filepath.Join(out, "scottish_fold"), filepath.Join(oldOut, "scottish_fold"), mtimeEqual)
	compareFiles(t, filepath.Join(out, "scottish_fold/maru-400.jpg"),
		filepath.Join(oldOut, "scottish_fold/maru-400.jpg"), contentsEqual|mtimeEqual)
	compareFiles(t, filepath.Join(out, "static.html"), filepath.Join(oldOut, "static.html"), contentsEqual|mtimeEqual)
	compareFiles(t, filepath.Join(out, "other/extra.html"), filepath.Join(oldOut, "other/extra.html"), contentsEqual|mtimeEqual)
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

// newTestSiteDir creates a new temporary directory and copies test data into it.
func newTestSiteDir() (string, error) {
	dir, err := ioutil.TempDir("", "build_test.")
	if err != nil {
		return "", err
	}
	const src = "../example"
	if err := copy.Copy(src, dir, copy.Options{
		// Make sure that we don't pick up any random output that happened to be
		// in the example directory.
		Skip: func(p string) bool {
			return strings.HasPrefix(p, src+"/out") || strings.HasPrefix(p, src+"/.out.")
		},
	}); err != nil {
		os.RemoveAll(dir)
		return "", fmt.Errorf("failed copying test data: %v", err)
	}
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
			t.Errorf("Page %v not matched by %v", filepath.Base(p), pat)
		}
	}
	for _, pat := range negPats {
		re := regexp.MustCompile(pat)
		if re.Match(b) {
			t.Errorf("Page %v unexpectedly matched by %v", filepath.Base(p), pat)
		}
	}
}

// getFileMtime returns p's mtime.
// If p is a .gz file, it additionally checks that the gzip header does not contain an mtime.
func getFileMtime(p string) (time.Time, error) {
	fi, err := os.Stat(p)
	if err != nil {
		return time.Time{}, err
	}

	if filepath.Ext(p) == ".gz" {
		f, err := os.Open(p)
		if err != nil {
			return time.Time{}, err
		}
		defer f.Close()

		r, err := gzip.NewReader(f)
		if err != nil {
			return time.Time{}, err
		}
		defer r.Close()

		if !r.Header.ModTime.IsZero() {
			return time.Time{}, fmt.Errorf("%v header contains non-zero mtime %v", p, r.Header.ModTime)
		}
	}

	return fi.ModTime(), nil
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
		if pm, err := getFileMtime(p); err != nil {
			t.Errorf("Couldn't stat out file: %v", err)
		} else if rm, err := getFileMtime(ref); err != nil {
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
