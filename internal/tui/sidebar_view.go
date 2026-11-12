package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// sidebarPanel renders the module list in a rounded border as tall as the
// column area. Every line is exactly frame.sidebarWidth cells wide.
func (model Model) sidebarPanel(frame layout) []string {
	if frame.sidebar == SidebarHidden {
		return nil
	}
	innerWidth := frame.sidebarWidth - sidebarBorderSize
	contentLines := max(frame.columnLines-panelBorderLineCount, 0)
	rows := make([]string, 0, int(moduleCount))
	for module := ModuleOverview; module < moduleCount; module++ {
		rows = append(rows, model.sidebarRow(module, frame.sidebar, innerWidth))
	}
	rows = firstLines(rows, contentLines)
	rows = append(rows, blankLines(contentLines-len(rows), innerWidth)...)
	border := lipgloss.RoundedBorder()
	side := model.styles.border.Render(border.Left)
	sideRight := model.styles.border.Render(border.Right)
	lines := make([]string, 0, len(rows)+panelBorderLineCount)
	lines = append(lines, model.styles.border.Render(border.TopLeft+strings.Repeat(border.Top, innerWidth)+border.TopRight))
	for _, row := range rows {
		lines = append(lines, side+row+sideRight)
	}
	return append(lines, model.styles.border.Render(border.BottomLeft+strings.Repeat(border.Bottom, innerWidth)+border.BottomRight))
}

// sidebarRow renders one module line of exactly innerWidth cells. The selected
// module gets the marker (a plain-text cue) and a highlighted background.
func (model Model) sidebarRow(module Module, mode SidebarMode, innerWidth int) string {
	selected := module == model.activeModule
	look := model.styles.plain
	markerLook := look
	if selected {
		look = look.Background(colorSelected).Bold(true)
		markerLook = model.styles.title.Background(colorSelected)
	}
	marker := sidebarNoMarker
	if selected {
		marker = selectionMark
	}
	if mode == SidebarCompact {
		label := fitWidth(module.compactLabel(), innerWidth-selectionMarkWidth)
		return markerLook.Render(marker) + look.Render(label)
	}
	badge := model.sidebarBadgeFor(module)
	left := marker + sidebarMarkerGap + module.Title()
	if badge.width() > 0 && lipgloss.Width(left)+len(sidebarBadgeJoin)+badge.width() > innerWidth {
		badge = sidebarBadge{}
	}
	name := fitWidth(module.Title(), innerWidth-selectionMarkWidth-len(sidebarMarkerGap)-badgeReserve(badge))
	padding := innerWidth - selectionMarkWidth - len(sidebarMarkerGap) - lipgloss.Width(name) - badge.width()
	return markerLook.Render(marker) + look.Render(sidebarMarkerGap+name+strings.Repeat(columnSeparator, max(padding, 0))) +
		model.renderBadge(badge, look)
}

// badgeReserve is the width kept for the badge and the gap before it.
func badgeReserve(badge sidebarBadge) int {
	if badge.width() == 0 {
		return 0
	}
	return badge.width() + len(sidebarBadgeJoin)
}

// renderBadge draws the badge text with the row look and its flag highlighted.
func (model Model) renderBadge(badge sidebarBadge, look lipgloss.Style) string {
	switch {
	case badge.width() == 0:
		return ""
	case badge.flag == "":
		return look.Render(badge.text)
	case badge.text == "":
		return model.styles.newBadge.Render(badge.flag)
	default:
		return look.Render(badge.text+sidebarBadgeJoin) + model.styles.newBadge.Render(badge.flag)
	}
}
