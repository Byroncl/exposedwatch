package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

func scanWithBaseline(testRun *testing.T, baselinePath string, procRoot string, extra ...string) (ExitCode, string, string) {
	testRun.Helper()
	arguments := append([]string{"--no-docker", "--proc-root", procRoot, "--baseline", baselinePath}, extra...)
	return executeScanWith(testRun, &fakeFirewallEvaluator{report: allowedFirewallReport()}, arguments...)
}

func TestScanBaselineWithoutDriftExitsZero(testRun *testing.T) {
	procRoot := writeProcWithListeners(testRun, procListener{wildcardHex, 8080})
	baselinePath := storeBaseline(testRun, hostEntry(8080, "0.0.0.0"))
	code, stdout, stderr := scanWithBaseline(testRun, baselinePath, procRoot)
	if code != ExitOK || stderr != "" {
		testRun.Fatalf("code=%d stderr=%q", code, stderr)
	}
	for _, want := range []string{"Drift", "No drift", baselinePath, "2026-11-11T10:00:00Z"} {
		if !strings.Contains(stdout, want) {
			testRun.Errorf("output lacks %q:\n%s", want, stdout)
		}
	}
	if strings.Contains(stdout, "NEW") || strings.Contains(stdout, "CLOSED") {
		testRun.Fatalf("no drift must not print NEW/CLOSED:\n%s", stdout)
	}
}

func TestScanBaselineNewPortExitsWithDriftCode(testRun *testing.T) {
	procRoot := writeProcWithListeners(testRun, procListener{wildcardHex, 8080}, procListener{wildcardHex, 23})
	baselinePath := storeBaseline(testRun, hostEntry(8080, "0.0.0.0"))
	code, stdout, _ := scanWithBaseline(testRun, baselinePath, procRoot)
	if code != ExitDrift {
		testRun.Fatalf("code=%d\n%s", code, stdout)
	}
	for _, want := range []string{"NEW (1)", "23", "critical", "Telnet", "baseline.json"} {
		if !strings.Contains(stdout, want) {
			testRun.Errorf("output lacks %q:\n%s", want, stdout)
		}
	}
}

func TestScanBaselineNeverUsesTheFindingsExitCode(testRun *testing.T) {
	procRoot := writeProcWithListeners(testRun, procListener{wildcardHex, 23})
	baselinePath := storeBaseline(testRun, hostEntry(23, "0.0.0.0"))
	code, _, _ := scanWithBaseline(testRun, baselinePath, procRoot)
	if code != ExitOK {
		testRun.Fatalf("accepted critical ports are expected in baseline mode, got %d", code)
	}
}

func TestScanBaselineMinSeverityDecidesWhatCountsAsDrift(testRun *testing.T) {
	procRoot := writeProcWithListeners(testRun, procListener{wildcardHex, 8080}, procListener{loopbackHex, 9999})
	baselinePath := storeBaseline(testRun)
	testCases := []struct {
		minimum string
		want    ExitCode
	}{
		{"info", ExitDrift}, {"low", ExitOK}, {"medium", ExitOK}, {"critical", ExitOK},
	}
	for _, testCase := range testCases {
		code, stdout, _ := scanWithBaseline(testRun, baselinePath, procRoot, "--min-severity", testCase.minimum)
		if code != testCase.want {
			testRun.Errorf("min-severity %s: code=%d want %d", testCase.minimum, code, testCase.want)
		}
		if !strings.Contains(stdout, "NEW (2)") {
			testRun.Errorf("min-severity %s: new ports are still listed:\n%s", testCase.minimum, stdout)
		}
	}
}

func TestScanBaselineLoopbackNewPortCountsAsInfo(testRun *testing.T) {
	procRoot := writeProcWithListeners(testRun, procListener{loopbackHex, 9999})
	baselinePath := storeBaseline(testRun)
	if code, _, _ := scanWithBaseline(testRun, baselinePath, procRoot); code != ExitDrift {
		testRun.Fatalf("code=%d", code)
	}
}

