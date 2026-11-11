package nmapscan

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
)

// CommandResult is what a finished process produced.
type CommandResult struct {
	Stdout   []byte
	Stderr   []byte
	ExitCode int
}

// CommandRunner runs an executable with an argument list. Implementations
// must never involve a shell.
type CommandRunner interface {
	Run(ctx context.Context, binary string, arguments []string) (CommandResult, error)
}

// ExecRunner runs real processes with os/exec. Arguments are passed as a
// slice straight to execve; no shell is ever used.
type ExecRunner struct{}

// NewExecRunner returns the production CommandRunner.
func NewExecRunner() ExecRunner {
	return ExecRunner{}
}

// Run resolves binary through PATH when needed and runs it. A non-zero exit is
// reported through CommandResult.ExitCode, not as an error. Context
// cancellation kills the process and returns the context error.
func (ExecRunner) Run(ctx context.Context, binary string, arguments []string) (CommandResult, error) {
	resolved, err := exec.LookPath(binary)
	if err != nil {
		return CommandResult{}, fmt.Errorf(unavailableFormat, ErrNmapUnavailable, err)
	}
	stdout := &boundedBuffer{limit: maxStdoutBytes}
	stderr := &boundedBuffer{limit: maxStderrBytes}
	command := exec.CommandContext(ctx, resolved, arguments...)
	command.Stdout = stdout
	command.Stderr = stderr
	command.WaitDelay = processWaitDelay
	runErr := command.Run()
	if contextErr := ctx.Err(); contextErr != nil {
		return CommandResult{}, contextErr
	}
	result := CommandResult{Stdout: stdout.Bytes(), Stderr: stderr.Bytes()}
	if runErr == nil {
		return result, nil
	}
	var exitError *exec.ExitError
	if errors.As(runErr, &exitError) {
		result.ExitCode = exitError.ExitCode()
		if result.ExitCode == noExitCode {
			result.ExitCode = killedExitStatus
		}
		return result, nil
	}
	return CommandResult{}, fmt.Errorf(startFailedFormat, runErr)
}
