package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// connectionCounterFitting returns the "conn N (in X / out Y)" counter, joined
// to the header line, when it fits the width; otherwise an empty string. It is
// also empty while connections are disabled or not read yet. The counts follow
// the loopback and inbound-only toggles, like the tab.
func (model Model) connectionCounterFitting(line string, width int) string {
	if !model.connectionsEnabled() || !model.hasConnections {
		return ""
	}
	summary := model.connectionSummary
	counter := headerJoin + model.styles.plain.Render(fmt.Sprintf(headerConnectionCounterFmt, summary.Total, summary.Inbound, summary.Outbound))
	if lipgloss.Width(line)+lipgloss.Width(counter) > width {
		return ""
	}
	return counter
}