func TestScanBaselineLoopbackToWildcardIsNew(testRun *testing.T) {
	procRoot := writeProcWithListeners(testRun, procListener{wildcardHex, 5432})
	baselinePath := storeBaseline(testRun, hostEntry(5432, "127.0.0.1"))
	code, stdout, _ := scanWithBaseline(testRun, baselinePath, procRoot)
	if code != ExitDrift || !strings.Contains(stdout, "NEW (1)") || !strings.Contains(stdout, "CLOSED") {
		testRun.Fatalf("code=%d\n%s", code, stdout)
	}
}

func TestScanBaselineClosedPortsAreInformational(testRun *testing.T) {
	procRoot := writeProcWithListeners(testRun)
	baselinePath := storeBaseline(testRun, hostEntry(3306, "0.0.0.0"))
	code, stdout, _ := scanWithBaseline(testRun, baselinePath, procRoot)
	if code != ExitOK {
		testRun.Fatalf("code=%d", code)
	}
	for _, want := range []string{"CLOSED (informational, 1)", "3306"} {
		if !strings.Contains(stdout, want) {
			testRun.Errorf("output lacks %q:\n%s", want, stdout)
		}
	}
}

func TestScanBaselineAFailedSourceIsNotReportedAsClosed(testRun *testing.T) {
	procRoot := writeProcWithListeners(testRun, procListener{wildcardHex, 8080})
	baselinePath := storeBaseline(testRun, hostEntry(8080, "0.0.0.0"), dockerEntry(9000, "0.0.0.0"))
	arguments := []string{"--proc-root", procRoot, "--docker-socket", missingDockerSocket(testRun), "--baseline", baselinePath}
	code, stdout, _ := executeScanWith(testRun, &fakeFirewallEvaluator{}, arguments...)
	if code != ExitRuntimeError {
		testRun.Fatalf("code=%d\n%s", code, stdout)
	}
	if strings.Contains(stdout, "CLOSED") || !strings.Contains(stdout, "docker") || !strings.Contains(stdout, "not checked") {
		testRun.Fatalf("docker entries must be marked unverified, not closed:\n%s", stdout)
	}
}

