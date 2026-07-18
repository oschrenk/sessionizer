package shell

import (
	"bytes"
	"os"
	"os/exec"
)

// Run a command and capture stdout, stderr
func Run(name string, args []string) (string, string, error) {
	// find executable in PATH, and get absolute path
	bin, err := exec.LookPath(name)
	if err != nil {
		return "", "", err
	}

	// prepare command
	cmd := exec.Command(bin, args...)

	// prepare capture stdout, stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// execute command
	err = cmd.Run()
	outStr, errStr := stdout.String(), stderr.String()

	return outStr, errStr, err
}

// RunInteractive runs a command with the real terminal wired up: the child
// inherits this process's stdin/stdout/stderr instead of having its output
// buffered. This is what allows an interactive `tmux attach` to take over the
// terminal. It uses cmd.Run() (not syscall.Exec) so it returns when the command
// exits — e.g. when the user detaches — letting the caller continue afterwards.
func RunInteractive(name string, args []string) error {
	// find executable in PATH, and get absolute path
	bin, err := exec.LookPath(name)
	if err != nil {
		return err
	}

	// hand over the real terminal
	cmd := exec.Command(bin, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// returns when the user detaches
	return cmd.Run()
}
