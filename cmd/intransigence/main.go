// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/derat/intransigence/build"
	"github.com/derat/intransigence/render"
)

func main() {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to get working dir:", err)
		os.Exit(1)
	}
	flag.StringVar(&dir, "dir", dir, "Site directory (defaults to working dir)")
	minifyPath := flag.String("minify", "", "Minify the named CSS or JS file and print it to stdout")
	out := flag.String("out", "", "Destination directory (site is built under -dir if empty)")
	pretty := flag.Bool("pretty", true, "Pretty-print HTML")
	prompt := flag.Bool("prompt", true, "Prompt with a diff before replacing dest dir (only if -out is empty)")
	serve := flag.Bool("serve", true, "Serve output over HTTP while displaying diff")
	validate := flag.Bool("validate", true, "Validate generated files")
	flag.Parse()

	os.Exit(func() int {
		if *minifyPath != "" {
			in, err := ioutil.ReadFile(*minifyPath)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Failed reading file:", err)
				return 1
			}
			out, err := render.MinifyData(string(in), filepath.Ext(*minifyPath))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed minifying %v: %v\n", *minifyPath, err)
				return 1
			}
			if _, err := io.WriteString(os.Stdout, out); err != nil {
				fmt.Fprintln(os.Stderr, "Failed writing data:", err)
				return 1
			}
			return 0
		}

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
			return 1
		}
		return 0
	}())
}
