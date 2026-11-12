package tui

import (
	"context"
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

const (
	testSecretToken      = "123456789:AA-fake_TOKEN+xyz/=="
	testSecretPart       = "AA-fake_TOKEN+xyz/=="
	testMaskedToken      = "123456789:****"
	testChatIdentifier   = "-1001234567890"
	testConfigPath       = "/home/user/.config/exposedwatch/config.json"
	testTokenEnvName     = "EXPOSEDWATCH_TELEGRAM_TOKEN"
	testChatEnvName      = "EXPOSEDWATCH_TELEGRAM_CHAT_ID"
	testFragmentLength   = 6
	testCommandDepthMax  = 8
	testDigitsOnlyInput  = "4a2b"
	testUnicodeChatInput = "日本"
)

// fakeSettingsStore never touches the disk.
type fakeSettingsStore struct {
	settings  AlertSettings
	loadErr   error
	saveErr   error
	saved     []AlertSettings
	loadCalls int
}

func (store *fakeSettingsStore) Load() (AlertSettings, error) {
	store.loadCalls++
	return store.settings, store.loadErr
}

func (store *fakeSettingsStore) Save(settings AlertSettings) error {
	store.saved = append(store.saved, settings)
	return store.saveErr
}

// fakeAlertTester never contacts a network.
type fakeAlertTester struct {
	calls []AlertSettings
	err   error
}

func (tester *fakeAlertTester) SendTest(_ context.Context, settings AlertSettings) error {
	tester.calls = append(tester.calls, settings)
	return tester.err
}

func configuredSettings() AlertSettings {
	return AlertSettings{
		Enabled:              true,
		Token:                domain.BotToken(testSecretToken),
		ChatID:               testChatIdentifier,
		MinSeverity:          "high",
		CooldownMinutes:      60,
		IntervalSeconds:      60,
		TokenSource:          SettingSourceFile,
		ChatSource:           SettingSourceFile,
		Path:                 testConfigPath,
		TokenEnvironmentName: testTokenEnvName,
		ChatEnvironmentName:  testChatEnvName,
	}
}

func unconfiguredSettings() AlertSettings {
	settings := configuredSettings()
	settings.Enabled = false
	settings.Token = ""
	settings.ChatID = ""
	settings.TokenSource = SettingSourceNone
	settings.ChatSource = SettingSourceNone
	return settings
}

// settingsModel builds a sized model with the fake store, already loaded.
func settingsModel(testRun *testing.T, store *fakeSettingsStore, tester *fakeAlertTester) Model {
	testRun.Helper()
	options := Options{Scanner: fakeScanner{}, Interval: testInterval, Version: testVersion, Settings: store}
	if tester != nil {
		options.AlertTester = tester
	}
	model := New(options)
	model = applyMessage(testRun, model, tea.WindowSizeMsg{Width: testWidth, Height: testHeight})
	model = applyMessage(testRun, model, RefreshResultMessage{Report: sampleReport(), At: testRefreshTime})
	return applyMessage(testRun, model, settingsLoadCommand(store)())
}

// telegramModel opens Settings > Telegram of a loaded model.
func telegramModel(testRun *testing.T, store *fakeSettingsStore, tester *fakeAlertTester) Model {
	testRun.Helper()
	return showView(testRun, settingsModel(testRun, store, tester), moduleView{ModuleSettings, 0})
}

// applyWithCommand applies one message and also returns the command it asked for.
func applyWithCommand(testRun *testing.T, model Model, incomingMessage tea.Msg) (Model, tea.Cmd) {
	testRun.Helper()
	updated, command := model.Update(incomingMessage)
	castModel, isModel := updated.(Model)
	if !isModel {
		testRun.Fatalf("Update returned %T, want Model", updated)
	}
	return castModel, command
}

// settle runs a command the way the runtime would and feeds its message back,
// following follow-up commands. Timer commands are not run.
func settle(testRun *testing.T, model Model, command tea.Cmd) Model {
	testRun.Helper()
	for depth := 0; command != nil && depth < testCommandDepthMax; depth++ {
		outcome := command()
		var next tea.Cmd
		model, next = applyWithCommand(testRun, model, outcome)
		command = next
	}
	return model
}

// press applies the keys, running every command they produce.
func press(testRun *testing.T, model Model, keys ...string) Model {
	testRun.Helper()
	for _, key := range keys {
		var command tea.Cmd
		model, command = applyWithCommand(testRun, model, keyMessage(key))
		model = settle(testRun, model, command)
	}
	return model
}

// pressWithoutRunning applies the keys and returns the last command unrun.
func pressWithoutRunning(testRun *testing.T, model Model, keys ...string) (Model, tea.Cmd) {
	testRun.Helper()
	var command tea.Cmd
	for _, key := range keys {
		model, command = applyWithCommand(testRun, model, keyMessage(key))
	}
	return model, command
}

// moveToField moves the form cursor to the field with j and k presses.
func moveToField(testRun *testing.T, model Model, field FormField) Model {
	testRun.Helper()
	for attempt := 0; attempt < int(formFieldCount) && model.settings.cursor != field; attempt++ {
		if model.settings.cursor < field {
			model = press(testRun, model, "j")
		} else {
			model = press(testRun, model, "k")
		}
	}
	if model.settings.cursor != field {
		testRun.Fatalf("cannot reach field %v", field)
	}
	return model
}

// containsSecretFragment reports whether text holds the token, or any
// fragment of its secret part that is at least testFragmentLength long.
func containsSecretFragment(text string) (string, bool) {
	// The variable names are public labels, and one of them contains "_TOKEN".
	text = strings.ReplaceAll(strings.ReplaceAll(text, testTokenEnvName, ""), testChatEnvName, "")
	if strings.Contains(text, testSecretToken) {
		return testSecretToken, true
	}
	runes := []rune(testSecretPart)
	for start := 0; start+testFragmentLength <= len(runes); start++ {
		fragment := string(runes[start : start+testFragmentLength])
		if strings.Contains(text, fragment) {
			return fragment, true
		}
	}
	return "", false
}

var errFakeSaveFailure = errors.New("config could not be saved: the disk is full")

func errFakeUntrusted() error {
	return errors.New("config file is not trusted: it is a symbolic link or changed while it was opened")
}

func domainToken(text string) domain.BotToken {
	return domain.BotToken(text)
}

// cursorLabelForTest is the text of the row under the form cursor.
func (state settingsState) cursorLabelForTest() string {
	return state.cursor.label()
}
