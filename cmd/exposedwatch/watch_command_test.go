package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Byroncl/exposedwatch/internal/adapters/configfile"
	"github.com/Byroncl/exposedwatch/internal/app"
	"github.com/Byroncl/exposedwatch/internal/domain"
	"github.com/Byroncl/exposedwatch/internal/ports"
)

const (
	testSecretToken = "123456789:AA-fake_TOKEN+xyz"
	testChatID      = "-1001234567890"
	enabledConfig   = `{"version":1,"telegram":{"enabled":true,"bot_token":"` + testSecretToken + `","chat_id":"` + testChatID + `"}}`
	disabledConfig  = `{"version":1,"telegram":{"enabled":false,"bot_token":"` + testSecretToken + `","chat_id":"` + testChatID + `"}}`
	enabledNoSecret = `{"version":1,"telegram":{"enabled":true}}`
)

// steppingClock only moves when a test (or the fake wait) advances it.
type steppingClock struct{ current time.Time }

func (clock *steppingClock) Now() time.Time { return clock.current }

// recordingNotifier fails the first failuresLeft sends with an error that
// deliberately contains the token, to prove the command masks it.
type recordingNotifier struct {
	failuresLeft int
	sent         []domain.Alert
}

func (notifier *recordingNotifier) Send(_ context.Context, alert domain.Alert) error {
	if notifier.failuresLeft > 0 {
		notifier.failuresLeft--
		return errors.New("simulated failure for " + testSecretToken)
	}
	notifier.sent = append(notifier.sent, alert)
	return nil
}

type watchFixture struct {
	runtime    commandRuntime
	notifier   *recordingNotifier
	scanCount  int
	waitCount  int
	waitBudget int // how many waits return true before the loop is stopped
	// waitAdvance is how far each wait moves the clock; zero means the
	// requested interval.
	waitAdvance time.Duration
	clock       *steppingClock
	reports     []app.Report
	scanError   error
	environ     map[string]string
	notifierID  []domain.ChatID
}

func newWatchFixture() *watchFixture {
	fixture := &watchFixture{
		notifier: &recordingNotifier{},
		environ:  map[string]string{},
		clock:    &steppingClock{current: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)},
	}
	fixture.runtime = commandRuntime{
		Environment: configfile.Environment{
			UserID: 1000,
			Getenv: func(name string) string { return fixture.environ[name] },
			HomeDirectory: func() (string, error) {
				return "/nonexistent-home", nil
			},
		},
		HostName: func() (string, error) { return "web-01", nil },
		NewNotifier: func(_ domain.BotToken, chatID domain.ChatID) ports.Notifier {
			fixture.notifierID = append(fixture.notifierID, chatID)
			return fixture.notifier
		},
		NewScanner: func(ScanOptions) scanFunc {
			return func(context.Context) (app.Report, error) {
				index := fixture.scanCount
				fixture.scanCount++
				if fixture.scanError != nil {
					return app.Report{}, fixture.scanError
				}
				if index >= len(fixture.reports) {
					index = len(fixture.reports) - 1
				}
				return fixture.reports[index], nil
			}
		},
		Clock: fixture.clock,
		Wait: func(_ context.Context, duration time.Duration) bool {
			fixture.waitCount++
			if fixture.waitAdvance != 0 {
				duration = fixture.waitAdvance
			}
			fixture.clock.current = fixture.clock.current.Add(duration)
			return fixture.waitCount <= fixture.waitBudget
		},
	}
	return fixture
}

func writeWatchConfig(testRun *testing.T, content string) string {
	testRun.Helper()
	path := filepath.Join(testRun.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		testRun.Fatal(err)
	}
	return path
}

func riskyReport(severity domain.Severity) app.Report {
	return app.Report{Findings: []domain.Finding{{
		Port:     domain.Port{Number: 5432, Protocol: domain.TCP, BindAddress: "0.0.0.0", Process: "postgres", Source: domain.SourceHost},
		Severity: severity, RuleID: "database-exposed", Title: "Database exposed", Recommendation: "Bind to loopback.",
	}}}
}

