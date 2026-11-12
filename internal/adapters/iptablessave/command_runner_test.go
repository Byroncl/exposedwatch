package iptablessave

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

const (
	helperModeEnvironment = "IPTABLESSAVE_HELPER_MODE"
	helperModeOutput      = "output"
	helperModeFail        = "fail"
	helperModeSleep       = "sleep"
	helperModeFlood       = "flood"
	helperStdoutText      = "hello-from-helper"
	helperStderrText      = "helper-stderr"
	helperFailureExitCode = 7
	helperFloodBytes      = 20 * 1024 * 1024
	helperSleepDuration   = 30 * time.Second
	cancelAfter           = 200 * time.Millisecond
)

// TestHelperProcess re-executes the test binary as a stand-in for iptables-save.
func TestHelperProcess(testRun *testing.T) {
	switch os.Getenv(helperModeEnvironment) {
	case helperModeOutput:
		fmt.Fprint(os.Stdout, helperStdoutText)
		os.Exit(0)
	case helperModeFail:
		fmt.Fprint(os.Stderr, helperStderrText)
		os.Exit(helperFailureExitCode)
	case helperModeSleep:
		time.Sleep(helperSleepDuration)
		os.Exit(0)
	case helperModeFlood:
		fmt.Fprint(os.Stdout, strings.Repeat("a", helperFloodBytes))
		os.Exit(0)
	}
}

func helperArguments() []string {
	return []string{"-test.run=TestHelperProcess"}
}

func TestExecRunnerCapturesOutput(testRun *testing.T) {
	testRun.Setenv(helperModeEnvironment, helperModeOutput)
	result, err := NewExecRunner().Run(context.Background(), os.Args[0], helperArguments())
	if err != nil || !strings.Contains(string(result.Stdout), helperStdoutText) || result.ExitCode != 0 || result.Truncated {
		testRun.Fatalf("result=%+v err=%v", result, err)
	}
}

func TestExecRunnerReportsNonZeroExit(testRun *testing.T) {
	testRun.Setenv(helperModeEnvironment, helperModeFail)
	result, err := NewExecRunner().Run(context.Background(), os.Args[0], helperArguments())
	if err != nil || result.ExitCode != helperFailureExitCode || !strings.Contains(string(result.Stderr), helperStderrText) {
		testRun.Fatalf("result=%+v err=%v", result, err)
	}
}

func TestExecRunnerStartFailureIsAnError(testRun *testing.T) {
	_, err := NewExecRunner().Run(context.Background(), "/definitely/not/a/binary", nil)
	if err == nil {
		testRun.Fatal("expected a start failure")
	}
}

func TestExecRunnerKillsProcessOnCancellation(testRun *testing.T) {
	testRun.Setenv(helperModeEnvironment, helperModeSleep)
	ctx, cancel := context.WithTimeout(context.Background(), cancelAfter)
	defer cancel()
	started := time.Now()
	_, err := NewExecRunner().Run(ctx, os.Args[0], helperArguments())
	if !errors.Is(err, context.DeadlineExceeded) {
		testRun.Fatalf("err=%v", err)
	}
	if time.Since(started) > helperSleepDuration/2 {
		testRun.Fatalf("process was not killed promptly: %v", time.Since(started))
	}
}

func TestExecRunnerBoundsCapturedOutput(testRun *testing.T) {
	testRun.Setenv(helperModeEnvironment, helperModeFlood)
	result, err := NewExecRunner().Run(context.Background(), os.Args[0], helperArguments())
	if err != nil {
		testRun.Fatal(err)
	}
	if !result.Truncated {
		testRun.Fatal("flooded output must be flagged as truncated")
	}
	if len(result.Stdout) != maxStdoutBytes {
		testRun.Fatalf("stdout = %d bytes, want exactly the %d byte bound", len(result.Stdout), maxStdoutBytes)
	}
}
