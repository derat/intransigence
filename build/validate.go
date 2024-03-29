// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package build

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/derat/intransigence/render"
	"github.com/derat/validate"
)

var htmlIgnore = []*regexp.Regexp{
	// Blackfriday uses align attributes on generated tables.
	regexp.MustCompile(`^The align attribute on the (td|th) element is obsolete. Use CSS instead\.$`),
	// The validator doesn't seem to understand image-set() yet.
	regexp.MustCompile(`background-image: image-set\(.*\) is not a background-image value\.$`),
}

var ampIgnore []*regexp.Regexp

var cssIgnore = []*regexp.Regexp{
	// Ingore e.g. 'foo is a vendor extension' and 'foo is a vendor extended pseudo-element'.
	// AMP uses '-amp-start.*' and '-[a-z]+-animation', and other extensions like -webkit-scrollbar,
	// -ms-high-contrast, etc. are also useful.
	regexp.MustCompile(` is a vendor `),
	// AMP's inlined styles use this.
	regexp.MustCompile(`^Unrecognized at-rule @-[a-z]+-keyframes$`),
	// As of 20210917, the W3C validator appears to be choking on more stuff from inlined styles.
	regexp.MustCompile(`^Parse Error  @(|-moz-|-ms-|-o-)keyframes -amp-start\{from\{visibility:hidden\}to\{visibility:visible\}\}$`),
	// The W3C recommends quoting font families with spaces,
	// but this isn't required: https://stackoverflow.com/a/13752149/6882947
	// The minify package removes quotes and lowercases families:
	// https://github.com/tdewolff/minify#css
	regexp.MustCompile(`^Family names containing whitespace should be quoted\. ` +
		`If quoting is omitted, any whitespace characters before and after the ` +
		`name are ignored and any sequence of whitespace characters inside the ` +
		`name is converted to a single space\.$`),
	// The validator doesn't seem to understand image-set() yet.
	regexp.MustCompile(`background-image image-set\(.*\) is not a background-image value`),
}

// validateFiles validates files at the supplied paths.
func validateFiles(ctx context.Context, paths []string) error {
	// AMP files are rejected by validate.HTML with e.g. "Attribute amp not allowed on
	// element html at this point.", so build up separate lists of files.
	var htmlPaths, ampPaths, cssPaths []string
	var baseDir string // longest common prefix among paths
	for _, p := range paths {
		var err error
		if p, err = filepath.Abs(p); err != nil {
			return err
		}
		if strings.HasSuffix(p, render.AMPExt) {
			ampPaths = append(ampPaths, p)
			cssPaths = append(cssPaths, p)
		} else if strings.HasSuffix(p, render.HTMLExt) {
			htmlPaths = append(htmlPaths, p)
			cssPaths = append(cssPaths, p)
		}
		for pre := filepath.Dir(p); ; pre = filepath.Dir(pre) {
			if baseDir == "" || strings.HasPrefix(baseDir, pre) {
				baseDir = pre
				break
			}
		}
	}

	// Perform validation tasks in parallel.
	htmlCh := startValidation(ctx, htmlPaths, &htmlValidator{})
	ampCh := startValidation(ctx, ampPaths, &ampValidator{})
	cssCh := startValidation(ctx, cssPaths, &cssValidator{})

	var failed bool
	var htmlRes, ampRes, cssRes int
	defer clearStatus()
	for htmlRes < len(htmlPaths) || ampRes < len(ampPaths) || cssRes < len(cssPaths) {
		statusf("Validating pages: HTML [%d/%d], AMP [%d/%d], CSS [%d/%d]",
			htmlRes, len(htmlPaths), ampRes, len(ampPaths), cssRes, len(cssPaths))
		select {
		case res := <-htmlCh:
			failed = !printValidateResult("HTML", baseDir, res, htmlIgnore) || failed
			htmlRes++
		case res := <-ampCh:
			failed = !printValidateResult("AMP", baseDir, res, ampIgnore) || failed
			ampRes++
		case res := <-cssCh:
			failed = !printValidateResult("CSS", baseDir, res, cssIgnore) || failed
			cssRes++
		}
	}
	if failed {
		return errors.New("validation failed")
	}
	return nil
}

// printValidateResult prints issues from res if they are not excluded by a regexp in ignore.
// vname describes the type of validation being performed, e.g. "HTML".
// baseDir contains a common prefix that is removed from file paths.
// Returns true if no problems were found.
func printValidateResult(vname, baseDir string, res validateResult, ignore []*regexp.Regexp) bool {
	good := true
	// Filter out ignored issues.
	var issues []validate.Issue
	for _, is := range res.issues {
		ignored := false
		for _, ig := range ignore {
			if ig.MatchString(is.Message) {
				ignored = true
				break
			}
		}
		if !ignored {
			issues = append(issues, is)
		}
	}
	fn := res.p[len(baseDir)+1:] // strip off common dir plus slash
	if res.err != nil {
		logf("%s: %s validator failed: %v\n", fn, vname, res.err)
		good = false
	}
	for _, is := range issues {
		logf("%s: %s %d:%d %s\n", fn, vname, is.Line, is.Col, is.Message)
		good = false
	}
	return good
}

// streamValidator validates the supplied data and returns a list of issues.
// The returned error is non-nil if the validation process itself fails.
type streamValidator interface {
	validateStream(context.Context, io.Reader) ([]validate.Issue, error)
}

// htmlValidator is a streamValidator for running validate.HTML.
type htmlValidator struct{}

func (v *htmlValidator) validateStream(ctx context.Context, r io.Reader) ([]validate.Issue, error) {
	issues, _, err := validate.HTML(ctx, r)
	return issues, err
}

// cssValidator is a streamValidator for running validate.CSS.
type cssValidator struct{}

func (v *cssValidator) validateStream(ctx context.Context, r io.Reader) ([]validate.Issue, error) {
	issues, _, err := validate.CSS(ctx, r, validate.HTMLDoc)
	return issues, err
}

// filesValidator validates multiple files in parallel.
type filesValidator interface {
	validateFiles(context.Context, []string) (map[string][]validate.Issue, error)
}

// ampValidator is a filesValidator for running validate.AMPFiles.
type ampValidator struct{}

func (v *ampValidator) validateFiles(ctx context.Context, files []string) (map[string][]validate.Issue, error) {
	return validate.AMPFiles(ctx, files)
}

// validateResult contains the results of a validation task.
type validateResult struct {
	p      string // file path
	issues []validate.Issue
	err    error
}

// startValidation asynchronously validates the supplied paths.
// v is a streamValidator or filesValidator.
// Results are streamed to the returned channel.
func startValidation(ctx context.Context, paths []string, v interface{}) <-chan validateResult {
	ch := make(chan validateResult, len(paths))
	go func() {
		switch tv := v.(type) {
		case streamValidator:
			for _, p := range paths {
				if ctx.Err() != nil {
					break
				}
				issues, err := validateFile(ctx, p, tv)
				ch <- validateResult{p, issues, err}
			}
		case filesValidator:
			fileIssues, err := tv.validateFiles(ctx, paths)
			for _, p := range paths {
				ch <- validateResult{p, fileIssues[p], err}
			}
		default:
			panic(fmt.Sprintf("Invalid validator type %T", tv))
		}
		// Avoid closing the channel so the select in validateFiles won't get zero values.
	}()
	return ch
}

// validateFile validates the file at the supplied path using v.
func validateFile(ctx context.Context, p string, v streamValidator) ([]validate.Issue, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return v.validateStream(ctx, f)
}
