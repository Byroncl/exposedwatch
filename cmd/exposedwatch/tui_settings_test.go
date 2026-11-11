package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Byroncl/exposedwatch/internal/adapters/configfile"
	"github.com/Byroncl/exposedwatch/internal/domain"
	"github.com/Byroncl/exposedwatch/internal/tui"
)

func newTestStore(fixture *watchFixture, path string) tui.SettingsStore {
	return newConfigSettingsStore(fixture.runtime, path)
}

func TestSettingsStoreLoadMapsTheEffectiveConfigAndSources(testRun *testing.T) {
	fixture := newWatchFixture()
	fixture.environ[configfile.EnvTelegramChatID] = "42"
	path := writeWatchConfig(testRun, enabledConfig)
	settings, err := newTestStore(fixture, path).Load()
	if err != nil {
		testRun.Fatal(err)
	}
	if !settings.Enabled || settings.Token.Reveal() != testSecretToken || settings.ChatID != "42" || settings.MinSeverity != "high" ||
		settings.CooldownMinutes != 60 || settings.IntervalSeconds != 60 || settings.Path != path {
		testRun.Fatalf("settings = %+v", settings)
	}
	if settings.TokenSource != tui.SettingSourceFile || settings.ChatSource != tui.SettingSourceEnvironment {
		testRun.Fatalf("sources = %v / %v", settings.TokenSource, settings.ChatSource)
	}
	if settings.TokenEnvironmentName != configfile.EnvTelegramToken || settings.ChatEnvironmentName != configfile.EnvTelegramChatID {
		testRun.Fatalf("environment names = %q / %q", settings.TokenEnvironmentName, settings.ChatEnvironmentName)
	}
}

func TestSettingsStoreLoadOfAMissingFileGivesDefaultsAtThatPath(testRun *testing.T) {
	fixture := newWatchFixture()
	path := filepath.Join(testRun.TempDir(), "absent", "config.json")
	settings, err := newTestStore(fixture, path).Load()
	if err != nil || settings.Path != path || settings.Enabled || !settings.Token.IsEmpty() || settings.TokenSource != tui.SettingSourceNone {
		testRun.Fatalf("settings=%+v err=%v", settings, err)
	}
}

func TestSettingsStoreLoadRefusesAMalformedFileWithoutQuotingIt(testRun *testing.T) {
	fixture := newWatchFixture()
	path := writeWatchConfig(testRun, `{"telegram": {"bot_token": "`+testSecretToken+`" oops`)
	_, err := newTestStore(fixture, path).Load()
	if err == nil || strings.Contains(err.Error(), "AA-fake") {
		testRun.Fatalf("err = %v", err)
	}
}

func TestSettingsStoreLoadUsesTheDefaultPathWhenNoneIsGiven(testRun *testing.T) {
	fixture := newWatchFixture()
	fixture.runtime.Environment.HomeDirectory = func() (string, error) { return testRun.TempDir(), nil }
	settings, err := newTestStore(fixture, "").Load()
	if err != nil || !strings.HasSuffix(settings.Path, filepath.Join(".config", "exposedwatch", "config.json")) {
		testRun.Fatalf("settings=%+v err=%v", settings, err)
	}
}

func readStoredConfig(testRun *testing.T, path string) configfile.Config {
	testRun.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		testRun.Fatal(err)
	}
	var stored configfile.Config
	if err := json.Unmarshal(content, &stored); err != nil {
		testRun.Fatal(err)
	}
	return stored
}

