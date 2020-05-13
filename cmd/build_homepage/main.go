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
	prompt := flag.Bool("prompt", true, "Prompt with a diff before replacing dest dir (only if -out is empty)")
	serve := flag.Bool("serve", true, "Serve output over HTTP while displaying diff")
	validate := flag.Bool("validate", true, "Validate generated files")
	flag.Parse()

	var flags build.Flags
	if *pretty {
		flags |= build.PrettyPrint
	}
	if *prompt {
		flags |= build.Prompt
	}
	if *serve {
		flags |= build.Serve
	}
	if *validate {
		flags |= build.Validate
	}
	if err := build.Build(context.Background(), dir, *out, flags); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to build site:", err)
		os.Exit(1)
	}
}
