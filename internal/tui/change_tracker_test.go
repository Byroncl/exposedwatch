package tui

import (
	"testing"
	"time"

	"github.com/Byroncl/exposedwatch/internal/app"
	"github.com/Byroncl/exposedwatch/internal/domain"
)

func hostPort(number int) domain.Port {
	return domain.Port{Number: number, Protocol: domain.TCP, BindAddress: "0.0.0.0", Source: domain.SourceHost, Process: "proc"}
}

func dockerPort(number int) domain.Port {
	return domain.Port{Number: number, Protocol: domain.TCP, BindAddress: "0.0.0.0", Source: domain.SourceDocker, Container: "box"}
}

func reportOfPorts(ports ...domain.Port) app.Report {
	return app.Report{Ports: ports}
}

func timeAfter(offset int) time.Time {
	return testRefreshTime.Add(time.Duration(offset) * time.Minute)
}

func TestTrackerFirstSnapshotIsBaselineWithoutChanges(testRun *testing.T) {
	tracker := newChangeTracker().observe(reportOfPorts(hostPort(22), hostPort(80)), timeAfter(0))
	if !tracker.hasBaseline || len(tracker.log) != 0 || tracker.openedTotal != 0 || tracker.closedTotal != 0 {
		testRun.Fatalf("baseline produced changes: %+v", tracker)
	}
	if tracker.isNew(app.IdentityOf(hostPort(22))) {
		testRun.Fatal("baseline ports must not be NEW")
	}
	firstSeen, known := tracker.firstSeenAt(app.IdentityOf(hostPort(22)))
	if !known || !firstSeen.Equal(timeAfter(0)) {
		testRun.Fatalf("first seen = %v known=%v", firstSeen, known)
	}
}

func TestTrackerOpenedPortIsNewForConfiguredRefreshCount(testRun *testing.T) {
	tracker := newChangeTracker().observe(reportOfPorts(hostPort(22)), timeAfter(0))
	opened := hostPort(9000)
	tracker = tracker.observe(reportOfPorts(hostPort(22), opened), timeAfter(1))
	identity := app.IdentityOf(opened)
	if !tracker.isNew(identity) || tracker.openedTotal != 1 {
		testRun.Fatalf("detection refresh: new=%v opened=%d", tracker.isNew(identity), tracker.openedTotal)
	}
	for refresh := 2; refresh <= newBadgeRefreshCount; refresh++ {
		tracker = tracker.observe(reportOfPorts(hostPort(22), opened), timeAfter(refresh))
		if !tracker.isNew(identity) {
			testRun.Fatalf("port expired too early at refresh %d", refresh)
		}
	}
	tracker = tracker.observe(reportOfPorts(hostPort(22), opened), timeAfter(newBadgeRefreshCount+1))
	if tracker.isNew(identity) {
		testRun.Fatalf("port still NEW after %d refreshes", newBadgeRefreshCount)
	}
	firstSeen, _ := tracker.firstSeenAt(identity)
	if !firstSeen.Equal(timeAfter(1)) {
		testRun.Fatalf("first seen = %v, want detection time", firstSeen)
	}
}

func TestTrackerClosedPortIsLoggedWithLastKnownSeverity(testRun *testing.T) {
	closing := hostPort(3306)
	baseline := app.Report{
		Ports:    []domain.Port{hostPort(22), closing},
		Findings: []domain.Finding{{Port: closing, Severity: domain.High}},
	}
	tracker := newChangeTracker().observe(baseline, timeAfter(0))
	tracker = tracker.observe(reportOfPorts(hostPort(22)), timeAfter(1))
	if len(tracker.log) != 1 || tracker.closedTotal != 1 {
		testRun.Fatalf("log = %+v closed=%d", tracker.log, tracker.closedTotal)
	}
	entry := tracker.log[0]
	if entry.Change.Kind != app.ChangeClosed || entry.Change.Port.Number != 3306 ||
		!entry.HasSeverity || entry.Severity != domain.High || !entry.At.Equal(timeAfter(1)) {
		testRun.Fatalf("entry = %+v", entry)
	}
	if _, known := tracker.firstSeenAt(app.IdentityOf(closing)); known {
		testRun.Fatal("closed port must be forgotten")
	}
}

func TestTrackerSeverityIsUnknownWhenPortHadNoFinding(testRun *testing.T) {
	tracker := newChangeTracker().observe(reportOfPorts(hostPort(22)), timeAfter(0))
	tracker = tracker.observe(reportOfPorts(), timeAfter(1))
	if len(tracker.log) != 1 || tracker.log[0].HasSeverity {
		testRun.Fatalf("log = %+v", tracker.log)
	}
}

func TestTrackerPartialSourceFailureCreatesNoClosedEvents(testRun *testing.T) {
	tracker := newChangeTracker().observe(reportOfPorts(hostPort(22), dockerPort(8080)), timeAfter(0))
	partial := reportOfPorts(hostPort(22))
	partial.SourceErrors = []app.SourceError{{Source: domain.SourceDocker, Err: errDockerDown}}
	tracker = tracker.observe(partial, timeAfter(1))
	if len(tracker.log) != 0 || tracker.closedTotal != 0 {
		testRun.Fatalf("false closed events: %+v", tracker.log)
	}
	tracker = tracker.observe(reportOfPorts(hostPort(22), dockerPort(8080)), timeAfter(2))
	if len(tracker.log) != 0 || tracker.openedTotal != 0 {
		testRun.Fatalf("recovery must not look like an opened port: %+v", tracker.log)
	}
	firstSeen, _ := tracker.firstSeenAt(app.IdentityOf(dockerPort(8080)))
	if !firstSeen.Equal(timeAfter(0)) {
		testRun.Fatalf("first seen must survive the outage, got %v", firstSeen)
	}
}

func TestTrackerPartialFailureStillReportsClosedPortsOfHealthySources(testRun *testing.T) {
	tracker := newChangeTracker().observe(reportOfPorts(hostPort(22), hostPort(80), dockerPort(8080)), timeAfter(0))
	partial := reportOfPorts(hostPort(22))
	partial.SourceErrors = []app.SourceError{{Source: domain.SourceDocker, Err: errDockerDown}}
	tracker = tracker.observe(partial, timeAfter(1))
	if len(tracker.log) != 1 || tracker.log[0].Change.Port.Number != 80 {
		testRun.Fatalf("log = %+v, want only port 80 closed", tracker.log)
	}
}

func TestTrackerLogIsBoundedAndDropsOldest(testRun *testing.T) {
	tracker := newChangeTracker().observe(reportOfPorts(), timeAfter(0))
	total := maxChangeLogEntries + 10
	for number := 1; number <= total; number++ {
		tracker = tracker.observe(reportOfPorts(hostPort(number)), timeAfter(number))
	}
	if len(tracker.log) != maxChangeLogEntries {
		testRun.Fatalf("log length = %d, want %d", len(tracker.log), maxChangeLogEntries)
	}
	if tracker.openedTotal != total {
		testRun.Fatalf("opened total = %d, want %d (counters are not capped)", tracker.openedTotal, total)
	}
}

func TestTrackerObserveDoesNotMutateThePreviousTracker(testRun *testing.T) {
	original := newChangeTracker().observe(reportOfPorts(hostPort(22)), timeAfter(0))
	_ = original.observe(reportOfPorts(), timeAfter(1))
	if len(original.log) != 0 || len(original.previousPorts) != 1 {
		testRun.Fatalf("original mutated: %+v", original)
	}
}
