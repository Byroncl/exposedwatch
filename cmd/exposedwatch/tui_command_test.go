package main

import (
	"strings"
	"testing"
	"time"
)

func TestTuiRefusesWhenStdoutIsNotATerminal(testRun *testing.T) {
	code, stdout, stderr := runCommand("tui", "--no-docker")
	if code != ExitUsage || stdout != "" || !strings.Contains(stderr, "interactive terminal") {
		testRun.Fatalf("code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
}

func TestTuiRejectsBadFlags(testRun *testing.T) {
	for _, arguments := range [][]string{
		{"tui", "--interval", "500ms"},
		{"tui", "--interval", "0s"},
		{"tui", "--nope"},
		{"tui", "extra"},
	} {
		code, _, stderr := runCommand(arguments...)
		if code != ExitUsage || stderr == "" {
			testRun.Errorf("%v: code=%d stderr=%q", arguments, code, stderr)
		}
	}
}

func TestParseTuiOptionsDefaults(testRun *testing.T) {
	options, err := parseTuiOptions(nil, &strings.Builder{})
	if err != nil {
		testRun.Fatal(err)
	}
	if options.Interval != 2*time.Second || options.ProcRoot != "/proc" || options.NoDocker {
		testRun.Fatalf("options = %+v", options)
	}
}

func TestParseTuiOptionsAcceptsMinimumInterval(testRun *testing.T) {
	options, err := parseTuiOptions([]string{"--interval", "1s", "--no-docker", "--proc-root", "/x"}, &strings.Builder{})
	if err != nil || options.Interval != time.Second || !options.NoDocker || options.ProcRoot != "/x" {
		testRun.Fatalf("options=%+v err=%v", options, err)
	}
}

func TestUsageMentionsTuiCommand(testRun *testing.T) {
	_, _, stderr := runCommand()
	if !strings.Contains(stderr, "tui") {
		testRun.Fatalf("usage lacks tui: %q", stderr)
	}
}
