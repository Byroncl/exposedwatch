package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Byroncl/exposedwatch/internal/app"
	"github.com/Byroncl/exposedwatch/internal/domain"
)

// overviewPanel renders the Summary dashboard from data already in the model.
func (model Model) overviewPanel(frame layout) []string {
	innerWidth := panelContentWidth(frame.contentWidth)
	lines := []string{
		model.overviewRiskLine(innerWidth),
		model.chipsLine(),
		model.overviewCountersLine(),
		model.overviewConnectionLine(),
		model.overviewFirewallLine(),
		model.overviewAlertsLine(),
		"",
	}
	lines = append(lines, model.worstFindingLines(innerWidth)...)
	lines = append(lines, "")
	lines = append(lines, model.recentChangeLines(innerWidth)...)
	if model.connectionsEnabled() {
		lines = append(lines, "")
		lines = append(lines, model.overviewTopRemoteLines(innerWidth)...)
	}
	lines = firstLines(lines, max(frame.panelLines-panelBorderLineCount, 0))
	lines = append(lines, blankLines(frame.panelLines-panelBorderLineCount-len(lines), innerWidth)...)
	return model.styles.renderPanel(panelOverviewTitle, lines, frame.contentWidth)
}

// overviewRiskLine draws the overall risk as a wide gauge followed by its word.
func (model Model) overviewRiskLine(innerWidth int) string {
	level := riskLevelOf(model.report)
	word := fmt.Sprintf(riskFormat, level)
	gaugeWidth := min(max(innerWidth-len(riskLabel)-len(word)-overviewGaugeMargin, overviewGaugeMinWidth), overviewGaugeMaxWidth)
	filled, empty := level.gaugeCellsFor(gaugeWidth)
	levelStyle := model.styles.safe
	if severity, hasSeverity := level.severity(); hasSeverity {
		levelStyle = model.styles.forSeverity(severity)
	}
	return model.styles.dim.Render(riskLabel) +
		levelStyle.Render(strings.Repeat(gaugeFilledGlyph, filled)) +
		model.styles.dim.Render(strings.Repeat(gaugeEmptyGlyph, empty)) +
		levelStyle.Bold(true).Render(word)
}

// overviewCountersLine lists the ports per source, the session counters and the
// connection counter (which follows the loopback and inbound-only toggles).
func (model Model) overviewCountersLine() string {
	counts := countBySource(model.report)
	sources := make([]string, 0, len(sourceOrder))
	for _, source := range sourceOrder {
		sources = append(sources, fmt.Sprintf(sourceCountFormat, sourceLabel(source), counts[source]))
	}
	changes := model.styles.opened.Render(fmt.Sprintf(openedCountFormat, model.tracker.openedTotal)) + headerJoin +
		model.styles.closed.Render(fmt.Sprintf(closedCountFormat, model.tracker.closedTotal))
	return columnSeparator + model.styles.dim.Render(strings.Join(sources, headerJoin)) + headerJoin + changes
}

// overviewConnectionLine shows the connection counter, or why there is none.
func (model Model) overviewConnectionLine() string {
	return columnSeparator + model.overviewConnectionCounter()
}

func (model Model) overviewConnectionCounter() string {
	switch {
	case !model.connectionsEnabled():
		return model.styles.dim.Render(summaryConnectionsDisabledMessage)
	case !model.hasConnections:
		return model.styles.dim.Render(summaryConnectionsReadingMessage)
	default:
		summary := model.connectionSummary
		return model.styles.plain.Render(fmt.Sprintf(headerConnectionCounterFmt, summary.Total, summary.Inbound, summary.Outbound))
	}
}

// worstFindingLines lists the most severe findings, highest severity first and
// lowest port first among equals.
func (model Model) worstFindingLines(innerWidth int) []string {
	lines := []string{model.styles.heading.Render(headingWorstFindings)}
	if !model.hasReport {
		return append(lines, model.styles.dim.Render(scanningMessage))
	}
	if len(model.report.Findings) == 0 {
		return append(lines, model.styles.dim.Render(noFindingsMessage))
	}
	for _, finding := range worstFindings(model.report.Findings, worstFindingsLimit) {
		lines = append(lines, model.findingRow(finding, innerWidth))
	}
	return lines
}

func worstFindings(findings []domain.Finding, limit int) []domain.Finding {
	ranked := append([]domain.Finding{}, findings...)
	sort.SliceStable(ranked, func(left, right int) bool {
		if ranked[left].Severity != ranked[right].Severity {
			return ranked[left].Severity > ranked[right].Severity
		}
		return ranked[left].Port.Number < ranked[right].Port.Number
	})
	return ranked[:min(limit, len(ranked))]
}

func (model Model) findingRow(finding domain.Finding, innerWidth int) string {
	chip := model.styles.chipFor(finding.Severity).Render(fitWidth(" "+strings.ToUpper(finding.Severity.String())+" ", findingChipWidth))
	port := fitWidth(fmt.Sprintf(portFormat, finding.Port.Number, finding.Port.Protocol), portColumnWidth)
	titleWidth := innerWidth - findingChipWidth - portColumnWidth - 2*len(columnSeparator)
	title := ""
	if titleWidth > 0 {
		title = fitWidth(finding.Title, titleWidth)
	}
	return fmt.Sprintf(findingRowFormat, chip, port, title)
}

// recentChangeLines lists the newest change log entries first.
func (model Model) recentChangeLines(innerWidth int) []string {
	lines := []string{model.styles.heading.Render(headingRecentChanges)}
	log := model.tracker.log
	if len(log) == 0 {
		return append(lines, model.styles.dim.Render(noRecentChangesMessage))
	}
	for index := len(log) - 1; index >= max(len(log)-recentChangesLimit, 0); index-- {
		lines = append(lines, model.changeRow(log[index], innerWidth))
	}
	return lines
}

func (model Model) changeRow(entry ChangeLogEntry, innerWidth int) string {
	kind, kindStyle := kindOpenedLabel, model.styles.opened
	if entry.Change.Kind == app.ChangeClosed {
		kind, kindStyle = kindClosedLabel, model.styles.closed
	}
	port := entry.Change.Port
	moment := entry.At.Format(timeLayoutShort)
	label := fitWidth(fmt.Sprintf(portFormat, port.Number, port.Protocol), portColumnWidth)
	detailWidth := innerWidth - timeColumnWidth - changeKindWidth - portColumnWidth - 3*len(columnSeparator)
	detail := ""
	if detailWidth > 0 {
		detail = fitWidth(processOrContainer(port), detailWidth)
	}
	return fmt.Sprintf(changeRowFormat, moment, kindStyle.Render(fitWidth(kind, changeKindWidth)), label, detail)
}

// overviewTopRemoteLines lists the busiest inbound remote addresses.
func (model Model) overviewTopRemoteLines(innerWidth int) []string {
	lines := []string{model.styles.heading.Render(headingTopRemote)}
	if !model.hasConnections {
		return append(lines, model.styles.dim.Render(connectionsReadingMessage))
	}
	entries := model.connectionSummary.TopRemoteAddresses
	if len(entries) == 0 {
		return append(lines, model.styles.dim.Render(noInboundConnectionsMessage))
	}
	for _, entry := range entries[:min(topRemoteLimit, len(entries))] {
		lines = append(lines, alignedPair(entry.Address, fmt.Sprintf(countFormat, entry.Count), innerWidth))
	}
	return lines
}
