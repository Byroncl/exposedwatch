package configfile

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

// SecretSource tells where a credential came from, for a future settings UI.
type SecretSource string

// Known secret sources.
const (
	SourceNone        SecretSource = "none"
	SourceFile        SecretSource = "file"
	SourceEnvironment SecretSource = "environment"
)

// Loaded is the outcome of Load.
type Loaded struct {
	Path string
	// Exists is false when no file was found and defaults are in use.
	Exists bool
	// File is the content as stored on disk. Save this, never Effective, so
	// secrets that came from the environment are not written to disk.
	File Config
	// Effective is File with the environment overrides applied.
	Effective   Config
	TokenSource SecretSource
	ChatSource  SecretSource
	// Warnings are non-fatal findings, such as loose file permissions.
	Warnings []string
}

// Load reads the config file (a missing file means defaults), validates it
// and applies the EXPOSEDWATCH_TELEGRAM_* environment overrides.
func Load(path string, getenv func(name string) string) (Loaded, error) {
	return LoadWithTrust(path, getenv, SystemTrust())
}

// LoadWithTrust is Load with an injectable trust policy. A file that is a
// symlink or has a foreign owner is refused; so is a file whose directory
// others can write to, when running as root (other users only get a warning).
func LoadWithTrust(path string, getenv func(name string) string, trust Trust) (Loaded, error) {
	loaded := Loaded{Path: path, File: Defaults(), TokenSource: SourceNone, ChatSource: SourceNone}
	content, info, trustWarnings, err := readTrusted(path, trust)
	loaded.Warnings = append(loaded.Warnings, trustWarnings...)
	switch {
	case errors.Is(err, os.ErrNotExist):
		// No file: defaults.
	case err != nil:
		return Loaded{}, err
	default:
		loaded.Exists = true
		if err := json.Unmarshal(content, &loaded.File); err != nil {
			// The decoder error can quote a fragment of the input; drop it.
			return Loaded{}, ErrMalformedConfig
		}
		if permissions := info.Mode().Perm(); permissions&groupAndOtherBits != 0 {
			loaded.Warnings = append(loaded.Warnings, fmt.Sprintf(looseModeFormat, path, permissions))
		}
	}
	if err := loaded.File.Validate(); err != nil {
		return Loaded{}, err
	}
	loaded.applyEnvironment(getenv)
	return loaded, nil
}

// readTrusted checks the file with Lstat, opens it, checks that the opened
// descriptor is the same non-symlink file with an acceptable owner, and reads
// through that descriptor, so the file cannot be swapped after the checks.
func readTrusted(path string, trust Trust) ([]byte, os.FileInfo, []string, error) {
	linkInfo, err := trust.Lstat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil, nil, err
		}
		return nil, nil, nil, fmt.Errorf(errorDetailFormat, ErrUnreadableConfig, err)
	}
	if linkInfo.Mode()&os.ModeSymlink != 0 {
		return nil, nil, nil, untrusted(symlinkMessage)
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, nil, fmt.Errorf(errorDetailFormat, ErrUnreadableConfig, err)
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return nil, nil, nil, fmt.Errorf(errorDetailFormat, ErrUnreadableConfig, err)
	}
	if err := trust.checkFile(linkInfo, info); err != nil {
		return nil, nil, nil, err
	}
	var warnings []string
	warning, err := trust.checkParent(path)
	if err != nil {
		return nil, nil, nil, err
	}
	if warning != "" {
		warnings = append(warnings, warning)
	}
	content, err := io.ReadAll(io.LimitReader(file, MaxFileBytes+1))
	if err != nil {
		return nil, nil, nil, fmt.Errorf(errorDetailFormat, ErrUnreadableConfig, err)
	}
	if len(content) > MaxFileBytes {
		return nil, nil, nil, fmt.Errorf(errorDetailFormat, ErrMalformedConfig, oversizedMessage)
	}
	return content, info, warnings, nil
}

func (loaded *Loaded) applyEnvironment(getenv func(name string) string) {
	loaded.Effective = loaded.File
	if loaded.File.Telegram.BotToken != "" {
		loaded.TokenSource = SourceFile
	}
	if loaded.File.Telegram.ChatID != "" {
		loaded.ChatSource = SourceFile
	}
	if getenv == nil {
		return
	}
	if token := strings.TrimSpace(getenv(EnvTelegramToken)); token != "" {
		loaded.Effective.Telegram.BotToken = domain.BotToken(token)
		loaded.TokenSource = SourceEnvironment
	}
	if chat := strings.TrimSpace(getenv(EnvTelegramChatID)); chat != "" {
		loaded.Effective.Telegram.ChatID = domain.ChatID(chat)
		loaded.ChatSource = SourceEnvironment
	}
}
