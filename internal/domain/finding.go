package domain

// Severity ranks how risky a finding is. Higher values are more severe.
type Severity int

// Severity levels.
const (
	Info Severity = iota
	Low
	Medium
	High
	Critical
)

// String returns a stable lowercase name.
func (s Severity) String() string {
	switch s {
	case Info:
		return "info"
	case Low:
		return "low"
	case Medium:
		return "medium"
	case High:
		return "high"
	case Critical:
		return "critical"
	default:
		return "unknown"
	}
}

// Finding is a risk assessment for one port.
type Finding struct {
	Port           Port
	Severity       Severity
	RuleID         string
	Title          string
	Recommendation string
}
