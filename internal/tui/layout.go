package tui

import "fmt"

// layout is the vertical and horizontal budget of one frame.
type layout struct {
	width    int
	height   int
	warnings []string
	// headerLines is the height of the header; columnLines is the height left
	// for the sidebar and the content area between the header and the footer.
	headerLines int
	columnLines int
	// sidebar is the sidebar mode, sidebarWidth its outer width (0 when hidden)
	// and contentWidth the width of the content area next to it.
	sidebar      SidebarMode
	sidebarWidth int
	contentWidth int
	// panelLines is the height of the content area below the tab strip.
	panelLines int
	// showDetail tells whether the bottom area (the Details panel, or the
	// summary panels of the Live connections tab) fits.
	showDetail bool
	tableRows  int
}

func (model Model) effectiveSize() (int, int) {
	width, height := model.width, model.height
	if width <= 0 {
		width = fallbackWidth
	}
	if height <= 0 {
		height = fallbackHeight
	}
	return width, height
}

// warningLines lists non-fatal problems: failed sources and refresh errors.
func (model Model) warningLines() []string {
	var lines []string
	for _, sourceError := range model.report.SourceErrors {
		lines = append(lines, fmt.Sprintf(sourceWarningFormat, sourceError.Error()))
	}
	if model.refreshErr != nil {
		lines = append(lines, fmt.Sprintf(refreshWarningFormat, model.refreshErr))
	}
	lines = append(lines, model.connectionWarningLines()...)
	lines = append(lines, model.firewallWarningLines()...)
	if len(lines) > maxWarningLines {
		hidden := len(lines) - (maxWarningLines - 1)
		lines = append(lines[:maxWarningLines-1], fmt.Sprintf(moreWarningsFormat, hidden))
	}
	return lines
}

// computeLayout decides the sidebar, the content width and how many rows the
// table and the detail panel get.
func (model Model) computeLayout() layout {
	width, height := model.effectiveSize()
	sidebar := sidebarModeFor(width, model.sidebarHidden)
	sidebarWidth := sidebar.cellWidth()
	contentWidth := width
	if sidebarWidth > 0 {
		contentWidth = width - sidebarWidth - sidebarContentGap
	}
	warnings := model.warningLines()
	headerLines := model.headerLineCount()
	columnLines := max(height-headerLines-len(warnings)-footerLineCount, 0)
	panelLines := max(columnLines-tabLineCount, 0)
	tableChrome := panelBorderLineCount + tableHeadingLines
	detailOuter := model.bottomOuterLines(contentWidth)
	showDetail := model.hasBottomArea() && panelLines-detailOuter-tableChrome >= minTableRowsToShow
	tableOuter := panelLines
	if showDetail {
		tableOuter -= detailOuter
	}
	return layout{
		width:        width,
		height:       height,
		warnings:     warnings,
		headerLines:  headerLines,
		columnLines:  columnLines,
		sidebar:      sidebar,
		sidebarWidth: sidebarWidth,
		contentWidth: contentWidth,
		panelLines:   panelLines,
		showDetail:   showDetail,
		tableRows:    max(tableOuter-tableChrome, 0),
	}
}

// fitLines keeps the footer first, then the header, then the body.
func fitLines(headerPart, bodyPart, footerPart []string, height int) []string {
	remaining := height
	footerPart = firstLines(footerPart, remaining)
	remaining -= len(footerPart)
	headerPart = firstLines(headerPart, remaining)
	remaining -= len(headerPart)
	bodyPart = firstLines(bodyPart, remaining)
	fitted := make([]string, 0, height)
	fitted = append(fitted, headerPart...)
	fitted = append(fitted, bodyPart...)
	return append(fitted, footerPart...)
}

func firstLines(lines []string, limit int) []string {
	if limit < 0 {
		limit = 0
	}
	if len(lines) > limit {
		return lines[:limit]
	}
	return lines
}
