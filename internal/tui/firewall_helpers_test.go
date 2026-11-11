package tui

import (
	"context"
	"errors"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

var errFirewallDown = errors.New("iptables-save timed out")

// fakeFirewall counts the evaluations a command runs.
type fakeFirewall struct {
	report domain.FirewallReport
	err    error
	calls  int
	ports  []domain.Port
}

func (evaluator *fakeFirewall) Evaluate(_ context.Context, ports []domain.Port) (domain.FirewallReport, error) {
	evaluator.calls++
	evaluator.ports = ports
	return evaluator.report, evaluator.err
}

// testClock is a clock the tests move by hand.
type testClock struct {
	now time.Time
}

func (clock *testClock) read() time.Time { return clock.now }

func (clock *testClock) advance(duration time.Duration) { clock.now = clock.now.Add(duration) }

func firewallFinding(port domain.Port, verdict domain.FirewallVerdict, reason string) domain.FirewallFinding {
	return domain.FirewallFinding{Port: port, Verdict: verdict, Reason: reason}
}

// sampleFirewallReport evaluates the ports of sampleReport: every verdict once
// where possible, using documentation-style reasons without rule text.
func sampleFirewallReport() domain.FirewallReport {
	report := sampleReport()
	verdicts := []struct {
		verdict domain.FirewallVerdict
		reason  string
	}{
		{domain.FirewallAllowed, "INPUT accepts new tcp/8080 from any source"},
		{domain.FirewallBlocked, "INPUT drops tcp/22 by the default policy"},
		{domain.FirewallRestricted, "INPUT accepts tcp/3306 only from 203.0.113.0/24"},
		{domain.FirewallUnknown, "rule uses a match the engine does not evaluate"},
		{domain.FirewallDockerPublished, "published by Docker; usually bypasses INPUT rules"},
	}
	var findings []domain.FirewallFinding
	for index, port := range report.Ports {
		findings = append(findings, firewallFinding(port, verdicts[index].verdict, verdicts[index].reason))
	}
	return domain.FirewallReport{Backend: "iptables-save", Available: true, Findings: findings, Notes: []string{"IPv6 rules were not readable"}}
}

func unavailableFirewallReport() domain.FirewallReport {
	report := sampleReport()
	var findings []domain.FirewallFinding
	for _, port := range report.Ports {
		findings = append(findings, firewallFinding(port, domain.FirewallUnknown, "needs root to read firewall rules"))
	}
	return domain.FirewallReport{Backend: domain.BackendUnknown, Findings: findings, Notes: []string{"needs root to read firewall rules"}}
}

// firewallTestModel builds a sized model with a firewall evaluator, a clock
// the test controls and the sample scan already applied. It returns the
// command that the first scan result started, if any.
func firewallTestModel(testRun *testing.T, evaluator FirewallEvaluator, clock *testClock) (Model, tea.Cmd) {
	testRun.Helper()
	model := New(Options{Scanner: fakeScanner{}, Firewall: evaluator, Interval: testInterval, Version: testVersion, Clock: clock.read})
	model = applyMessage(testRun, model, tea.WindowSizeMsg{Width: testWidth, Height: testHeight})
	updated, command := model.Update(RefreshResultMessage{Report: sampleReportWithQuietPort(), At: clock.now})
	return updated.(Model), command
}

// evaluatedModel is firewallTestModel after the firewall answered with report.
func evaluatedModel(testRun *testing.T, report domain.FirewallReport) Model {
	testRun.Helper()
	clock := &testClock{now: testRefreshTime}
	model, _ := firewallTestModel(testRun, &fakeFirewall{report: report}, clock)
	return applyMessage(testRun, model, FirewallResultMessage{Report: report, At: clock.now})
}

func updateWith(testRun *testing.T, model Model, incoming tea.Msg) (Model, tea.Cmd) {
	testRun.Helper()
	updated, command := model.Update(incoming)
	castModel, isModel := updated.(Model)
	if !isModel {
		testRun.Fatalf("Update returned %T, want Model", updated)
	}
	return castModel, command
}
