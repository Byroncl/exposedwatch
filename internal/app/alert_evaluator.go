package app

import (
	"sort"
	"time"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

// trackedFinding is what the evaluator remembers about one finding.
type trackedFinding struct {
	lastAlertedAt time.Time
	present       bool
}

// EvaluatedAlert is an alert together with the finding identities it covers,
// so a caller that gives up on delivering it can Forget them.
type EvaluatedAlert struct {
	Alert domain.Alert
	// Identities is empty for summaries.
	Identities []FindingIdentity
}

// AlertEvaluator turns successive reports into the alerts to send now. It is
// pure apart from its in-memory state and the injected clock. It is not safe
// for concurrent use.
type AlertEvaluator struct {
	policy       AlertPolicy
	clock        Clock
	hostName     string
	tracked      map[FindingIdentity]*trackedFinding
	baselineDone bool
	// unbaselined holds the sources that could not be read during the
	// baseline. Their existing findings are summarised, not alerted one by
	// one, when they first become visible.
	unbaselined map[domain.Source]bool
}

// NewAlertEvaluator builds an evaluator for one process lifetime.
func NewAlertEvaluator(policy AlertPolicy, clock Clock, hostName string) *AlertEvaluator {
	return &AlertEvaluator{
		policy:      policy,
		clock:       clock,
		hostName:    hostName,
		tracked:     make(map[FindingIdentity]*trackedFinding),
		unbaselined: make(map[domain.Source]bool),
	}
}

// Evaluate returns the alerts for this report. See EvaluateDetailed.
func (evaluator *AlertEvaluator) Evaluate(report Report) []domain.Alert {
	evaluated := evaluator.EvaluateDetailed(report)
	alerts := make([]domain.Alert, 0, len(evaluated))
	for _, entry := range evaluated {
		alerts = append(alerts, entry.Alert)
	}
	return alerts
}

// EvaluateDetailed returns the alerts for this report with their identities.
//
// The first call records everything that already exists and returns one
// summary alert (or none). A source that failed during that first call is
// baselined when it first loads: its existing findings become one summary,
// never one alert each. Later calls alert once per finding whose identity
// is not currently known; a finding that disappeared is remembered until the
// cooldown expires so a flapping port cannot spam. Findings of a source that
// failed in this report are never forgotten.
func (evaluator *AlertEvaluator) EvaluateDetailed(report Report) []EvaluatedAlert {
	now := evaluator.clock.Now()
	qualifying := evaluator.qualifyingFindings(report)
	failed := failedSources(report)
	if !evaluator.baselineDone {
		return evaluator.baseline(qualifying, failed, now)
	}

	recovered := make(map[domain.Source]bool)
	for source := range evaluator.unbaselined {
		if !failed[source] {
			recovered[source] = true
			delete(evaluator.unbaselined, source)
		}
	}
	var alerts []EvaluatedAlert
	recoveredFindings := make(map[domain.Source][]domain.Finding)
	currentIdentities := make(map[FindingIdentity]bool, len(qualifying))
	for _, finding := range qualifying {
		identity := FindingIdentityOf(finding)
		currentIdentities[identity] = true
		if recovered[finding.Port.Source] {
			evaluator.tracked[identity] = &trackedFinding{lastAlertedAt: now, present: true}
			recoveredFindings[finding.Port.Source] = append(recoveredFindings[finding.Port.Source], finding)
			continue
		}
		if evaluator.shouldAlert(identity, now) {
			alerts = append(alerts, EvaluatedAlert{
				Alert:      newFindingAlert(evaluator.hostName, finding),
				Identities: []FindingIdentity{identity},
			})
		}
	}
	for _, source := range sortedSources(recoveredFindings) {
		alerts = append(alerts, EvaluatedAlert{
			Alert: sourceSummaryAlert(evaluator.hostName, source, recoveredFindings[source]),
		})
	}
	evaluator.forgetMissing(currentIdentities, failed, now)
	return alerts
}

// Forget removes identities from the known set, so the findings alert again
// if they are still present at a later evaluation. Call it when an alert for
// them was given up on.
func (evaluator *AlertEvaluator) Forget(identities []FindingIdentity) {
	for _, identity := range identities {
		delete(evaluator.tracked, identity)
	}
}

// baseline records the existing findings and returns the start-up summary.
func (evaluator *AlertEvaluator) baseline(qualifying []domain.Finding, failed map[domain.Source]bool, now time.Time) []EvaluatedAlert {
	evaluator.baselineDone = true
	for source := range failed {
		evaluator.unbaselined[source] = true
	}
	for _, finding := range qualifying {
		evaluator.tracked[FindingIdentityOf(finding)] = &trackedFinding{lastAlertedAt: now, present: true}
	}
	if len(qualifying) == 0 {
		return nil
	}
	return []EvaluatedAlert{{Alert: startupSummaryAlert(evaluator.hostName, evaluator.policy.MinSeverity, qualifying)}}
}

// shouldAlert updates the state for a finding seen now and reports whether
// it deserves an alert.
func (evaluator *AlertEvaluator) shouldAlert(identity FindingIdentity, now time.Time) bool {
	entry, known := evaluator.tracked[identity]
	if !known {
		evaluator.tracked[identity] = &trackedFinding{lastAlertedAt: now, present: true}
		return true
	}
	if entry.present {
		return false
	}
	entry.present = true
	if now.Sub(entry.lastAlertedAt) < evaluator.policy.Cooldown {
		return false
	}
	entry.lastAlertedAt = now
	return true
}

// forgetMissing marks absent findings as gone and drops those whose
// cooldown expired. Findings of failed sources are left untouched.
func (evaluator *AlertEvaluator) forgetMissing(current map[FindingIdentity]bool, failed map[domain.Source]bool, now time.Time) {
	for identity, entry := range evaluator.tracked {
		if current[identity] || failed[identity.Port.Source] {
			continue
		}
		entry.present = false
		if now.Sub(entry.lastAlertedAt) >= evaluator.policy.Cooldown {
			delete(evaluator.tracked, identity)
		}
	}
}

// qualifyingFindings keeps findings at or above the threshold, once per identity.
func (evaluator *AlertEvaluator) qualifyingFindings(report Report) []domain.Finding {
	seen := make(map[FindingIdentity]bool, len(report.Findings))
	var qualifying []domain.Finding
	for _, finding := range report.Findings {
		identity := FindingIdentityOf(finding)
		if finding.Severity < evaluator.policy.MinSeverity || seen[identity] {
			continue
		}
		seen[identity] = true
		qualifying = append(qualifying, finding)
	}
	return qualifying
}

func failedSources(report Report) map[domain.Source]bool {
	failed := make(map[domain.Source]bool, len(report.SourceErrors))
	for _, sourceError := range report.SourceErrors {
		failed[sourceError.Source] = true
	}
	return failed
}

// sortedSources lists the map keys in a stable order.
func sortedSources(findings map[domain.Source][]domain.Finding) []domain.Source {
	sources := make([]domain.Source, 0, len(findings))
	for source := range findings {
		sources = append(sources, source)
	}
	sort.Slice(sources, func(first, second int) bool { return sources[first] < sources[second] })
	return sources
}
