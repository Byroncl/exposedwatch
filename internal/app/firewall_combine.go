package app

import "github.com/Byroncl/exposedwatch/internal/domain"

// combineVerdicts merges the verdicts of the two address families of a
// dual-stack socket. A connection only needs one family to get through, so
// Allowed wins; Blocked needs both families to block. Anything uncertain stays
// Unknown rather than being rounded towards an assurance.
func combineVerdicts(first domain.FirewallVerdict, second domain.FirewallVerdict) domain.FirewallVerdict {
	switch {
	case first == domain.FirewallAllowed || second == domain.FirewallAllowed:
		return domain.FirewallAllowed
	case !isEvaluatedVerdict(first) || !isEvaluatedVerdict(second):
		return domain.FirewallUnknown
	case first == domain.FirewallUnknown || second == domain.FirewallUnknown:
		return domain.FirewallUnknown
	case first == domain.FirewallRestricted || second == domain.FirewallRestricted:
		return domain.FirewallRestricted
	default:
		return domain.FirewallBlocked
	}
}

// isEvaluatedVerdict reports whether the verdict is one an INPUT evaluation
// can produce.
func isEvaluatedVerdict(verdict domain.FirewallVerdict) bool {
	switch verdict {
	case domain.FirewallAllowed, domain.FirewallRestricted, domain.FirewallBlocked, domain.FirewallUnknown:
		return true
	default:
		return false
	}
}
