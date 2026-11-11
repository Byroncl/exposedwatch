package tui

import (
	"github.com/Byroncl/exposedwatch/internal/app"
	"github.com/Byroncl/exposedwatch/internal/domain"
)

// countBySeverity tallies findings per severity.
func countBySeverity(findings []domain.Finding) map[domain.Severity]int {
	counts := make(map[domain.Severity]int)
	for _, finding := range findings {
		counts[finding.Severity]++
	}
	return counts
}

// countBySource tallies ports per discovery source.
func countBySource(report app.Report) map[domain.Source]int {
	counts := make(map[domain.Source]int)
	for _, port := range report.Ports {
		counts[port.Source]++
	}
	return counts
}
