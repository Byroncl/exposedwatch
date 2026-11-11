package main

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Byroncl/exposedwatch/internal/adapters/baselinefile"
	"github.com/Byroncl/exposedwatch/internal/app"
	"github.com/Byroncl/exposedwatch/internal/domain"
	"github.com/Byroncl/exposedwatch/internal/ports"
)

// baselineStep is one result of the fake baseline store.
type baselineStep struct {
	entries  []domain.BaselineEntry
	err      error
	warnings []string
	savedAt  time.Time // zero means the default date
}

// fakeBaselineStore answers successive loads from its steps; the last step
// repeats once they run out.
type fakeBaselineStore struct {
	steps []baselineStep
	calls int
}

func (store *fakeBaselineStore) load(ports.BaselinePath) (ports.BaselineLoad, error) {
	index := min(store.calls, len(store.steps)-1)
	store.calls++
	step := store.steps[index]
	if step.err != nil {
		return ports.BaselineLoad{}, step.err
	}
	savedAt := step.savedAt
	if savedAt.IsZero() {
		savedAt = time.Date(2026, 11, 11, 10, 0, 0, 0, time.UTC)
	}
	return ports.BaselineLoad{Baseline: domain.Baseline{
		Version: domain.BaselineVersion, CreatedAt: savedAt,
		ToolVersion: "test", Entries: step.entries,
	}, Warnings: step.warnings}, nil
}

const fakeBaselineFile = "/fake/baseline.json"

func realStoreLoad(path ports.BaselinePath) (ports.BaselineLoad, error) {
	return baselinefile.NewStore(baselinefile.SystemTrust()).Load(path)
}

func disabledNotificationsConfig(testRun *testing.T) string {
	return writeWatchConfig(testRun, disabledConfig)
}

func sshAndPostgresReport() app.Report {
	report := riskyReport(domain.Critical)
	report.Findings = append(report.Findings, domain.Finding{
		Port:     domain.Port{Number: 22, Protocol: domain.TCP, BindAddress: "0.0.0.0", Process: "sshd", Source: domain.SourceHost},
		Severity: domain.High, RuleID: "ssh-exposed", Title: "SSH exposed", Recommendation: "Restrict it.",
	})
	return report
}

func TestWatchRejectsAnEmptyBaselineFlag(testRun *testing.T) {
	fixture := newWatchFixture()
	fixture.runtime.LoadBaseline = realStoreLoad
	code, _, stderr := execute(fixture, context.Background(), "--baseline", "", "--once")
	if code != ExitUsage || !strings.Contains(stderr, "baseline") {
		testRun.Fatalf("code=%d stderr=%q", code, stderr)
	}
}

func TestWatchRefusesToStartWithAnUnusableBaseline(testRun *testing.T) {
	trustedPath := storeBaseline(testRun)
	writable := storeBaseline(testRun)
	if err := os.Chmod(writable, 0o666); err != nil {
		testRun.Fatal(err)
	}
	corrupt := storeBaseline(testRun)
	if err := os.WriteFile(corrupt, []byte("not json"), 0o600); err != nil {
		testRun.Fatal(err)
	}
	link := trustedPath + ".link"
	if err := os.Symlink(trustedPath, link); err != nil {
		testRun.Fatal(err)
	}
	testCases := map[string]string{
		"missing":        trustedPath + ".missing",
		"group writable": writable,
		"corrupt":        corrupt,
		"symlink":        link,
	}
	for name, path := range testCases {
		testRun.Run(name, func(testRun *testing.T) {
			fixture := newWatchFixture()
			fixture.runtime.LoadBaseline = realStoreLoad
			fixture.reports = []app.Report{riskyReport(domain.Critical)}
			config := writeWatchConfig(testRun, enabledConfig)
			code, _, stderr := execute(fixture, context.Background(), "--config", config, "--baseline", path, "--once")
			if code != ExitRuntimeError || !strings.Contains(stderr, "exposedwatch baseline save") {
				testRun.Fatalf("code=%d stderr=%q", code, stderr)
			}
			if fixture.scanCount != 0 || len(fixture.notifier.sent) != 0 {
				testRun.Fatalf("must not scan or alert: scans=%d sent=%d", fixture.scanCount, len(fixture.notifier.sent))
			}
			if strings.Contains(stderr, "not json") || strings.Contains(stderr, testSecretToken) {
				testRun.Fatalf("leak: %q", stderr)
			}
		})
	}
}

