package main

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

type fakeFirewallEvaluator struct {
	report domain.FirewallReport
	err    error
	calls  int
	ports  []domain.Port
}

func (evaluator *fakeFirewallEvaluator) Evaluate(_ context.Context, ports []domain.Port) (domain.FirewallReport, error) {
	evaluator.calls++
	evaluator.ports = ports
	return evaluator.report, evaluator.err
}

func executeScanWith(testRun *testing.T, evaluator firewallEvaluator, arguments ...string) (ExitCode, string, string) {
	testRun.Helper()
	var stdout, stderr bytes.Buffer
	code := executeScan(arguments, &stdout, &stderr, scanDependencies{firewall: func() firewallEvaluator { return evaluator }})
	return code, stdout.String(), stderr.String()
}

func fakeProcScanArguments(testRun *testing.T, extra ...string) []string {
	testRun.Helper()
	procRoot := testRun.TempDir()
	writeFakeProc(testRun, procRoot)
	return append([]string{"--no-docker", "--proc-root", procRoot}, extra...)
}

func allowedFirewallReport() domain.FirewallReport {
	port := domain.Port{Number: 8080, Protocol: domain.TCP, BindAddress: "0.0.0.0", Source: domain.SourceHost}
	return domain.FirewallReport{
		Backend: "iptables-save", Available: true,
		Findings: []domain.FirewallFinding{{Port: port, Verdict: domain.FirewallAllowed, Reason: "INPUT accepts new tcp/8080"}},
	}
}

func TestFirewallFlagIsOffByDefault(testRun *testing.T) {
	options, err := parseScanOptions(nil, &bytes.Buffer{})
	if err != nil || options.Firewall {
		testRun.Fatalf("options=%+v err=%v", options, err)
	}
	options, err = parseScanOptions([]string{"--firewall"}, &bytes.Buffer{})
	if err != nil || !options.Firewall {
		testRun.Fatalf("options=%+v err=%v", options, err)
	}
}

func TestScanWithoutFirewallFlagNeverReadsTheFirewall(testRun *testing.T) {
	evaluator := &fakeFirewallEvaluator{report: allowedFirewallReport()}
	code, stdout, _ := executeScanWith(testRun, evaluator, fakeProcScanArguments(testRun)...)
	if code != ExitFindings || evaluator.calls != 0 {
		testRun.Fatalf("code=%d calls=%d", code, evaluator.calls)
	}
	if strings.Contains(stdout, "FIREWALL") || strings.Contains(stdout, "informational") {
		testRun.Fatalf("output must be unchanged without --firewall:\n%s", stdout)
	}
	_, document, _ := executeScanWith(testRun, evaluator, fakeProcScanArguments(testRun, "--format", "json")...)
	if strings.Contains(document, `"firewall"`) {
		testRun.Fatalf("json must be unchanged without --firewall:\n%s", document)
	}
}

func TestScanWithFirewallShowsVerdictsInTableAndJSON(testRun *testing.T) {
	evaluator := &fakeFirewallEvaluator{report: allowedFirewallReport()}
	code, stdout, _ := executeScanWith(testRun, evaluator, fakeProcScanArguments(testRun, "--firewall")...)
	if code != ExitFindings || evaluator.calls != 1 || len(evaluator.ports) != 1 || evaluator.ports[0].Number != 8080 {
		testRun.Fatalf("code=%d calls=%d ports=%v", code, evaluator.calls, evaluator.ports)
	}
	for _, want := range []string{"FIREWALL", "allowed", "backend: iptables-save", "available: yes", "Firewall verdicts are informational"} {
		if !strings.Contains(stdout, want) {
			testRun.Errorf("table lacks %q:\n%s", want, stdout)
		}
	}
	_, document, _ := executeScanWith(testRun, evaluator, fakeProcScanArguments(testRun, "--firewall", "--format", "json")...)
	for _, want := range []string{`"firewall": {`, `"verdict": "allowed"`, `"available": true`} {
		if !strings.Contains(document, want) {
			testRun.Errorf("json lacks %q:\n%s", want, document)
		}
	}
}

func TestScanWithAnUnavailableFirewallIsNotFatalAndKeepsTheExitCode(testRun *testing.T) {
	unavailable := domain.FirewallReport{
		Backend: domain.BackendUnknown, Notes: []string{"needs root to read firewall rules"},
		Findings: []domain.FirewallFinding{{
			Port:    domain.Port{Number: 8080, Protocol: domain.TCP, BindAddress: "0.0.0.0", Source: domain.SourceHost},
			Verdict: domain.FirewallUnknown, Reason: "needs root to read firewall rules",
		}},
	}
	code, stdout, stderr := executeScanWith(testRun, &fakeFirewallEvaluator{report: unavailable}, fakeProcScanArguments(testRun, "--firewall")...)
	if code != ExitFindings || stderr != "" {
		testRun.Fatalf("code=%d stderr=%q", code, stderr)
	}
	for _, want := range []string{"unknown", "available: no", "needs root to read firewall rules"} {
		if !strings.Contains(stdout, want) {
			testRun.Errorf("output lacks %q:\n%s", want, stdout)
		}
	}
	clean, _, _ := executeScanWith(testRun, &fakeFirewallEvaluator{report: unavailable}, fakeProcScanArguments(testRun, "--firewall", "--min-severity", "critical")...)
	if clean != ExitOK {
		testRun.Fatalf("exit code must not depend on the firewall, got %d", clean)
	}
}

func TestScanTurnsAnEvaluationErrorIntoAnUnknownNote(testRun *testing.T) {
	evaluator := &fakeFirewallEvaluator{err: errors.New("iptables-save exploded")}
	code, stdout, stderr := executeScanWith(testRun, evaluator, fakeProcScanArguments(testRun, "--firewall")...)
	if code != ExitFindings || stderr != "" {
		testRun.Fatalf("code=%d stderr=%q", code, stderr)
	}
	for _, want := range []string{"unknown", "available: no", "could not be read", "iptables-save exploded"} {
		if !strings.Contains(stdout, want) {
			testRun.Errorf("output lacks %q:\n%s", want, stdout)
		}
	}
}

func TestScanFirewallMissingProcStillFailsWithRuntimeError(testRun *testing.T) {
	evaluator := &fakeFirewallEvaluator{report: allowedFirewallReport()}
	code, _, stderr := executeScanWith(testRun, evaluator, "--no-docker", "--proc-root", testRun.TempDir(), "--firewall")
	if code != ExitRuntimeError || stderr == "" || evaluator.calls != 0 {
		testRun.Fatalf("code=%d stderr=%q calls=%d", code, stderr, evaluator.calls)
	}
}

func TestUsageMentionsTheFirewallFlags(testRun *testing.T) {
	_, stdout, _ := runCommand("help")
	for _, want := range []string{"--firewall", "--no-firewall"} {
		if !strings.Contains(stdout, want) {
			testRun.Errorf("usage lacks %s: %q", want, stdout)
		}
	}
}
