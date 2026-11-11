package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Byroncl/exposedwatch/internal/adapters/baselinefile"
	"github.com/Byroncl/exposedwatch/internal/ports"
)

func TestBaselineSaveWritesTheFileAndPrintsOneLine(testRun *testing.T) {
	procRoot := writeProcWithListeners(testRun, procListener{wildcardHex, 8080}, procListener{loopbackHex, 5432})
	path := filepath.Join(testRun.TempDir(), "state", "baseline.json")
	code, stdout, stderr := executeBaselineWith(
		[]string{"save", "--file", path, "--no-docker", "--proc-root", procRoot}, fixedBaselineDependencies(""))
	if code != ExitOK || stderr != "" {
		testRun.Fatalf("code=%d stderr=%q", code, stderr)
	}
	if want := "Baseline saved to " + path + " (2 ports)\n"; stdout != want {
		testRun.Fatalf("stdout %q want %q", stdout, want)
	}
	loaded, err := baselinefile.NewStore(baselinefile.SystemTrust()).Load(ports.BaselinePath(path))
	if err != nil || len(loaded.Baseline.Entries) != 2 {
		testRun.Fatalf("err=%v entries=%v", err, loaded.Baseline.Entries)
	}
	if !loaded.Baseline.CreatedAt.Equal(time.Date(2026, 11, 11, 12, 30, 0, 0, time.UTC)) || loaded.Baseline.ToolVersion != version {
		testRun.Fatalf("header: %+v", loaded.Baseline)
	}
}

func TestBaselineSaveUsesTheDefaultPath(testRun *testing.T) {
	stateHome := testRun.TempDir()
	procRoot := writeProcWithListeners(testRun, procListener{wildcardHex, 8080})
	code, stdout, _ := executeBaselineWith([]string{"save", "--no-docker", "--proc-root", procRoot}, fixedBaselineDependencies(stateHome))
	wantPath := filepath.Join(stateHome, "exposedwatch", "baseline.json")
	if code != ExitOK || !strings.Contains(stdout, wantPath) {
		testRun.Fatalf("code=%d stdout=%q", code, stdout)
	}
	if _, err := os.Stat(wantPath); err != nil {
		testRun.Fatal(err)
	}
}

func TestBaselineSaveRefusesAPartialBaseline(testRun *testing.T) {
	procRoot := writeProcWithListeners(testRun, procListener{wildcardHex, 8080})
	path := filepath.Join(testRun.TempDir(), "baseline.json")
	code, stdout, stderr := executeBaselineWith(
		[]string{"save", "--file", path, "--proc-root", procRoot, "--docker-socket", missingDockerSocket(testRun)}, fixedBaselineDependencies(""))
	if code != ExitRuntimeError || stdout != "" {
		testRun.Fatalf("code=%d stdout=%q", code, stdout)
	}
	for _, want := range []string{"docker", "--allow-partial"} {
		if !strings.Contains(stderr, want) {
			testRun.Errorf("stderr lacks %q: %s", want, stderr)
		}
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		testRun.Fatalf("a partial baseline must not be written, stat=%v", err)
	}
}

