package configfile

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

const (
	fileToken   = "111111111:FILE-token_secret"
	envToken    = "222222222:ENV-token_secret"
	fileChatID  = "-100111"
	envChatID   = "-100222"
	emptyEnvVar = ""
)

func noEnvironment(string) string { return emptyEnvVar }

func environmentOf(values map[string]string) func(string) string {
	return func(name string) string { return values[name] }
}

func writeConfigFile(testRun *testing.T, directory string, content string, mode os.FileMode) string {
	testRun.Helper()
	path := filepath.Join(directory, "config.json")
	// The sandbox may create temp directories with a permissive umask.
	if err := os.Chmod(directory, 0o700); err != nil {
		testRun.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), mode); err != nil {
		testRun.Fatal(err)
	}
	if err := os.Chmod(path, mode); err != nil {
		testRun.Fatal(err)
	}
	return path
}

func TestLoadMissingFileReturnsDefaults(testRun *testing.T) {
	loaded, err := Load(filepath.Join(testRun.TempDir(), "absent.json"), noEnvironment)
	if err != nil {
		testRun.Fatal(err)
	}
	if loaded.Effective != Defaults() || loaded.TokenSource != SourceNone || loaded.ChatSource != SourceNone || loaded.Exists {
		testRun.Fatalf("%+v", loaded)
	}
	if Defaults().Alerts.MinSeverity != "high" || Defaults().Alerts.CooldownMinutes != 60 || Defaults().Alerts.IntervalSeconds != 60 {
		testRun.Fatalf("defaults %+v", Defaults())
	}
}

func TestLoadKeepsDefaultsForMissingKeys(testRun *testing.T) {
	path := writeConfigFile(testRun, testRun.TempDir(), `{"version":1,"telegram":{"enabled":true,"bot_token":"`+fileToken+`","chat_id":"`+fileChatID+`"}}`, 0o600)
	loaded, err := Load(path, noEnvironment)
	if err != nil {
		testRun.Fatal(err)
	}
	if loaded.Effective.Alerts != Defaults().Alerts || !loaded.Effective.Telegram.Enabled {
		testRun.Fatalf("%+v", loaded.Effective)
	}
	if loaded.TokenSource != SourceFile || loaded.ChatSource != SourceFile || len(loaded.Warnings) != 0 {
		testRun.Fatalf("%+v", loaded)
	}
}

func TestEnvironmentOverridesFileAndTracksSource(testRun *testing.T) {
	testCases := []struct {
		name       string
		environ    map[string]string
		wantToken  string
		wantChat   string
		wantSource SecretSource
		wantChatSr SecretSource
	}{
		{"nothing in env", nil, fileToken, fileChatID, SourceFile, SourceFile},
		{"token only", map[string]string{EnvTelegramToken: envToken}, envToken, fileChatID, SourceEnvironment, SourceFile},
		{"chat only", map[string]string{EnvTelegramChatID: envChatID}, fileToken, envChatID, SourceFile, SourceEnvironment},
		{"both", map[string]string{EnvTelegramToken: " " + envToken + "\n", EnvTelegramChatID: envChatID}, envToken, envChatID, SourceEnvironment, SourceEnvironment},
	}
	for _, testCase := range testCases {
		testRun.Run(testCase.name, func(testRun *testing.T) {
			path := writeConfigFile(testRun, testRun.TempDir(),
				`{"version":1,"telegram":{"enabled":true,"bot_token":"`+fileToken+`","chat_id":"`+fileChatID+`"}}`, 0o600)
			loaded, err := Load(path, environmentOf(testCase.environ))
			if err != nil {
				testRun.Fatal(err)
			}
			if loaded.Effective.Telegram.BotToken.Reveal() != testCase.wantToken || string(loaded.Effective.Telegram.ChatID) != testCase.wantChat {
				testRun.Fatalf("effective credentials differ")
			}
			if loaded.TokenSource != testCase.wantSource || loaded.ChatSource != testCase.wantChatSr {
				testRun.Fatalf("sources %v %v", loaded.TokenSource, loaded.ChatSource)
			}
			if loaded.File.Telegram.BotToken.Reveal() != fileToken {
				testRun.Fatal("File must keep the on-disk value so Save never persists env secrets")
			}
		})
	}
}

