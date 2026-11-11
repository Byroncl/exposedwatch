package tui

import (
	"testing"
	"time"

	"github.com/Byroncl/exposedwatch/internal/app"
	"github.com/Byroncl/exposedwatch/internal/domain"
)

func TestConnectionTrackerBaselineFlagsNothing(testRun *testing.T) {
	tracker := newConnectionTracker().observe(fullConnections(), testRefreshTime)
	for _, connection := range fullConnections() {
		identity := app.ConnectionIdentityOf(connection)
		if tracker.isNew(identity) {
			testRun.Errorf("baseline connection %v flagged NEW", identity)
		}
		if seenAt, known := tracker.firstSeenAt(identity); !known || !seenAt.Equal(testRefreshTime) {
			testRun.Errorf("first seen of %v = %v, %v", identity, seenAt, known)
		}
	}
}

func TestConnectionTrackerFlagsOnlyNewInboundConnections(testRun *testing.T) {
	baseline := newConnectionTracker().observe(baselineConnections(), testRefreshTime)
	later := testRefreshTime.Add(time.Minute)
	next := baseline.observe(fullConnections(), later)
	if !next.isNew(app.ConnectionIdentityOf(connectionMetricsInbound())) {
		testRun.Error("new inbound connection must be NEW")
	}
	if next.isNew(app.ConnectionIdentityOf(connectionDatabaseOutbound())) {
		testRun.Error("new outbound connection must not be NEW")
	}
	if next.isNew(app.ConnectionIdentityOf(connectionSshInbound())) {
		testRun.Error("baseline connection must not be NEW")
	}
	if seenAt, _ := next.firstSeenAt(app.ConnectionIdentityOf(connectionSshInbound())); !seenAt.Equal(testRefreshTime) {
		testRun.Errorf("first seen of a persisting connection changed to %v", seenAt)
	}
}

func TestConnectionTrackerNewBadgeExpires(testRun *testing.T) {
	tracker := newConnectionTracker().observe(baselineConnections(), testRefreshTime)
	tracker = tracker.observe(fullConnections(), testRefreshTime.Add(time.Second))
	identity := app.ConnectionIdentityOf(connectionMetricsInbound())
	for snapshot := 1; snapshot < connectionNewBadgeRefreshCount; snapshot++ {
		if !tracker.isNew(identity) {
			testRun.Fatalf("badge expired too early at snapshot %d", snapshot)
		}
		tracker = tracker.observe(fullConnections(), testRefreshTime.Add(time.Duration(snapshot+1)*time.Second))
	}
	if !tracker.isNew(identity) {
		testRun.Fatal("badge must last for the configured number of snapshots, the detecting one included")
	}
	tracker = tracker.observe(fullConnections(), testRefreshTime.Add(time.Hour))
	if tracker.isNew(identity) {
		testRun.Fatal("badge must expire after the configured number of snapshots")
	}
}

func TestConnectionTrackerForgetsClosedConnectionsAndDoesNotMutate(testRun *testing.T) {
	first := newConnectionTracker().observe(fullConnections(), testRefreshTime)
	second := first.observe(baselineConnections(), testRefreshTime.Add(time.Second))
	closed := app.ConnectionIdentityOf(connectionDatabaseOutbound())
	if _, known := second.firstSeenAt(closed); known {
		testRun.Error("closed connection must be forgotten")
	}
	if _, known := first.firstSeenAt(closed); !known {
		testRun.Error("observe must not mutate the receiver")
	}
}

func TestConnectionTrackerEmptyFirstSnapshotIsStillABaseline(testRun *testing.T) {
	tracker := newConnectionTracker().observe([]domain.Connection{}, testRefreshTime)
	next := tracker.observe(baselineConnections(), testRefreshTime.Add(time.Second))
	if !next.isNew(app.ConnectionIdentityOf(connectionSshInbound())) {
		testRun.Fatal("inbound connection after an empty baseline must be NEW")
	}
}
