package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// tabCounts returns the number shown next to each tab title.
func (model Model) tabCounts() [tabCount]int {
	return [tabCount]int{len(model.report.Findings), len(model.report.Ports), len(model.tracker.log)}
}

// tabLine renders the tab strip. The active tab is bracketed and highlighted.
// When the full labels do not fit, only the tab numbers are shown.
func (model Model) tabLine(width int) string {
	line := model.renderTabs(false)
	if lipgloss.Width(line) > width {
		line = model.renderTabs(true)
	}
	return clampWidth(line, width)
}

func (model Model) renderTabs(compact bool) string {
	counts := model.tabCounts()
	parts := make([]string, 0, int(tabCount))
	for tab := TabFindings; tab < tabCount; tab++ {
		active := tab == model.activeTab
		var label string
		switch {
		case compact && active:
			label = fmt.Sprintf(compactActiveTab, int(tab)+1)
		case compact:
			label = fmt.Sprintf(compactInactiveTab, int(tab)+1)
		case active:
			label = fmt.Sprintf(activeTabFormat, int(tab)+1, tab.Title(), counts[tab])
		default:
			label = fmt.Sprintf(inactiveTabFormat, int(tab)+1, tab.Title(), counts[tab])
		}
		if active {
			parts = append(parts, model.styles.tabActive.Render(label))
		} else {
			parts = append(parts, model.styles.tabInactive.Render(label))
		}
	}
	// A dimmed divider with a gap on both sides keeps the tabs from looking
	// glued together; it has the same width whichever tab is active.
	separator := tabGap + model.styles.dim.Render(tabDivider) + tabGap
	return strings.Join(parts, separator)
}
