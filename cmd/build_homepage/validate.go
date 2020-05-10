// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/derat/homepage/render"
	"github.com/derat/validate"
)

// validateDir validates all files within dir (including in subdirs).
func validateDir(ctx context.Context, dir string) error {
	// AMP files are rejected by validate.HTML with e.g. "Attribute amp not allowed on
	// element html at this point.", so build up separate lists of files.
	var htmlPaths, ampPaths, cssPaths []string
	if err := filepath.Walk(dir, func(p string, _ os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(p, render.AMPExt) {
			ampPaths = append(ampPaths, p)
			cssPaths = append(cssPaths, p)
		} else if strings.HasSuffix(p, render.HTMLExt) {
			htmlPaths = append(htmlPaths, p)
			cssPaths = append(cssPaths, p)
		}
		return nil
	}); err != nil {
		return err
	}

	// Perform validation tasks in parallel.
	htmlCh := startValidation(ctx, htmlPaths, validateHTML)
	ampCh := startValidation(ctx, ampPaths, validate.AMP)
	cssCh := startValidation(ctx, cssPaths, validateCSS)

	results := make(map[string]bool)
	results["HTML"] = printValidateResults("HTML", dir, htmlCh, []*regexp.Regexp{
		// Blackfriday uses align attributes on generated tables.
		regexp.MustCompile(`^The align attribute on the (td|th) element is obsolete. Use CSS instead\.$`),
	})
	results["AMP"] = printValidateResults("AMP", dir, ampCh, nil)
	results["CSS"] = printValidateResults("CSS", dir, cssCh, []*regexp.Regexp{
		// AMP's inlined styles use these.
		regexp.MustCompile(`-amp-start.* is an unknown vendor extension`),
		regexp.MustCompile(`^Unrecognized at-rule @-[a-z]+-keyframes$`),
		regexp.MustCompile(`^-[a-z]+-animation is an unknown vendor extension$`),
	})

	var failed []string
	for name, res := range results {
		if !res {
			failed = append(failed, name)
		}
	}
	if len(failed) != 0 {
		return fmt.Errorf("validation failed (%v)", strings.Join(failed, ", "))
	}
	return nil
}

// printValidateResults reads all results from ch and prints errors to stdout.
// Issues are printed if they are not excluded by a regexp in ignore.
// vname describes the type of validation being performed, e.g. "HTML".
// dir should be the output directory passed to validateDir.
// Returns true if all documents validated successfully.
func printValidateResults(vname, dir string, ch <-chan validateResult, ignore []*regexp.Regexp) bool {
	succeeded := true
	for res := range ch {
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
		fn := res.p[len(dir)+1:] // strip off dir but keep subdir(s)
		if res.err != nil {
			fmt.Printf("%s: %s validator failed: %v\n", fn, vname, res.err)
			succeeded = false
		}
		for _, is := range issues {
			fmt.Printf("%s: %s %d:%d %s\n", fn, vname, is.Line, is.Col, is.Message)
			succeeded = false
		}
	}
	return succeeded
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
		close(ch)
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
