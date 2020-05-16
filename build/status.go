// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package build

import (
	"fmt"
	"io"
	"os"
)

// statusDest is the io.Writer used by statusf and logf.
var statusDest = os.Stderr

// statusLine is the status text most recently written via statusf.
var statusLine string

// statusf replaces the status line with the supplied format string and args.
func statusf(format string, args ...interface{}) {
	clearStatus()
	statusLine = fmt.Sprintf(format, args...)
	writeStatus()
}

// logf writes the supplied format string and args and rewrites the status line after it.
// A trailing newline is added to format if not already present.
func logf(format string, args ...interface{}) {
	if len(format) > 0 && format[len(format)-1] != '\n' {
		format += "\n"
	}
	clearStatus()
	fmt.Fprintf(statusDest, format, args...)
	writeStatus()
}

func writeStatus() {
	io.WriteString(statusDest, statusLine)
}

func clearStatus() {
	// https://unix.stackexchange.com/questions/26576/how-to-delete-line-with-echo
	io.WriteString(statusDest, "\x1b[2K\r")
}
