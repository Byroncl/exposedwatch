package tui

import (
	"fmt"
	"time"
)

// statusKind colors a status line.
type statusKind int

const (
	statusNone statusKind = iota
	statusInfo
	statusSuccess
	statusError
)

// formStatus is the last thing the form has to say about an action.
type formStatus struct {
	kind statusKind
	text string
}

// testPhase is the lifecycle of the test message.
type testPhase int

const (
	testPhaseIdle testPhase = iota
	testPhaseSending
	testPhaseSucceeded
	testPhaseFailed
)

// testState is the test message of this session.
type testState struct {
	phase      testPhase
	finishedAt time.Time
	// failure is the redacted, single-line reason of a failed test.
	failure string
}

// settingsState is everything the Settings and Alerts modules keep in the
// model. It holds the bot token, so Format keeps it out of fmt output.
type settingsState struct {
	store  SettingsStore
	tester AlertTester

	// loaded is true once the store answered; loadFailure is the sanitized
	// reason when it could not be read, which keeps the form read-only.
	loaded      bool
	loadFailed  bool
	loadFailure string
	// meta is the loaded settings without the editable values: path,
	// sources, environment names and warnings.
	meta AlertSettings

	// form is what the rows show, baseline what is on disk (or in the
	// environment) and pending what the running save is writing.
	form     formValues
	baseline formValues
	pending  formValues

	cursor FormField
	editor editBuffer
	status formStatus
	saving bool
	// refreshFailure is why the re-read after a successful save failed.
	refreshFailure string

	testMessage testState
}

// Format prints a placeholder whatever the verb.
func (state settingsState) Format(formatState fmt.State, _ rune) {
	_, _ = fmt.Fprint(formatState, settingsStatePlaceholder)
}

// isAvailable tells whether a store was configured at all.
func (state settingsState) isAvailable() bool {
	return state.store != nil
}

// isEditable tells whether the form can be changed: it needs a store and a
// config that loaded.
func (state settingsState) isEditable() bool {
	return state.isAvailable() && state.loaded && !state.loadFailed
}

// isDirty tells whether the form differs from what was loaded or saved.
func (state settingsState) isDirty() bool {
	return state.form != state.baseline
}

// isTokenReadOnly and isChatReadOnly tell whether the environment owns the value.
func (state settingsState) isTokenReadOnly() bool {
	return state.meta.TokenSource == SettingSourceEnvironment
}

func (state settingsState) isChatReadOnly() bool {
	return state.meta.ChatSource == SettingSourceEnvironment
}

// isReadOnly tells whether the row cannot be edited because of its source.
func (state settingsState) isReadOnly(field FormField) bool {
	switch field {
	case FieldToken:
		return state.isTokenReadOnly()
	case FieldChatID:
		return state.isChatReadOnly()
	default:
		return false
	}
}
