package main

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/Byroncl/exposedwatch/internal/adapters/nmapscan"
)

func TestExternalIsOffByDefault(testRun *testing.T) {
	options, err := parseScanOptions(nil, &bytes.Buffer{})
	if err != nil || options.External || options.ConfirmOwnership || len(options.ExternalTargets) != 0 {
		testRun.Fatalf("options=%+v err=%v", options, err)
	}
	if options.ExternalTimeout != nmapscan.DefaultTimeout || options.NmapPath != nmapscan.BinaryName {
		testRun.Fatalf("defaults wrong: %+v", options)
	}
}

func TestExternalFlagsParse(testRun *testing.T) {
	options, err := parseScanOptions([]string{
		"--external", "--external-target", "10.0.0.1,example.com", "--external-target", "10.0.0.0/30",
		"--confirm-ownership", "--external-ports", "22,8000-8100", "--external-timeout", "30s",
		"--nmap-path", "/opt/nmap",
	}, &bytes.Buffer{})
	if err != nil {
		testRun.Fatal(err)
	}
	wantTargets := nmapscan.TargetList{"10.0.0.1", "example.com", "10.0.0.0/30"}
	if !options.External || !options.ConfirmOwnership || len(options.ExternalTargets) != len(wantTargets) {
		testRun.Fatalf("options=%+v", options)
	}
	for index := range wantTargets {
		if options.ExternalTargets[index] != wantTargets[index] {
			testRun.Fatalf("targets=%v", options.ExternalTargets)
		}
	}
	if options.ExternalPorts != "22,8000-8100" || options.ExternalTimeout != 30*time.Second || options.NmapPath != "/opt/nmap" {
		testRun.Fatalf("options=%+v", options)
	}
}

func TestExternalTargetRequiresConfirmOwnership(testRun *testing.T) {
	code, _, stderr := runCommand("scan", "--no-docker", "--external", "--external-target", "10.0.0.1")
	if code != ExitUsage || !strings.Contains(stderr, "--confirm-ownership") {
		testRun.Fatalf("code=%d stderr=%q", code, stderr)
	}
}

func TestExternalTargetRequiresExternal(testRun *testing.T) {
	code, _, stderr := runCommand("scan", "--external-target", "10.0.0.1", "--confirm-ownership")
	if code != ExitUsage || !strings.Contains(stderr, "--external") {
		testRun.Fatalf("code=%d stderr=%q", code, stderr)
	}
}

func TestExternalRejectsBadInputsAsUsageErrors(testRun *testing.T) {
	badArguments := [][]string{
		{"--external", "--confirm-ownership", "--external-target", "--script=x"},
		{"--external", "--confirm-ownership", "--external-target", "1.1.1.1; rm -rf /"},
		{"--external", "--confirm-ownership", "--external-target", "10.0.0.0/8"},
		{"--external", "--confirm-ownership", "--external-target", "::1"},
		{"--external", "--confirm-ownership", "--external-target", "0x7f000001"},
		{"--external", "--external-ports", "22;ls"},
		{"--external", "--external-timeout", "1s"},
		{"--external", "--external-timeout", "99h"},
	}
	for _, badSet := range badArguments {
		code, _, stderr := runCommand(append([]string{"scan", "--no-docker"}, badSet...)...)
		if code != ExitUsage || stderr == "" {
			testRun.Errorf("%v: code=%d stderr=%q", badSet, code, stderr)
		}
	}
}

func TestExternalConfirmHelpMentionsLegalWarning(testRun *testing.T) {
	var errorOutput bytes.Buffer
	_, _ = parseScanOptions([]string{"-h"}, &errorOutput)
	help := errorOutput.String()
	for _, expected := range []string{"--external", "--confirm-ownership", "illegal", "TCP"} {
		if !strings.Contains(help, expected) {
			testRun.Errorf("help is missing %q:\n%s", expected, help)
		}
	}
}

func TestExternalWithMissingNmapDegradesGracefully(testRun *testing.T) {
	procRoot := testRun.TempDir()
	writeFakeProc(testRun, procRoot)
	code, stdout, stderr := runCommand("scan", "--no-docker", "--proc-root", procRoot, "--external",
		"--nmap-path", "definitely-not-an-installed-binary-xyz", "--format", "json")
	if code == ExitRuntimeError {
		testRun.Fatalf("host results must still print: code=%d stderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, `"port": 8080`) || !strings.Contains(stdout, `"source": "external"`) {
		testRun.Fatalf("host port missing from output: %q", stdout)
	}
}

func TestUsageMentionsExternalScan(testRun *testing.T) {
	if !strings.Contains(usageText, "--external") {
		testRun.Fatal("usage text does not mention --external")
	}
}
