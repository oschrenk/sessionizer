package shell

import (
	"bytes"
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