func TestWatchOnceWithBaselineAlertsOnlyDrift(testRun *testing.T) {
	fixture := newWatchFixture()
	path := storeBaseline(testRun, hostEntry(22, "0.0.0.0"))
	fixture.runtime.LoadBaseline = realStoreLoad
	fixture.reports = []app.Report{sshAndPostgresReport()}
	config := writeWatchConfig(testRun, enabledConfig)
	code, _, stderr := execute(fixture, context.Background(), "--config", config, "--baseline", path, "--once")
	if code != ExitOK || len(fixture.notifier.sent) != 1 {
		testRun.Fatalf("code=%d sent=%+v stderr=%q", code, fixture.notifier.sent, stderr)
	}
	alert := fixture.notifier.sent[0]
	if alert.Kind != domain.AlertKindNewFinding || alert.Port != 5432 {
		testRun.Fatalf("%+v", alert)
	}
	if !strings.Contains(alert.Body, "Not in baseline (saved 2026-11-11)") || strings.Contains(alert.Body, path) {
		testRun.Fatalf("body %q", alert.Body)
	}
	if fixture.scanCount != 1 || fixture.waitCount != 0 {
		testRun.Fatalf("scans=%d waits=%d", fixture.scanCount, fixture.waitCount)
	}
}

func TestWatchOnceWithBaselineAndNoDriftSendsNothing(testRun *testing.T) {
	fixture := newWatchFixture()
	path := storeBaseline(testRun, hostEntry(22, "0.0.0.0"), hostEntry(5432, "0.0.0.0"))
	fixture.runtime.LoadBaseline = realStoreLoad
	fixture.reports = []app.Report{sshAndPostgresReport()}
	config := writeWatchConfig(testRun, enabledConfig)
	code, _, _ := execute(fixture, context.Background(), "--config", config, "--baseline", path, "--once")
	if code != ExitOK || len(fixture.notifier.sent) != 0 {
		testRun.Fatalf("code=%d sent=%+v", code, fixture.notifier.sent)
	}
}

func TestWatchOnceSendFailureStillExitsOneWithBaseline(testRun *testing.T) {
	fixture := newWatchFixture()
	fixture.notifier.failuresLeft = 5
	path := storeBaseline(testRun)
	fixture.runtime.LoadBaseline = realStoreLoad
	fixture.reports = []app.Report{riskyReport(domain.Critical)}
	config := writeWatchConfig(testRun, enabledConfig)
	code, _, stderr := execute(fixture, context.Background(), "--config", config, "--baseline", path, "--once")
	if code != ExitRuntimeError || strings.Contains(stderr, testSecretToken) {
		testRun.Fatalf("code=%d stderr=%q", code, stderr)
	}
}

func TestWatchOnceLoadsTheBaselineExactlyOnce(testRun *testing.T) {
	fixture := newWatchFixture()
	store := &fakeBaselineStore{steps: []baselineStep{{}}}
	fixture.runtime.LoadBaseline = store.load
	fixture.reports = []app.Report{{}}
	config := disabledNotificationsConfig(testRun)
	execute(fixture, context.Background(), "--config", config, "--baseline", fakeBaselineFile, "--once")
	if store.calls != 1 {
		testRun.Fatalf("loads=%d", store.calls)
	}
}

func TestWatchLoopKeepsLastGoodBaselineAndLogsEachStateChangeOnce(testRun *testing.T) {
	fixture := newWatchFixture()
	fixture.waitBudget = 5
	broken := errors.New("baseline file is group-writable")
	store := &fakeBaselineStore{steps: []baselineStep{
		{entries: []domain.BaselineEntry{hostEntry(5432, "0.0.0.0")}}, // startup
		{entries: []domain.BaselineEntry{hostEntry(5432, "0.0.0.0")}}, // cycle 2
		{err: broken}, {err: broken}, {err: broken},
		{entries: []domain.BaselineEntry{hostEntry(5432, "0.0.0.0")}},
	}}
	fixture.runtime.LoadBaseline = store.load
	fixture.reports = []app.Report{riskyReport(domain.Critical)}
	config := disabledNotificationsConfig(testRun)
	code, _, stderr := execute(fixture, context.Background(), "--config", config, "--baseline", fakeBaselineFile)
	if code != ExitOK {
		testRun.Fatalf("code=%d", code)
	}
	if strings.Count(stderr, "baseline reload failed: baseline file is group-writable; using the last good baseline") != 1 {
		testRun.Fatalf("stderr %q", stderr)
	}
	if strings.Count(stderr, "baseline reload recovered") != 1 {
		testRun.Fatalf("stderr %q", stderr)
	}
	if strings.Contains(stderr, "alert:") {
		testRun.Fatalf("the last good baseline must keep accepting the port: %q", stderr)
	}
}

