package tui

import (
	"fmt"
	"strings"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

// telegramFormPanel renders the Telegram inner tab of the Settings module.
func (model Model) telegramFormPanel(frame layout) []string {
	innerWidth := panelContentWidth(frame.contentWidth)
	limit := max(frame.panelLines-panelBorderLineCount, 0)
	lines, selectedLine := model.telegramFormLines(innerWidth)
	lines = windowAround(lines, selectedLine, limit)
	lines = append(lines, blankLines(limit-len(lines), innerWidth)...)
	return model.styles.renderPanel(panelTelegramTitle, lines, frame.contentWidth)
}

// windowAround keeps at most limit lines and always includes the selected one.
func windowAround(lines []string, selected, limit int) []string {
	if limit <= 0 {
		return nil
	}
	start := 0
	if selected >= limit {
		start = selected - limit + 1
	}
	end := min(start+limit, len(lines))
	return lines[min(start, end):end]
}

// telegramFormLines lists every line of the form and the index of the one
// that holds the selection.
func (model Model) telegramFormLines(innerWidth int) ([]string, int) {
	state := model.settings
	switch {
	case !state.isAvailable():
		return wrapDim(model.styles, settingsNotAvailable, innerWidth), 0
	case !state.loaded:
		return wrapDim(model.styles, settingsLoadingText, innerWidth), 0
	case state.loadFailed:
		return model.loadFailureLines(innerWidth), 0
	}
	var lines []string
	selectedLine := 0
	for field := FieldEnabled; field < formFieldCount; field++ {
		if field == FieldSave {
			lines = append(lines, "")
		}
		if field == state.cursor {
			selectedLine = len(lines)
		}
		lines = append(lines, model.formRowLine(field, innerWidth))
		if note := model.sourceNote(field); note != "" {
			lines = append(lines, model.styles.dim.Render(fitWidth(note, innerWidth)))
		}
	}
	lines = append(lines, "")
	return append(lines, model.formInfoLines(innerWidth)...), selectedLine
}

func (model Model) loadFailureLines(innerWidth int) []string {
	lines := wrapStyled(model.styles.warning, fmt.Sprintf(loadFailedFormat, model.redacted(model.settings.loadFailure)), innerWidth)
	return append(lines, wrapDim(model.styles, loadFailedHint, innerWidth)...)
}

// wrapStyled wraps plain text to the width and renders every line in the style.
func wrapStyled(look styleRenderer, text string, width int) []string {
	wrapped := wrapText(text, width)
	for index, line := range wrapped {
		wrapped[index] = look.Render(line)
	}
	return wrapped
}

// formRowLine renders one row: the marker, the label and the value, or the
// button. The selected row is highlighted and the one being edited is warned.
func (model Model) formRowLine(field FormField, innerWidth int) string {
	state := model.settings
	selected := field == state.cursor
	marker := noSelectionMark
	if selected {
		marker = selectionMark
	}
	text := marker + columnSeparator + field.label()
	if !field.isAction() {
		text = marker + columnSeparator + fitWidth(field.label(), formLabelWidth) + columnSeparator + model.formValueText(field)
	}
	text = fitWidth(text, innerWidth)
	switch {
	case selected && model.viewState == ViewFieldEdit:
		return model.styles.warning.Bold(true).Render(text)
	case selected:
		return model.styles.selected.Render(text)
	case field.isAction():
		return model.styles.heading.Render(text)
	default:
		return model.styles.plain.Render(text)
	}
}

// formValueText is what a row shows after its label. The token is only ever
// shown masked, or as one bullet per typed rune while it is being typed.
func (model Model) formValueText(field FormField) string {
	state := model.settings
	editing := model.viewState == ViewFieldEdit && state.cursor == field
	switch field {
	case FieldEnabled:
		if state.form.enabled {
			return formEnabledText
		}
		return formDisabledText
	case FieldToken:
		return model.tokenValueText(editing)
	case FieldChatID:
		if editing {
			return state.editor.text() + formEditCursor
		}
		return model.chatValueText()
	case FieldMinSeverity:
		return state.form.minSeverity.String()
	default:
		if editing {
			return state.editor.text() + formEditCursor
		}
		return state.form.numberText(field)
	}
}

func (model Model) tokenValueText(editing bool) string {
	state := model.settings
	if editing {
		return strings.Repeat(secretBullet, state.editor.length()) + formEditCursor
	}
	masked := maskToken(state.form.token)
	if masked == "" {
		masked = formNotSetText
	}
	return masked
}

func (model Model) chatValueText() string {
	state := model.settings
	value := model.redacted(state.form.chatID)
	if value == "" {
		value = formNotSetText
	}
	return value
}

// sourceNote explains, under a row, that the environment owns its value.
func (model Model) sourceNote(field FormField) string {
	state := model.settings
	if !state.isReadOnly(field) {
		return ""
	}
	name := state.meta.ChatEnvironmentName
	if field == FieldToken {
		name = state.meta.TokenEnvironmentName
	}
	name = model.redacted(name)
	return strings.Repeat(columnSeparator, selectionMarkWidth+len(columnSeparator)+formLabelWidth+len(columnSeparator)) +
		fmt.Sprintf(fromEnvironmentFmt, name)
}

// formInfoLines is the area under the buttons: the unsaved marker, the result
// of the last action, the test message, the file and its warnings.
func (model Model) formInfoLines(innerWidth int) []string {
	state := model.settings
	var lines []string
	if state.isDirty() {
		lines = append(lines, model.styles.warning.Bold(true).Render(unsavedChangesText))
	}
	if state.status.text != "" {
		lines = append(lines, wrapStyled(model.statusStyle(state.status.kind), model.redacted(state.status.text), innerWidth)...)
	}
	lines = append(lines, model.testMessageLines(innerWidth)...)
	if state.meta.Path != "" {
		lines = append(lines, wrapDim(model.styles, fmt.Sprintf(configPathFormat, model.redacted(state.meta.Path)), innerWidth)...)
	}
	if state.refreshFailure != "" {
		text := fmt.Sprintf(warningLineFormat, fmt.Sprintf(refreshFailedFormat, model.redacted(state.refreshFailure)))
		lines = append(lines, wrapStyled(model.styles.warning, text, innerWidth)...)
	}
	return append(lines, model.warningLinesOf(innerWidth)...)
}

// warningLinesOf lists the load warnings of the config file, redacted.
func (model Model) warningLinesOf(innerWidth int) []string {
	var lines []string
	for _, warning := range model.settings.meta.Warnings {
		text := fmt.Sprintf(warningLineFormat, model.redacted(warning))
		lines = append(lines, wrapStyled(model.styles.warning, text, innerWidth)...)
	}
	return lines
}

// redacted removes the bot token from text coming from outside the model and
// then every control and invisible character, so a config value, a warning or
// a path can never send a terminal escape sequence.
func (model Model) redacted(text string) string {
	return domain.SanitizeLine(redactSecrets(text, model.settings.form.token, model.settings.baseline.token))
}

func (model Model) statusStyle(kind statusKind) styleRenderer {
	switch kind {
	case statusSuccess:
		return model.styles.safe
	case statusError:
		return model.styles.warning
	default:
		return model.styles.plain
	}
}

// testMessageLines shows the state of the test message of this session.
func (model Model) testMessageLines(innerWidth int) []string {
	lastTest := model.settings.testMessage
	switch lastTest.phase {
	case testPhaseSending:
		return wrapStyled(model.styles.plain, testSendingText, innerWidth)
	case testPhaseSucceeded:
		return wrapStyled(model.styles.safe, fmt.Sprintf(testSentFormat, lastTest.finishedAt.Format(timeLayout)), innerWidth)
	case testPhaseFailed:
		return wrapStyled(model.styles.warning, fmt.Sprintf(testFailedFormat, lastTest.finishedAt.Format(timeLayout), lastTest.failure), innerWidth)
	default:
		return nil
	}
}
