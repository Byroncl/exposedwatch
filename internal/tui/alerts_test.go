package tui

import (
	"strings"
	"testing"

	"github.com/Byroncl/exposedwatch/internal/app"
)

func alertsModel(testRun *testing.T, settings AlertSettings, inner InnerTab) Model {
	testRun.Helper()
	model := settingsModel(testRun, &fakeSettingsStore{settings: settings}, &fakeAlertTester{})
	return showView(testRun, model, moduleView{ModuleAlerts, inner})
}

func TestAlertsStatusShowsConfigurationMaskedAndTheHonestNote(testRun *testing.T) {
	rendered := alertsModel(testRun, configuredSettings(), 0).View()
	for _, want := range []string{"configured", testMaskedToken, "enabled", "high", "60", testConfigPath, "does not send alerts itself", "exposedwatch watch", "none this session"} {
		if !strings.Contains(strings.Join(strings.Fields(rendered), " "), want) {
			testRun.Errorf("status lacks %q:\n%s", want, rendered)
		}
	}
	assertNoSecret(testRun, rendered, "alerts status")
	if strings.Contains(rendered, testChatIdentifier) {
		testRun.Error("the chat id is only reported as present")
	}
}

func TestAlertsStatusExplainsAMissingConfigurationAndADisabledSwitch(testRun *testing.T) {
	missing := alertsModel(testRun, unconfiguredSettings(), 0).View()
	if !strings.Contains(missing, "not configured") || !strings.Contains(missing, "disabled") {
		testRun.Errorf("unconfigured status:\n%s", missing)
	}
	disabled := configuredSettings()
	disabled.Enabled = false
	if rendered := alertsModel(testRun, disabled, 0).View(); !strings.Contains(rendered, "disabled") || !strings.Contains(rendered, "configured") {
		testRun.Errorf("disabled status:\n%s", rendered)
	}
}

func TestAlertsStatusShowsWarningsAndTheLoadError(testRun *testing.T) {
	settings := configuredSettings()
	settings.Warnings = []string{"directory /x can be written by group or others"}
	if rendered := alertsModel(testRun, settings, 0).View(); !strings.Contains(rendered, "can be written by group") {
		testRun.Errorf("warnings missing:\n%s", rendered)
	}
	failed := showView(testRun, settingsModel(testRun, &fakeSettingsStore{loadErr: errFakeUntrusted()}, nil), moduleView{ModuleAlerts, 0})
	if rendered := failed.View(); !strings.Contains(rendered, "not trusted") {
		testRun.Errorf("load error missing:\n%s", rendered)
	}
}

func TestAlertsStatusReportsTheLastTestResultOfTheSession(testRun *testing.T) {
	tester := &fakeAlertTester{}
	model := settingsModel(testRun, &fakeSettingsStore{settings: configuredSettings()}, tester)
	model = showView(testRun, model, moduleView{ModuleSettings, 0})
	model = press(testRun, moveToField(testRun, model, FieldTest), "enter")
	model = applyMessage(testRun, model, AlertTestResultMessage{At: testRefreshTime})
	status := showView(testRun, model, moduleView{ModuleAlerts, 0}).View()
	if !strings.Contains(status, "03:04:05") || !strings.Contains(status, "sent") {
		testRun.Errorf("last test missing:\n%s", status)
	}
	failing := applyMessage(testRun, model, AlertTestResultMessage{Err: errFakeSaveFailure, At: timeAfter(1)})
	if rendered := showView(testRun, failing, moduleView{ModuleAlerts, 0}).View(); !strings.Contains(rendered, "failed") || !strings.Contains(rendered, "03:05:05") {
		testRun.Errorf("failed test missing:\n%s", rendered)
	}
}

func TestAlertsPreviewListsFindingsAtOrAboveTheMinimumSeverity(testRun *testing.T) {
	rendered := alertsModel(testRun, configuredSettings(), 1).View()
	for _, want := range []string{"Would alert now (3)", "CRITICAL", "3306/tcp", "mysqld", "Title for mysqld", "22/tcp", "sshd", "443/tcp", "nginx",
		"one summary at start", "only new findings"} {
		if !strings.Contains(strings.Join(strings.Fields(rendered), " "), want) {
			testRun.Errorf("preview lacks %q:\n%s", want, rendered)
		}
	}
	for _, hidden := range []string{"8080/tcp", "53/tcp"} {
		if strings.Contains(rendered, hidden) {
			testRun.Errorf("preview must hide findings below the minimum severity: %q", hidden)
		}
	}
}