func TestBaselineSaveAllowPartialWritesAndWarns(testRun *testing.T) {
	procRoot := writeProcWithListeners(testRun, procListener{wildcardHex, 8080})
	path := filepath.Join(testRun.TempDir(), "baseline.json")
	code, stdout, stderr := executeBaselineWith(
		[]string{"save", "--allow-partial", "--file", path, "--proc-root", procRoot, "--docker-socket", missingDockerSocket(testRun)}, fixedBaselineDependencies(""))
	if code != ExitOK || !strings.Contains(stdout, "(1 ports)") || !strings.Contains(stderr, "docker") {
		testRun.Fatalf("code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	if _, err := os.Stat(path); err != nil {
		testRun.Fatal(err)
	}
}

func TestBaselineSaveFailsWhenEverySourceFails(testRun *testing.T) {
	path := filepath.Join(testRun.TempDir(), "baseline.json")
	code, _, _ := executeBaselineWith(
		[]string{"save", "--allow-partial", "--file", path, "--proc-root", filepath.Join(testRun.TempDir(), "nope"), "--docker-socket", missingDockerSocket(testRun)},
		fixedBaselineDependencies(""))
	if code != ExitRuntimeError {
		testRun.Fatalf("code=%d", code)
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		testRun.Fatal("nothing may be written")
	}
}

func TestBaselineSaveReplacesAnExistingBaseline(testRun *testing.T) {
	path := storeBaseline(testRun, hostEntry(22, "0.0.0.0"))
	procRoot := writeProcWithListeners(testRun, procListener{wildcardHex, 8080})
	code, _, _ := executeBaselineWith([]string{"save", "--file", path, "--no-docker", "--proc-root", procRoot}, fixedBaselineDependencies(""))
	loaded, err := baselinefile.NewStore(baselinefile.SystemTrust()).Load(ports.BaselinePath(path))
	if code != ExitOK || err != nil || len(loaded.Baseline.Entries) != 1 || loaded.Baseline.Entries[0].Port != 8080 {
		testRun.Fatalf("code=%d err=%v entries=%v", code, err, loaded.Baseline.Entries)
	}
}

func TestBaselineUsageErrors(testRun *testing.T) {
	testCases := []struct {
		name      string
		arguments []string
	}{
		{"no subcommand", []string{"baseline"}},
		{"unknown subcommand", []string{"baseline", "frobnicate"}},
		{"unknown flag", []string{"baseline", "save", "--bogus"}},
		{"unexpected argument", []string{"baseline", "save", "extra"}},
		{"empty file flag", []string{"baseline", "save", "--file", ""}},
	}
	for _, testCase := range testCases {
		testRun.Run(testCase.name, func(testRun *testing.T) {
			var stdout, stderr strings.Builder
			code := run(testCase.arguments, &stdout, &stderr)
			if code != ExitUsage || stdout.String() != "" || stderr.Len() == 0 {
				testRun.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
			}
		})
	}
}

func TestBaselineSaveHelpExitsZero(testRun *testing.T) {
	var stdout, stderr strings.Builder
	if code := run([]string{"baseline", "save", "-h"}, &stdout, &stderr); code != ExitOK {
		testRun.Fatalf("code=%d", code)
	}
}

func TestUsageTextDocumentsTheBaselineCommandAndExitCode(testRun *testing.T) {
	for _, want := range []string{"baseline", "baseline save", "--baseline", "4 drift"} {
		if !strings.Contains(usageText, want) {
			testRun.Errorf("usage text lacks %q", want)
		}
	}
	if ExitDrift != 4 {
		testRun.Fatalf("ExitDrift=%d", ExitDrift)
	}
}

func TestBaselineSaveRefusesToOverwriteANonBaselineFile(testRun *testing.T) {
	directory := testRun.TempDir()
	path := filepath.Join(directory, "passwd")
	if err := os.WriteFile(path, []byte("root:x:0:0\n"), 0o644); err != nil {
		testRun.Fatal(err)
	}
	procRoot := writeProcWithListeners(testRun, procListener{wildcardHex, 8080})
	arguments := []string{"save", "--file", path, "--no-docker", "--proc-root", procRoot}
	code, stdout, stderr := executeBaselineWith(arguments, fixedBaselineDependencies(""))
	want := "refusing to overwrite " + path + ": it is not a baseline file (use --force to replace it anyway)"
	if code != ExitRuntimeError || stdout != "" || !strings.Contains(stderr, want) {
		testRun.Fatalf("code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	if content, _ := os.ReadFile(path); string(content) != "root:x:0:0\n" {
		testRun.Fatal("the file must be untouched")
	}
	code, _, _ = executeBaselineWith(append(arguments, "--force"), fixedBaselineDependencies(""))
	if code != ExitOK {
		testRun.Fatalf("--force: code=%d", code)
	}
}
