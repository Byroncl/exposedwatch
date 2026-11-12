package tui

import (
	"fmt"
	"math"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// moveFormCursor moves the selection of the form. It does nothing while the
// form cannot be used.
func (model *Model) moveFormCursor(action Action) {
	if !model.settings.isEditable() {
		return
	}
	switch action {
	case ActionMoveUp:
		model.settings.cursor = model.settings.cursor.previous()
	case ActionMoveDown:
		model.settings.cursor = model.settings.cursor.next()
	case ActionJumpTop:
		model.settings.cursor = FieldEnabled
	case ActionJumpBottom:
		model.settings.cursor = formFieldCount - 1
	}
}

// activateField reacts to enter on the selected row: toggles and cycles act
// at once, text rows start editing and the buttons run their action.
func (model Model) activateField() (Model, tea.Cmd) {
	state := &model.settings
	if !state.isEditable() {
		if state.isAvailable() {
			state.status = formStatus{kind: statusError, text: statusUnavailableText}
		}
		return model, nil
	}
	field := state.cursor
	switch field.kind() {
	case kindToggle:
		state.form = state.form.withToggledEnabled()
	case kindCycle:
		state.form = state.form.withCycledSeverity()
	case kindAction:
		return model.runAction(field)
	default:
		model.beginEditing(field)
	}
	return model, nil
}

// runAction runs the Save or the Send test message button.
func (model Model) runAction(field FormField) (Model, tea.Cmd) {
	if field == FieldSave {
		return model.startSave()
	}
	return model.startTest()
}

// beginEditing opens the editor on a text row. Secrets start empty: the
// current token is never copied into the buffer. Rows owned by the
// environment are refused with a hint.
func (model *Model) beginEditing(field FormField) {
	state := &model.settings
	if state.isReadOnly(field) {
		state.status = formStatus{kind: statusInfo, text: model.environmentHint(field)}
		return
	}
	switch field.kind() {
	case kindText:
		state.editor, _ = editBuffer(nil).appended([]rune(state.form.chatID), maxEditLength, false)
	case kindNumber:
		state.editor = editBuffer(state.form.numberText(field))
	default:
		state.editor = nil
	}
	state.status = formStatus{}
	model.viewState = ViewFieldEdit
}

// environmentHint explains which variable owns a read-only row.
func (model Model) environmentHint(field FormField) string {
	if field == FieldToken {
		return fmt.Sprintf(statusEnvironmentFormat, field.label(), model.settings.meta.TokenEnvironmentName)
	}
	return fmt.Sprintf(statusEnvironmentFormat, field.label(), model.settings.meta.ChatEnvironmentName)
}

// startSave writes the form. Only one save runs at a time.
func (model Model) startSave() (Model, tea.Cmd) {
	state := &model.settings
	if state.saving {
		return model, nil
	}
	if !state.isDirty() {
		state.status = formStatus{kind: statusInfo, text: statusNoChanges}
		return model, nil
	}
	state.refreshFailure = ""
	state.saving = true
	state.pending = state.form
	state.status = formStatus{kind: statusInfo, text: statusSavingText}
	return model, settingsSaveCommand(state.store, state.form.alertSettings(state.meta))
}

// startTest sends a test message with the values in the form, saved or not.
// Only one is in flight at a time.
func (model Model) startTest() (Model, tea.Cmd) {
	state := &model.settings
	switch {
	case state.tester == nil:
		state.status = formStatus{kind: statusInfo, text: statusNoTester}
	case state.testMessage.phase == testPhaseSending:
		state.status = formStatus{kind: statusInfo, text: statusAlreadySending}
	case model.testCooldownLeft() > 0:
		seconds := int(math.Ceil(model.testCooldownLeft().Seconds()))
		state.status = formStatus{kind: statusInfo, text: fmt.Sprintf(statusWaitFormat, seconds)}
	case !state.form.isConfigured():
		state.status = formStatus{kind: statusError, text: statusNeedsCredentials}
	default:
		state.testMessage.phase = testPhaseSending
		state.status = formStatus{}
		return model, alertTestCommand(state.tester, state.form.alertSettings(state.meta), model.clock)
	}
	return model, nil
}

// testCooldownLeft is how long a new test message must still wait.
func (model Model) testCooldownLeft() time.Duration {
	finished := model.settings.testMessage.finishedAt
	if finished.IsZero() {
		return 0
	}
	return testMessageCooldown - model.clock().Sub(finished)
}
