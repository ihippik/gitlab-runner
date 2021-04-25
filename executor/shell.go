// Package executor contain various component implementations for launching jobs.
package executor

import "os/exec"

// ShellExecutor represent executor which runs scripts on the host machine.
type ShellExecutor struct{}

// NewShellExecutor create new instance of shell executor.
func NewShellExecutor() *ShellExecutor {
	return &ShellExecutor{}
}

// Execute implements interface and execute job.
func (s ShellExecutor) Execute(command string) (string, error) {
	// TODO (k.makarov): linux edition
	cmd := exec.Command("bash", "-c", command)

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(output), nil
}