func TestSettingsStoreSaveWritesTheFormValuesWithPrivatePermissions(testRun *testing.T) {
	fixture := newWatchFixture()
	path := filepath.Join(testRun.TempDir(), "nested", "config.json")
	store := newTestStore(fixture, path)
	settings := tui.AlertSettings{
		Enabled: true, Token: domain.BotToken(testSecretToken), ChatID: testChatID,
		MinSeverity: "medium", CooldownMinutes: 5, IntervalSeconds: 30,
	}
	if err := store.Save(settings); err != nil {
		testRun.Fatal(err)
	}
	stored := readStoredConfig(testRun, path)
	if !stored.Telegram.Enabled || stored.Telegram.BotToken.Reveal() != testSecretToken || stored.Telegram.ChatID != testChatID ||
		stored.Alerts.MinSeverity != "medium" || stored.Alerts.CooldownMinutes != 5 || stored.Alerts.IntervalSeconds != 30 {
		testRun.Fatalf("stored = %+v", stored)
	}
	info, err := os.Stat(path)
	if err != nil || info.Mode().Perm() != 0o600 {
		testRun.Fatalf("mode = %v err = %v", info.Mode().Perm(), err)
	}
}

func TestSettingsStoreSaveNeverWritesEnvironmentSecrets(testRun *testing.T) {
	fixture := newWatchFixture()
	fixture.environ[configfile.EnvTelegramToken] = testSecretToken
	fixture.environ[configfile.EnvTelegramChatID] = "777"
	path := writeWatchConfig(testRun, `{"version":1,"telegram":{"enabled":false,"bot_token":"file-token","chat_id":"file-chat"}}`)
	store := newTestStore(fixture, path)
	settings, err := store.Load()
	if err != nil {
		testRun.Fatal(err)
	}
	settings.Enabled = true
	settings.MinSeverity = "low"
	if err := store.Save(settings); err != nil {
		testRun.Fatal(err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		testRun.Fatal(err)
	}
	if strings.Contains(string(content), "AA-fake") || strings.Contains(string(content), "777") {
		testRun.Fatalf("environment secrets reached the disk: %s", content)
	}
	stored := readStoredConfig(testRun, path)
	if stored.Telegram.BotToken.Reveal() != "file-token" || stored.Telegram.ChatID != "file-chat" || !stored.Telegram.Enabled || stored.Alerts.MinSeverity != "low" {
		testRun.Fatalf("stored = %+v", stored)
	}
}

func TestSettingsStoreSaveIgnoresAForgedSourceAndStillProtectsTheEnvironment(testRun *testing.T) {
	fixture := newWatchFixture()
	fixture.environ[configfile.EnvTelegramToken] = testSecretToken
	path := writeWatchConfig(testRun, enabledNoSecret)
	forged := tui.AlertSettings{Enabled: true, Token: domain.BotToken(testSecretToken), TokenSource: tui.SettingSourceFile,
		MinSeverity: "high", CooldownMinutes: 60, IntervalSeconds: 60}
	if err := newTestStore(fixture, path).Save(forged); err != nil {
		testRun.Fatal(err)
	}
	if content, _ := os.ReadFile(path); strings.Contains(string(content), "AA-fake") {
		testRun.Fatalf("the store must decide from the live environment, not from the caller: %s", content)
	}
}

func TestSettingsStoreSaveReusesTheConfigValidationAndNeverLeaksTheToken(testRun *testing.T) {
	fixture := newWatchFixture()
	path := writeWatchConfig(testRun, enabledConfig)
	store := newTestStore(fixture, path)
	cases := map[string]tui.AlertSettings{
		"severity": {Token: domain.BotToken(testSecretToken), MinSeverity: "extreme", CooldownMinutes: 60, IntervalSeconds: 60},
		"cooldown": {Token: domain.BotToken(testSecretToken), MinSeverity: "high", CooldownMinutes: configfile.MaxCooldownMinutes + 1, IntervalSeconds: 60},
		"interval": {Token: domain.BotToken(testSecretToken), MinSeverity: "high", CooldownMinutes: 60, IntervalSeconds: configfile.MinIntervalSeconds - 1},
	}
	before, _ := os.ReadFile(path)
	for name, settings := range cases {
		err := store.Save(settings)
		if err == nil || strings.Contains(err.Error(), "AA-fake") {
			testRun.Errorf("%s: err = %v", name, err)
		}
	}
	if after, _ := os.ReadFile(path); !bytes.Equal(before, after) {
		testRun.Fatal("a rejected save must leave the file untouched")
	}
}

func TestSettingsStoreSaveRefusesAFileThatCannotBeLoaded(testRun *testing.T) {
	fixture := newWatchFixture()
	path := writeWatchConfig(testRun, "not json")
	if err := newTestStore(fixture, path).Save(tui.AlertSettings{MinSeverity: "high", CooldownMinutes: 60, IntervalSeconds: 60}); err == nil {
		testRun.Fatal("saving over a file that cannot be read would destroy it")
	}
}

func TestAlertTesterSendsTheTestAlertWithTheFormValues(testRun *testing.T) {
	fixture := newWatchFixture()
	tester := newTelegramAlertTester(fixture.runtime)
	err := tester.SendTest(context.Background(), tui.AlertSettings{Token: domain.BotToken(testSecretToken), ChatID: testChatID})
	if err != nil || len(fixture.notifier.sent) != 1 {
		testRun.Fatalf("err=%v sent=%d", err, len(fixture.notifier.sent))
	}
	sent := fixture.notifier.sent[0]
	if sent.Kind != domain.AlertKindTest || sent.Severity != domain.Info || !strings.Contains(sent.Title, "web-01") {
		testRun.Fatalf("alert = %+v", sent)
	}
	if len(fixture.notifierID) != 1 || fixture.notifierID[0] != domain.ChatID(testChatID) {
		testRun.Fatalf("chat = %v", fixture.notifierID)
	}
}

func TestAlertTesterRequiresCredentialsAndNeverSendsWithoutThem(testRun *testing.T) {
	fixture := newWatchFixture()
	tester := newTelegramAlertTester(fixture.runtime)
	for _, settings := range []tui.AlertSettings{{}, {Token: domain.BotToken(testSecretToken)}, {ChatID: testChatID}} {
		if err := tester.SendTest(context.Background(), settings); err == nil {
			testRun.Errorf("%+v must be refused", settings)
		}
	}
	if len(fixture.notifier.sent) != 0 || len(fixture.notifierID) != 0 {
		testRun.Fatal("nothing may be built or sent without credentials")
	}
}

func TestAlertTesterMasksTheTokenInFailures(testRun *testing.T) {
	fixture := newWatchFixture()
	fixture.notifier.failuresLeft = 1
	err := newTelegramAlertTester(fixture.runtime).SendTest(context.Background(), tui.AlertSettings{Token: domain.BotToken(testSecretToken), ChatID: testChatID})
	if err == nil || strings.Contains(err.Error(), testSecretToken) {
		testRun.Fatalf("err = %v", err)
	}
}

func TestAlertTesterHasANamedTimeout(testRun *testing.T) {
	if alertTestTimeout.Seconds() != 15 {
		testRun.Fatalf("timeout = %v", alertTestTimeout)
	}
}

func TestParseTuiOptionsAcceptsTheConfigFlag(testRun *testing.T) {
	options, err := parseTuiOptions([]string{"--config", "/tmp/custom.json"}, &strings.Builder{})
	if err != nil || options.ConfigPath != "/tmp/custom.json" {
		testRun.Fatalf("options=%+v err=%v", options, err)
	}
	defaults, err := parseTuiOptions(nil, &strings.Builder{})
	if err != nil || defaults.ConfigPath != "" {
		testRun.Fatalf("defaults=%+v err=%v", defaults, err)
	}
}

func TestBuildTuiOptionsWiresTheSettingsStoreAndTesterWithTheConfigPath(testRun *testing.T) {
	fixture := newWatchFixture()
	path := writeWatchConfig(testRun, enabledConfig)
	built := buildTuiOptions(TuiOptions{ConfigPath: path, Interval: tui.DefaultInterval}, fixture.runtime)
	if built.Settings == nil || built.AlertTester == nil {
		testRun.Fatal("the tui must receive a settings store and an alert tester")
	}
	settings, err := built.Settings.Load()
	if err != nil || settings.Path != path {
		testRun.Fatalf("the store must read the --config path: %+v %v", settings, err)
	}
}

func TestRunTuiLaunchesWithTheWiredOptionsOnATerminal(testRun *testing.T) {
	fixture := newWatchFixture()
	path := writeWatchConfig(testRun, enabledConfig)
	var launched []tui.Options
	var stdout, stderr bytes.Buffer
	code := runTuiWith(fixture.runtime, func(options tui.Options) error { launched = append(launched, options); return nil },
		func(io.Writer) bool { return true }, []string{"--config", path, "--no-docker"}, &stdout, &stderr)
	if code != ExitOK || len(launched) != 1 || launched[0].Settings == nil || launched[0].AlertTester == nil {
		testRun.Fatalf("code=%d launched=%d stderr=%q", code, len(launched), stderr.String())
	}
}

func TestRunTuiDoesNotLaunchOutsideATerminalOrWithBadFlags(testRun *testing.T) {
	fixture := newWatchFixture()
	launch := func(tui.Options) error {
		testRun.Error("must not launch")
		return nil
	}
	var stdout, stderr bytes.Buffer
	if code := runTuiWith(fixture.runtime, launch, func(io.Writer) bool { return false }, nil, &stdout, &stderr); code != ExitUsage {
		testRun.Fatalf("non-terminal code = %d", code)
	}
	if code := runTuiWith(fixture.runtime, launch, func(io.Writer) bool { return true }, []string{"--bogus"}, &stdout, &stderr); code != ExitUsage {
		testRun.Fatalf("bad flag code = %d", code)
	}
}

func TestUsageMentionsTheTuiConfigFlag(testRun *testing.T) {
	_, stdout, _ := runCommand("help")
	if !strings.Contains(stdout, "--config") {
		testRun.Fatalf("usage lacks --config for the tui: %q", stdout)
	}
}

// writableDirectoryInfo reports a group- and other-writable directory.
type writableDirectoryInfo struct{ os.FileInfo }

func (info writableDirectoryInfo) Mode() os.FileMode { return os.ModeDir | 0o777 }
func (info writableDirectoryInfo) IsDir() bool       { return true }

func storeWithWritableParent(fixture *watchFixture, path string, userID int) configSettingsStore {
	fixture.runtime.Environment.UserID = userID
	return configSettingsStore{runtime: fixture.runtime, configPath: path, statPath: func(string) (os.FileInfo, error) {
		return writableDirectoryInfo{}, nil
	}}
}

func validSettings() tui.AlertSettings {
	return tui.AlertSettings{MinSeverity: "high", CooldownMinutes: 60, IntervalSeconds: 60, Token: domain.BotToken(testSecretToken), ChatID: testChatID}
}

func TestSettingsStoreSaveRefusesAWritableDirectoryAsRoot(testRun *testing.T) {
	fixture := newWatchFixture()
	path := filepath.Join(testRun.TempDir(), "config.json")
	err := storeWithWritableParent(fixture, path, 0).Save(validSettings())
	if err == nil || !strings.Contains(err.Error(), "the config directory is writable by group or others; fix its permissions first") {
		testRun.Fatalf("err = %v", err)
	}
	if _, statErr := os.Stat(path); statErr == nil {
		testRun.Fatal("nothing may be written")
	}
}

func TestSettingsStoreSaveAllowsAWritableDirectoryForANormalUserAndLoadWarns(testRun *testing.T) {
	fixture := newWatchFixture()
	path := filepath.Join(testRun.TempDir(), "config.json")
	if err := storeWithWritableParent(fixture, path, 1000).Save(validSettings()); err != nil {
		testRun.Fatal(err)
	}
	if err := os.Chmod(filepath.Dir(path), 0o777); err != nil {
		testRun.Fatal(err)
	}
	settings, err := newTestStore(fixture, path).Load()
	if err != nil || len(settings.Warnings) == 0 || !strings.Contains(strings.Join(settings.Warnings, " "), "can be written by group or others") {
		testRun.Fatalf("settings=%+v err=%v", settings, err)
	}
}
