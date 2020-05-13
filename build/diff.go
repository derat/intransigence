// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package build

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// showDiffAndPrompt displays differences between directories a and b and
// prompts the user to accept the changes. The user's response is returned.
func showDiffAndPrompt(a, b string) (ok bool, err error) {
	for {
		if err := showDiff(a, b); err != nil {
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
func showDiff(a, b string) error {
	pager := os.Getenv("PAGER")
	if pager == "" {
		pager = "less"
	}
	pagerCmd := exec.Command(pager)
	pagerStdin, err := pagerCmd.StdinPipe()
	if err != nil {
		return err
	}
	pagerCmd.Stdout = os.Stdout
	pagerCmd.Stderr = os.Stderr
	if err := pagerCmd.Start(); err != nil {
		return fmt.Errorf("failed starting %q: %v", strings.Join(pagerCmd.Args, " "), err)
	}

	diffCmd := exec.Command("diff", "-r", "-u", "--color=always", a, b)
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
