package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// workArea renders the sidebar and the content area side by side. The content
// area is the inner tab strip followed by the view of the active inner tab.
func (model Model) workArea(frame layout) []string {
	content := append([]string{model.tabLine(frame.contentWidth)}, model.contentPanels(frame)...)
	if frame.sidebar == SidebarHidden {
		return content
	}
	return joinColumns(model.sidebarPanel(frame), frame.sidebarWidth, content)
}

// contentPanels renders the view of the active inner tab.
func (model Model) contentPanels(frame layout) []string {
	switch model.activeTab {
	case TabSummary:
		return model.overviewPanel(frame)
	case TabTopRemote:
		return model.topRemoteView(frame)
	case TabFirewall:
		return model.firewallStatusPanel(frame)
	case TabAlerts, TabSettings:
		return model.placeholderPanel(frame)
	default:
		panels := model.tablePanel(frame)
		if frame.showDetail {
			panels = append(panels, model.bottomPanels(frame.contentWidth)...)
		}
		return panels
	}
}

// joinColumns puts the left column, padded to leftWidth, then a gap, then the
// right column on each line. The shorter column is padded with blank lines.
func joinColumns(left []string, leftWidth int, right []string) []string {
	rows := max(len(left), len(right))
	gap := strings.Repeat(columnSeparator, sidebarContentGap)
	joined := make([]string, 0, rows)
	for index := range rows {
		leftLine := strings.Repeat(columnSeparator, leftWidth)
		if index < len(left) {
			leftLine = padWidth(left[index], leftWidth)
		}
		rightLine := ""
		if index < len(right) {
			rightLine = right[index]
		}
		joined = append(joined, leftLine+gap+rightLine)
	}
	return joined
}

// padWidth cuts a line to width cells, without an ellipsis, or pads it with spaces.
func padWidth(line string, width int) string {
	clamped := clampWidth(line, width)
	return clamped + strings.Repeat(columnSeparator, max(width-lipgloss.Width(clamped), 0))
}
