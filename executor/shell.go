// Package executor contain various component implementations for launching jobs.
package executor

import (
	"context"
	"os/exec"
)

// ShellExecutor represent executor which runs scripts on the host machine.
type ShellExecutor struct {
	homeDir string
}

// NewShellExecutor create new instance of shell executor.
func NewShellExecutor() *ShellExecutor {
	return &ShellExecutor{}
}

// Execute implements interface and execute job.
func (s *ShellExecutor) Execute(_ context.Context, command string) (string, error) {
	// TODO (k.makarov): linux edition
	cmd := exec.Command("bash", "-c", command)
	cmd.Dir = s.homeDir
	output, err := cmd.CombinedOutput()

	if err != nil {
		return string(output), err
	}

	if len(output) == 0 {
		output = []byte("ok")
	}

	return string(output), nil
}

// HomeDirectory set home directory.
func (s *ShellExecutor) HomeDirectory(dir string) {
	s.homeDir = dir
}
