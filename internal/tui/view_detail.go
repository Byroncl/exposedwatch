package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// detailLines renders the pane for the selected finding, detailPaneHeight lines.
func (model Model) detailLines(width int) []string {
	lines := []string{model.styles.dim.MaxWidth(width).Render(strings.Repeat(detailSeparatorGlyph, width))}
	if model.cursor < 0 || model.cursor >= len(model.visibleFindings) {
		lines = append(lines, model.styles.dim.MaxWidth(width).Render(detailEmptyMessage))
		return lines
	}
	finding := model.visibleFindings[model.cursor]
	port := finding.Port
	text := []string{
		model.styles.forSeverity(finding.Severity).Render(finding.Title),
		fmt.Sprintf(detailRuleFormat, finding.RuleID, finding.Severity, port.Source),
		fmt.Sprintf(detailWhereFormat, bindLabel(port.BindAddress, port.Number, string(port.Protocol)), port.Exposure(), processOrContainer(port)),
		fmt.Sprintf(detailFixFormat, finding.Recommendation),
	}
	wrapper := lipgloss.NewStyle().Width(width)
	for _, paragraph := range text {
		lines = append(lines, strings.Split(wrapper.Render(paragraph), "\n")...)
	}
	return firstLines(lines, detailPaneHeight)
}

func bindLabel(address string, number int, protocol string) string {
	return fmt.Sprintf("%s:%d/%s", address, number, protocol)
}
