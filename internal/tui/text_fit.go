package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// clampWidth truncates a (possibly styled) single line to at most width cells.
func clampWidth(line string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(line) <= width {
		return line
	}
	return lipgloss.NewStyle().MaxWidth(width).Render(line)
}

// ellipsis marks text that was cut to fit.
const ellipsis = "…"

// truncateWithEllipsis cuts a (possibly styled) line to at most width cells,
// ending it with a single ellipsis when anything was removed.
func truncateWithEllipsis(line string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(line) <= width {
		return line
	}
	if width == lipgloss.Width(ellipsis) {
		return ellipsis
	}
	return clampWidth(line, width-lipgloss.Width(ellipsis)) + ellipsis
}

// fitWidth truncates (with an ellipsis) or pads a single line to exactly width cells.
func fitWidth(line string, width int) string {
	clamped := truncateWithEllipsis(line, width)
	if padding := width - lipgloss.Width(clamped); padding > 0 {
		return clamped + strings.Repeat(" ", padding)
	}
	return clamped
}

// wrapText word-wraps plain text to the width, one string per line.
func wrapText(text string, width int) []string {
	if width <= 0 {
		return nil
	}
	return strings.Split(lipgloss.NewStyle().Width(width).Render(text), "\n")
}

// blankLines returns count empty lines of the given width.
func blankLines(count, width int) []string {
	lines := make([]string, 0, max(count, 0))
	for range max(count, 0) {
		lines = append(lines, strings.Repeat(" ", max(width, 0)))
	}
	return lines
}
