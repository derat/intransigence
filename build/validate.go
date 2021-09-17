// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package build

import (
	"context"
	"errors"
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
}

var ampIgnore []*regexp.Regexp

var cssIgnore = []*regexp.Regexp{
	// AMP's inlined styles use these.
	regexp.MustCompile(`-amp-start.* is an unknown vendor extension`),
	regexp.MustCompile(`^Unrecognized at-rule @-[a-z]+-keyframes$`),
	regexp.MustCompile(`^-[a-z]+-animation is an unknown vendor extension$`),
	// Permit styling scrollbars for WebKit-based browsers.
	regexp.MustCompile(`-webkit-scrollbar is an unknown vendor extended pseudo-element$`),
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
	htmlCh := startValidation(ctx, htmlPaths, validateHTML)
	ampCh := startValidation(ctx, ampPaths, validate.AMP)
	cssCh := startValidation(ctx, cssPaths, validateCSS)

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

// validateFunc validates the supplied data and returns a list of issues.
// The returned error is non-nil if the validation process itself fails.
type validateFunc func(context.Context, io.Reader) ([]validate.Issue, error)

// validateHTML is a validateFunc for running validate.HTML.
func validateHTML(ctx context.Context, r io.Reader) ([]validate.Issue, error) {
	issues, _, err := validate.HTML(ctx, r)
	return issues, err
}

// validateCSS is a validateFunc for running validate.CSS.
func validateCSS(ctx context.Context, r io.Reader) ([]validate.Issue, error) {
	issues, _, err := validate.CSS(ctx, r, validate.HTMLDoc)
	return issues, err
}

// validateResult contains the results of a validation task.
type validateResult struct {
	p      string // file path
	issues []validate.Issue
	err    error
}

// startValidation asynchronously validates the supplied paths.
// Results are streamed to the returned channel.
func startValidation(ctx context.Context, paths []string, vf validateFunc) <-chan validateResult {
	ch := make(chan validateResult, len(paths))
	go func() {
		for _, p := range paths {
			if ctx.Err() != nil {
				break
			}
			issues, err := validateFile(ctx, p, vf)
			ch <- validateResult{p, issues, err}
		}
		// Avoid closing the channel so the select in validateFiles won't get zero values.
	}()
	return ch
}

// validateFile validates the file at the supplied path using vf.
func validateFile(ctx context.Context, p string, vf validateFunc) ([]validate.Issue, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return vf(ctx, f)
}