func TestAlertsPreviewHasHonestEmptyStates(testRun *testing.T) {
	critical := configuredSettings()
	critical.MinSeverity = "critical"
	empty := alertsModel(testRun, critical, 1)
	empty = applyMessage(testRun, empty, RefreshResultMessage{Report: app.Report{}, At: testRefreshTime})
	if rendered := empty.View(); !strings.Contains(rendered, "No findings at or above critical") {
		testRun.Errorf("no-findings state:\n%s", rendered)
	}
	disabled := configuredSettings()
	disabled.Enabled = false
	if rendered := alertsModel(testRun, disabled, 1).View(); !strings.Contains(rendered, "Alerts are disabled") || !strings.Contains(rendered, "Would alert if enabled") {
		testRun.Errorf("disabled state:\n%s", rendered)
	}
	if rendered := alertsModel(testRun, unconfiguredSettings(), 1).View(); !strings.Contains(rendered, "Not configured") {
		testRun.Errorf("unconfigured state:\n%s", rendered)
	}
	noStore := showView(testRun, newTestModel(testRun), moduleView{ModuleAlerts, 1})
	if rendered := noStore.View(); !strings.Contains(rendered, "not available") {
		testRun.Errorf("no store:\n%s", rendered)
	}
}

func TestAlertsPreviewCollapsesRowsThatDoNotFit(testRun *testing.T) {
	model := sizedModel(testRun, alertsModel(testRun, configuredSettings(), 1), 110, 12)
	rendered := model.View()
	if !strings.Contains(rendered, "more") {
		testRun.Errorf("hidden rows are counted:\n%s", rendered)
	}
	assertRendersWithin(testRun, model, 110, 12)
}

func TestSidebarBadgeOfAlerts(testRun *testing.T) {
	disabled := configuredSettings()
	disabled.Enabled = false
	cases := []struct {
		name  string
		model Model
		want  string
	}{
		{"enabled and configured", settingsModel(testRun, &fakeSettingsStore{settings: configuredSettings()}, nil), "on"},
		{"disabled", settingsModel(testRun, &fakeSettingsStore{settings: disabled}, nil), "off"},
		{"not configured", settingsModel(testRun, &fakeSettingsStore{settings: unconfiguredSettings()}, nil), "off"},
		{"load error", settingsModel(testRun, &fakeSettingsStore{loadErr: errFakeUntrusted()}, nil), "n/a"},
		{"no store", newTestModel(testRun), "off"},
	}
	for _, testCase := range cases {
		if got := testCase.model.sidebarBadgeFor(ModuleAlerts); got.text != testCase.want {
			testRun.Errorf("%s: badge %+v, want %q", testCase.name, got, testCase.want)
		}
	}
	pending := New(Options{Scanner: fakeScanner{}, Settings: &fakeSettingsStore{settings: configuredSettings()}})
	if got := pending.sidebarBadgeFor(ModuleAlerts); got.text != "" {
		testRun.Errorf("before the first load the badge is empty, got %+v", got)
	}
}

func TestSidebarBadgeFollowsTheSavedValuesNotTheUnsavedForm(testRun *testing.T) {
	model := settingsModel(testRun, &fakeSettingsStore{settings: configuredSettings()}, nil)
	model = press(testRun, showView(testRun, model, moduleView{ModuleSettings, 0}), "enter")
	if model.sidebarBadgeFor(ModuleAlerts).text != "on" {
		testRun.Fatal("unsaved edits do not change what watch will read")
	}
}

func TestOverviewShowsTheAlertsLine(testRun *testing.T) {
	cases := []struct {
		name     string
		model    Model
		expected string
	}{
		{"on", settingsModel(testRun, &fakeSettingsStore{settings: configuredSettings()}, nil), "Alerts: on (min high)"},
		{"off", settingsModel(testRun, &fakeSettingsStore{settings: unconfiguredSettings()}, nil), "Alerts: off (min high)"},
		{"load error", settingsModel(testRun, &fakeSettingsStore{loadErr: errFakeUntrusted()}, nil), "Alerts: n/a"},
	}
	for _, testCase := range cases {
		rendered := showModule(testRun, testCase.model, ModuleOverview).View()
		if !strings.Contains(rendered, testCase.expected) {
			testRun.Errorf("%s: overview lacks %q:\n%s", testCase.name, testCase.expected, rendered)
		}
	}
}

func TestAlertsFooterHidesFormKeys(testRun *testing.T) {
	footer := strings.Join(alertsModel(testRun, configuredSettings(), 0).footerHintTexts(testWidth), footerHintJoin)
	for _, hidden := range []string{"j/k", "enter", "s sort"} {
		if strings.Contains(footer, hidden) {
			testRun.Errorf("alerts footer must hide %q: %q", hidden, footer)
		}
	}
}
