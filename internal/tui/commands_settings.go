package tui

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// settingsLoadCommand reads the config file off the UI loop.
func settingsLoadCommand(store SettingsStore) tea.Cmd {
	return func() tea.Msg {
		settings, err := store.Load()
		return SettingsLoadedMessage{Settings: settings, Err: err}
	}
}

// settingsRefreshCommand re-reads the file after a save to refresh the metadata.
func settingsRefreshCommand(store SettingsStore) tea.Cmd {
	return func() tea.Msg {
		settings, err := store.Load()
		return SettingsLoadedMessage{Settings: settings, Err: err, Refresh: true}
	}
}

// settingsSaveCommand writes the config file off the UI loop.
func settingsSaveCommand(store SettingsStore, settings AlertSettings) tea.Cmd {
	return func() tea.Msg {
		return SettingsSavedMessage{Err: store.Save(settings)}
	}
}

// alertTestCommand sends the test message off the UI loop. The tester owns the
// network timeout; the context only guards against a tester that never returns.
func alertTestCommand(tester AlertTester, settings AlertSettings, clock func() time.Time) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), alertTestGuardTimeout)
		defer cancel()
		err := tester.SendTest(ctx, settings)
		return AlertTestResultMessage{Err: err, At: clock()}
	}
}
