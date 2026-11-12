package app

import "github.com/Byroncl/exposedwatch/internal/domain"

// FindingIdentity identifies a finding across evaluations: rule id plus the
// port identity (protocol, number, bind address and source). Titles and
// process names are descriptive only.
type FindingIdentity struct {
	RuleID string
	Port   PortIdentity
}

// FindingIdentityOf returns the identity of a finding.
func FindingIdentityOf(finding domain.Finding) FindingIdentity {
	return FindingIdentity{RuleID: finding.RuleID, Port: IdentityOf(finding.Port)}
}