func TestWatchLoopNeverFallsBackToNoBaselineWhenTheFirstReloadFails(testRun *testing.T) {
	fixture := newWatchFixture()
	fixture.waitBudget = 2
	store := &fakeBaselineStore{steps: []baselineStep{
		{entries: []domain.BaselineEntry{hostEntry(5432, "0.0.0.0")}},
		{err: errors.New("gone")},
	}}
	fixture.runtime.LoadBaseline = store.load
	fixture.reports = []app.Report{riskyReport(domain.Critical)}
	config := disabledNotificationsConfig(testRun)
	_, _, stderr := execute(fixture, context.Background(), "--config", config, "--baseline", fakeBaselineFile)
	if strings.Contains(stderr, "alert:") || strings.Contains(stderr, "already present") {
		testRun.Fatalf("stderr %q", stderr)
	}
}

func TestWatchLoopAppliesAChangedBaselineImmediately(testRun *testing.T) {
	fixture := newWatchFixture()
	fixture.waitBudget = 2
	store := &fakeBaselineStore{steps: []baselineStep{
		{}, // startup: nothing accepted
		{}, // cycle 2
		{entries: []domain.BaselineEntry{hostEntry(5432, "0.0.0.0")}}, // cycle 3: accepted now
	}}
	fixture.runtime.LoadBaseline = store.load
	fixture.reports = []app.Report{riskyReport(domain.Critical)}
	config := disabledNotificationsConfig(testRun)
	_, _, stderr := execute(fixture, context.Background(), "--config", config, "--baseline", fakeBaselineFile)
	if strings.Count(stderr, "alert: CRITICAL") != 1 {
		testRun.Fatalf("one alert in the first cycle, none after the baseline accepts it: %q", stderr)
	}
}

func TestWatchLoopAlertsForAPortRemovedFromTheBaseline(testRun *testing.T) {
	fixture := newWatchFixture()
	fixture.waitBudget = 2
	store := &fakeBaselineStore{steps: []baselineStep{
		{entries: []domain.BaselineEntry{hostEntry(5432, "0.0.0.0")}}, // startup
		{entries: []domain.BaselineEntry{hostEntry(5432, "0.0.0.0")}}, // cycle 2
		{}, // cycle 3: no longer accepted
	}}
	fixture.runtime.LoadBaseline = store.load
	fixture.reports = []app.Report{riskyReport(domain.Critical)}
	config := disabledNotificationsConfig(testRun)
	_, _, stderr := execute(fixture, context.Background(), "--config", config, "--baseline", fakeBaselineFile)
	if strings.Count(stderr, "alert: CRITICAL") != 1 {
		testRun.Fatalf("stderr %q", stderr)
	}
}

func TestWatchLoopLogsThePathOnlyInTheStartupLine(testRun *testing.T) {
	fixture := newWatchFixture()
	fixture.waitBudget = 2
	store := &fakeBaselineStore{steps: []baselineStep{{}, {err: errors.New("denied")}, {}}}
	fixture.runtime.LoadBaseline = store.load
	fixture.reports = []app.Report{riskyReport(domain.Critical)}
	config := disabledNotificationsConfig(testRun)
	_, _, stderr := execute(fixture, context.Background(), "--config", config, "--baseline", fakeBaselineFile)
	if strings.Count(stderr, fakeBaselineFile) != 1 || !strings.Contains(stderr, "baseline: "+fakeBaselineFile) {
		testRun.Fatalf("stderr %q", stderr)
	}
}

func TestWatchWithBaselineStopsOnShutdown(testRun *testing.T) {
	fixture := newWatchFixture()
	store := &fakeBaselineStore{steps: []baselineStep{{}}}
	fixture.runtime.LoadBaseline = store.load
	fixture.reports = []app.Report{{}}
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	config := disabledNotificationsConfig(testRun)
	code, _, stderr := execute(fixture, cancelled, "--config", config, "--baseline", fakeBaselineFile)
	if code != ExitOK || !strings.Contains(stderr, "shutting down") || fixture.scanCount != 0 {
		testRun.Fatalf("code=%d scans=%d stderr=%q", code, fixture.scanCount, stderr)
	}
}

func TestWatchWithoutBaselineStillSendsTheSummary(testRun *testing.T) {
	fixture := newWatchFixture()
	fixture.reports = []app.Report{riskyReport(domain.Critical)}
	config := writeWatchConfig(testRun, enabledConfig)
	code, _, _ := execute(fixture, context.Background(), "--config", config, "--once")
	if code != ExitOK || len(fixture.notifier.sent) != 1 || fixture.notifier.sent[0].Kind != domain.AlertKindSummary {
		testRun.Fatalf("code=%d sent=%+v", code, fixture.notifier.sent)
	}
}

func runReloadSteps(testRun *testing.T, steps []baselineStep) string {
	testRun.Helper()
	fixture := newWatchFixture()
	fixture.waitBudget = len(steps) - 1
	store := &fakeBaselineStore{steps: steps}
	fixture.runtime.LoadBaseline = store.load
	fixture.reports = []app.Report{{}}
	config := disabledNotificationsConfig(testRun)
	_, _, stderr := execute(fixture, context.Background(), "--config", config, "--baseline", fakeBaselineFile)
	return stderr
}

