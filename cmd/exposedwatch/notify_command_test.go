package main

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/Byroncl/exposedwatch/internal/adapters/configfile"
	"github.com/Byroncl/exposedwatch/internal/domain"
)

func executeNotifyTest(fixture *watchFixture, arguments ...string) (ExitCode, string, string) {
	var stdout, stderr bytes.Buffer
	code := runNotifyTestWith(context.Background(), fixture.runtime, arguments, &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

func TestNotifyTestSendsATestAlertAndShowsOnlyTheMaskedToken(testRun *testing.T) {
	fixture := newWatchFixture()
	fixture.environ[configfile.EnvTelegramToken] = testSecretToken
	path := writeWatchConfig(testRun, disabledConfig)
	code, stdout, stderr := executeNotifyTest(fixture, "--config", path)
	if code != ExitOK || len(fixture.notifier.sent) != 1 {
		testRun.Fatalf("code=%d stderr=%q", code, stderr)
	}
	sent := fixture.notifier.sent[0]
	if sent.Kind != domain.AlertKindTest || !strings.Contains(sent.Title, "exposedwatch test message from web-01") {
		testRun.Fatalf("%+v", sent)
	}
	if !strings.Contains(stdout, "123456789:****") || !strings.Contains(stdout, "environment") || !strings.Contains(stdout, "file") {
		testRun.Fatalf("stdout %q", stdout)
	}
	if strings.Contains(stdout+stderr, "AA-fake") || strings.Contains(stdout, "OK") == false {
		testRun.Fatalf("stdout %q stderr %q", stdout, stderr)
	}
}

func TestNotifyTestFailureIsClearAndMasked(testRun *testing.T) {
	fixture := newWatchFixture()
	fixture.notifier.failuresLeft = 1
	path := writeWatchConfig(testRun, enabledConfig)
	code, stdout, stderr := executeNotifyTest(fixture, "--config", path)
	if code != ExitRuntimeError || !strings.Contains(stderr, "FAILED") {
		testRun.Fatalf("code=%d stderr=%q", code, stderr)
	}
	if strings.Contains(stdout+stderr, testSecretToken) {
		testRun.Fatal("output leaks the token")
	}
}

func TestNotifyTestWithoutCredentialsExplainsHowToConfigure(testRun *testing.T) {
	fixture := newWatchFixture()
	path := writeWatchConfig(testRun, enabledNoSecret)
	code, _, stderr := executeNotifyTest(fixture, "--config", path)
	if code != ExitRuntimeError || !strings.Contains(stderr, configfile.EnvTelegramToken) || len(fixture.notifier.sent) != 0 {
		testRun.Fatalf("code=%d stderr=%q", code, stderr)
	}
}

func TestNotifyTestRejectsBadFlags(testRun *testing.T) {
	fixture := newWatchFixture()
	if code, _, _ := executeNotifyTest(fixture, "--bogus"); code != ExitUsage {
		testRun.Fatalf("code=%d", code)
	}
}
