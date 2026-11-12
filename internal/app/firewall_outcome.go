package app

import (
	"fmt"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

// outcomeRecord remembers the first terminal decision of one kind.
type outcomeRecord struct {
	present bool
	reason  string
	detail  string
}

func (record *outcomeRecord) remember(reason string, detail string) {
	if record.present {
		return
	}
	*record = outcomeRecord{present: true, reason: reason, detail: detail}
}

// walkOutcome accumulates what can happen to a new inbound connection while
// the INPUT chain is walked. Decisions for only some sources are kept apart
// from decisions for every source.
type walkOutcome struct {
	acceptEverywhere  outcomeRecord
	acceptSomeSources outcomeRecord
	dropEverywhere    outcomeRecord
	dropSomeSources   outcomeRecord
	// uncertainty is set as soon as a rule cannot be evaluated; it makes the
	// verdict Unknown whatever else was recorded.
	uncertainty string
}

func (outcome *walkOutcome) recordAccept(restriction string, reason string) {
	if restriction == "" {
		outcome.acceptEverywhere.remember(reason, "")
		return
	}
	outcome.acceptSomeSources.remember(reason, restriction)
}

func (outcome *walkOutcome) recordDrop(restriction string, reason string) {
	if restriction == "" {
		outcome.dropEverywhere.remember(reason, "")
		return
	}
	outcome.dropSomeSources.remember(reason, restriction)
}

func (outcome *walkOutcome) markUncertain(reason string) {
	if outcome.uncertainty == "" {
		outcome.uncertainty = reason
	}
}

// verdict maps the accumulated decisions to a verdict and its reason.
func (outcome walkOutcome) verdict() (domain.FirewallVerdict, string) {
	if outcome.uncertainty != "" {
		return domain.FirewallUnknown, outcome.uncertainty
	}
	anyAccept := outcome.acceptEverywhere.present || outcome.acceptSomeSources.present
	anyDrop := outcome.dropEverywhere.present || outcome.dropSomeSources.present
	switch {
	case anyAccept && anyDrop:
		if outcome.acceptSomeSources.present {
			return domain.FirewallRestricted, fmt.Sprintf(reasonRestrictedFormat, outcome.acceptSomeSources.detail)
		}
		return domain.FirewallRestricted, fmt.Sprintf(reasonBlockedForFormat, outcome.dropSomeSources.detail)
	case anyAccept:
		if outcome.acceptEverywhere.present {
			return domain.FirewallAllowed, outcome.acceptEverywhere.reason
		}
		return domain.FirewallAllowed, outcome.acceptSomeSources.reason
	case anyDrop:
		if outcome.dropEverywhere.present {
			return domain.FirewallBlocked, outcome.dropEverywhere.reason
		}
		return domain.FirewallBlocked, outcome.dropSomeSources.reason
	default:
		return domain.FirewallUnknown, reasonNoDecision
	}
}
