// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package build

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

const serverAddr = ":8888"

// prompt displays differences between directories a and b and
// prompts the user to accept the changes. The user's response is returned.
// If serveB is true, an HTTP server is started at serveAddr to serve the contents of b.
func prompt(ctx context.Context, a, b string, serveB bool) (ok bool, err error) {
	var msg string
	var srv *http.Server
	var sch <-chan error
	if serveB {
		msg = fmt.Sprintf("Serving %v at %v\n\n", b, serverAddr)
		srv, sch = startServer(b, serverAddr)
	}

	ok, perr := showDiffAndPrompt(ctx, a, b, msg)

	var serr error
	if srv != nil {
		serr = srv.Shutdown(ctx)
		if err := <-sch; err != nil && err != http.ErrServerClosed {
			serr = err
		}
	}

	if perr != nil {
		return ok, perr
	}
	return ok, serr
}

// startServer starts a new HTTP server at addr to serve the files in dir.
// The return value from ListenAndServe will be written to the returned channel.
func startServer(dir, addr string) (*http.Server, <-chan error) {
	srv := &http.Server{
		Addr:    addr,
		Handler: http.FileServer(http.Dir(dir)),
	}
	ch := make(chan error, 1)
	go func() {
		ch <- srv.ListenAndServe()
	}()
	return srv, ch
}

// showDiffAndPrompt displays differences between directories a and b and
// prompts the user to accept the changes. The user's response is returned.
// msg is printed above the diff.
func showDiffAndPrompt(ctx context.Context, a, b, msg string) (ok bool, err error) {
	for {
		if err := showDiff(ctx, a, b, msg); err != nil {
			return false, err
		}

		r := bufio.NewReader(os.Stdin)
		fmt.Print("Replace output dir (y/N/diff)? ")
		s, _ := r.ReadString('\n')
		s = strings.ToLower(s)
		switch {
		case strings.HasPrefix(s, "y"):
			return true, nil
		case strings.HasPrefix(s, "d"):
			continue // show diff again
		default:
			return false, nil
		}
	}
}

// showDiff displays differences between directories a and b.
// header is written above the diff.
func showDiff(ctx context.Context, a, b, header string) error {
	pager := os.Getenv("PAGER")
	if pager == "" {
		pager = "less"
	}
	pagerCmd := exec.CommandContext(ctx, pager)
	pagerStdin, err := pagerCmd.StdinPipe()
	if err != nil {
		return err
	}
	pagerCmd.Stdout = os.Stdout
	pagerCmd.Stderr = os.Stderr
	if err := pagerCmd.Start(); err != nil {
		return fmt.Errorf("failed starting %q: %v", strings.Join(pagerCmd.Args, " "), err)
	}

	io.WriteString(pagerStdin, header)

	diffCmd := exec.CommandContext(ctx, "diff", "-r", "-u", "--color=always", a, b)
	diffCmd.Stdout = pagerStdin
	var diffStderr bytes.Buffer
	diffCmd.Stderr = &diffStderr

	// diff(1): "Exit status is 0 if inputs are the same, 1 if different, 2 if trouble."
	diffErr := diffCmd.Run()
	if diffErr == nil {
		io.WriteString(pagerStdin, "No differences.\n")
	} else if exitErr, ok := diffErr.(*exec.ExitError); ok {
		if exitErr.ExitCode() == 1 {
			diffErr = nil // differences found
		} else if ws, ok := exitErr.Sys().(syscall.WaitStatus); ok && ws.Signal() == syscall.SIGPIPE {
			diffErr = nil // pager exited before it read entire diff
		}
	}
	if diffErr != nil {
		io.WriteString(pagerStdin, diffStderr.String())
		diffErr = fmt.Errorf("%q failed: %v", strings.Join(diffCmd.Args, " "), diffErr)
	}

	if err := pagerStdin.Close(); err != nil {
		return err
	}
	if err := pagerCmd.Wait(); err != nil {
		return fmt.Errorf("failed waiting for %q: %v", strings.Join(pagerCmd.Args, " "), err)
	}
	return diffErr
}
