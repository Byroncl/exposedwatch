package tui

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

// column is one table column and the terminal width from which it shows.
type column struct {
	heading   string
	width     int
	fromWidth int
	value     func(finding domain.Finding) string
}

var tableColumns = []column{
	{headingSeverity, severityColumnWidth, 0, func(finding domain.Finding) string { return finding.Severity.String() }},
	{headingPort, portColumnWidth, 0, func(finding domain.Finding) string { return strconv.Itoa(finding.Port.Number) }},
	{headingProtocol, protocolColumnWidth, showProtocolFromWidth, func(finding domain.Finding) string { return string(finding.Port.Protocol) }},
	{headingBind, bindColumnWidth, showBindFromWidth, func(finding domain.Finding) string { return finding.Port.BindAddress }},
	{headingExposure, exposureColumnWidth, showExposureFromWidth, func(finding domain.Finding) string { return finding.Port.Exposure().String() }},
	{headingSource, sourceColumnWidth, showSourceFromWidth, func(finding domain.Finding) string { return string(finding.Port.Source) }},
	{headingProcess, processColumnWidth, showProcessFromWidth, func(finding domain.Finding) string { return processOrContainer(finding.Port) }},
}

func visibleColumns(width int) []column {
	shown := make([]column, 0, len(tableColumns))
	for _, candidate := range tableColumns {
		if width >= candidate.fromWidth {
			shown = append(shown, candidate)
		}
	}
	return shown
}

func titleColumnWidth(width int, shown []column) int {
	used := 0
	for _, candidate := range shown {
		used += candidate.width + len(columnSeparator)
	}
	remaining := width - used
	if remaining < minTitleColumnWidth {
		return minTitleColumnWidth
	}
	return remaining
}

// fixedCell pads or truncates text to exactly width cells.
func fixedCell(text string, width int) string {
	return lipgloss.NewStyle().Inline(true).Width(width).MaxWidth(width).Render(text)
}

func (model Model) tableLines(frame layout) []string {
	if frame.tableRows < 0 {
		return nil
	}
	shown := visibleColumns(frame.width)
	titleWidth := titleColumnWidth(frame.width, shown)
	lines := []string{model.tableHeadingLine(shown, titleWidth, frame.width)}
	if len(model.visibleFindings) == 0 {
		return append(lines, model.emptyMessageLine(frame.width))
	}
	end := model.scrollOffset + frame.tableRows
	if end > len(model.visibleFindings) {
		end = len(model.visibleFindings)
	}
	for index := model.scrollOffset; index < end; index++ {
		lines = append(lines, model.tableRowLine(index, shown, titleWidth, frame.width))
	}
	return lines
}

func (model Model) emptyMessageLine(width int) string {
	message := noMatchMessage
	switch {
	case !model.hasReport:
		message = scanningMessage
	case len(model.report.Findings) == 0:
		message = noFindingsMessage
	}
	return model.styles.dim.MaxWidth(width).Render(message)
}

func (model Model) tableHeadingLine(shown []column, titleWidth, width int) string {
	cells := make([]string, 0, len(shown)+1)
	for _, candidate := range shown {
		cells = append(cells, fixedCell(candidate.heading, candidate.width))
	}
	cells = append(cells, fixedCell(headingTitle, titleWidth))
	return model.styles.heading.MaxWidth(width).Render(strings.Join(cells, columnSeparator))
}

func (model Model) tableRowLine(index int, shown []column, titleWidth, width int) string {
	finding := model.visibleFindings[index]
	selected := index == model.cursor
	cells := make([]string, 0, len(shown)+1)
	for position, candidate := range shown {
		cell := fixedCell(candidate.value(finding), candidate.width)
		if position == 0 && !selected {
			cell = model.styles.forSeverity(finding.Severity).Render(cell)
		}
		cells = append(cells, cell)
	}
	cells = append(cells, fixedCell(finding.Title, titleWidth))
	line := strings.Join(cells, columnSeparator)
	if selected {
		return model.styles.selected.MaxWidth(width).Render(line)
	}
	return model.styles.plain.MaxWidth(width).Render(line)
}
