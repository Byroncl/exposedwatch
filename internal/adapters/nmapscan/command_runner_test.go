package nmapscan

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

const (
	helperModeEnvironment = "NMAPSCAN_HELPER_MODE"
	helperModeOutput      = "output"
	helperModeFail        = "fail"
	helperModeSleep       = "sleep"
	helperModeFlood       = "flood"
	helperModeOrphan      = "orphan"
	helperModeOrphanChild = "orphan-child"
	helperStdoutText      = "hello-from-helper"
	helperStderrText      = "helper-stderr"
	helperFailureExitCode = 7
	helperFloodBytes      = 2 * 1024 * 1024
	helperSleepDuration   = 30 * time.Second
	helperOrphanLifetime  = 6 * time.Second
	orphanKillTolerance   = 2 * time.Second
)

// TestHelperProcess re-executes the test binary as a stand-in for nmap.
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
	case helperModeOrphan:
		child := exec.Command(os.Args[0], helperArguments()...)
		child.Env = append(os.Environ(), helperModeEnvironment+"="+helperModeOrphanChild)
		child.Stdout = os.Stdout
		child.Stderr = os.Stderr
		if err := child.Start(); err != nil {
			os.Exit(helperFailureExitCode)
		}
		time.Sleep(helperSleepDuration)
		os.Exit(0)
	case helperModeOrphanChild:
		time.Sleep(helperOrphanLifetime)
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
	if err != nil || !strings.Contains(string(result.Stdout), helperStdoutText) || result.ExitCode != 0 {
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

func TestExecRunnerMissingBinaryIsUnavailable(testRun *testing.T) {
	_, err := NewExecRunner().Run(context.Background(), "definitely-not-an-installed-binary-xyz", nil)
	if !errors.Is(err, ErrNmapUnavailable) {
		testRun.Fatalf("err=%v", err)
	}
}

func TestExecRunnerKillsProcessOnTimeout(testRun *testing.T) {
	testRun.Setenv(helperModeEnvironment, helperModeSleep)
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
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
	if len(result.Stdout) > maxStdoutBytes {
		testRun.Fatalf("stdout not bounded: %d", len(result.Stdout))
	}
}

func TestExecRunnerDoesNotHangWhenGrandchildHoldsThePipes(testRun *testing.T) {
	testRun.Setenv(helperModeEnvironment, helperModeOrphan)
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	started := time.Now()
	_, err := NewExecRunner().Run(ctx, os.Args[0], helperArguments())
	if !errors.Is(err, context.DeadlineExceeded) {
		testRun.Fatalf("err=%v", err)
	}
	if elapsed := time.Since(started); elapsed > processWaitDelay+orphanKillTolerance {
		testRun.Fatalf("Run blocked on inherited pipes: %v", elapsed)
	}
}
