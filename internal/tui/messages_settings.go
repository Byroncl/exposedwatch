package tui

import "time"

// SettingsLoadedMessage carries the outcome of reading the config file.
// Refresh is set for the re-read after a save, which only updates the
// metadata (path, sources, warnings) and never the values being edited.
type SettingsLoadedMessage struct {
	Settings AlertSettings
	Err      error
	Refresh  bool
}

// SettingsSavedMessage carries the outcome of saving the config file.
type SettingsSavedMessage struct {
	Err error
}

// AlertTestResultMessage carries the outcome of one test message. It is
// separate from the scan messages so a slow network never holds back a scan.
type AlertTestResultMessage struct {
	Err error
	At  time.Time
}
