package tui

import "fmt"

// alertsStatusPanel renders the Status inner tab of the Alerts module.
func (model Model) alertsStatusPanel(frame layout) []string {
	innerWidth := panelContentWidth(frame.contentWidth)
	lines := model.alertsStatusLines(innerWidth)
	lines = firstLines(lines, max(frame.panelLines-panelBorderLineCount, 0))
	lines = append(lines, blankLines(frame.panelLines-panelBorderLineCount-len(lines), innerWidth)...)
	return model.styles.renderPanel(panelAlertsStatusTitle, lines, frame.contentWidth)
}

func (model Model) alertsStatusLines(innerWidth int) []string {
	state := model.settings
	switch {
	case !state.isAvailable():
		return wrapDim(model.styles, alertsNotAvailableText, innerWidth)
	case !state.loaded:
		return wrapDim(model.styles, settingsLoadingText, innerWidth)
	case state.loadFailed:
		lines := wrapStyled(model.styles.warning, fmt.Sprintf(alertsLoadFailedFormat, state.loadFailure), innerWidth)
		return append(append(lines, ""), wrapDim(model.styles, alertsHonestNote, innerWidth)...)
	}
	saved := state.baseline
	telegram := alertsNotConfiguredText
	if saved.isConfigured() {
		telegram = fmt.Sprintf(alertsConfiguredFormat, maskToken(saved.token))
	}
	enabled := alertsDisabledText
	if saved.enabled {
		enabled = alertsEnabledText
	}
	lines := []string{
		model.alertsRow(alertsLabelTelegram, telegram),
		model.alertsRow(alertsLabelAlerts, enabled),
		model.alertsRow(alertsLabelMinimum, saved.minSeverity.String()),
		model.alertsRow(alertsLabelCooldown, fmt.Sprintf(alertsCooldownFormat, saved.cooldownMinutes)),
		model.alertsRow(alertsLabelInterval, fmt.Sprintf(alertsIntervalFormat, saved.intervalSeconds)),
		model.alertsRow(alertsLabelFile, state.meta.Path),
		model.alertsRow(alertsLabelLastTest, model.lastTestText()),
	}
	lines = append(lines, model.warningLinesOf(innerWidth)...)
	lines = append(lines, "")
	return append(lines, wrapDim(model.styles, alertsHonestNote, innerWidth)...)
}

func (model Model) alertsRow(label, value string) string {
	return model.styles.plain.Render(fmt.Sprintf(alertsLabelFormat, label, model.redacted(value)))
}

// lastTestText describes the test message of this session.
func (model Model) lastTestText() string {
	lastTest := model.settings.testMessage
	switch lastTest.phase {
	case testPhaseSending:
		return alertsTestSendingText
	case testPhaseSucceeded:
		return fmt.Sprintf(alertsTestSentFormat, lastTest.finishedAt.Format(timeLayout))
	case testPhaseFailed:
		return fmt.Sprintf(alertsTestFailedFormat, lastTest.finishedAt.Format(timeLayout), lastTest.failure)
	default:
		return alertsNoTestText
	}
}
