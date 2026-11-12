package app

import "github.com/Byroncl/exposedwatch/internal/domain"

// UseBaseline switches the evaluator to drift-only alerting: findings whose
// port is accepted never alert, there is no first-run summary, and every
// other finding alerts individually. Calling it again replaces the accepted
// set (a baseline reload); nil restores the default behavior.
func (evaluator *AlertEvaluator) UseBaseline(acceptance *BaselineAcceptance) {
	evaluator.acceptance = acceptance
	if acceptance != nil {
		evaluator.baselineDone = true
	}
}

// A reload that accepts a finding only affects later evaluations: an alert
// that is already queued for it is still delivered, which is intended (it was
// true when it was raised).

// withoutAccepted drops the accepted findings and silently forgets any that
// were tracked before they became accepted.
func (evaluator *AlertEvaluator) withoutAccepted(findings []domain.Finding) []domain.Finding {
	for identity := range evaluator.tracked {
		if evaluator.acceptance.Accepts(identityPort(identity)) {
			delete(evaluator.tracked, identity)
		}
	}
	remaining := findings[:0:0]
	for _, finding := range findings {
		if !evaluator.acceptance.Accepts(finding.Port) {
			remaining = append(remaining, finding)
		}
	}
	return remaining
}

// findingAlert builds the alert of a new finding, with the baseline line in
// baseline mode.
func (evaluator *AlertEvaluator) findingAlert(finding domain.Finding) domain.Alert {
	alert := newFindingAlert(evaluator.hostName, finding)
	if evaluator.acceptance != nil {
		if alert.Body != "" {
			alert.Body += bodyLineSeparator
		}
		alert.Body += evaluator.acceptance.NotAcceptedLine()
	}
	return alert
}

// identityPort rebuilds a comparable port from a tracked identity.
func identityPort(identity FindingIdentity) domain.Port {
	return domain.Port{
		Protocol:    identity.Port.Protocol,
		Number:      identity.Port.Number,
		BindAddress: identity.Port.BindAddress,
		Source:      identity.Port.Source,
	}
}
