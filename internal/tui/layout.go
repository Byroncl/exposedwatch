package tui

import "fmt"

// layout is the vertical and horizontal budget of one frame.
type layout struct {
	width    int
	height   int
	warnings []string
	// showDetail tells whether the bottom area (the Details panel, or the
	// summary panels of the Connections tab) fits.
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
	if len(lines) > maxWarningLines {
		hidden := len(lines) - (maxWarningLines - 1)
		lines = append(lines[:maxWarningLines-1], fmt.Sprintf(moreWarningsFormat, hidden))
	}
	return lines
}

// computeLayout decides how many rows the table and the detail panel get.
func (model Model) computeLayout() layout {
	width, height := model.effectiveSize()
	warnings := model.warningLines()
	fixedLines := headerLineCount + tabLineCount + len(warnings) + footerLineCount
	bodyLines := height - fixedLines
	tableChrome := panelBorderLineCount + tableHeadingLines
	detailOuter := model.bottomOuterLines(width)
	showDetail := model.hasBottomArea() && bodyLines-detailOuter-tableChrome >= minTableRowsToShow
	tableOuter := bodyLines
	if showDetail {
		tableOuter -= detailOuter
	}
	return layout{
		width:      width,
		height:     height,
		warnings:   warnings,
		showDetail: showDetail,
		tableRows:  max(tableOuter-tableChrome, 0),
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
