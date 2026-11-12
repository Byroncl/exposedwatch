package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderCell draws one cell padded to exactly width cells.
func (palette styles) renderCell(content cell, width int, selected bool) string {
	look := content.look
	if selected {
		look = look.Background(colorSelected)
	}
	if !content.chip {
		return look.Render(fitWidth(content.text, width))
	}
	label := clampWidth(strings.Repeat(columnSeparator, chipSidePadding)+content.text+strings.Repeat(columnSeparator, chipSidePadding), width)
	gap := strings.Repeat(" ", max(width-lipgloss.Width(label), 0))
	return look.Render(label) + palette.rowGap(selected).Render(gap)
}

// rowGap styles the spaces between cells.
func (palette styles) rowGap(selected bool) lipgloss.Style {
	if selected {
		return palette.selected
	}
	return palette.plain
}

// tableContent returns the heading and the visible rows, padded to the table
// height so the detail panel keeps its place.
func (model Model) tableContent(frame layout) []string {
	contentWidth := panelContentWidth(frame.contentWidth)
	spec := tableSpecFor(model.activeTab)
	shown, flexWidth := spec.layoutColumns(contentWidth)
	lines := []string{model.tableHeadingLine(spec, shown, flexWidth)}
	if len(model.visibleRows) == 0 {
		lines = append(lines, model.styles.dim.Render(model.emptyMessage()))
	}
	end := min(model.scrollOffset+frame.tableRows, len(model.visibleRows))
	for index := model.scrollOffset; index < end; index++ {
		lines = append(lines, model.tableRowLine(index, spec, shown, flexWidth))
	}
	missing := tableHeadingLines + frame.tableRows - len(lines)
	return append(lines, blankLines(missing, contentWidth)...)
}

func (model Model) emptyMessage() string {
	switch {
	case model.activeTab == TabConnections:
		return model.connectionsEmptyMessage()
	case model.activeTab == TabFirewallPorts:
		return model.firewallEmptyMessage()
	case !model.hasReport:
		return scanningMessage
	case model.activeTab == TabChanges:
		return noChangesMessage
	case model.activeTab == TabFindings && len(model.report.Findings) == 0:
		return noFindingsMessage
	case model.activeTab == TabAllPorts && len(model.report.Ports) == 0:
		return noPortsMessage
	default:
		return noMatchMessage
	}
}

func (model Model) tableHeadingLine(spec tableSpec, shown []column, flexWidth int) string {
	cells := make([]string, 0, len(shown)+1)
	for _, candidate := range shown {
		cells = append(cells, fitWidth(strings.Repeat(columnSeparator, candidate.headingIndent)+candidate.heading, candidate.width))
	}
	cells = append(cells, fitWidth(spec.flexHeading, flexWidth))
	leading := noSelectionMark + columnSeparator
	return model.styles.heading.Render(leading + strings.Join(cells, columnSeparator))
}

func (model Model) tableRowLine(index int, spec tableSpec, shown []column, flexWidth int) string {
	row := model.visibleRows[index]
	selected := index == model.cursor
	gap := model.styles.rowGap(selected).Render(columnSeparator)
	mark := model.styles.plain.Render(noSelectionMark)
	if selected {
		mark = model.styles.title.Background(colorSelected).Render(selectionMark)
	}
	var builder strings.Builder
	builder.WriteString(mark)
	builder.WriteString(gap)
	for _, candidate := range shown {
		builder.WriteString(model.styles.renderCell(candidate.content(model.styles, row), candidate.width, selected))
		builder.WriteString(gap)
	}
	builder.WriteString(model.styles.renderCell(spec.flexContent(model.styles, row), flexWidth, selected))
	return builder.String()
}

// tablePanel wraps the table in a titled rounded border.
func (model Model) tablePanel(frame layout) []string {
	return model.styles.renderPanel(model.tablePanelTitle(), model.tableContent(frame), frame.contentWidth)
}

// tablePanelTitle shows the tab, the cursor position and the active view options.
func (model Model) tablePanelTitle() string {
	position := fmt.Sprintf(panelPositionEmpty, model.activeTab.Title())
	if len(model.visibleRows) > 0 {
		position = fmt.Sprintf(panelPositionFormat, model.activeTab.Title(), model.cursor+1, len(model.visibleRows))
	}
	var status string
	switch model.activeTab {
	case TabChanges:
		return position + panelChangesNote + columnSeparator
	case TabConnections:
		status = fmt.Sprintf(connectionStatusFormat, model.connectionSortMode, model.loopbackState(), model.inboundState())
	case TabFirewallPorts:
		status = fmt.Sprintf(panelFirewallStatusFmt, model.firewallSortMode)
	default:
		status = fmt.Sprintf(panelStatusFormat, model.sortMode, model.minSeverity)
	}
	if filterText := model.activeFilterText(); filterText != "" {
		status += fmt.Sprintf(panelFilterFormat, filterText)
	}
	return position + status + columnSeparator
}
