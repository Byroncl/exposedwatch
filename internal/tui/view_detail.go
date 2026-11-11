package tui

import (
	"fmt"

	"github.com/Byroncl/exposedwatch/internal/app"
)

// detailPanel renders the pane for the selected row, always the same height.
func (model Model) detailPanel(width int) []string {
	contentWidth := panelContentWidth(width)
	lines := model.detailLines(contentWidth)
	lines = firstLines(lines, detailContentLines)
	lines = append(lines, blankLines(detailContentLines-len(lines), contentWidth)...)
	return model.styles.renderPanel(panelDetailsTitle, lines, width)
}

func (model Model) detailLines(contentWidth int) []string {
	if model.cursor < 0 || model.cursor >= len(model.visibleRows) {
		return []string{model.styles.dim.Render(detailEmptyMessage)}
	}
	row := model.visibleRows[model.cursor]
	if row.isChange {
		return model.changeDetailLines(row)
	}
	return model.portDetailLines(row, contentWidth)
}

func (model Model) portDetailLines(row displayRow, contentWidth int) []string {
	port := row.port
	headline := model.styles.dim.Render(detailNoFindingMessage)
	ruleID, severity := emptyValue, emptyValue
	if row.hasSeverity {
		headline = model.styles.forSeverity(row.severity).Bold(true).Render(row.title)
		ruleID, severity = row.ruleID, row.severity.String()
	}
	if row.isNew {
		headline += columnSeparator + model.styles.newBadge.Render(columnSeparator+newBadgeText+columnSeparator)
	}
	lines := []string{
		headline,
		fmt.Sprintf(detailRuleFormat, ruleID, severity, port.Source),
		model.bindDetail(row),
		model.processDetail(row),
		fmt.Sprintf(detailFirstSeenFormat, firstSeenText(row), stateText(row)),
	}
	if row.recommendation != "" {
		lines = append(lines, wrapText(fmt.Sprintf(detailFixFormat, row.recommendation), contentWidth)...)
	}
	return lines
}

func (model Model) changeDetailLines(row displayRow) []string {
	headline := model.styles.opened.Render(detailOpenedTitle)
	if row.changeKind == app.ChangeClosed {
		headline = model.styles.closed.Render(detailClosedTitle)
	}
	severity := emptyValue
	if row.hasSeverity {
		severity = row.severity.String()
	}
	return []string{
		headline,
		fmt.Sprintf(detailChangeFormat, row.changedAt.Format(timeLayout), row.changeKind, severity),
		model.bindDetail(row),
		model.processDetail(row),
		fmt.Sprintf(detailSourceFormat, row.port.Source),
	}
}

func (model Model) bindDetail(row displayRow) string {
	address := fmt.Sprintf(portBindFormat, row.port.BindAddress, row.port.Number, row.port.Protocol)
	return fmt.Sprintf(detailBindFormat, address, row.port.Exposure())
}

func (model Model) processDetail(row displayRow) string {
	process, container := row.port.Process, row.port.Container
	if process == "" {
		process = emptyValue
	}
	if container == "" {
		container = emptyValue
	}
	return fmt.Sprintf(detailProcessFormat, process, container)
}

func firstSeenText(row displayRow) string {
	if !row.hasFirstSeen {
		return detailUnknownFirstSeen
	}
	return row.firstSeen.Format(timeLayout)
}

func stateText(row displayRow) string {
	if row.isNew {
		return detailStateNew
	}
	return detailStateStable
}
