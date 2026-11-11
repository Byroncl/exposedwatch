package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// viewCount returns the number shown next to an inner tab title, and whether
// that tab has one.
func (model Model) viewCount(view TabID) (int, bool) {
	switch view {
	case TabFindings:
		return len(model.report.Findings), true
	case TabAllPorts:
		return len(model.report.Ports), true
	case TabChanges:
		return len(model.tracker.log), true
	case TabConnections:
		return model.connectionDisplayCount, true
	default:
		return 0, false
	}
}

// tabLine renders the inner tab strip of the active module. The active tab is
// bracketed and highlighted. When the full labels do not fit, only the tab
// numbers are shown.
func (model Model) tabLine(width int) string {
	line := model.renderTabs(false)
	if lipgloss.Width(line) > width {
		line = model.renderTabs(true)
	}
	return clampWidth(line, width)
}

func (model Model) renderTabs(compact bool) string {
	specs := model.activeModule.innerTabs()
	parts := make([]string, 0, len(specs))
	for position, spec := range specs {
		active := InnerTab(position) == model.innerTabs[model.activeModule]
		label := model.tabLabel(position+1, spec, active, compact)
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

func (model Model) tabLabel(number int, spec innerTabSpec, active, compact bool) string {
	count, hasCount := model.viewCount(spec.view)
	switch {
	case compact && active:
		return fmt.Sprintf(compactActiveTab, number)
	case compact:
		return fmt.Sprintf(compactInactiveTab, number)
	case active && hasCount:
		return fmt.Sprintf(activeTabFormat, number, spec.title, count)
	case active:
		return fmt.Sprintf(activeTabPlainFormat, number, spec.title)
	case hasCount:
		return fmt.Sprintf(inactiveTabFormat, number, spec.title, count)
	default:
		return fmt.Sprintf(inactiveTabPlainFormat, number, spec.title)
	}
}
