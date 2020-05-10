// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/derat/homepage/build"
)

func main() {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to get working dir:", err)
		os.Exit(1)
	}
	flag.StringVar(&dir, "dir", dir, "Site directory (defaults to working dir)")
	out := flag.String("out", "", "Destination directory (site is built under -dir if empty)")
	pretty := flag.Bool("pretty", true, "Pretty-print HTML")
	validate := flag.Bool("validate", true, "Validate generated files")
	flag.Parse()

	// TODO: Permit this once everything works.
	if *out == "" {
		fmt.Fprintln(os.Stderr, "-out must be explicitly specified", err)
		os.Exit(2)
	}

	var flags build.Flags
	if *pretty {
		flags |= build.PrettyPrint
	}
	if *validate {
		flags |= build.Validate
	}
	if err := build.Build(context.Background(), dir, *out, flags); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to build site:", err)
		os.Exit(1)
	}
}