func TestEnvironmentWorksWithoutAnyFile(testRun *testing.T) {
	loaded, err := Load(filepath.Join(testRun.TempDir(), "absent.json"),
		environmentOf(map[string]string{EnvTelegramToken: envToken, EnvTelegramChatID: envChatID}))
	if err != nil || loaded.TokenSource != SourceEnvironment || loaded.ChatSource != SourceEnvironment {
		testRun.Fatalf("%v %+v", err, loaded)
	}
}

func TestLoadWarnsOnLoosePermissions(testRun *testing.T) {
	path := writeConfigFile(testRun, testRun.TempDir(), `{"version":1}`, 0o644)
	loaded, err := Load(path, noEnvironment)
	if err != nil {
		testRun.Fatal(err)
	}
	if len(loaded.Warnings) != 1 || !strings.Contains(loaded.Warnings[0], "0644") {
		testRun.Fatalf("warnings %v", loaded.Warnings)
	}
	if err := Save(path, loaded.File); err != nil {
		testRun.Fatal(err)
	}
	assertMode(testRun, path, 0o600)
}

func TestMalformedJSONDoesNotLeakContents(testRun *testing.T) {
	secretLine := `{"telegram":{"bot_token":"` + fileToken + `" "oops"`
	path := writeConfigFile(testRun, testRun.TempDir(), secretLine, 0o600)
	_, err := Load(path, noEnvironment)
	if !errors.Is(err, ErrMalformedConfig) {
		testRun.Fatalf("got %v", err)
	}
	if strings.Contains(err.Error(), "FILE-token") || strings.Contains(err.Error(), "oops") {
		testRun.Fatalf("error leaks file contents: %v", err)
	}
}

func TestTypeMismatchDoesNotLeakContents(testRun *testing.T) {
	path := writeConfigFile(testRun, testRun.TempDir(), `{"telegram":{"bot_token":12345678,"chat_id":"x"}}`, 0o600)
	_, err := Load(path, noEnvironment)
	if !errors.Is(err, ErrMalformedConfig) || strings.Contains(err.Error(), "12345678") {
		testRun.Fatalf("got %v", err)
	}
}

func TestLoadRejectsInvalidValues(testRun *testing.T) {
	testCases := map[string]string{
		"unknown severity":   `{"alerts":{"min_severity":"urgent"}}`,
		"interval too small": `{"alerts":{"interval_seconds":5}}`,
		"interval too large": `{"alerts":{"interval_seconds":999999}}`,
		"negative cooldown":  `{"alerts":{"cooldown_minutes":-1}}`,
		"huge cooldown":      `{"alerts":{"cooldown_minutes":99999999}}`,
		"future version":     `{"version":2}`,
	}
	for name, content := range testCases {
		testRun.Run(name, func(testRun *testing.T) {
			path := writeConfigFile(testRun, testRun.TempDir(), content, 0o600)
			if _, err := Load(path, noEnvironment); !errors.Is(err, ErrInvalidConfig) {
				testRun.Fatalf("got %v", err)
			}
		})
	}
}

func TestLoadRejectsOversizedFile(testRun *testing.T) {
	path := writeConfigFile(testRun, testRun.TempDir(), `{"version":1,"padding":"`+strings.Repeat("x", MaxFileBytes)+`"}`, 0o600)
	if _, err := Load(path, noEnvironment); !errors.Is(err, ErrMalformedConfig) {
		testRun.Fatalf("got %v", err)
	}
}

func TestSaveCreatesPrivateDirectoryAndFile(testRun *testing.T) {
	path := filepath.Join(testRun.TempDir(), "nested", "exposedwatch", "config.json")
	config := Defaults()
	config.Telegram = TelegramSettings{Enabled: true, BotToken: domain.BotToken(fileToken), ChatID: domain.ChatID(fileChatID)}
	if err := Save(path, config); err != nil {
		testRun.Fatal(err)
	}
	assertMode(testRun, path, 0o600)
	assertMode(testRun, filepath.Dir(path), 0o700)
	loaded, err := Load(path, noEnvironment)
	if err != nil || loaded.File != config || len(loaded.Warnings) != 0 {
		testRun.Fatalf("%v %+v", err, loaded)
	}
}

