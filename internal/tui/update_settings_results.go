package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

// handleSettingsLoaded stores the config the store read. A failed read keeps
// the form read-only. The re-read after a save only refreshes the metadata.
func (model Model) handleSettingsLoaded(result SettingsLoadedMessage) Model {
	state := &model.settings
	if result.Err != nil {
		if result.Refresh {
			state.refreshFailure = model.failureText(result.Err)
		} else {
			state.loaded = true
			state.loadFailed = true
			state.loadFailure = model.failureText(result.Err)
		}
		return model
	}
	values := formValuesFrom(result.Settings)
	// The metadata never keeps a second copy of the token.
	result.Settings.Token = ""
	state.meta = result.Settings
	if result.Refresh {
		state.refreshFailure = ""
		return model
	}
	state.loaded = true
	state.loadFailed = false
	state.loadFailure = ""
	state.form = values
	state.baseline = values
	return model
}

// handleSettingsSaved reports the outcome of a save and, on success, makes
// the saved values the baseline and re-reads the file for its warnings.
func (model Model) handleSettingsSaved(result SettingsSavedMessage) (Model, tea.Cmd) {
	state := &model.settings
	state.saving = false
	if result.Err != nil {
		state.status = formStatus{kind: statusError, text: fmt.Sprintf(statusSaveFailedFormat, model.failureText(result.Err))}
		return model, nil
	}
	state.baseline = state.pending
	state.status = formStatus{kind: statusSuccess, text: fmt.Sprintf(statusSavedFormat, state.meta.Path)}
	return model, settingsRefreshCommand(state.store)
}

// handleAlertTestResult stores the outcome of the test message.
func (model Model) handleAlertTestResult(result AlertTestResultMessage) Model {
	state := &model.settings
	state.testMessage.finishedAt = result.At
	if result.Err != nil {
		state.testMessage.phase = testPhaseFailed
		state.testMessage.failure = model.failureText(result.Err)
		return model
	}
	state.testMessage.phase = testPhaseSucceeded
	state.testMessage.failure = ""
	return model
}

// failureText turns an error into one safe line: the bot token is removed
// even though the adapters already redact, and control characters are dropped.
func (model Model) failureText(err error) string {
	text := redactSecrets(err.Error(), model.settings.form.token, model.settings.baseline.token, model.settings.pending.token)
	return truncateRunes(domain.SanitizeLine(text), maxFailureRunes)
}

// truncateRunes cuts text to limit runes, ending it with an ellipsis.
func truncateRunes(text string, limit int) string {
	runes := []rune(text)
	if len(runes) <= limit {
		return text
	}
	return string(runes[:limit-1]) + ellipsis
}
