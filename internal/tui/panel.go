package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// panelContentWidth is the usable width inside a bordered panel.
func panelContentWidth(outerWidth int) int {
	return outerWidth - panelBorderWidth - panelPaddingWidth
}

// renderPanel draws a rounded border with the title on the top edge around the
// content lines. Every output line is exactly outerWidth cells wide. Panels
// narrower than minPanelWidth are not drawn.
func (palette styles) renderPanel(title string, content []string, outerWidth int) []string {
	if outerWidth < minPanelWidth {
		return nil
	}
	border := lipgloss.RoundedBorder()
	contentWidth := panelContentWidth(outerWidth)
	innerWidth := outerWidth - panelBorderWidth
	lines := make([]string, 0, len(content)+panelBorderLineCount)
	lines = append(lines, palette.topEdge(border, title, innerWidth))
	sideLeft := palette.border.Render(border.Left) + " "
	sideRight := " " + palette.border.Render(border.Right)
	for _, line := range content {
		lines = append(lines, sideLeft+fitWidth(line, contentWidth)+sideRight)
	}
	bottom := border.BottomLeft + strings.Repeat(border.Bottom, innerWidth) + border.BottomRight
	return append(lines, palette.border.Render(bottom))
}

func (palette styles) topEdge(border lipgloss.Border, title string, innerWidth int) string {
	available := innerWidth - panelTitleMargin
	if available < 1 || title == "" {
		return palette.border.Render(border.TopLeft + strings.Repeat(border.Top, innerWidth) + border.TopRight)
	}
	shown := clampWidth(title, available)
	fill := innerWidth - lipgloss.Width(shown) - 1
	return palette.border.Render(border.TopLeft+border.Top) + palette.panelTitle.Render(shown) +
		palette.border.Render(strings.Repeat(border.Top, fill)+border.TopRight)
}
