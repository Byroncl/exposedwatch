package app

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

type fakeClock struct{ current time.Time }

func (clock *fakeClock) Now() time.Time                 { return clock.current }
func (clock *fakeClock) advance(duration time.Duration) { clock.current = clock.current.Add(duration) }

func newTestClock() *fakeClock {
	return &fakeClock{current: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)}
}

func findingOn(port int, severity domain.Severity, source domain.Source) domain.Finding {
	return domain.Finding{
		Port: domain.Port{
			Number: port, Protocol: domain.TCP, BindAddress: "0.0.0.0", Process: "svc", Source: source,
		},
		Severity:       severity,
		RuleID:         "rule-x",
		Title:          "Title of port " + string(rune('0'+port%10)),
		Recommendation: "Fix it.",
	}
}

func reportOf(findings ...domain.Finding) Report {
	return Report{Findings: findings}
}

func newTestEvaluator(clock *fakeClock) *AlertEvaluator {
	return NewAlertEvaluator(AlertPolicy{MinSeverity: domain.High, Cooldown: time.Hour}, clock, "web-01")
}

func TestFirstEvaluationSendsOneSummaryAndNoIndividualAlerts(testRun *testing.T) {
	clock := newTestClock()
	evaluator := newTestEvaluator(clock)
	existing := reportOf(
		findingOn(5432, domain.Critical, domain.SourceHost),
		findingOn(6379, domain.High, domain.SourceHost),
		findingOn(8080, domain.Low, domain.SourceHost),
	)
	alerts := evaluator.Evaluate(existing)
	if len(alerts) != 1 || alerts[0].Kind != domain.AlertKindSummary {
		testRun.Fatalf("%+v", alerts)
	}
	if alerts[0].Severity != domain.Critical || alerts[0].Host != "web-01" {
		testRun.Fatalf("%+v", alerts[0])
	}
	if !strings.Contains(alerts[0].Title, "2 findings at or above high") {
		testRun.Fatalf("title %q", alerts[0].Title)
	}
	if again := evaluator.Evaluate(existing); len(again) != 0 {
		testRun.Fatalf("known findings must stay silent: %+v", again)
	}
}

func TestFirstEvaluationWithoutQualifyingFindingsIsSilent(testRun *testing.T) {
	evaluator := newTestEvaluator(newTestClock())
	if alerts := evaluator.Evaluate(reportOf(findingOn(80, domain.Low, domain.SourceHost))); len(alerts) != 0 {
		testRun.Fatalf("%+v", alerts)
	}
	if alerts := evaluator.Evaluate(Report{}); len(alerts) != 0 {
		testRun.Fatalf("%+v", alerts)
	}
}

func TestSummaryListsAtMostFiveTitles(testRun *testing.T) {
	evaluator := newTestEvaluator(newTestClock())
	var findings []domain.Finding
	for port := 1; port <= 8; port++ {
		findings = append(findings, findingOn(port, domain.High, domain.SourceHost))
	}
	alerts := evaluator.Evaluate(reportOf(findings...))
	if len(alerts) != 1 {
		testRun.Fatalf("%+v", alerts)
	}
	if got := strings.Count(alerts[0].Body, "\n- "); got != MaxSummaryTitles-1 {
		testRun.Fatalf("listed %d titles after the first: %q", got, alerts[0].Body)
	}
	if !strings.Contains(alerts[0].Body, "and 3 more") {
		testRun.Fatalf("body %q", alerts[0].Body)
	}
}

func TestSingularSummaryTitle(testRun *testing.T) {
	alerts := newTestEvaluator(newTestClock()).Evaluate(reportOf(findingOn(80, domain.High, domain.SourceHost)))
	if len(alerts) != 1 || !strings.HasPrefix(alerts[0].Title, "1 finding at or above") {
		testRun.Fatalf("%+v", alerts)
	}
}

func TestLaterEvaluationAlertsOncePerNewFinding(testRun *testing.T) {
	clock := newTestClock()
	evaluator := newTestEvaluator(clock)
	evaluator.Evaluate(Report{}) // baseline: nothing existing
	opened := findingOn(5432, domain.Critical, domain.SourceHost)
	alerts := evaluator.Evaluate(reportOf(opened, findingOn(80, domain.Low, domain.SourceHost)))
	if len(alerts) != 1 {
		testRun.Fatalf("%+v", alerts)
	}
	alert := alerts[0]
	if alert.Kind != domain.AlertKindNewFinding || alert.Port != 5432 || alert.Protocol != domain.TCP ||
		alert.Process != "svc" || alert.RuleID != "rule-x" || alert.Body != "Fix it." || alert.Host != "web-01" ||
		alert.Exposure != domain.AllInterfaces || alert.Severity != domain.Critical {
		testRun.Fatalf("%+v", alert)
	}
	if again := evaluator.Evaluate(reportOf(opened)); len(again) != 0 {
		testRun.Fatalf("still-present finding alerted again: %+v", again)
	}
}

