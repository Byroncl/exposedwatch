package domain

import (
	"fmt"
	"strings"
)

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

// ParseSeverity converts a case-insensitive severity name into a Severity.
func ParseSeverity(name string) (Severity, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case severityNameInfo:
		return Info, nil
	case severityNameLow:
		return Low, nil
	case severityNameMedium:
		return Medium, nil
	case severityNameHigh:
		return High, nil
	case severityNameCritical:
		return Critical, nil
	default:
		return Info, fmt.Errorf(unknownSeverityMessage, name)
	}
}