func TestSaveRejectsInvalidConfigBeforeTouchingDisk(testRun *testing.T) {
	path := writeConfigFile(testRun, testRun.TempDir(), `{"version":1}`, 0o600)
	config := Defaults()
	config.Alerts.IntervalSeconds = 1
	if err := Save(path, config); !errors.Is(err, ErrInvalidConfig) {
		testRun.Fatalf("got %v", err)
	}
	assertContent(testRun, path, `{"version":1}`)
}

func TestAtomicWriteKeepsPreviousFileWhenRenameFails(testRun *testing.T) {
	directory := testRun.TempDir()
	path := writeConfigFile(testRun, directory, `previous`, 0o600)
	failingRename := func(string, string) error { return errors.New("simulated rename failure") }
	if err := writeAtomically(path, []byte("replacement"), failingRename); err == nil {
		testRun.Fatal("expected failure")
	}
	assertContent(testRun, path, "previous")
	assertNoTemporaryFiles(testRun, directory)
}

func TestAtomicWriteKeepsPreviousFileWhenDirectoryIsUnwritable(testRun *testing.T) {
	if os.Geteuid() == 0 {
		testRun.Skip("root ignores directory permissions")
	}
	directory := testRun.TempDir()
	path := writeConfigFile(testRun, directory, `previous`, 0o600)
	if err := os.Chmod(directory, 0o500); err != nil {
		testRun.Fatal(err)
	}
	defer os.Chmod(directory, 0o700)
	if err := writeAtomically(path, []byte("replacement"), os.Rename); err == nil {
		testRun.Fatal("expected failure")
	}
	assertContent(testRun, path, "previous")
}

func TestDefaultPath(testRun *testing.T) {
	testCases := []struct {
		name        string
		environment Environment
		want        string
		wantErr     bool
	}{
		{"root", Environment{UserID: 0, Getenv: noEnvironment}, "/etc/exposedwatch/config.json", false},
		{"xdg", Environment{UserID: 1000, Getenv: environmentOf(map[string]string{"XDG_CONFIG_HOME": "/xdg"})}, "/xdg/exposedwatch/config.json", false},
		{"relative xdg is ignored", Environment{UserID: 1000, Getenv: environmentOf(map[string]string{"XDG_CONFIG_HOME": "relative"}),
			HomeDirectory: func() (string, error) { return "/home/user", nil }}, "/home/user/.config/exposedwatch/config.json", false},
		{"home", Environment{UserID: 1000, Getenv: noEnvironment,
			HomeDirectory: func() (string, error) { return "/home/user", nil }}, "/home/user/.config/exposedwatch/config.json", false},
		{"no home", Environment{UserID: 1000, Getenv: noEnvironment,
			HomeDirectory: func() (string, error) { return "", errors.New("none") }}, "", true},
	}
	for _, testCase := range testCases {
		testRun.Run(testCase.name, func(testRun *testing.T) {
			got, err := DefaultPath(testCase.environment)
			if (err != nil) != testCase.wantErr || got != testCase.want {
				testRun.Fatalf("got %q, %v", got, err)
			}
		})
	}
}

func TestMaskToken(testRun *testing.T) {
	testCases := map[string]string{
		"":                       "",
		"123456789:AA-secret_x+": "123456789:****",
		"nocolonsecret":          "****",
	}
	for token, want := range testCases {
		got := MaskToken(domain.BotToken(token))
		if got != want || (token != "" && strings.Contains(got, "secret")) {
			testRun.Errorf("MaskToken -> %q, want %q", got, want)
		}
	}
}

func assertMode(testRun *testing.T, path string, want os.FileMode) {
	testRun.Helper()
	info, err := os.Stat(path)
	if err != nil {
		testRun.Fatal(err)
	}
	if info.Mode().Perm() != want {
		testRun.Fatalf("%s mode %v, want %v", path, info.Mode().Perm(), want)
	}
}

func assertContent(testRun *testing.T, path string, want string) {
	testRun.Helper()
	content, err := os.ReadFile(path)
	if err != nil || string(content) != want {
		testRun.Fatalf("content %q err %v", content, err)
	}
}

func assertNoTemporaryFiles(testRun *testing.T, directory string) {
	testRun.Helper()
	entries, err := os.ReadDir(directory)
	if err != nil {
		testRun.Fatal(err)
	}
	if len(entries) != 1 {
		testRun.Fatalf("leftover files in %s: %d entries", directory, len(entries))
	}
}
