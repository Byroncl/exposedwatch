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

// sourceOrder lists discovery sources in header order.
var sourceOrder = []domain.Source{domain.SourceHost, domain.SourceDocker, domain.SourceExternal}

// headerLineCount is the header height: the Overview shows the summary in its
// own panel, so its header is the title line alone.
func (model Model) headerLineCount() int {
	if model.activeTab == TabSummary {
		return overviewHeaderLineCount
	}
	return headerLineCount
}

// headerLines renders the title line, the severity chips and the risk line.
func (model Model) headerLines(width int) []string {
	if model.activeTab == TabSummary {
		return []string{clampWidth(model.titleLine(), width)}
	}
	return []string{
		clampWidth(model.titleLine(), width),
		clampWidth(model.chipsLine(), width),
		clampWidth(model.riskLine(width), width),
	}
}

func (model Model) titleLine() string {
	updated := updatedPlaceholder
	if !model.lastRefresh.IsZero() {
		updated = model.lastRefresh.Format(timeLayout)
	}
	line := model.styles.title.Render(fmt.Sprintf(titleFormat, appTitle, model.version)) +
		model.styles.dim.Render(fmt.Sprintf(updatedFormat, updated)) +
		model.styles.dim.Render(fmt.Sprintf(intervalFormat, model.interval))
	if model.refreshing {
		line += model.styles.warning.Render(refreshingText)
	}
	return line
}

// chipsLine shows one colored chip per severity and the total port count.
func (model Model) chipsLine() string {
	counts := countBySeverity(model.report.Findings)
	parts := make([]string, 0, len(severityOrder)+1)
	for _, severity := range severityOrder {
		label := " " + strings.ToUpper(fmt.Sprintf(chipCountFormat, severity, counts[severity])) + " "
		if counts[severity] == 0 {
			parts = append(parts, model.styles.dim.Render(label))
			continue
		}
		parts = append(parts, model.styles.chipFor(severity).Render(label))
	}
	ports := model.styles.plain.Render(fmt.Sprintf(portsCountFormat, len(model.report.Ports)))
	return strings.Join(parts, "") + headerJoin + ports
}

// riskLine shows the risk gauge, the per-source counters and the session
// counters of opened and closed ports.
func (model Model) riskLine(width int) string {
	level := riskLevelOf(model.report)
	filled, empty := level.gaugeCells()
	levelStyle := model.styles.safe
	if severity, hasSeverity := level.severity(); hasSeverity {
		levelStyle = model.styles.forSeverity(severity)
	}
	gauge := levelStyle.Render(strings.Repeat(gaugeFilledGlyph, filled)) +
		model.styles.dim.Render(strings.Repeat(gaugeEmptyGlyph, empty))
	risk := model.styles.dim.Render(riskLabel) + gauge + levelStyle.Bold(true).Render(fmt.Sprintf(riskFormat, level))
	sources := make([]string, 0, len(sourceOrder))
	counts := countBySource(model.report)
	for _, source := range sourceOrder {
		sources = append(sources, fmt.Sprintf(sourceCountFormat, sourceLabel(source), counts[source]))
	}
	changes := model.styles.opened.Render(fmt.Sprintf(openedCountFormat, model.tracker.openedTotal)) + headerJoin +
		model.styles.closed.Render(fmt.Sprintf(closedCountFormat, model.tracker.closedTotal))
	line := risk + headerJoin + model.styles.dim.Render(strings.Join(sources, headerJoin)) + headerJoin + changes
	return line + model.connectionCounterFitting(line, width)
}

func sourceLabel(source domain.Source) string {
	if source == domain.SourceExternal {
		return sourceLabelExternal
	}
	return string(source)
}