func TestScanBaselineMissingFileIsARuntimeErrorWithAHint(testRun *testing.T) {
	procRoot := writeProcWithListeners(testRun)
	missing := filepath.Join(testRun.TempDir(), "absent.json")
	code, stdout, stderr := scanWithBaseline(testRun, missing, procRoot)
	if code != ExitRuntimeError || stdout != "" || !strings.Contains(stderr, "exposedwatch baseline save") {
		testRun.Fatalf("code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
}

func TestScanBaselineUntrustedOrCorruptFilesAreRuntimeErrors(testRun *testing.T) {
	procRoot := writeProcWithListeners(testRun)
	good := storeBaseline(testRun)
	link := filepath.Join(testRun.TempDir(), "link.json")
	if err := os.Symlink(good, link); err != nil {
		testRun.Fatal(err)
	}
	corrupt := filepath.Join(testRun.TempDir(), "corrupt.json")
	if err := os.WriteFile(corrupt, []byte("SECRET-TEXT"), 0o600); err != nil {
		testRun.Fatal(err)
	}
	for name, path := range map[string]string{"symlink": link, "corrupt": corrupt} {
		code, stdout, stderr := scanWithBaseline(testRun, path, procRoot)
		if code != ExitRuntimeError || stdout != "" {
			testRun.Errorf("%s: code=%d stdout=%q", name, code, stdout)
		}
		if !strings.Contains(stderr, "exposedwatch baseline save") || strings.Contains(stderr, "SECRET-TEXT") {
			testRun.Errorf("%s: stderr=%q", name, stderr)
		}
	}
}

func TestScanBaselineJSONShape(testRun *testing.T) {
	procRoot := writeProcWithListeners(testRun, procListener{wildcardHex, 8080})
	baselinePath := storeBaseline(testRun, hostEntry(3306, "0.0.0.0"))
	code, stdout, _ := scanWithBaseline(testRun, baselinePath, procRoot, "--format", "json")
	if code != ExitDrift {
		testRun.Fatalf("code=%d", code)
	}
	var document struct {
		Drift struct {
			BaselineFile string           `json:"baseline_file"`
			CreatedAt    string           `json:"created_at"`
			New          []map[string]any `json:"new"`
			Closed       []map[string]any `json:"closed"`
			Notes        []string         `json:"notes"`
		} `json:"drift"`
	}
	if err := json.Unmarshal([]byte(stdout), &document); err != nil {
		testRun.Fatal(err)
	}
	drift := document.Drift
	if drift.BaselineFile != baselinePath || drift.CreatedAt != "2026-11-11T10:00:00Z" || len(drift.New) != 1 || len(drift.Closed) != 1 || drift.Notes == nil {
		testRun.Fatalf("drift: %+v", drift)
	}
	if drift.New[0]["port"] != float64(8080) || drift.New[0]["severity"] != "info" || drift.Closed[0]["port"] != float64(3306) {
		testRun.Fatalf("entries: %+v", drift)
	}
	if !strings.Contains(stdout, `"notes": []`) {
		testRun.Fatalf("empty arrays must be [] not null:\n%s", stdout)
	}
}

func TestScanBaselineJSONEmptyDriftHasEmptyArrays(testRun *testing.T) {
	procRoot := writeProcWithListeners(testRun)
	baselinePath := storeBaseline(testRun)
	_, stdout, _ := scanWithBaseline(testRun, baselinePath, procRoot, "--format", "json")
	for _, want := range []string{`"new": []`, `"closed": []`, `"notes": []`} {
		if !strings.Contains(stdout, want) {
			testRun.Errorf("json lacks %q:\n%s", want, stdout)
		}
	}
}

func TestScanWithoutBaselineIsUnchanged(testRun *testing.T) {
	procRoot := writeProcWithListeners(testRun, procListener{wildcardHex, 8080})
	for _, format := range []string{"table", "json"} {
		code, stdout, _ := executeScanWith(testRun, &fakeFirewallEvaluator{}, "--no-docker", "--proc-root", procRoot, "--format", format)
		if code != ExitFindings {
			testRun.Fatalf("%s: code=%d", format, code)
		}
		if strings.Contains(stdout, "Drift") || strings.Contains(stdout, `"drift"`) || strings.Contains(stdout, "baseline") {
			testRun.Fatalf("%s: output must not change without --baseline:\n%s", format, stdout)
		}
	}
	_, table, _ := executeScanWith(testRun, &fakeFirewallEvaluator{}, "--no-docker", "--proc-root", procRoot)
	wantTail := "\n1 port(s) scanned, 1 finding(s).\n"
	if !strings.HasSuffix(table, wantTail) {
		testRun.Fatalf("table changed:\n%s", table)
	}
}

func TestScanBaselineCombinesWithFirewall(testRun *testing.T) {
	procRoot := writeProcWithListeners(testRun, procListener{wildcardHex, 8080})
	baselinePath := storeBaseline(testRun)
	evaluator := &fakeFirewallEvaluator{report: allowedFirewallReport()}
	code, stdout, _ := executeScanWith(testRun, evaluator, "--no-docker", "--proc-root", procRoot, "--baseline", baselinePath, "--firewall")
	if code != ExitDrift || evaluator.calls != 1 {
		testRun.Fatalf("code=%d calls=%d", code, evaluator.calls)
	}
	for _, want := range []string{"FIREWALL", "Firewall verdicts are informational", "NEW (1)"} {
		if !strings.Contains(stdout, want) {
			testRun.Errorf("output lacks %q:\n%s", want, stdout)
		}
	}
	_, document, _ := executeScanWith(testRun, evaluator, "--no-docker", "--proc-root", procRoot, "--baseline", baselinePath, "--firewall", "--format", "json")
	var parsed map[string]any
	if err := json.Unmarshal([]byte(document), &parsed); err != nil || parsed["firewall"] == nil || parsed["drift"] == nil {
		testRun.Fatalf("err=%v keys=%v", err, parsed)
	}
}

func TestScanBaselineFlagValidation(testRun *testing.T) {
	var errorOutput bytes.Buffer
	options, err := parseScanOptions([]string{"--baseline", "/tmp/base.json"}, &errorOutput)
	if err != nil || options.BaselineFile != "/tmp/base.json" {
		testRun.Fatalf("options=%+v err=%v", options, err)
	}
	if options, err = parseScanOptions(nil, &errorOutput); err != nil || options.BaselineFile != "" {
		testRun.Fatalf("options=%+v err=%v", options, err)
	}
	if _, err = parseScanOptions([]string{"--baseline", ""}, &errorOutput); err == nil {
		testRun.Fatal("an empty --baseline must be a usage error")
	}
	external := []string{"--baseline", "/tmp/base.json", "--external", "--min-severity", "high", "--firewall", "--no-docker"}
	if options, err = parseScanOptions(external, &errorOutput); err != nil || !options.External || options.MinSeverity != domain.High {
		testRun.Fatalf("options=%+v err=%v", options, err)
	}
}

func TestScanBaselineFailedSourceExitCodes(testRun *testing.T) {
	failedDocker := func(procRoot, baselinePath string) []string {
		return []string{"--proc-root", procRoot, "--docker-socket", missingDockerSocket(testRun), "--baseline", baselinePath}
	}
	testCases := []struct {
		name      string
		listeners []procListener
		entries   []domain.BaselineEntry
		want      ExitCode
	}{
		{"failed docker, no new port", []procListener{{wildcardHex, 8080}}, []domain.BaselineEntry{hostEntry(8080, "0.0.0.0"), dockerEntry(9000, "0.0.0.0")}, ExitRuntimeError},
		{"failed docker, new port wins", []procListener{{wildcardHex, 8080}}, nil, ExitDrift},
		{"failed docker without baseline entries for it", []procListener{{wildcardHex, 8080}}, []domain.BaselineEntry{hostEntry(8080, "0.0.0.0")}, ExitRuntimeError},
	}
	for _, testCase := range testCases {
		testRun.Run(testCase.name, func(testRun *testing.T) {
			procRoot := writeProcWithListeners(testRun, testCase.listeners...)
			baselinePath := storeBaseline(testRun, testCase.entries...)
			code, stdout, stderr := executeScanWith(testRun, &fakeFirewallEvaluator{}, failedDocker(procRoot, baselinePath)...)
			if code != testCase.want {
				testRun.Fatalf("code=%d want %d\n%s", code, testCase.want, stdout)
			}
			if !strings.Contains(stderr, "unverified sources: docker") || !strings.Contains(stdout, "docker") {
				testRun.Fatalf("stderr=%q stdout=%s", stderr, stdout)
			}
		})
	}
	procRoot := writeProcWithListeners(testRun, procListener{wildcardHex, 8080})
	baselinePath := storeBaseline(testRun, hostEntry(8080, "0.0.0.0"))
	if code, _, stderr := scanWithBaseline(testRun, baselinePath, procRoot); code != ExitOK || stderr != "" {
		testRun.Fatalf("all sources fine: code=%d stderr=%q", code, stderr)
	}
}

func TestScanBaselineFailedSourceIsInTheJSONNotes(testRun *testing.T) {
	procRoot := writeProcWithListeners(testRun)
	baselinePath := storeBaseline(testRun)
	_, stdout, _ := executeScanWith(testRun, &fakeFirewallEvaluator{}, "--proc-root", procRoot, "--docker-socket", missingDockerSocket(testRun), "--baseline", baselinePath, "--format", "json")
	var document struct {
		Drift struct {
			Notes []string `json:"notes"`
		} `json:"drift"`
	}
	if err := json.Unmarshal([]byte(stdout), &document); err != nil || len(document.Drift.Notes) != 1 || !strings.Contains(document.Drift.Notes[0], "docker") {
		testRun.Fatalf("err=%v notes=%v", err, document.Drift.Notes)
	}
}
