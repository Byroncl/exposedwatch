package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

// handleFieldEditKey handles every key while a row is being edited. Typed
// runes edit the buffer, so no global key (quit, tab, help, digits...) acts
// here; only ctrl+c still quits, as in the filter prompt.
func (model Model) handleFieldEditKey(keyPress tea.KeyMsg) (tea.Model, tea.Cmd) {
	state := &model.settings
	switch keyPress.Type {
	case tea.KeyCtrlC:
		return model, tea.Quit
	case tea.KeyEnter:
		model.commitEdit()
	case tea.KeyEsc:
		state.editor = nil
		model.viewState = ViewTable
	case tea.KeyBackspace:
		state.editor = state.editor.withoutLast()
	case tea.KeyRunes:
		if !keyPress.Alt {
			isNumber := state.cursor.kind() == kindNumber
			var rejected bool
			state.editor, rejected = state.editor.appended(keyPress.Runes, model.editLimit(), isNumber)
			if rejected && isNumber {
				state.status = formStatus{kind: statusInfo, text: statusDigitsOnly}
			}
		}
	}
	return model, nil
}

// editLimit is how many runes the edited row accepts.
func (model Model) editLimit() int {
	if model.settings.cursor.kind() == kindNumber {
		return maxNumberLength
	}
	return maxEditLength
}

// commitEdit stores the typed text. A number that cannot be read keeps the
// editor open with a hint; every other row closes it.
func (model *Model) commitEdit() {
	state := &model.settings
	text := state.editor.text()
	field := state.cursor
	switch field.kind() {
	case kindSecret:
		if text != "" {
			state.form = state.form.withToken(domain.BotToken(text))
		}
	case kindText:
		state.form = state.form.withChatID(text)
	case kindNumber:
		updated, isNumber := state.form.withNumber(field, text)
		if !isNumber {
			state.status = formStatus{kind: statusError, text: statusNotWholeNumber}
			return
		}
		state.form = updated
	}
	state.editor = nil
	state.status = formStatus{}
	model.viewState = ViewTable
}
