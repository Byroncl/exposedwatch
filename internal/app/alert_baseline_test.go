package app

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

func dockerDown() []SourceError {
	return []SourceError{{Source: domain.SourceDocker, Err: errors.New("socket down")}}
}

func TestSourceThatRecoversAfterAPartialFirstScanIsSummarizedOnce(testRun *testing.T) {
	clock := newTestClock()
	evaluator := newTestEvaluator(clock)
	// The first scan has Docker failing: the host baseline is recorded.
	hostFinding := findingOn(22, domain.High, domain.SourceHost)
	first := evaluator.Evaluate(Report{Findings: []domain.Finding{hostFinding}, SourceErrors: dockerDown()})
	if len(first) != 1 || first[0].Kind != domain.AlertKindSummary {
		testRun.Fatalf("%+v", first)
	}
	// Docker recovers with two findings that were already there.
	clock.advance(time.Minute)
	recovered := evaluator.Evaluate(reportOf(
		hostFinding,
		findingOn(6379, domain.High, domain.SourceDocker),
		findingOn(5432, domain.Critical, domain.SourceDocker),
	))
	if len(recovered) != 1 || recovered[0].Kind != domain.AlertKindSummary {
		testRun.Fatalf("want one summary and no individual alerts: %+v", recovered)
	}
	if !strings.Contains(recovered[0].Title, "2 existing findings from docker now visible") || recovered[0].Severity != domain.Critical {
		testRun.Fatalf("%+v", recovered[0])
	}
	// A genuinely new finding afterwards still alerts, once.
	clock.advance(time.Minute)
	genuinelyNew := findingOn(3306, domain.High, domain.SourceDocker)
	later := evaluator.Evaluate(reportOf(hostFinding, findingOn(6379, domain.High, domain.SourceDocker),
		findingOn(5432, domain.Critical, domain.SourceDocker), genuinelyNew))
	if len(later) != 1 || later[0].Kind != domain.AlertKindNewFinding || later[0].Port != 3306 {
		testRun.Fatalf("%+v", later)
	}
}

func TestSourceStillFailingStaysPendingUntilItRecovers(testRun *testing.T) {
	evaluator := newTestEvaluator(newTestClock())
	evaluator.Evaluate(Report{SourceErrors: dockerDown()})
	if alerts := evaluator.Evaluate(Report{SourceErrors: dockerDown()}); len(alerts) != 0 {
		testRun.Fatalf("%+v", alerts)
	}
	alerts := evaluator.Evaluate(reportOf(findingOn(6379, domain.High, domain.SourceDocker)))
	if len(alerts) != 1 || alerts[0].Kind != domain.AlertKindSummary || !strings.Contains(alerts[0].Title, "1 existing finding from docker now visible") {
		testRun.Fatalf("%+v", alerts)
	}
}

func TestRecoveredSourceWithoutFindingsIsSilent(testRun *testing.T) {
	evaluator := newTestEvaluator(newTestClock())
	evaluator.Evaluate(Report{SourceErrors: dockerDown()})
	if alerts := evaluator.Evaluate(Report{}); len(alerts) != 0 {
		testRun.Fatalf("%+v", alerts)
	}
	if alerts := evaluator.Evaluate(reportOf(findingOn(6379, domain.High, domain.SourceDocker))); len(alerts) != 1 || alerts[0].Kind != domain.AlertKindNewFinding {
		testRun.Fatalf("a source baselined as empty must alert on new findings: %+v", alerts)
	}
}

func TestForgottenIdentityAlertsAgain(testRun *testing.T) {
	clock := newTestClock()
	evaluator := newTestEvaluator(clock)
	evaluator.Evaluate(Report{})
	finding := findingOn(5432, domain.High, domain.SourceHost)
	detailed := evaluator.EvaluateDetailed(reportOf(finding))
	if len(detailed) != 1 || len(detailed[0].Identities) != 1 || detailed[0].Identities[0] != FindingIdentityOf(finding) {
		testRun.Fatalf("%+v", detailed)
	}
	if again := evaluator.Evaluate(reportOf(finding)); len(again) != 0 {
		testRun.Fatalf("%+v", again)
	}
	evaluator.Forget(detailed[0].Identities)
	clock.advance(time.Minute) // well inside the cooldown
	if after := evaluator.Evaluate(reportOf(finding)); len(after) != 1 {
		testRun.Fatalf("a forgotten finding must alert again: %+v", after)
	}
}

func TestSummariesCarryNoIdentitiesToForget(testRun *testing.T) {
	evaluator := newTestEvaluator(newTestClock())
	detailed := evaluator.EvaluateDetailed(reportOf(findingOn(22, domain.High, domain.SourceHost)))
	if len(detailed) != 1 || len(detailed[0].Identities) != 0 {
		testRun.Fatalf("%+v", detailed)
	}
}
