package tui

import (
	"strings"
	"testing"

	"github.com/Byroncl/exposedwatch/internal/app"
	"github.com/Byroncl/exposedwatch/internal/domain"
)

const freshPortNumber = 9999

func findRow(testRun *testing.T, model Model, number int) displayRow {
	testRun.Helper()
	for _, row := range model.visibleRows {
		if row.port.Number == number {
			return row
		}
	}
	testRun.Fatalf("no visible row for port %d in %v", number, visiblePorts(model))
	return displayRow{}
}

func TestFirstSuccessfulScanIsTheBaselineAndFlagsNothing(testRun *testing.T) {
	model := newTestModel(testRun)
	for _, row := range model.visibleRows {
		if row.isNew {
			testRun.Errorf("port %d flagged NEW on the baseline", row.port.Number)
		}
	}
	if strings.Contains(model.View(), "NEW") || !strings.Contains(model.View(), "+0 new") {
		testRun.Fatalf("baseline view must show +0 new and no badge:\n%s", model.View())
	}
}

func TestOpenedPortGetsNewBadgeAndCounter(testRun *testing.T) {
	model := newTestModel(testRun)
	model = refreshWith(testRun, model, reportWith(sampleReport(), freshPortNumber), 1)
	if !findRow(testRun, model, freshPortNumber).isNew {
		testRun.Fatal("opened port must be flagged NEW")
	}
	rendered := model.View()
	if !strings.Contains(rendered, "NEW") || !strings.Contains(rendered, "+1 new") {
		testRun.Fatalf("view lacks NEW badge or counter:\n%s", rendered)
	}
	if findRow(testRun, model, 22).isNew {
		testRun.Fatal("pre-existing port must not be NEW")
	}
}

func TestNewBadgeExpiresAfterConfiguredRefreshes(testRun *testing.T) {
	model := newTestModel(testRun)
	report := reportWith(sampleReport(), freshPortNumber)
	model = refreshWith(testRun, model, report, 1)
	for refresh := 2; refresh <= newBadgeRefreshCount; refresh++ {
		model = refreshWith(testRun, model, report, refresh)
		if !findRow(testRun, model, freshPortNumber).isNew {
			testRun.Fatalf("badge expired too early at refresh %d", refresh)
		}
	}
	model = refreshWith(testRun, model, report, newBadgeRefreshCount+1)
	if findRow(testRun, model, freshPortNumber).isNew {
		testRun.Fatal("badge must expire")
	}
}

func TestClosedPortGoesToTheChangeLog(testRun *testing.T) {
	model := newTestModel(testRun)
	model = refreshWith(testRun, model, reportWithout(sampleReport(), 3306), 1)
	if len(model.tracker.log) != 1 {
		testRun.Fatalf("log = %+v", model.tracker.log)
	}
	entry := model.tracker.log[0]
	if entry.Change.Kind != app.ChangeClosed || entry.Change.Port.Number != 3306 ||
		entry.Severity != domain.Critical || entry.Change.Port.Process != "mysqld" {
		testRun.Fatalf("entry = %+v", entry)
	}
	model = typeKeys(testRun, model, "3")
	rendered := model.View()
	for _, expected := range []string{"CLOSED", "3306", "mysqld", "CRITICAL", "03:05:05"} {
		if !strings.Contains(rendered, expected) {
			testRun.Errorf("changes tab lacks %q:\n%s", expected, rendered)
		}
	}
	if !strings.Contains(rendered, "-1 closed") {
		testRun.Errorf("header lacks -1 closed:\n%s", rendered)
	}
}

func TestChangesTabListsNewestFirst(testRun *testing.T) {
	model := newTestModel(testRun)
	model = refreshWith(testRun, model, reportWithout(sampleReport(), 3306), 1)
	model = refreshWith(testRun, model, reportWithout(sampleReport(), 3306, 53), 2)
	model = typeKeys(testRun, model, "3")
	if !equalInts(visiblePorts(model), []int{53, 3306}) {
		testRun.Fatalf("order = %v, want newest first", visiblePorts(model))
	}
}

func TestFailedScanCreatesNoClosedEvents(testRun *testing.T) {
	model := newTestModel(testRun)
	model = applyMessage(testRun, model, RefreshResultMessage{Err: errDockerDown, At: timeAfter(1)})
	if len(model.tracker.log) != 0 || model.tracker.closedTotal != 0 {
		testRun.Fatalf("failed scan created events: %+v", model.tracker.log)
	}
	model = refreshWith(testRun, model, sampleReport(), 2)
	if len(model.tracker.log) != 0 {
		testRun.Fatalf("recovery created events: %+v", model.tracker.log)
	}
}

func TestPartialFailureCreatesNoClosedEvents(testRun *testing.T) {
	dockerFinding := makeFinding(8080, domain.Low, "webapp")
	dockerFinding.Port.Source = domain.SourceDocker
	baseline := sampleReport()
	baseline.Ports = append(baseline.Ports, dockerFinding.Port)
	model := New(Options{Scanner: fakeScanner{}, Interval: testInterval, Version: testVersion})
	model = refreshWith(testRun, model, baseline, 0)
	partial := reportWithout(sampleReport(), 3306)
	partial.SourceErrors = []app.SourceError{{Source: domain.SourceDocker, Err: errDockerDown}}
	model = refreshWith(testRun, model, partial, 2)
	if len(model.tracker.log) != 1 || model.tracker.log[0].Change.Port.Number != 3306 {
		testRun.Fatalf("log = %+v, want only the host port 3306 closed", model.tracker.log)
	}
}

func TestDetailPaneShowsFirstSeenAndNewState(testRun *testing.T) {
	model := newTestModel(testRun)
	model = refreshWith(testRun, model, reportWith(sampleReport(), freshPortNumber), 1)
	model = typeKeys(testRun, model, "/", "f", "r", "e", "s", "h", "enter")
	rendered := model.View()
	for _, expected := range []string{"First seen", "03:05:05", "NEW", "Title for fresh", "RULE-fresh"} {
		if !strings.Contains(rendered, expected) {
			testRun.Errorf("detail lacks %q:\n%s", expected, rendered)
		}
	}
}
