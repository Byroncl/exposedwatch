package app

import (
	"strings"
	"testing"
	"time"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

var testBaselineCreatedAt = time.Date(2026, 11, 11, 10, 30, 0, 0, time.UTC)

func savedBaselineOf(entries ...domain.BaselineEntry) domain.Baseline {
	return domain.Baseline{Version: domain.BaselineVersion, CreatedAt: testBaselineCreatedAt, ToolVersion: "test", Entries: entries}
}

func hostEntryOn(port int, bind string) domain.BaselineEntry {
	return domain.BaselineEntry{Protocol: domain.TCP, Port: port, BindAddress: bind, Source: domain.SourceHost}
}

func newBaselineEvaluator(clock *fakeClock, entries ...domain.BaselineEntry) *AlertEvaluator {
	evaluator := newTestEvaluator(clock)
	evaluator.UseBaseline(NewBaselineAcceptance(savedBaselineOf(entries...)))
	return evaluator
}

func TestBaselineModeAcceptedFindingNeverAlerts(testRun *testing.T) {
	clock := newTestClock()
	evaluator := newBaselineEvaluator(clock, hostEntryOn(22, "0.0.0.0"))
	accepted := findingOn(22, domain.Critical, domain.SourceHost)
	for cycle := 0; cycle < 3; cycle++ {
		if alerts := evaluator.Evaluate(reportOf(accepted)); len(alerts) != 0 {
			testRun.Fatalf("cycle %d: %+v", cycle, alerts)
		}
		clock.advance(2 * time.Hour)
	}
}

func TestBaselineModeHasNoSummaryAndAlertsIndividuallyAtFirstEvaluation(testRun *testing.T) {
	evaluator := newBaselineEvaluator(newTestClock(), hostEntryOn(22, "0.0.0.0"))
	alerts := evaluator.Evaluate(reportOf(
		findingOn(22, domain.High, domain.SourceHost),
		findingOn(5432, domain.Critical, domain.SourceHost),
		findingOn(6379, domain.High, domain.SourceHost),
	))
	if len(alerts) != 2 {
		testRun.Fatalf("%+v", alerts)
	}
	for _, alert := range alerts {
		if alert.Kind != domain.AlertKindNewFinding {
			testRun.Fatalf("no summary expected: %+v", alert)
		}
	}
}

func TestBaselineModeBodyLineOnlyOnNonAcceptedFindings(testRun *testing.T) {
	evaluator := newBaselineEvaluator(newTestClock())
	alerts := evaluator.Evaluate(reportOf(findingOn(5432, domain.High, domain.SourceHost)))
	if len(alerts) != 1 {
		testRun.Fatalf("%+v", alerts)
	}
	want := "Fix it.\nNot in baseline (saved 2026-11-11)"
	if alerts[0].Body != want {
		testRun.Fatalf("body %q, want %q", alerts[0].Body, want)
	}
	plain := newTestEvaluator(newTestClock())
	plain.Evaluate(Report{})
	plainAlerts := plain.Evaluate(reportOf(findingOn(5432, domain.High, domain.SourceHost)))
	if strings.Contains(plainAlerts[0].Body, "baseline") {
		testRun.Fatalf("no baseline line without a baseline: %q", plainAlerts[0].Body)
	}
}

func TestBaselineModeBodyLineHasNoControlCharacters(testRun *testing.T) {
	acceptance := NewBaselineAcceptance(savedBaselineOf())
	if strings.ContainsAny(acceptance.NotAcceptedLine(), "\r\t\x1b/") {
		testRun.Fatalf("%q", acceptance.NotAcceptedLine())
	}
}

func TestBaselineModeAcceptsCanonicalBindVariants(testRun *testing.T) {
	evaluator := newBaselineEvaluator(newTestClock(), hostEntryOn(8080, "10.0.0.5"))
	mapped := findingOn(8080, domain.High, domain.SourceHost)
	mapped.Port.BindAddress = "::FFFF:10.0.0.5"
	if alerts := evaluator.Evaluate(reportOf(mapped)); len(alerts) != 0 {
		testRun.Fatalf("IPv4-mapped IPv6 form must match its canonical entry: %+v", alerts)
	}
	wildcard := newBaselineEvaluator(newTestClock(), hostEntryOn(80, "0.0.0.0"))
	empty := findingOn(80, domain.High, domain.SourceHost)
	empty.Port.BindAddress = ""
	if alerts := wildcard.Evaluate(reportOf(empty)); len(alerts) != 0 {
		testRun.Fatalf("empty bind is the wildcard: %+v", alerts)
	}
}

func TestBaselineEntryNeverSuppressesADifferentPortIdentity(testRun *testing.T) {
	base := hostEntryOn(5432, "127.0.0.1")
	testCases := []struct {
		name   string
		mutate func(finding *domain.Finding)
	}{
		{"different bind", func(finding *domain.Finding) { finding.Port.BindAddress = "0.0.0.0" }},
		{"different source", func(finding *domain.Finding) { finding.Port.Source = domain.SourceDocker }},
		{"different port", func(finding *domain.Finding) { finding.Port.Number = 5433 }},
		{"different protocol", func(finding *domain.Finding) { finding.Port.Protocol = domain.UDP }},
		{"external source", func(finding *domain.Finding) { finding.Port.Source = domain.SourceExternal }},
	}
	for _, testCase := range testCases {
		testRun.Run(testCase.name, func(testRun *testing.T) {
			evaluator := newBaselineEvaluator(newTestClock(), base)
			finding := findingOn(5432, domain.High, domain.SourceHost)
			finding.Port.BindAddress = "127.0.0.1"
			testCase.mutate(&finding)
			if alerts := evaluator.Evaluate(reportOf(finding)); len(alerts) != 1 {
				testRun.Fatalf("must alert: %+v", alerts)
			}
		})
	}
}

func TestBaselineModeSeverityThresholdStillApplies(testRun *testing.T) {
	evaluator := newBaselineEvaluator(newTestClock())
	if alerts := evaluator.Evaluate(reportOf(findingOn(8080, domain.Low, domain.SourceHost))); len(alerts) != 0 {
		testRun.Fatalf("%+v", alerts)
	}
}

func TestBaselineModeDedupeCooldownAndForgetStillWork(testRun *testing.T) {
	clock := newTestClock()
	evaluator := newBaselineEvaluator(clock)
	finding := findingOn(5432, domain.High, domain.SourceHost)
	detailed := evaluator.EvaluateDetailed(reportOf(finding))
	if len(detailed) != 1 {
		testRun.Fatalf("%+v", detailed)
	}
	if again := evaluator.Evaluate(reportOf(finding)); len(again) != 0 {
		testRun.Fatalf("dedupe: %+v", again)
	}
	// Disappears and returns inside the cooldown: silent.
	clock.advance(time.Minute)
	evaluator.Evaluate(Report{})
	clock.advance(time.Minute)
	if back := evaluator.Evaluate(reportOf(finding)); len(back) != 0 {
		testRun.Fatalf("cooldown: %+v", back)
	}
	evaluator.Forget(detailed[0].Identities)
	if after := evaluator.Evaluate(reportOf(finding)); len(after) != 1 {
		testRun.Fatalf("forgotten finding must alert again: %+v", after)
	}
}

func TestFindingThatBecomesAcceptedAfterReloadIsForgottenSilently(testRun *testing.T) {
	clock := newTestClock()
	evaluator := newBaselineEvaluator(clock)
	finding := findingOn(5432, domain.High, domain.SourceHost)
	if alerts := evaluator.Evaluate(reportOf(finding)); len(alerts) != 1 {
		testRun.Fatalf("%+v", alerts)
	}
	evaluator.UseBaseline(NewBaselineAcceptance(savedBaselineOf(hostEntryOn(5432, "0.0.0.0"))))
	clock.advance(time.Minute)
	if alerts := evaluator.Evaluate(reportOf(finding)); len(alerts) != 0 {
		testRun.Fatalf("accepted finding must stay silent: %+v", alerts)
	}
	// Dropping the port from a later baseline makes it new again at once.
	evaluator.UseBaseline(NewBaselineAcceptance(savedBaselineOf()))
	clock.advance(time.Minute)
	if alerts := evaluator.Evaluate(reportOf(finding)); len(alerts) != 1 {
		testRun.Fatalf("no longer accepted: %+v", alerts)
	}
}

func TestBaselineModePartialSourceDoesNotFlapOrSummarize(testRun *testing.T) {
	clock := newTestClock()
	evaluator := newBaselineEvaluator(clock)
	docker := findingOn(6379, domain.High, domain.SourceDocker)
	if alerts := evaluator.Evaluate(reportOf(docker)); len(alerts) != 1 {
		testRun.Fatalf("%+v", alerts)
	}
	clock.advance(3 * time.Hour)
	if alerts := evaluator.Evaluate(Report{SourceErrors: dockerDown()}); len(alerts) != 0 {
		testRun.Fatalf("%+v", alerts)
	}
	clock.advance(3 * time.Hour)
	if alerts := evaluator.Evaluate(reportOf(docker)); len(alerts) != 0 {
		testRun.Fatalf("a source that came back must not re-alert nor summarize: %+v", alerts)
	}
}

func TestEvaluatorWithoutBaselineKeepsSummaryBehavior(testRun *testing.T) {
	evaluator := newTestEvaluator(newTestClock())
	evaluator.UseBaseline(nil)
	alerts := evaluator.Evaluate(reportOf(findingOn(22, domain.High, domain.SourceHost)))
	if len(alerts) != 1 || alerts[0].Kind != domain.AlertKindSummary {
		testRun.Fatalf("%+v", alerts)
	}
}
