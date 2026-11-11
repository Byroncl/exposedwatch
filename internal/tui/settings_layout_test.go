package tui

import (
	"errors"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

const longTestSentence = "something went wrong while talking to the service and this explanation is intentionally much longer than any panel"

// layoutSettingsModels are loaded models in the states that render the most text.
func layoutSettingsModels(testRun *testing.T) map[string]Model {
	testRun.Helper()
	environment := configuredSettings()
	environment.TokenSource = SettingSourceEnvironment
	environment.ChatSource = SettingSourceEnvironment
	environment.Warnings = []string{longTestSentence, longTestSentence, longTestSentence}
	tester := &fakeAlertTester{err: errors.New(longTestSentence + longTestSentence)}
	failedTest := press(testRun, moveToField(testRun, telegramModel(testRun, &fakeSettingsStore{settings: configuredSettings(), saveErr: errors.New(longTestSentence)}, tester), FieldSave), "enter")
	failedTest = press(testRun, moveToField(testRun, failedTest, FieldTest), "enter")
	chatEditing := press(testRun, moveToField(testRun, telegramModel(testRun, &fakeSettingsStore{settings: configuredSettings()}, tester), FieldChatID), "enter", longTestSentence)
	tokenEditing := press(testRun, moveToField(testRun, telegramModel(testRun, &fakeSettingsStore{settings: unconfiguredSettings()}, tester), FieldToken), "enter", longTestSentence)
	return map[string]Model{
		"configured":   telegramModel(testRun, &fakeSettingsStore{settings: configuredSettings()}, tester),
		"unconfigured": telegramModel(testRun, &fakeSettingsStore{settings: unconfiguredSettings()}, nil),
		"environment":  telegramModel(testRun, &fakeSettingsStore{settings: environment}, tester),
		"load error":   telegramModel(testRun, &fakeSettingsStore{loadErr: errors.New(longTestSentence)}, tester),
		"failed test":  failedTest,
		"chat editing": chatEditing,
		"token typing": tokenEditing,
	}
}

func TestSettingsAndAlertsViewsStayInsideEverySizeInEveryState(testRun *testing.T) {
	for _, base := range layoutSettingsModels(testRun) {
		for _, size := range layoutSizes {
			sized := applyMessage(testRun, base, size)
			assertRendersWithin(testRun, sized, size.Width, size.Height)
			if sized.viewState == ViewFieldEdit {
				continue
			}
			for _, target := range []moduleView{{ModuleSettings, 0}, {ModuleSettings, 1}, {ModuleAlerts, 0}, {ModuleAlerts, 1}, {ModuleOverview, 0}} {
				shown := showView(testRun, sized, target)
				assertRendersWithin(testRun, shown, size.Width, size.Height)
				assertRendersWithin(testRun, press(testRun, shown, "?"), size.Width, size.Height)
				assertRendersWithin(testRun, press(testRun, shown, "b"), size.Width, size.Height)
			}
			if !sized.settings.isEditable() {
				continue
			}
			editing := press(testRun, moveToField(testRun, showView(testRun, sized, moduleView{ModuleSettings, 0}), FieldChatID), "enter", "x")
			assertRendersWithin(testRun, editing, size.Width, size.Height)
		}
	}
}

// TestTheSelectedFormRowStaysVisibleOnShortTerminals checks the scroll window.
func TestTheSelectedFormRowStaysVisibleOnShortTerminals(testRun *testing.T) {
	base := telegramModel(testRun, &fakeSettingsStore{settings: configuredSettings()}, &fakeAlertTester{})
	for _, height := range []int{9, 10, 12, 14, 18} {
		sized := sizedModel(testRun, base, 100, height)
		for field := FieldEnabled; field < formFieldCount; field++ {
			shown := moveToField(testRun, sized, field)
			if !strings.Contains(shown.View(), shown.settings.cursorLabelForTest()) {
				testRun.Errorf("height %d: the %v row is scrolled out of view:\n%s", height, field, shown.View())
			}
		}
	}
}

// TestFormValuesStartAtTheSameColumn checks the label and value alignment.
func TestFormValuesStartAtTheSameColumn(testRun *testing.T) {
	model := sizedModel(testRun, telegramModel(testRun, &fakeSettingsStore{settings: configuredSettings()}, nil), 110, 28)
	values := map[FormField]string{
		FieldEnabled: "[x]", FieldToken: testMaskedToken, FieldChatID: testChatIdentifier,
		FieldMinSeverity: "high", FieldCooldown: "60", FieldInterval: "60",
	}
	lines := strings.Split(model.View(), "\n")
	found := 0
	for field, want := range values {
		for _, line := range lines {
			index := strings.Index(line, field.label())
			if index < 0 {
				continue
			}
			found++
			rest := line[index:]
			valueIndex := strings.Index(rest, want)
			if valueIndex < 0 || lipgloss.Width(rest[:valueIndex]) != formLabelWidth+len(columnSeparator) {
				testRun.Errorf("%q value %q does not start %d cells after the label: %q", field.label(), want, formLabelWidth+len(columnSeparator), line)
			}
		}
	}
	if found != len(values) {
		testRun.Fatalf("found %d of %d rows:\n%s", found, len(values), model.View())
	}
}

const (
	hostileText    = "\x1b]52;c;QUJD\x07evil"
	escapeByte     = "\x1b"
	bellByte       = "\x07"
	hostilePathTxt = "/tmp/\x1b]0;pwn\x07/config.json"
)

func TestNoEscapeSequenceFromConfigValuesReachesAnyView(testRun *testing.T) {
	file := configuredSettings()
	file.ChatID = hostileText
	file.Path = hostilePathTxt
	file.Warnings = []string{"warning " + hostileText}
	environment := file
	environment.ChatSource = SettingSourceEnvironment
	environment.ChatEnvironmentName = "NAME" + hostileText
	environment.TokenEnvironmentName = "NAME" + hostileText
	environment.TokenSource = SettingSourceEnvironment
	for name, settings := range map[string]AlertSettings{"file": file, "environment": environment} {
		base := settingsModel(testRun, &fakeSettingsStore{settings: settings}, &fakeAlertTester{})
		base = press(testRun, base, "enter")
		for _, size := range secretSweepSizes {
			sized := applyMessage(testRun, base, size)
			for _, target := range everyModuleView() {
				shown := showView(testRun, sized, target)
				for _, rendered := range []string{shown.View(), press(testRun, shown, "?").View()} {
					if strings.Contains(rendered, escapeByte) || strings.Contains(rendered, bellByte) {
						testRun.Errorf("%s %v/%d %dx%d leaks a control byte:\n%q", name, target.module, target.inner, size.Width, size.Height, rendered)
					}
				}
			}
			saved := press(testRun, moveToField(testRun, showView(testRun, sized, moduleView{ModuleSettings, 0}), FieldSave), "enter")
			if strings.Contains(saved.View(), escapeByte) {
				testRun.Errorf("%s: status leaks an escape byte", name)
			}
		}
	}
}
