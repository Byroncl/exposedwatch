package tui

import (
	"time"

	"github.com/Byroncl/exposedwatch/internal/app"
	"github.com/Byroncl/exposedwatch/internal/domain"
)

// changeTracker follows the port snapshots of one session. The first
// successful snapshot is the baseline and flags nothing. It is a value type:
// observe returns a new tracker and never mutates the receiver.
type changeTracker struct {
	hasBaseline   bool
	previousPorts []domain.Port
	firstSeen     map[app.PortIdentity]time.Time
	newAge        map[app.PortIdentity]int
	severities    map[app.PortIdentity]domain.Severity
	log           []ChangeLogEntry
	openedTotal   int
	closedTotal   int
}

func newChangeTracker() changeTracker {
	return changeTracker{
		firstSeen:  map[app.PortIdentity]time.Time{},
		newAge:     map[app.PortIdentity]int{},
		severities: map[app.PortIdentity]domain.Severity{},
	}
}

// isNew reports whether the port still carries the NEW badge.
func (tracker changeTracker) isNew(identity app.PortIdentity) bool {
	_, flagged := tracker.newAge[identity]
	return flagged
}

// firstSeenAt returns when this session first observed the port.
func (tracker changeTracker) firstSeenAt(identity app.PortIdentity) (time.Time, bool) {
	seenAt, known := tracker.firstSeen[identity]
	return seenAt, known
}

// observe folds one successful scan into the tracker. Ports that belong to a
// source that failed in this scan are carried over untouched, so a partial
// failure never produces a false "closed" event.
func (tracker changeTracker) observe(report app.Report, observedAt time.Time) changeTracker {
	next := tracker.cloned()
	failedSources := failedSourceSet(report)
	currentSeverities := highestSeverityByPort(report)
	if !tracker.hasBaseline {
		next.hasBaseline = true
		for _, port := range report.Ports {
			next.firstSeen[app.IdentityOf(port)] = observedAt
		}
		next.previousPorts = carryOver(report.Ports, nil, failedSources)
		next.severities = currentSeverities
		return next
	}
	next.ageNewBadges()
	for _, change := range app.DetectChanges(tracker.previousPorts, report.Ports) {
		identity := app.IdentityOf(change.Port)
		if change.Kind == app.ChangeClosed && failedSources[change.Port.Source] {
			continue
		}
		severity, hasSeverity := tracker.severities[identity]
		if change.Kind == app.ChangeOpened {
			severity, hasSeverity = currentSeverities[identity]
			next.firstSeen[identity] = observedAt
			next.newAge[identity] = 0
			next.openedTotal++
		} else {
			delete(next.firstSeen, identity)
			delete(next.newAge, identity)
			next.closedTotal++
		}
		next.appendLog(ChangeLogEntry{At: observedAt, Change: change, Severity: severity, HasSeverity: hasSeverity})
	}
	next.previousPorts = carryOver(report.Ports, tracker.previousPorts, failedSources)
	next.severities = mergeSeverities(currentSeverities, tracker.severities, failedSources, tracker.previousPorts)
	return next
}

func (tracker changeTracker) cloned() changeTracker {
	clone := tracker
	clone.firstSeen = make(map[app.PortIdentity]time.Time, len(tracker.firstSeen))
	for identity, seenAt := range tracker.firstSeen {
		clone.firstSeen[identity] = seenAt
	}
	clone.newAge = make(map[app.PortIdentity]int, len(tracker.newAge))
	for identity, age := range tracker.newAge {
		clone.newAge[identity] = age
	}
	clone.log = append([]ChangeLogEntry(nil), tracker.log...)
	return clone
}

// ageNewBadges advances every NEW badge by one snapshot and drops expired ones.
func (tracker *changeTracker) ageNewBadges() {
	for identity, age := range tracker.newAge {
		age++
		if age >= newBadgeRefreshCount {
			delete(tracker.newAge, identity)
			continue
		}
		tracker.newAge[identity] = age
	}
}

func (tracker *changeTracker) appendLog(entry ChangeLogEntry) {
	tracker.log = append(tracker.log, entry)
	if overflow := len(tracker.log) - maxChangeLogEntries; overflow > 0 {
		tracker.log = append([]ChangeLogEntry(nil), tracker.log[overflow:]...)
	}
}

func failedSourceSet(report app.Report) map[domain.Source]bool {
	failed := make(map[domain.Source]bool, len(report.SourceErrors))
	for _, sourceError := range report.SourceErrors {
		failed[sourceError.Source] = true
	}
	return failed
}

// carryOver returns the next baseline: the current ports plus the previous
// ports of every failed source. The result is never nil.
func carryOver(current, previous []domain.Port, failedSources map[domain.Source]bool) []domain.Port {
	carried := append(make([]domain.Port, 0, len(current)), current...)
	for _, port := range previous {
		if failedSources[port.Source] {
			carried = append(carried, port)
		}
	}
	return carried
}

// highestSeverityByPort maps each port to its most severe finding.
func highestSeverityByPort(report app.Report) map[app.PortIdentity]domain.Severity {
	severities := make(map[app.PortIdentity]domain.Severity, len(report.Findings))
	for _, finding := range report.Findings {
		identity := app.IdentityOf(finding.Port)
		if existing, known := severities[identity]; !known || finding.Severity > existing {
			severities[identity] = finding.Severity
		}
	}
	return severities
}

// mergeSeverities keeps the last known severity of ports whose source failed.
func mergeSeverities(
	current, previous map[app.PortIdentity]domain.Severity,
	failedSources map[domain.Source]bool,
	previousPorts []domain.Port,
) map[app.PortIdentity]domain.Severity {
	merged := current
	for _, port := range previousPorts {
		identity := app.IdentityOf(port)
		if severity, known := previous[identity]; known && failedSources[port.Source] {
			merged[identity] = severity
		}
	}
	return merged
}
