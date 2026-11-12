package tui

import (
	"errors"
	"strings"
	"testing"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

func TestSaveSendsTheFormValuesAndReportsThePathAndTheRestartReminder(testRun *testing.T) {
	store := &fakeSettingsStore{settings: configuredSettings()}
	model := telegramModel(testRun, store, nil)
	model = press(testRun, model, "enter") // disable alerts
	model = moveToField(testRun, model, FieldSave)
	model = press(testRun, model, "enter")
	if len(store.saved) != 1 {
		testRun.Fatalf("saves = %d", len(store.saved))
	}
	saved := store.saved[0]
	if saved.Enabled || saved.Token.Reveal() != testSecretToken || saved.ChatID != testChatIdentifier ||
		saved.MinSeverity != "high" || saved.CooldownMinutes != 60 || saved.IntervalSeconds != 60 {
		testRun.Fatalf("saved = %+v", saved)
	}
	rendered := model.View()
	for _, want := range []string{"Saved to " + testConfigPath, "restart watch to apply"} {
		if !strings.Contains(strings.Join(strings.Fields(rendered), " "), want) {
			testRun.Errorf("status lacks %q:\n%s", want, rendered)
		}
	}
	if strings.Contains(rendered, "unsaved changes") {
		testRun.Fatal("a successful save clears the marker")
	}
	if store.loadCalls < 2 {
		testRun.Fatal("a successful save reloads the metadata (permissions warnings)")
	}
}

func TestSaveFailureShowsTheErrorAndKeepsTheUnsavedMarker(testRun *testing.T) {
	store := &fakeSettingsStore{settings: configuredSettings(), saveErr: errFakeSaveFailure}
	model := press(testRun, telegramModel(testRun, store, nil), "enter")
	model = press(testRun, moveToField(testRun, model, FieldSave), "enter")
	rendered := model.View()
	if !strings.Contains(rendered, "the disk is full") || !strings.Contains(rendered, "unsaved changes") {
		testRun.Fatalf("failure feedback:\n%s", rendered)
	}
	if strings.Contains(rendered, "Saved to") {
		testRun.Fatal("a failed save must not claim success")
	}
}

func TestValidationErrorsFromTheStoreAreShownAndNothingIsClaimedSaved(testRun *testing.T) {
	validation := errors.New("invalid config: alerts.cooldown_minutes must be between 0 and 10080")
	store := &fakeSettingsStore{settings: configuredSettings(), saveErr: validation}
	model := moveToField(testRun, telegramModel(testRun, store, nil), FieldCooldown)
	model = press(testRun, model, "enter", "backspace", "backspace", "9", "9", "9", "9", "9", "9", "enter")
	model = press(testRun, moveToField(testRun, model, FieldSave), "enter")
	if rendered := strings.Join(strings.Fields(model.View()), " "); !strings.Contains(rendered, "cooldown_minutes must be between 0 and 10080") {
		testRun.Fatalf("validation error missing:\n%s", rendered)
	}
	if store.saved[0].CooldownMinutes != 999999 {
		testRun.Fatalf("the TUI must not clamp or validate bounds itself: %+v", store.saved[0])
	}
}

func TestLoadWarningsAreShownInTheStatusArea(testRun *testing.T) {
	settings := configuredSettings()
	settings.Warnings = []string{"config file /x is accessible by other users (mode 0644); saving it again restricts it to its owner"}
	model := telegramModel(testRun, &fakeSettingsStore{settings: settings}, nil)
	if rendered := strings.Join(strings.Fields(model.View()), " "); !strings.Contains(rendered, "accessible by other users") {
		testRun.Fatalf("warnings missing:\n%s", rendered)
	}
}

func TestOnlyOneSaveRunsAtATime(testRun *testing.T) {
	store := &fakeSettingsStore{settings: configuredSettings()}
	model := press(testRun, telegramModel(testRun, store, nil), "enter")
	model = moveToField(testRun, model, FieldSave)
	model, command := pressWithoutRunning(testRun, model, "enter")
	if command == nil || !model.settings.saving {
		testRun.Fatal("save starts a command")
	}
	_, second := pressWithoutRunning(testRun, model, "enter")
	if second != nil {
		testRun.Fatal("a second save while one is running must be ignored")
	}
}

func TestSaveWithoutChangesIsANoOpWithAHint(testRun *testing.T) {
	store := &fakeSettingsStore{settings: configuredSettings()}
	model := moveToField(testRun, telegramModel(testRun, store, nil), FieldSave)
	model = press(testRun, model, "enter", "enter", "enter")
	if len(store.saved) != 0 || !strings.Contains(model.View(), "No changes to save") {
		testRun.Fatalf("saves=%d\n%s", len(store.saved), model.View())
	}
}

func TestEnvironmentSourcedValuesAreHandedToTheStoreAsSources(testRun *testing.T) {
	settings := configuredSettings()
	settings.TokenSource = SettingSourceEnvironment
	store := &fakeSettingsStore{settings: settings}
	model := press(testRun, telegramModel(testRun, store, nil), "enter")
	model = press(testRun, moveToField(testRun, model, FieldSave), "enter")
	if len(store.saved) != 1 || store.saved[0].TokenSource != SettingSourceEnvironment {
		testRun.Fatalf("the store must see that the token is environment-sourced: %+v", store.saved)
	}
	_ = model
}

func TestSeverityIsSavedAsItsLowercaseName(testRun *testing.T) {
	store := &fakeSettingsStore{settings: configuredSettings()}
	model := press(testRun, moveToField(testRun, telegramModel(testRun, store, nil), FieldMinSeverity), "enter")
	model = press(testRun, moveToField(testRun, model, FieldSave), "enter")
	if store.saved[0].MinSeverity != domain.Critical.String() {
		testRun.Fatalf("severity = %q", store.saved[0].MinSeverity)
	}
}

func TestSavingErrorsNeverEchoTheToken(testRun *testing.T) {
	leaky := errors.New("write failed for " + testSecretToken)
	store := &fakeSettingsStore{settings: configuredSettings(), saveErr: leaky}
	model := press(testRun, telegramModel(testRun, store, nil), "enter")
	model = press(testRun, moveToField(testRun, model, FieldSave), "enter")
	if fragment, leaked := containsSecretFragment(model.View()); leaked {
		testRun.Fatalf("save error leaked %q:\n%s", fragment, model.View())
	}
}

func TestAFailedRefreshAfterSavingIsShownAsAWarning(testRun *testing.T) {
	store := &fakeSettingsStore{settings: configuredSettings()}
	model := press(testRun, telegramModel(testRun, store, nil), "enter")
	model = moveToField(testRun, model, FieldSave)
	model, command := pressWithoutRunning(testRun, model, "enter")
	model, refresh := applyWithCommand(testRun, model, command())
	if refresh == nil {
		testRun.Fatal("a successful save re-reads the file")
	}
	store.loadErr = errFakeUntrusted()
	model = settle(testRun, model, refresh)
	rendered := strings.Join(strings.Fields(model.View()), " ")
	if !strings.Contains(rendered, "Saved to") || !strings.Contains(rendered, "not trusted") {
		testRun.Fatalf("refresh failure hidden:\n%s", rendered)
	}
}
