package tui

import (
	"time"

	"github.com/Byroncl/exposedwatch/internal/app"
	"github.com/Byroncl/exposedwatch/internal/domain"
)

// connectionTracker follows the connection snapshots of one session, in
// memory only. The first successful snapshot is the baseline and flags
// nothing. Only inbound connections get a NEW badge. It is a value type:
// observe returns a new tracker and never mutates the receiver.
type connectionTracker struct {
	hasBaseline bool
	previous    []domain.Connection
	firstSeen   map[app.ConnectionIdentity]time.Time
	newAge      map[app.ConnectionIdentity]int
}

func newConnectionTracker() connectionTracker {
	return connectionTracker{
		firstSeen: map[app.ConnectionIdentity]time.Time{},
		newAge:    map[app.ConnectionIdentity]int{},
	}
}

// isNew reports whether the connection still carries the NEW badge.
func (tracker connectionTracker) isNew(identity app.ConnectionIdentity) bool {
	_, flagged := tracker.newAge[identity]
	return flagged
}

// firstSeenAt returns when this session first observed the connection.
func (tracker connectionTracker) firstSeenAt(identity app.ConnectionIdentity) (time.Time, bool) {
	seenAt, known := tracker.firstSeen[identity]
	return seenAt, known
}

// observe folds one successful snapshot into the tracker. Connections that
// are gone are forgotten, so memory stays bounded by the current snapshot.
func (tracker connectionTracker) observe(connections []domain.Connection, observedAt time.Time) connectionTracker {
	next := connectionTracker{
		hasBaseline: true,
		previous:    append(make([]domain.Connection, 0, len(connections)), connections...),
		firstSeen:   make(map[app.ConnectionIdentity]time.Time, len(connections)),
		newAge:      make(map[app.ConnectionIdentity]int),
	}
	for _, connection := range connections {
		identity := app.ConnectionIdentityOf(connection)
		if seenAt, known := tracker.firstSeen[identity]; known {
			next.firstSeen[identity] = seenAt
			continue
		}
		next.firstSeen[identity] = observedAt
	}
	if !tracker.hasBaseline {
		return next
	}
	for identity, age := range tracker.newAge {
		if _, stillOpen := next.firstSeen[identity]; stillOpen && age+1 < connectionNewBadgeRefreshCount {
			next.newAge[identity] = age + 1
		}
	}
	for _, change := range app.DetectConnectionChanges(tracker.previous, connections) {
		if change.Kind == app.ChangeOpened && change.Connection.Direction == domain.ConnectionInbound {
			next.newAge[app.ConnectionIdentityOf(change.Connection)] = 0
		}
	}
	return next
}
