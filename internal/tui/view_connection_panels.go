package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// bottomOuterLines is the height of the bottom area, borders included: the
// Details panel, or the connection summary panels.
func (model Model) bottomOuterLines(width int) int {
	if model.activeTab != TabConnections {
		return detailContentLines + panelBorderLineCount
	}
	if width >= connectionSideBySideMinWidth {
		return detailContentLines + panelBorderLineCount
	}
	return 2 * (connectionStackedContentLines + panelBorderLineCount)
}

// hasBottomArea tells whether the active view has a bottom area at all: only
// the table views do, and the disabled Live tab has nothing to summarize.
func (model Model) hasBottomArea() bool {
	if !model.activeTab.isTable() {
		return false
	}
	switch model.activeTab {
	case TabConnections:
		return model.connectionsEnabled()
	case TabFirewallPorts:
		return model.firewallEnabled()
	default:
		return true
	}
}

// bottomPanels renders the bottom area of the active tab.
func (model Model) bottomPanels(width int) []string {
	if model.activeTab == TabConnections {
		return model.connectionPanels(width, detailContentLines, connectionStackedContentLines)
	}
	return model.detailPanel(width)
}

// connectionPanels renders "Top remote addresses" and "Inbound by local port"
// side by side when wide enough (sideBySideLines rows each), stacked otherwise
// (stackedLines rows each).
func (model Model) connectionPanels(width, sideBySideLines, stackedLines int) []string {
	if width < connectionSideBySideMinWidth {
		stacked := model.topRemotePanel(width, stackedLines)
		return append(stacked, model.inboundPortsPanel(width, stackedLines)...)
	}
	leftWidth := (width - connectionPanelGap) / 2
	rightWidth := width - connectionPanelGap - leftWidth
	left := model.topRemotePanel(leftWidth, sideBySideLines)
	right := model.inboundPortsPanel(rightWidth, sideBySideLines)
	if len(left) != len(right) {
		return nil
	}
	joined := make([]string, 0, len(left))
	for index := range left {
		joined = append(joined, left[index]+strings.Repeat(columnSeparator, connectionPanelGap)+right[index])
	}
	return joined
}

func (model Model) topRemotePanel(width, contentLines int) []string {
	entries := make([]string, 0, len(model.connectionSummary.TopRemoteAddresses))
	for _, entry := range model.connectionSummary.TopRemoteAddresses {
		entries = append(entries, alignedPair(entry.Address, fmt.Sprintf(countFormat, entry.Count), panelContentWidth(width)))
	}
	return model.summaryPanel(panelTopRemoteTitle, entries, width, contentLines)
}

func (model Model) inboundPortsPanel(width, contentLines int) []string {
	entries := make([]string, 0, len(model.connectionSummary.PerLocalPort))
	for _, entry := range model.connectionSummary.PerLocalPort {
		entries = append(entries, alignedPair(fmt.Sprintf(inboundPortFormat, entry.LocalPort), fmt.Sprintf(countFormat, entry.Count), panelContentWidth(width)))
	}
	return model.summaryPanel(panelInboundPortsTitle, entries, width, contentLines)
}

// summaryPanel draws a panel of fixed height listing the entries, or a note
// when there are none. Entries beyond the height collapse into "... and N more".
func (model Model) summaryPanel(title string, entries []string, width, contentLines int) []string {
	contentWidth := panelContentWidth(width)
	var lines []string
	switch {
	case contentLines < 1:
		lines = nil
	case len(entries) == 0:
		lines = []string{model.styles.dim.Render(noInboundConnectionsMessage)}
	case len(entries) > contentLines:
		lines = append(lines, entries[:contentLines-1]...)
		lines = append(lines, model.styles.dim.Render(fmt.Sprintf(moreEntriesFormat, len(entries)-(contentLines-1))))
	default:
		lines = entries
	}
	lines = firstLines(lines, contentLines)
	lines = append(lines, blankLines(contentLines-len(lines), contentWidth)...)
	return model.styles.renderPanel(title, lines, width)
}

// alignedPair puts left at the start and right at the end of a line of the
// given width, cutting left with an ellipsis when they do not both fit.
func alignedPair(left, right string, width int) string {
	available := width - lipgloss.Width(right) - pairGapWidth
	if available < 1 {
		return fitWidth(left, width)
	}
	return fitWidth(left, available) + columnSeparator + right
}
