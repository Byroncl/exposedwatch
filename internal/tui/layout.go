package tui

import "fmt"

// layout is the vertical and horizontal budget of one frame.
type layout struct {
	width      int
	height     int
	warnings   []string
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
	if len(lines) > maxWarningLines {
		hidden := len(lines) - (maxWarningLines - 1)
		lines = append(lines[:maxWarningLines-1], fmt.Sprintf(moreWarningsFormat, hidden))
	}
	return lines
}

// computeLayout decides how many rows the table and detail pane get.
func (model Model) computeLayout() layout {
	width, height := model.effectiveSize()
	warnings := model.warningLines()
	fixedLines := headerLineCount + len(warnings) + footerLineCount + tableHeaderLineCount
	showDetail := height-fixedLines-detailPaneHeight >= minTableRowsToShow
	tableRows := height - fixedLines
	if showDetail {
		tableRows -= detailPaneHeight
	}
	if tableRows < 0 {
		tableRows = 0
	}
	return layout{width: width, height: height, warnings: warnings, showDetail: showDetail, tableRows: tableRows}
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
