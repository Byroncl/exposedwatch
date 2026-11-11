package domain

// Severity ranks how risky a finding is. Higher values are more severe.
type Severity int

// String returns a stable lowercase name.
func (severity Severity) String() string {
	switch severity {
	case Info:
		return severityNameInfo
	case Low:
		return severityNameLow
	case Medium:
		return severityNameMedium
	case High:
		return severityNameHigh
	case Critical:
		return severityNameCritical
	default:
		return severityNameUnknown
	}
}
