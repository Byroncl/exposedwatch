package main

import (
	"bytes"
	"strings"
	"testing"
)

func runCommand(arguments ...string) (ExitCode, string, string) {
	var stdout, stderr bytes.Buffer
	code := run(arguments, &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

func TestNoArgumentsOpensTui(testRun *testing.T) {
	// The test writer is not a terminal, so the tui refuses to start; reaching
	// that refusal proves the empty command line was routed to the tui.
	code, _, stderr := runCommand()
	if code != ExitUsage || !strings.Contains(stderr, "interactive terminal") {
		testRun.Fatalf("code=%d stderr=%q", code, stderr)
	}
}

func TestLeadingFlagsOpenTui(testRun *testing.T) {
	code, _, stderr := runCommand("--interval", "5s")
	if code != ExitUsage || !strings.Contains(stderr, "interactive terminal") {
		testRun.Fatalf("code=%d stderr=%q", code, stderr)
	}
}

func TestHelpFlagsPrintUsage(testRun *testing.T) {
	for _, helpArgument := range []string{"-h", "--help", "help"} {
		code, stdout, _ := runCommand(helpArgument)
		if code != ExitOK || !strings.Contains(stdout, "Usage") {
			testRun.Errorf("%s: code=%d stdout=%q", helpArgument, code, stdout)
		}
	}
}

func TestVersionFlag(testRun *testing.T) {
	code, stdout, _ := runCommand("--version")
	if code != ExitOK || !strings.Contains(stdout, version) {
		testRun.Fatalf("code=%d stdout=%q", code, stdout)
	}
}

func TestUnknownSubcommandIsUsageError(testRun *testing.T) {
	code, _, stderr := runCommand("bogus")
	if code != ExitUsage || !strings.Contains(stderr, "Usage") {
		testRun.Fatalf("code=%d stderr=%q", code, stderr)
	}
}

func TestVersionSubcommand(testRun *testing.T) {
	code, stdout, _ := runCommand("version")
	if code != ExitOK || !strings.Contains(stdout, version) {
		testRun.Fatalf("code=%d stdout=%q", code, stdout)
	}
}

func TestScanRejectsInvalidFlags(testRun *testing.T) {
	for _, arguments := range [][]string{
		{"scan", "--format", "xml"},
		{"scan", "--min-severity", "severe"},
		{"scan", "--nope"},
	} {
		code, _, stderr := runCommand(arguments...)
		if code != ExitUsage || stderr == "" {
			testRun.Errorf("%v: code=%d stderr=%q", arguments, code, stderr)
		}
	}
}

func TestParseScanOptionsDefaults(testRun *testing.T) {
	options, err := parseScanOptions(nil, &bytes.Buffer{})
	if err != nil {
		testRun.Fatal(err)
	}
	if options.Format != FormatTable || options.ProcRoot != "/proc" || options.DockerSocket != "/var/run/docker.sock" || options.NoDocker {
		testRun.Fatalf("options = %+v", options)
	}
}

func TestScanFromFakeProcWithoutDocker(testRun *testing.T) {
	procRoot := testRun.TempDir()
	writeFakeProc(testRun, procRoot)
	code, stdout, _ := runCommand("scan", "--no-docker", "--proc-root", procRoot, "--format", "json")
	if code != ExitFindings {
		testRun.Fatalf("code=%d stdout=%s", code, stdout)
	}
	if !strings.Contains(stdout, `"port": 8080`) {
		testRun.Fatalf("stdout=%s", stdout)
	}
}

func TestScanMinSeverityControlsExitCode(testRun *testing.T) {
	procRoot := testRun.TempDir()
	writeFakeProc(testRun, procRoot)
	code, _, _ := runCommand("scan", "--no-docker", "--proc-root", procRoot, "--min-severity", "critical")
	if code != ExitOK {
		testRun.Fatalf("code=%d, want ExitOK", code)
	}
}

func TestScanRuntimeErrorWhenProcMissing(testRun *testing.T) {
	code, _, stderr := runCommand("scan", "--no-docker", "--proc-root", testRun.TempDir())
	if code != ExitRuntimeError || stderr == "" {
		testRun.Fatalf("code=%d stderr=%q", code, stderr)
	}
}
