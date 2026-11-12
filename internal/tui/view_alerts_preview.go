package tui

import (
	"fmt"
	"strings"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

// alertsPreviewPanel renders the Preview inner tab of the Alerts module: the
// findings that would alert now, computed from the live report.
func (model Model) alertsPreviewPanel(frame layout) []string {
	innerWidth := panelContentWidth(frame.contentWidth)
	limit := max(frame.panelLines-panelBorderLineCount, 0)
	lines := model.alertsPreviewLines(innerWidth, limit)
	lines = firstLines(lines, limit)
	lines = append(lines, blankLines(limit-len(lines), innerWidth)...)
	return model.styles.renderPanel(panelAlertsPreviewTitle, lines, frame.contentWidth)
}

func (model Model) alertsPreviewLines(innerWidth, limit int) []string {
	state := model.settings
	switch {
	case !state.isAvailable():
		return wrapDim(model.styles, alertsNotAvailableText, innerWidth)
	case !state.loaded:
		return wrapDim(model.styles, settingsLoadingText, innerWidth)
	case state.loadFailed:
		return wrapStyled(model.styles.warning, fmt.Sprintf(alertsLoadFailedFormat, state.loadFailure), innerWidth)
	}
	var head []string
	heading := previewWouldAlertFormat
	switch {
	case !model.alertsConfigured():
		head = wrapStyled(model.styles.warning, previewNotConfiguredText, innerWidth)
		heading = previewWouldIfFormat
	case !state.baseline.enabled:
		head = wrapStyled(model.styles.warning, previewDisabledText, innerWidth)
		heading = previewWouldIfFormat
	}
	matching := model.alertingFindings()
	tail := append([]string{""}, wrapDim(model.styles, previewSemanticsText, innerWidth)...)
	lines := append(head, model.styles.heading.Render(fmt.Sprintf(heading, len(matching))))
	rowBudget := limit - len(lines) - len(tail)
	return append(append(lines, model.previewRows(matching, rowBudget, innerWidth)...), tail...)
}

// alertingFindings lists the current findings at or above the saved minimum
// severity, worst first.
func (model Model) alertingFindings() []domain.Finding {
	var matching []domain.Finding
	for _, finding := range model.report.Findings {
		if finding.Severity >= model.settings.baseline.minSeverity {
			matching = append(matching, finding)
		}
	}
	return worstFindings(matching, len(matching))
}

// previewRows renders the findings that fit the budget and counts the rest.
func (model Model) previewRows(matching []domain.Finding, budget, innerWidth int) []string {
	switch {
	case !model.hasReport:
		return []string{model.styles.dim.Render(scanningMessage)}
	case len(matching) == 0:
		return []string{model.styles.dim.Render(fmt.Sprintf(previewNoFindingsFormat, model.settings.baseline.minSeverity))}
	}
	shown := len(matching)
	if shown > budget {
		shown = max(budget-1, 0)
	}
	rows := make([]string, 0, shown+1)
	for _, finding := range matching[:shown] {
		rows = append(rows, model.previewRow(finding, innerWidth))
	}
	if shown < len(matching) {
		rows = append(rows, model.styles.dim.Render(fmt.Sprintf(previewMoreFormat, len(matching)-shown)))
	}
	return rows
}

// previewRow is the severity chip, the port, the process and the title.
func (model Model) previewRow(finding domain.Finding, innerWidth int) string {
	chip := model.styles.chipFor(finding.Severity).Render(fitWidth(" "+strings.ToUpper(finding.Severity.String())+" ", findingChipWidth))
	port := fitWidth(fmt.Sprintf(portFormat, finding.Port.Number, finding.Port.Protocol), portColumnWidth)
	process := fitWidth(processOrContainer(finding.Port), processColumnWidth)
	titleWidth := innerWidth - findingChipWidth - portColumnWidth - processColumnWidth - 3*len(columnSeparator)
	title := ""
	if titleWidth > 0 {
		title = fitWidth(finding.Title, titleWidth)
	}
	return fmt.Sprintf(previewRowFormat, chip, port, process, title)
}
