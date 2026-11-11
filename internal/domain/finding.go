package domain

// Finding is a risk assessment for one port.
type Finding struct {
	Port           Port
	Severity       Severity
	RuleID         string
	Title          string
	Recommendation string
}
