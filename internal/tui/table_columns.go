package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/Byroncl/exposedwatch/internal/app"
)

// cell is the content and look of one table cell. A chip is drawn with a
// colored background and one cell of padding on each side.
type cell struct {
	text string
	look lipgloss.Style
	chip bool
}

// column is one fixed-width table column. Lower priority numbers are kept
// first when the terminal is too narrow; priorityMandatory is always kept.
type column struct {
	heading  string
	width    int
	priority int
	// headingIndent shifts the heading right so it sits above chip text.
	headingIndent int
	content       func(palette styles, row displayRow) cell
}

// tableSpec is the column set of one tab: fixed columns plus one flexible
// last column that takes the remaining width.
type tableSpec struct {
	fixed       []column
	flexHeading string
	flexContent func(palette styles, row displayRow) cell
	// minFlexWidth is the width kept for the flexible column before optional
	// fixed columns are added.
	minFlexWidth int
}

func tableSpecFor(tab TabID) tableSpec {
	switch tab {
	case TabChanges:
		return changesSpec()
	case TabConnections:
		return connectionsSpec()
	default:
		return portsSpec()
	}
}

func portsSpec() tableSpec {
	return tableSpec{
		fixed: []column{
			{headingSeverity, severityColumnWidth, priorityMandatory, chipSidePadding, severityCell},
			{headingNew, newColumnWidth, priorityMandatory, 0, newBadgeCell},
			{headingPort, portColumnWidth, priorityMandatory, 0, portCell},
			{headingExposure, exposureColumnWidth, priorityExposure, 0, exposureCell},
			{headingBind, bindColumnWidth, priorityBind, 0, bindCell},
			{headingSource, sourceColumnWidth, prioritySource, 0, sourceCell},
			{headingProcessShort, processColumnWidth, priorityProcess, 0, processCell},
		},
		flexHeading:  headingFinding,
		flexContent:  findingTitleCell,
		minFlexWidth: minTitleWidth,
	}
}

func changesSpec() tableSpec {
	return tableSpec{
		fixed: []column{
			{headingTime, timeColumnWidth, priorityMandatory, 0, changeTimeCell},
			{headingKind, kindColumnWidth, priorityMandatory, chipSidePadding, changeKindCell},
			{headingPort, portColumnWidth, priorityMandatory, 0, portCell},
			{headingSeverity, severityColumnWidth, priorityExposure, chipSidePadding, severityCell},
			{headingBind, bindColumnWidth, priorityBind, 0, bindCell},
			{headingSource, sourceColumnWidth, prioritySource, 0, sourceCell},
		},
		flexHeading:  headingProcess,
		flexContent:  processCell,
		minFlexWidth: minTitleWidth,
	}
}

// layoutColumns keeps the columns that fit, in priority order, while leaving
// at least minFlexWidth cells for the flexible column. It returns the kept
// columns in display order and the width of the flexible column.
func (spec tableSpec) layoutColumns(contentWidth int) ([]column, int) {
	used := selectionMarkWidth + len(columnSeparator)
	kept := make(map[int]bool, len(spec.fixed))
	order := make([]int, len(spec.fixed))
	for index := range order {
		order[index] = index
	}
	sort.SliceStable(order, func(left, right int) bool {
		return spec.fixed[order[left]].priority < spec.fixed[order[right]].priority
	})
	for _, index := range order {
		candidate := spec.fixed[index]
		cost := candidate.width + len(columnSeparator)
		if candidate.priority != priorityMandatory && contentWidth-(used+cost) < spec.minFlexWidth {
			continue
		}
		kept[index] = true
		used += cost
	}
	shown := make([]column, 0, len(kept))
	for index, candidate := range spec.fixed {
		if kept[index] {
			shown = append(shown, candidate)
		}
	}
	return shown, max(contentWidth-used, 0)
}

func severityCell(palette styles, row displayRow) cell {
	if !row.hasSeverity {
		return cell{text: strings.Repeat(columnSeparator, chipSidePadding) + emptyValue, look: palette.dim}
	}
	return cell{text: strings.ToUpper(row.severity.String()), look: palette.chipFor(row.severity), chip: true}
}

func newBadgeCell(palette styles, row displayRow) cell {
	if !row.isNew {
		return cell{look: palette.plain}
	}
	return cell{text: newBadgeText, look: palette.newBadge}
}

func portCell(palette styles, row displayRow) cell {
	return cell{text: fmt.Sprintf(portFormat, row.port.Number, row.port.Protocol), look: palette.plain}
}

func exposureCell(palette styles, row displayRow) cell {
	return cell{text: row.port.Exposure().String(), look: palette.plain}
}

func bindCell(palette styles, row displayRow) cell {
	return cell{text: row.port.BindAddress, look: palette.plain}
}

func sourceCell(palette styles, row displayRow) cell {
	return cell{text: string(row.port.Source), look: palette.plain}
}

func processCell(palette styles, row displayRow) cell {
	return cell{text: processOrContainer(row.port), look: palette.plain}
}

func findingTitleCell(palette styles, row displayRow) cell {
	if !row.hasSeverity {
		return cell{text: emptyValue, look: palette.dim}
	}
	return cell{text: row.title, look: palette.plain}
}

func changeTimeCell(palette styles, row displayRow) cell {
	return cell{text: row.changedAt.Format(timeLayoutShort), look: palette.dim}
}

func changeKindCell(palette styles, row displayRow) cell {
	if row.changeKind == app.ChangeClosed {
		return cell{text: kindClosedLabel, look: palette.chipText.Background(colorClosed), chip: true}
	}
	return cell{text: kindOpenedLabel, look: palette.chipText.Background(colorOpened), chip: true}
}