func TestIdentityDistinguishesPortBindProtocolSourceAndRule(testRun *testing.T) {
	clock := newTestClock()
	evaluator := newTestEvaluator(clock)
	evaluator.Evaluate(Report{})
	base := findingOn(5432, domain.High, domain.SourceHost)
	evaluator.Evaluate(reportOf(base))

	otherBind := base
	otherBind.Port.BindAddress = "10.0.0.5"
	otherProtocol := base
	otherProtocol.Port.Protocol = domain.UDP
	otherSource := base
	otherSource.Port.Source = domain.SourceDocker
	otherRule := base
	otherRule.RuleID = "another-rule"
	otherPort := findingOn(5433, domain.High, domain.SourceHost)

	for name, variant := range map[string]domain.Finding{
		"bind": otherBind, "protocol": otherProtocol, "source": otherSource, "rule": otherRule, "port": otherPort,
	} {
		if alerts := evaluator.Evaluate(reportOf(base, variant)); len(alerts) != 1 {
			testRun.Errorf("%s: got %d alerts", name, len(alerts))
		}
	}
}

func TestProcessNameChangeIsNotANewFinding(testRun *testing.T) {
	evaluator := newTestEvaluator(newTestClock())
	evaluator.Evaluate(Report{})
	first := findingOn(5432, domain.High, domain.SourceHost)
	evaluator.Evaluate(reportOf(first))
	renamed := first
	renamed.Port.Process = "other"
	if alerts := evaluator.Evaluate(reportOf(renamed)); len(alerts) != 0 {
		testRun.Fatalf("%+v", alerts)
	}
}

func TestReappearanceRespectsCooldown(testRun *testing.T) {
	testCases := []struct {
		name       string
		absentFor  time.Duration
		wantAlerts int
	}{
		{"within cooldown stays quiet", 30 * time.Minute, 0},
		{"exactly at cooldown alerts", time.Hour, 1},
		{"after cooldown alerts", 2 * time.Hour, 1},
	}
	for _, testCase := range testCases {
		testRun.Run(testCase.name, func(testRun *testing.T) {
			clock := newTestClock()
			evaluator := newTestEvaluator(clock)
			evaluator.Evaluate(Report{})
			flapping := findingOn(5432, domain.High, domain.SourceHost)
			if len(evaluator.Evaluate(reportOf(flapping))) != 1 {
				testRun.Fatal("first appearance must alert")
			}
			evaluator.Evaluate(Report{}) // closed: forgotten
			clock.advance(testCase.absentFor)
			if got := len(evaluator.Evaluate(reportOf(flapping))); got != testCase.wantAlerts {
				testRun.Fatalf("got %d alerts", got)
			}
		})
	}
}

func TestZeroCooldownAlertsOnEveryReappearance(testRun *testing.T) {
	clock := newTestClock()
	evaluator := NewAlertEvaluator(AlertPolicy{MinSeverity: domain.High}, clock, "web-01")
	evaluator.Evaluate(Report{})
	flapping := findingOn(5432, domain.High, domain.SourceHost)
	for round := 0; round < 3; round++ {
		if len(evaluator.Evaluate(reportOf(flapping))) != 1 {
			testRun.Fatalf("round %d", round)
		}
		evaluator.Evaluate(Report{})
	}
}

func TestPartialReportDoesNotForgetFailedSource(testRun *testing.T) {
	clock := newTestClock()
	evaluator := newTestEvaluator(clock)
	evaluator.Evaluate(Report{})
	dockerFinding := findingOn(6379, domain.High, domain.SourceDocker)
	evaluator.Evaluate(reportOf(dockerFinding))

	clock.advance(3 * time.Hour)
	partial := Report{SourceErrors: []SourceError{{Source: domain.SourceDocker, Err: errors.New("socket down")}}}
	if alerts := evaluator.Evaluate(partial); len(alerts) != 0 {
		testRun.Fatalf("%+v", alerts)
	}
	// Docker recovers much later: no flapping alert, the finding never left.
	clock.advance(3 * time.Hour)
	if alerts := evaluator.Evaluate(reportOf(dockerFinding)); len(alerts) != 0 {
		testRun.Fatalf("flapping alert: %+v", alerts)
	}
}

func TestHealthySourceStillForgetsClosedFindings(testRun *testing.T) {
	clock := newTestClock()
	evaluator := newTestEvaluator(clock)
	evaluator.Evaluate(Report{})
	hostFinding := findingOn(22, domain.High, domain.SourceHost)
	evaluator.Evaluate(reportOf(hostFinding))
	partial := Report{SourceErrors: []SourceError{{Source: domain.SourceDocker, Err: errors.New("down")}}}
	evaluator.Evaluate(partial) // host healthy, finding closed
	clock.advance(2 * time.Hour)
	if alerts := evaluator.Evaluate(reportOf(hostFinding)); len(alerts) != 1 {
		testRun.Fatalf("%+v", alerts)
	}
}

func TestDuplicateFindingsInOneReportAlertOnce(testRun *testing.T) {
	evaluator := newTestEvaluator(newTestClock())
	evaluator.Evaluate(Report{})
	duplicate := findingOn(5432, domain.High, domain.SourceHost)
	if alerts := evaluator.Evaluate(reportOf(duplicate, duplicate)); len(alerts) != 1 {
		testRun.Fatalf("%+v", alerts)
	}
}

func TestMinSeverityIsInclusive(testRun *testing.T) {
	evaluator := newTestEvaluator(newTestClock())
	evaluator.Evaluate(Report{})
	alerts := evaluator.Evaluate(reportOf(
		findingOn(1, domain.Medium, domain.SourceHost),
		findingOn(2, domain.High, domain.SourceHost),
	))
	if len(alerts) != 1 || alerts[0].Port != 2 {
		testRun.Fatalf("%+v", alerts)
	}
}
