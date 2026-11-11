package tui

import (
	"fmt"
	"strings"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

// severityOrder lists severities from most to least severe for the header.
var severityOrder = []domain.Severity{
	domain.Critical, domain.High, domain.Medium, domain.Low, domain.Info,
}

func (model Model) headerLines(width int) []string {
	updated := updatedPlaceholder
	if !model.lastRefresh.IsZero() {
		updated = model.lastRefresh.Format(timeLayout)
	}
	titleLine := model.styles.title.Render(fmt.Sprintf(titleFormat, appTitle, model.version)) +
		model.styles.dim.Render(fmt.Sprintf(updatedFormat, updated))
	if model.refreshing {
		titleLine += model.styles.warning.Render(refreshingIndicator)
	}
	counts := countBySeverity(model.report.Findings)
	parts := []string{model.styles.plain.Render(fmt.Sprintf(portsCountFormat, len(model.report.Ports)))}
	for _, severity := range severityOrder {
		label := fmt.Sprintf(severityCountFormat, severity, counts[severity])
		parts = append(parts, model.styles.forSeverity(severity).Render(label))
	}
	countsLine := strings.Join(parts, countSeparator)
	return []string{
		model.styles.plain.MaxWidth(width).Render(titleLine),
		model.styles.plain.MaxWidth(width).Render(countsLine),
	}
}
