package tui

import (
	"context"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

// SettingSource tells where a setting came from. The Telegram credentials may
// come from the environment, in which case the form shows them read-only and
// never writes them.
type SettingSource string

// Known setting sources.
const (
	SettingSourceNone        SettingSource = "none"
	SettingSourceFile        SettingSource = "file"
	SettingSourceEnvironment SettingSource = "environment"
)

// AlertSettings is what the TUI knows about the Telegram alert configuration.
// Token is a domain.BotToken, so printing it through fmt never reveals it.
type AlertSettings struct {
	Enabled         bool
	Token           domain.BotToken
	ChatID          string
	MinSeverity     string
	CooldownMinutes int
	IntervalSeconds int
	TokenSource     SettingSource
	ChatSource      SettingSource
	// TokenEnvironmentName and ChatEnvironmentName name the environment
	// variables that override the file, for the read-only hint.
	TokenEnvironmentName string
	ChatEnvironmentName  string
	// Path is the config file the settings were read from and are saved to.
	Path string
	// Warnings are non-fatal findings about the file, such as loose permissions.
	Warnings []string
}

// SettingsStore loads and saves the alert configuration. Save validates with
// the rules of the config file and never persists values that come from the
// environment. Its errors must not contain secrets.
type SettingsStore interface {
	Load() (AlertSettings, error)
	Save(settings AlertSettings) error
}

// AlertTester sends one test message with the given (possibly unsaved)
// settings. Its errors must not contain secrets.
type AlertTester interface {
	SendTest(ctx context.Context, settings AlertSettings) error
}
