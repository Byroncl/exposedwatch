package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Byroncl/exposedwatch/internal/adapters/configfile"
	"github.com/Byroncl/exposedwatch/internal/domain"
	"github.com/Byroncl/exposedwatch/internal/tui"
)

// configSettingsStore implements tui.SettingsStore over the config file. All
// validation, trust checks and permissions come from the configfile package.
type configSettingsStore struct {
	runtime    commandRuntime
	configPath string
	// statPath inspects the config directory; os.Stat unless a test replaces it.
	statPath func(path string) (os.FileInfo, error)
}

// newConfigSettingsStore reads and writes the file at configPath, or at the
// default path when it is empty.
func newConfigSettingsStore(runtime commandRuntime, configPath string) tui.SettingsStore {
	return configSettingsStore{runtime: runtime, configPath: configPath, statPath: os.Stat}
}

// Load returns the effective settings with where each credential came from.
func (store configSettingsStore) Load() (tui.AlertSettings, error) {
	loaded, err := store.runtime.loadConfiguration(store.configPath)
	if err != nil {
		return tui.AlertSettings{}, err
	}
	effective := loaded.Effective
	return tui.AlertSettings{
		Enabled:              effective.Telegram.Enabled,
		Token:                effective.Telegram.BotToken,
		ChatID:               string(effective.Telegram.ChatID),
		MinSeverity:          effective.Alerts.MinSeverity,
		CooldownMinutes:      effective.Alerts.CooldownMinutes,
		IntervalSeconds:      effective.Alerts.IntervalSeconds,
		TokenSource:          tui.SettingSource(loaded.TokenSource),
		ChatSource:           tui.SettingSource(loaded.ChatSource),
		TokenEnvironmentName: configfile.EnvTelegramToken,
		ChatEnvironmentName:  configfile.EnvTelegramChatID,
		Path:                 loaded.Path,
		Warnings:             loaded.Warnings,
	}, nil
}

// Save writes the form values. The file is read again first: a file that
// cannot be loaded is never overwritten, and a credential that the
// environment provides keeps the value the file already has, so environment
// secrets never reach the disk. The caller's source fields are not trusted.
func (store configSettingsStore) Save(settings tui.AlertSettings) error {
	loaded, err := store.runtime.loadConfiguration(store.configPath)
	if err != nil {
		return sanitizedError(fmt.Errorf(saveRefusedFormat, err), settings.Token)
	}
	if err := store.checkDirectory(loaded.Path); err != nil {
		return err
	}
	file := loaded.File
	file.Telegram.Enabled = settings.Enabled
	if loaded.TokenSource != configfile.SourceEnvironment {
		file.Telegram.BotToken = settings.Token
	}
	if loaded.ChatSource != configfile.SourceEnvironment {
		file.Telegram.ChatID = domain.ChatID(settings.ChatID)
	}
	file.Alerts.MinSeverity = settings.MinSeverity
	file.Alerts.CooldownMinutes = settings.CooldownMinutes
	file.Alerts.IntervalSeconds = settings.IntervalSeconds
	if err := configfile.Save(loaded.Path, file); err != nil {
		return sanitizedError(err, settings.Token)
	}
	return nil
}

// sanitizedError drops the token from an error message, as defence in depth:
// the config package never puts it there.
func sanitizedError(err error, token domain.BotToken) error {
	return errors.New(maskSecret(err.Error(), token))
}

// checkDirectory refuses to create or replace the config file inside a
// directory that group or others can write to when running as root. The
// config package only checks the directory of files that already exist, so
// this closes the gap for a first save. A normal user may save; the re-read
// after saving reports the directory as a warning.
func (store configSettingsStore) checkDirectory(configPath string) error {
	if store.statPath == nil || store.runtime.Environment.UserID != rootUserID {
		return nil
	}
	info, err := store.statPath(filepath.Dir(configPath))
	if err != nil || !info.IsDir() || info.Mode().Perm()&directoryWriteMask == 0 {
		return nil
	}
	return errors.New(directoryWritableMessage)
}