func entriesUpTo(count int) []domain.BaselineEntry {
	entries := make([]domain.BaselineEntry, 0, count)
	for port := 1; port <= count; port++ {
		entries = append(entries, hostEntry(port, "0.0.0.0"))
	}
	return entries
}

func TestReloadWithUnchangedBaselineLogsNothing(testRun *testing.T) {
	stderr := runReloadSteps(testRun, []baselineStep{{entries: entriesUpTo(3)}, {entries: entriesUpTo(3)}, {entries: entriesUpTo(3)}, {entries: entriesUpTo(3)}})
	if strings.Contains(stderr, "baseline reloaded") {
		testRun.Fatalf("stderr %q", stderr)
	}
}

func TestReloadWithChangedEntryCountLogsOneLine(testRun *testing.T) {
	stderr := runReloadSteps(testRun, []baselineStep{{entries: entriesUpTo(3)}, {entries: entriesUpTo(4)}, {entries: entriesUpTo(4)}, {entries: entriesUpTo(4)}})
	if strings.Count(stderr, "baseline reloaded: 4 entries (was 3), saved 2026-11-11") != 1 || strings.Contains(stderr, "WIDENED") {
		testRun.Fatalf("stderr %q", stderr)
	}
}

func TestReloadWithBigGrowthIsMarkedWidened(testRun *testing.T) {
	stderr := runReloadSteps(testRun, []baselineStep{{entries: entriesUpTo(100)}, {entries: entriesUpTo(150)}, {entries: entriesUpTo(150)}})
	if strings.Count(stderr, "WIDENED") != 1 || !strings.Contains(stderr, "baseline reloaded: 150 entries (was 100)") {
		testRun.Fatalf("stderr %q", stderr)
	}
	doubled := runReloadSteps(testRun, []baselineStep{{entries: entriesUpTo(3)}, {entries: entriesUpTo(7)}, {entries: entriesUpTo(7)}})
	if strings.Count(doubled, "WIDENED") != 1 {
		testRun.Fatalf("more than doubling is widening: %q", doubled)
	}
	shrunk := runReloadSteps(testRun, []baselineStep{{entries: entriesUpTo(150)}, {entries: entriesUpTo(2)}, {entries: entriesUpTo(2)}})
	if strings.Contains(shrunk, "WIDENED") || !strings.Contains(shrunk, "baseline reloaded: 2 entries (was 150)") {
		testRun.Fatalf("%q", shrunk)
	}
}

func TestReloadWithOnlyANewDateLogsOneLine(testRun *testing.T) {
	later := time.Date(2026, 12, 1, 9, 0, 0, 0, time.UTC)
	stderr := runReloadSteps(testRun, []baselineStep{{entries: entriesUpTo(3)}, {entries: entriesUpTo(3), savedAt: later}, {entries: entriesUpTo(3), savedAt: later}, {entries: entriesUpTo(3), savedAt: later}})
	if strings.Count(stderr, "baseline reloaded: 3 entries (was 3), saved 2026-12-01") != 1 {
		testRun.Fatalf("stderr %q", stderr)
	}
}

func TestReloadFailureLogsAgainOnlyWhenTheReasonChanges(testRun *testing.T) {
	missing, untrusted := errors.New("file is missing"), errors.New("file is group-writable")
	good := baselineStep{entries: entriesUpTo(3)}
	stderr := runReloadSteps(testRun, []baselineStep{good, {err: missing}, {err: missing}, {err: untrusted}, {err: untrusted}, good})
	if strings.Count(stderr, "file is missing") != 1 || strings.Count(stderr, "file is group-writable") != 1 {
		testRun.Fatalf("stderr %q", stderr)
	}
	if strings.Count(stderr, "baseline reload failed") != 2 || strings.Count(stderr, "baseline reload recovered") != 1 {
		testRun.Fatalf("stderr %q", stderr)
	}
}

func TestReloadWarningsAreLoggedOncePerChange(testRun *testing.T) {
	unsafeParent := []string{"parent directory is group-writable"}
	good := baselineStep{entries: entriesUpTo(3), warnings: unsafeParent}
	changed := baselineStep{entries: entriesUpTo(3), warnings: []string{"another warning"}}
	stderr := runReloadSteps(testRun, []baselineStep{good, good, good, changed, changed, {entries: entriesUpTo(3)}})
	if strings.Count(stderr, "baseline warning: parent directory is group-writable") != 1 {
		testRun.Fatalf("startup warning once: %q", stderr)
	}
	if strings.Count(stderr, "baseline warning: another warning") != 1 {
		testRun.Fatalf("changed warning once: %q", stderr)
	}
}
