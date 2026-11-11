package tui

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

var secretSweepSizes = []tea.WindowSizeMsg{
	{Width: 120, Height: 40}, {Width: 110, Height: 28}, {Width: 70, Height: 20}, {Width: 50, Height: 16}, {Width: 30, Height: 6}, {Width: 1, Height: 1},
}

// assertNoSecret fails when the rendered text holds the token or a fragment.
func assertNoSecret(testRun *testing.T, rendered, label string) {
	testRun.Helper()
	if fragment, leaked := containsSecretFragment(rendered); leaked {
		testRun.Errorf("%s leaks %q:\n%s", label, fragment, rendered)
	}
}

// secretModels returns loaded models that hold the token in different places:
// stored in the file, from the environment and being typed into the editor.
func secretModels(testRun *testing.T) map[string]Model {
	testRun.Helper()
	tester := &fakeAlertTester{err: errors.New("send failed for " + testSecretToken)}
	fileModel := telegramModel(testRun, &fakeSettingsStore{settings: configuredSettings(), saveErr: errors.New("cannot write " + testSecretToken)}, tester)
	environment := configuredSettings()
	environment.TokenSource = SettingSourceEnvironment
	environment.Warnings = []string{"watch out " + testSecretToken}
	typing := moveToField(testRun, telegramModel(testRun, &fakeSettingsStore{settings: unconfiguredSettings()}, tester), FieldToken)
	typing = press(testRun, typing, "enter", testSecretToken)
	return map[string]Model{
		"file":                fileModel,
		"environment":         telegramModel(testRun, &fakeSettingsStore{settings: environment}, tester),
		"typing":              typing,
		"after save and test": press(testRun, moveToField(testRun, press(testRun, moveToField(testRun, fileModel, FieldSave), "enter"), FieldTest), "enter"),
		"load error":          telegramModel(testRun, &fakeSettingsStore{loadErr: errors.New("config file could not be read")}, tester),
	}
}

func TestNoViewOfAnyModuleTabHelpOrPromptContainsTheToken(testRun *testing.T) {
	for name, base := range secretModels(testRun) {
		for _, size := range secretSweepSizes {
			sized := applyMessage(testRun, base, size)
			assertNoSecret(testRun, sized.View(), fmt.Sprintf("%s %dx%d", name, size.Width, size.Height))
			assertNoSecret(testRun, sized.footerLine(size.Width), name+" footer")
			if sized.viewState == ViewFieldEdit {
				// Global keys are text while editing; only the view is checked.
				continue
			}
			assertNoSecret(testRun, press(testRun, sized, "?").View(), name+" help")
			for _, target := range everyModuleView() {
				assertNoSecret(testRun, showView(testRun, sized, target).View(), fmt.Sprintf("%s %v/%d %dx%d", name, target.module, target.inner, size.Width, size.Height))
			}
		}
	}
}

func TestFormatOfTheModelNeverContainsTheToken(testRun *testing.T) {
	for name, model := range secretModels(testRun) {
		for _, verb := range []string{"%v", "%+v", "%#v", "%s", "%q", "%x", "%d"} {
			assertNoSecret(testRun, fmt.Sprintf(verb, model), name+" "+verb)
			assertNoSecret(testRun, fmt.Sprintf(verb, &model), name+" pointer "+verb)
			assertNoSecret(testRun, fmt.Sprintf(verb, model.settings), name+" settings "+verb)
			assertNoSecret(testRun, fmt.Sprintf(verb, model.settings.form), name+" form "+verb)
			assertNoSecret(testRun, fmt.Sprintf(verb, model.settings.editor), name+" editor "+verb)
		}
	}
}

func TestFormatOfTheLoadedSettingsMessageNeverContainsTheToken(testRun *testing.T) {
	message := SettingsLoadedMessage{Settings: configuredSettings()}
	for _, verb := range []string{"%v", "%+v", "%#v"} {
		assertNoSecret(testRun, fmt.Sprintf(verb, message), "message "+verb)
		assertNoSecret(testRun, fmt.Sprintf(verb, configuredSettings()), "settings "+verb)
	}
}

func TestTheMaskedTokenIsShownWhenNotEditing(testRun *testing.T) {
	model := telegramModel(testRun, &fakeSettingsStore{settings: configuredSettings()}, nil)
	if !strings.Contains(model.View(), testMaskedToken) {
		testRun.Fatalf("masked token missing:\n%s", model.View())
	}
}

func TestEditingTheTokenShowsOneBulletPerRune(testRun *testing.T) {
	model := moveToField(testRun, telegramModel(testRun, &fakeSettingsStore{settings: unconfiguredSettings()}, nil), FieldToken)
	model = press(testRun, model, "enter", "a", "é", "日")
	if rendered := model.View(); !strings.Contains(rendered, "•••") || strings.Contains(rendered, "é") || strings.Contains(rendered, "日") {
		testRun.Fatalf("bullets only while editing a secret:\n%s", rendered)
	}
}

func TestMaskTokenKeepsOnlyThePublicBotIdentifier(testRun *testing.T) {
	cases := map[string]string{"": "", "no-colon-secret": "****", testSecretToken: testMaskedToken}
	for token, want := range cases {
		if got := maskToken(domainToken(token)); got != want {
			testRun.Errorf("maskToken(%q) = %q, want %q", token, got, want)
		}
	}
}

func TestRedactSecretsReplacesEveryOccurrence(testRun *testing.T) {
	text := redactSecrets("a "+testSecretToken+" b "+testSecretToken, domainToken(testSecretToken))
	if strings.Contains(text, testSecretPart) || strings.Count(text, redactedText) != 2 {
		testRun.Fatalf("redacted = %q", text)
	}
}