func execute(fixture *watchFixture, ctx context.Context, arguments ...string) (ExitCode, string, string) {
	var stdout, stderr bytes.Buffer
	code := runWatchWith(ctx, fixture.runtime, arguments, &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

func TestWatchOnceSendsTheSummaryAndExitsZero(testRun *testing.T) {
	fixture := newWatchFixture()
	fixture.reports = []app.Report{riskyReport(domain.Critical)}
	path := writeWatchConfig(testRun, enabledConfig)
	code, _, stderr := execute(fixture, context.Background(), "--config", path, "--once")
	if code != ExitOK || len(fixture.notifier.sent) != 1 || fixture.notifier.sent[0].Kind != domain.AlertKindSummary {
		testRun.Fatalf("code=%d sent=%+v stderr=%q", code, fixture.notifier.sent, stderr)
	}
	if strings.Contains(stderr, testSecretToken) {
		testRun.Fatal("log leaks the token")
	}
	if fixture.scanCount != 1 || fixture.waitCount != 0 {
		testRun.Fatalf("scans=%d waits=%d", fixture.scanCount, fixture.waitCount)
	}
}

func TestWatchWithNotificationsDisabledStillRunsAndLogsOnce(testRun *testing.T) {
	fixture := newWatchFixture()
	fixture.reports = []app.Report{riskyReport(domain.Critical)}
	fixture.waitBudget = 2
	path := writeWatchConfig(testRun, disabledConfig)
	code, _, stderr := execute(fixture, context.Background(), "--config", path)
	if code != ExitOK || fixture.scanCount != 3 || len(fixture.notifier.sent) != 0 || len(fixture.notifierID) != 0 {
		testRun.Fatalf("code=%d scans=%d", code, fixture.scanCount)
	}
	if strings.Count(stderr, "notifications disabled") != 1 {
		testRun.Fatalf("stderr %q", stderr)
	}
	if !strings.Contains(stderr, "alert: CRITICAL") || strings.Contains(stderr, testSecretToken) {
		testRun.Fatalf("stderr %q", stderr)
	}
}

func TestWatchWithoutCredentialsIsDisabled(testRun *testing.T) {
	fixture := newWatchFixture()
	fixture.reports = []app.Report{{}}
	path := writeWatchConfig(testRun, enabledNoSecret)
	code, _, stderr := execute(fixture, context.Background(), "--config", path, "--once")
	if code != ExitOK || !strings.Contains(stderr, "notifications disabled") {
		testRun.Fatalf("code=%d stderr=%q", code, stderr)
	}
}

func TestWatchTakesSecretsFromTheEnvironment(testRun *testing.T) {
	fixture := newWatchFixture()
	fixture.environ[configfile.EnvTelegramToken] = testSecretToken
	fixture.environ[configfile.EnvTelegramChatID] = testChatID
	fixture.reports = []app.Report{riskyReport(domain.High)}
	path := writeWatchConfig(testRun, enabledNoSecret)
	code, _, _ := execute(fixture, context.Background(), "--config", path, "--once")
	if code != ExitOK || len(fixture.notifier.sent) != 1 || fixture.notifierID[0] != testChatID {
		testRun.Fatalf("code=%d sent=%d", code, len(fixture.notifier.sent))
	}
}

func TestWatchFlagOverridesConfigMinSeverity(testRun *testing.T) {
	fixture := newWatchFixture()
	fixture.reports = []app.Report{riskyReport(domain.High)}
	path := writeWatchConfig(testRun, enabledConfig)
	code, _, _ := execute(fixture, context.Background(), "--config", path, "--once", "--min-severity", "critical")
	if code != ExitOK || len(fixture.notifier.sent) != 0 {
		testRun.Fatalf("code=%d sent=%d", code, len(fixture.notifier.sent))
	}
}

func TestWatchOnceSendFailureExitsOneAndMasksTheToken(testRun *testing.T) {
	fixture := newWatchFixture()
	fixture.notifier.failuresLeft = 5
	fixture.reports = []app.Report{riskyReport(domain.Critical)}
	path := writeWatchConfig(testRun, enabledConfig)
	code, stdout, stderr := execute(fixture, context.Background(), "--config", path, "--once")
	if code != ExitRuntimeError {
		testRun.Fatalf("code=%d", code)
	}
	if strings.Contains(stderr+stdout, testSecretToken) || !strings.Contains(stderr, "alert delivery failed") {
		testRun.Fatalf("stderr %q", stderr)
	}
}

func TestWatchRetriesAFailedAlertOnTheNextInterval(testRun *testing.T) {
	fixture := newWatchFixture()
	fixture.notifier.failuresLeft = 1
	fixture.waitBudget = 1
	fixture.reports = []app.Report{riskyReport(domain.Critical)}
	path := writeWatchConfig(testRun, enabledConfig)
	code, _, stderr := execute(fixture, context.Background(), "--config", path)
	if code != ExitOK || fixture.scanCount != 2 || len(fixture.notifier.sent) != 1 {
		testRun.Fatalf("code=%d scans=%d sent=%d stderr=%q", code, fixture.scanCount, len(fixture.notifier.sent), stderr)
	}
	if !strings.Contains(stderr, "will retry") || strings.Contains(stderr, testSecretToken) {
		testRun.Fatalf("stderr %q", stderr)
	}
}

func TestWatchBacksOffInsteadOfGivingUpDuringAnOutage(testRun *testing.T) {
	fixture := newWatchFixture()
	fixture.notifier.failuresLeft = 6 // far more than the old 3 attempts
	fixture.waitBudget = 12
	fixture.waitAdvance = maxRetryBackoff // every wait makes the retry due
	fixture.reports = []app.Report{riskyReport(domain.Critical)}
	path := writeWatchConfig(testRun, enabledConfig)
	_, _, stderr := execute(fixture, context.Background(), "--config", path)
	if len(fixture.notifier.sent) != 1 {
		testRun.Fatalf("the alert must survive the outage: sent=%d stderr=%q", len(fixture.notifier.sent), stderr)
	}
	if strings.Contains(stderr, "giving up") || strings.Contains(stderr, testSecretToken) {
		testRun.Fatalf("stderr %q", stderr)
	}
}

func TestWatchDropsAnExpiredAlertAndLetsTheFindingAlertAgain(testRun *testing.T) {
	fixture := newWatchFixture()
	fixture.notifier.failuresLeft = 1
	fixture.waitAdvance = 25 * time.Hour // beyond the alert lifetime
	fixture.waitBudget = 3
	fixture.reports = []app.Report{{}, riskyReport(domain.Critical)}
	path := writeWatchConfig(testRun, enabledConfig)
	_, _, stderr := execute(fixture, context.Background(), "--config", path)
	if !strings.Contains(stderr, "giving up") {
		testRun.Fatalf("expected the expired alert to be dropped: %q", stderr)
	}
	if len(fixture.notifier.sent) != 1 || fixture.notifier.sent[0].Kind != domain.AlertKindNewFinding {
		testRun.Fatalf("the finding must alert again after the drop: sent=%+v stderr=%q", fixture.notifier.sent, stderr)
	}
}

func TestWatchScanFailureKeepsTheLoopAlive(testRun *testing.T) {
	fixture := newWatchFixture()
	fixture.scanError = errors.New("every source failed")
	fixture.waitBudget = 1
	path := writeWatchConfig(testRun, enabledConfig)
	code, _, stderr := execute(fixture, context.Background(), "--config", path)
	if code != ExitOK || fixture.scanCount != 2 || !strings.Contains(stderr, "scan failed") {
		testRun.Fatalf("code=%d scans=%d stderr=%q", code, fixture.scanCount, stderr)
	}
	onceCode, _, _ := execute(newFailingScanFixture(), context.Background(), "--config", path, "--once")
	if onceCode != ExitRuntimeError {
		testRun.Fatalf("--once with a failing scan: %d", onceCode)
	}
}

func newFailingScanFixture() *watchFixture {
	fixture := newWatchFixture()
	fixture.scanError = errors.New("every source failed")
	return fixture
}

func TestWatchStopsWhenTheContextIsCanceled(testRun *testing.T) {
	fixture := newWatchFixture()
	fixture.reports = []app.Report{{}}
	path := writeWatchConfig(testRun, enabledConfig)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	code, _, stderr := execute(fixture, ctx, "--config", path)
	if code != ExitOK || fixture.scanCount != 0 || !strings.Contains(stderr, "shutting down") {
		testRun.Fatalf("code=%d scans=%d", code, fixture.scanCount)
	}
}

func TestWatchStopsWhenWaitingIsInterrupted(testRun *testing.T) {
	fixture := newWatchFixture()
	fixture.reports = []app.Report{{}}
	fixture.waitBudget = 0 // wait reports "context ended" immediately
	path := writeWatchConfig(testRun, enabledConfig)
	code, _, _ := execute(fixture, context.Background(), "--config", path)
	if code != ExitOK || fixture.scanCount != 1 {
		testRun.Fatalf("code=%d scans=%d", code, fixture.scanCount)
	}
}

func TestWatchReportsConfigProblems(testRun *testing.T) {
	fixture := newWatchFixture()
	path := writeWatchConfig(testRun, `{"alerts":{"interval_seconds":1}}`)
	code, _, stderr := execute(fixture, context.Background(), "--config", path, "--once")
	if code != ExitRuntimeError || !strings.Contains(stderr, "config error") {
		testRun.Fatalf("code=%d stderr=%q", code, stderr)
	}
	loose := filepath.Join(testRun.TempDir(), "config.json")
	if err := os.WriteFile(loose, []byte(disabledConfig), 0o644); err != nil {
		testRun.Fatal(err)
	}
	fixture.reports = []app.Report{{}}
	code, _, stderr = execute(fixture, context.Background(), "--config", loose, "--once")
	if code != ExitOK || !strings.Contains(stderr, "config warning") || strings.Contains(stderr, testSecretToken) {
		testRun.Fatalf("code=%d stderr=%q", code, stderr)
	}
}

func TestParseWatchOptions(testRun *testing.T) {
	testCases := []struct {
		name      string
		arguments []string
		wantError bool
		check     func(WatchOptions) bool
	}{
		{"defaults", nil, false, func(options WatchOptions) bool {
			return options.Interval == 0 && options.MinSeverity == "" && !options.Once && options.ProcRoot == "/proc" && !options.NoDocker
		}},
		{"all flags", []string{"--config", "/tmp/c.json", "--interval", "30s", "--min-severity", "medium", "--once",
			"--proc-root", "/fake", "--docker-socket", "/s.sock", "--no-docker"}, false, func(options WatchOptions) bool {
			return options.ConfigPath == "/tmp/c.json" && options.Interval == 30*time.Second && options.MinSeverity == "medium" &&
				options.Once && options.ProcRoot == "/fake" && options.DockerSock == "/s.sock" && options.NoDocker
		}},
		{"interval too short", []string{"--interval", "5s"}, true, nil},
		{"bad severity", []string{"--min-severity", "urgent"}, true, nil},
		{"unexpected argument", []string{"extra"}, true, nil},
		{"unknown flag", []string{"--external"}, true, nil},
	}
	for _, testCase := range testCases {
		testRun.Run(testCase.name, func(testRun *testing.T) {
			options, err := parseWatchOptions(testCase.arguments, &bytes.Buffer{})
			if (err != nil) != testCase.wantError {
				testRun.Fatalf("err=%v", err)
			}
			if err == nil && !testCase.check(options) {
				testRun.Fatalf("%+v", options)
			}
		})
	}
}

func TestWatchUsageErrorsExitTwoThroughRun(testRun *testing.T) {
	for _, arguments := range [][]string{
		{"watch", "--interval", "1s"}, {"watch", "--bogus"}, {"watch", "extra"},
		{"notify-test", "--bogus"}, {"notify-test", "extra"},
	} {
		if code, _, _ := runCommand(arguments...); code != ExitUsage {
			testRun.Errorf("%v exited %d", arguments, code)
		}
	}
}

func TestUsageMentionsTheNewCommands(testRun *testing.T) {
	_, stdout, _ := runCommand("help")
	for _, want := range []string{"watch", "notify-test"} {
		if !strings.Contains(stdout, want) {
			testRun.Errorf("usage lacks %q", want)
		}
	}
}
