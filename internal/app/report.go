package app

import (
	"fmt"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

// SourceError records why one port source could not be read.
type SourceError struct {
	Source domain.Source
	Err    error
}

// Error implements the error interface.
func (sourceError SourceError) Error() string {
	return fmt.Sprintf(sourceErrorFormat, sourceError.Source, sourceError.Err)
}

// Unwrap exposes the underlying error.
func (sourceError SourceError) Unwrap() error {
	return sourceError.Err
}

// Report is the outcome of one scan.
type Report struct {
	Ports        []domain.Port
	Findings     []domain.Finding
	SourceErrors []SourceError
}

// FilterBySeverity returns a copy of the report keeping only findings at or
// above the minimum severity. Ports and source errors are kept untouched.
func (report Report) FilterBySeverity(minimum domain.Severity) Report {
	filtered := report
	filtered.Findings = make([]domain.Finding, 0, len(report.Findings))
	for _, finding := range report.Findings {
		if finding.Severity >= minimum {
			filtered.Findings = append(filtered.Findings, finding)
		}
	}
	return filtered
}

// HasFindingsAtOrAbove reports whether any finding reaches the threshold.
func (report Report) HasFindingsAtOrAbove(threshold domain.Severity) bool {
	for _, finding := range report.Findings {
		if finding.Severity >= threshold {
			return true
		}
	}
	return false
}

// HighestSeverity returns the most severe finding level, or Info when empty.
func (report Report) HighestSeverity() domain.Severity {
	highest := domain.Info
	for _, finding := range report.Findings {
		if finding.Severity > highest {
			highest = finding.Severity
		}
	}
	return highest
}
