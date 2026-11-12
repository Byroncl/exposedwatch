package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

func TestTelegramFormNavigationMovesBetweenRowsAndClampsAtTheEnds(testRun *testing.T) {
	model := telegramModel(testRun, &fakeSettingsStore{settings: configuredSettings()}, nil)
	if model.settings.cursor != FieldEnabled {
		testRun.Fatalf("cursor starts at %v", model.settings.cursor)
	}
	model = press(testRun, model, "k", "up")
	if model.settings.cursor != FieldEnabled {
		testRun.Fatalf("moving above the first row must clamp, cursor %v", model.settings.cursor)
	}
	model = press(testRun, model, "j", "down")
	if model.settings.cursor != FieldChatID {
		testRun.Fatalf("j and down move two rows, cursor %v", model.settings.cursor)
	}
	model = press(testRun, model, "k")
	if model.settings.cursor != FieldToken {
		testRun.Fatalf("k moves up, cursor %v", model.settings.cursor)
	}
	for range int(formFieldCount) + 2 {
		model = press(testRun, model, "j")
	}
	if model.settings.cursor != FieldTest {
		testRun.Fatalf("moving below the last row must clamp, cursor %v", model.settings.cursor)
	}
}

func TestEnabledRowTogglesImmediatelyAndMarksUnsavedChanges(testRun *testing.T) {
	model := telegramModel(testRun, &fakeSettingsStore{settings: configuredSettings()}, nil)
	if strings.Contains(model.View(), "unsaved changes") {
		testRun.Fatal("a freshly loaded form has no unsaved changes")
	}
	toggled := press(testRun, model, "enter")
	if toggled.settings.form.enabled || !strings.Contains(toggled.View(), "unsaved changes") {
		testRun.Fatalf("enter must toggle and show the marker:\n%s", toggled.View())
	}
	back := press(testRun, toggled, "enter")
	if !back.settings.form.enabled || strings.Contains(back.View(), "unsaved changes") {
		testRun.Fatal("toggling back must clear the marker")
	}
}

func TestMinimumSeverityRowCyclesThroughEverySeverity(testRun *testing.T) {
	model := moveToField(testRun, telegramModel(testRun, &fakeSettingsStore{settings: configuredSettings()}, nil), FieldMinSeverity)
	var seen []string
	for range 5 {
		model = press(testRun, model, "enter")
		seen = append(seen, model.settings.form.minSeverity.String())
	}
	if strings.Join(seen, ",") != "critical,info,low,medium,high" {
		testRun.Fatalf("cycle = %v", seen)
	}
}

func TestEditingTheChatIdentifierCommitsWithEnterAndEditsRunes(testRun *testing.T) {
	model := moveToField(testRun, telegramModel(testRun, &fakeSettingsStore{settings: configuredSettings()}, nil), FieldChatID)
	model = press(testRun, model, "enter")
	if model.viewState != ViewFieldEdit {
		testRun.Fatalf("enter on a text row starts editing, state %v", model.viewState)
	}
	model = press(testRun, model, "backspace", "backspace", "x", testUnicodeChatInput)
	if model.settings.editor.text() != "-10012345678x"+testUnicodeChatInput {
		testRun.Fatalf("buffer = %q", model.settings.editor.text())
	}
	model = press(testRun, model, "backspace")
	if !strings.HasSuffix(model.settings.editor.text(), "日") {
		testRun.Fatalf("backspace removes one rune, buffer %q", model.settings.editor.text())
	}
	committed := press(testRun, model, "enter")
	if committed.viewState != ViewTable || committed.settings.form.chatID != "-10012345678x日" {
		testRun.Fatalf("state %v chat %q", committed.viewState, committed.settings.form.chatID)
	}
	if !strings.Contains(committed.View(), "unsaved changes") {
		testRun.Fatal("a committed edit is unsaved")
	}
}

func TestEscapeCancelsAnEditAndRestoresThePreviousValue(testRun *testing.T) {
	model := moveToField(testRun, telegramModel(testRun, &fakeSettingsStore{settings: configuredSettings()}, nil), FieldChatID)
	model = press(testRun, model, "enter", "z", "z", "esc")
	if model.viewState != ViewTable || model.settings.form.chatID != testChatIdentifier {
		testRun.Fatalf("state %v chat %q", model.viewState, model.settings.form.chatID)
	}
	if strings.Contains(model.View(), "unsaved changes") {
		testRun.Fatal("a cancelled edit leaves nothing unsaved")
	}
}

func TestNoGlobalKeyActsWhileEditing(testRun *testing.T) {
	model := moveToField(testRun, telegramModel(testRun, &fakeSettingsStore{settings: configuredSettings()}, nil), FieldChatID)
	model = press(testRun, model, "enter", "backspace", "backspace", "backspace", "backspace", "backspace", "backspace",
		"backspace", "backspace", "backspace", "backspace", "backspace", "backspace", "backspace", "backspace")
	module, tab := model.activeModule, model.activeTab
	for _, key := range []string{"q", "b", "?", "1", "2", "[", "]", "j", "k", "g", "G", "r", "s", "f", "l", "i", "/"} {
		var command tea.Cmd
		model, command = applyWithCommand(testRun, model, keyMessage(key))
		if command != nil {
			testRun.Fatalf("key %q must not produce a command while editing", key)
		}
	}
	for _, special := range []tea.KeyMsg{{Type: tea.KeyTab}, {Type: tea.KeyShiftTab}, {Type: tea.KeyUp}, {Type: tea.KeyDown}} {
		var command tea.Cmd
		model, command = applyWithCommand(testRun, model, special)
		if command != nil {
			testRun.Fatalf("key %v must not produce a command while editing", special)
		}
	}
	if model.viewState != ViewFieldEdit || model.activeModule != module || model.activeTab != tab || model.sidebarHidden {
		testRun.Fatalf("a global key leaked: state %v module %v tab %v", model.viewState, model.activeModule, model.activeTab)
	}
	if model.settings.editor.text() != "qb?12[]jkgGrsfli/" {
		testRun.Fatalf("typed keys must edit the buffer, got %q", model.settings.editor.text())
	}
}

func TestPasteInsertsTextAndDropsControlCharacters(testRun *testing.T) {
	model := moveToField(testRun, telegramModel(testRun, &fakeSettingsStore{settings: unconfiguredSettings()}, nil), FieldToken)
	model = press(testRun, model, "enter")
	pasted := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" " + testSecretToken + "\r\n\t"), Paste: true}
	model, command := applyWithCommand(testRun, model, pasted)
	if command != nil || model.settings.editor.text() != testSecretToken {
		testRun.Fatalf("paste buffer = %q (command %v)", model.settings.editor.text(), command != nil)
	}
	model = press(testRun, model, "enter")
	if model.settings.form.token.Reveal() != testSecretToken {
		testRun.Fatal("the pasted token must be committed")
	}
}

func TestPasteOutsideEditingIsIgnored(testRun *testing.T) {
	model := telegramModel(testRun, &fakeSettingsStore{settings: configuredSettings()}, nil)
	model, command := applyWithCommand(testRun, model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q"), Paste: true})
	if command != nil || model.viewState != ViewTable {
		testRun.Fatal("a paste must never act as a key outside editing")
	}
}

func TestEditingTheTokenStartsEmptyAndAnEmptyCommitKeepsTheToken(testRun *testing.T) {
	model := moveToField(testRun, telegramModel(testRun, &fakeSettingsStore{settings: configuredSettings()}, nil), FieldToken)
	model = press(testRun, model, "enter")
	if model.settings.editor.text() != "" {
		testRun.Fatal("the secret is never copied into the edit buffer")
	}
	model = press(testRun, model, "enter")
	if model.settings.form.token.Reveal() != testSecretToken || strings.Contains(model.View(), "unsaved changes") {
		testRun.Fatal("an empty commit keeps the previous token")
	}
}

func TestNumberRowsAcceptOnlyDigitsAndRejectEmptyValues(testRun *testing.T) {
	model := moveToField(testRun, telegramModel(testRun, &fakeSettingsStore{settings: configuredSettings()}, nil), FieldCooldown)
	model = press(testRun, model, "enter", "backspace", "backspace", "x", "-", ".")
	for _, character := range testDigitsOnlyInput {
		model = press(testRun, model, string(character))
	}
	if model.settings.editor.text() != "42" {
		testRun.Fatalf("only digits are accepted, buffer %q", model.settings.editor.text())
	}
	model = press(testRun, model, "enter")
	if model.settings.form.cooldownMinutes != 42 {
		testRun.Fatalf("cooldown = %d", model.settings.form.cooldownMinutes)
	}
	empty := press(testRun, model, "enter", "backspace", "backspace", "enter")
	if empty.settings.form.cooldownMinutes != 42 || !strings.Contains(empty.View(), "whole number") {
		testRun.Fatalf("an empty number is rejected with a hint:\n%s", empty.View())
	}
}

func TestEditingTheIntervalRow(testRun *testing.T) {
	model := moveToField(testRun, telegramModel(testRun, &fakeSettingsStore{settings: configuredSettings()}, nil), FieldInterval)
	model = press(testRun, model, "enter", "backspace", "backspace", "3", "0", "enter")
	if model.settings.form.intervalSeconds != 30 {
		testRun.Fatalf("interval = %d", model.settings.form.intervalSeconds)
	}
}

func TestCtrlCStillQuitsWhileEditing(testRun *testing.T) {
	model := moveToField(testRun, telegramModel(testRun, &fakeSettingsStore{settings: configuredSettings()}, nil), FieldChatID)
	model = press(testRun, model, "enter")
	_, command := applyWithCommand(testRun, model, keyMessage("ctrl+c"))
	if command == nil {
		testRun.Fatal("ctrl+c must still quit, as in the filter prompt")
	}
}

func TestNonFormKeysStillWorkOnTheTelegramTabWhenNotEditing(testRun *testing.T) {
	model := telegramModel(testRun, &fakeSettingsStore{settings: configuredSettings()}, nil)
	for _, ignored := range []string{"s", "f", "l", "i", "/"} {
		model = press(testRun, model, ignored)
		if model.viewState != ViewTable || model.sortMode != SortBySeverity || model.minSeverity != domain.Info {
			testRun.Fatalf("key %q must be ignored on the form", ignored)
		}
	}
	if model = press(testRun, model, "]"); model.activeTab != TabSettingsDisplay {
		testRun.Fatalf("] moves to the Display tab, got %v", model.activeTab)
	}
	if model = press(testRun, model, "b"); !model.sidebarHidden {
		testRun.Fatal("b still hides the sidebar outside editing")
	}
}

func TestEnvironmentSourcedRowsAreReadOnlyAndExplained(testRun *testing.T) {
	settings := configuredSettings()
	settings.TokenSource = SettingSourceEnvironment
	settings.ChatSource = SettingSourceEnvironment
	model := moveToField(testRun, telegramModel(testRun, &fakeSettingsStore{settings: settings}, nil), FieldToken)
	rendered := model.View()
	for _, want := range []string{"from environment (EXPOSEDWATCH_TELEGRAM_TOKEN)", "from environment (EXPOSEDWATCH_TELEGRAM_CHAT_ID)", testMaskedToken} {
		if !strings.Contains(rendered, want) {
			testRun.Errorf("view lacks %q:\n%s", want, rendered)
		}
	}
	editAttempt := press(testRun, model, "enter")
	if editAttempt.viewState != ViewTable || !strings.Contains(editAttempt.View(), testTokenEnvName) {
		testRun.Fatalf("editing an environment secret is refused with a hint:\n%s", editAttempt.View())
	}
	chatAttempt := press(testRun, moveToField(testRun, model, FieldChatID), "enter")
	if chatAttempt.viewState != ViewTable || !strings.Contains(chatAttempt.View(), testChatEnvName) {
		testRun.Fatal("editing an environment chat id is refused with a hint")
	}
}

func TestAConfigThatCannotBeLoadedKeepsTheFormReadOnlyWithASanitizedError(testRun *testing.T) {
	store := &fakeSettingsStore{loadErr: errFakeUntrusted()}
	model := telegramModel(testRun, store, &fakeAlertTester{})
	rendered := model.View()
	if !strings.Contains(rendered, "config file is not trusted") || !strings.Contains(rendered, "read-only") {
		testRun.Fatalf("load failure must be explained:\n%s", rendered)
	}
	if after := press(testRun, model, "j", "enter"); after.viewState != ViewTable || len(store.saved) != 0 {
		testRun.Fatal("a form that failed to load cannot be edited")
	}
}

func TestSettingsWithoutAStoreAreUnavailable(testRun *testing.T) {
	model := showView(testRun, newTestModel(testRun), moduleView{ModuleSettings, 0})
	if rendered := model.View(); !strings.Contains(rendered, "not available") {
		testRun.Fatalf("no store, no form:\n%s", rendered)
	}
	if after := press(testRun, model, "j", "enter"); after.viewState != ViewTable {
		testRun.Fatal("nothing to edit without a store")
	}
}

func TestDisplayTabKeepsTheReadOnlyValues(testRun *testing.T) {
	model := press(testRun, telegramModel(testRun, &fakeSettingsStore{settings: configuredSettings()}, nil), "2")
	rendered := model.View()
	for _, want := range []string{"Theme", "Refresh interval", "Connections", "Firewall"} {
		if !strings.Contains(rendered, want) {
			testRun.Errorf("display lacks %q:\n%s", want, rendered)
		}
	}
}

func TestHelpListsTheSettingsFormKeys(testRun *testing.T) {
	rendered := press(testRun, telegramModel(testRun, &fakeSettingsStore{settings: configuredSettings()}, nil), "?").View()
	for _, want := range []string{"Settings", "esc"} {
		if !strings.Contains(rendered, want) {
			testRun.Errorf("help lacks %q:\n%s", want, rendered)
		}
	}
}

func TestFooterHintsFollowTheFormApplicability(testRun *testing.T) {
	loaded := telegramModel(testRun, &fakeSettingsStore{settings: configuredSettings()}, nil)
	footer := strings.Join(loaded.footerHintTexts(testWidth), footerHintJoin)
	if !strings.Contains(footer, "j/k move") || !strings.Contains(footer, "enter") {
		testRun.Errorf("form footer lacks its hints: %q", footer)
	}
	display := strings.Join(press(testRun, loaded, "2").footerHintTexts(testWidth), footerHintJoin)
	if strings.Contains(display, "j/k") || strings.Contains(display, "enter") {
		testRun.Errorf("the Display tab has no form keys: %q", display)
	}
	editing := press(testRun, moveToField(testRun, loaded, FieldChatID), "enter")
	line := editing.footerLine(testWidth)
	if !strings.Contains(line, "enter save value") || !strings.Contains(line, "esc cancel") || strings.Contains(line, "q quit") {
		testRun.Errorf("editing footer = %q", line)
	}
}

func TestNumberRowsHintThatOnlyDigitsAreAccepted(testRun *testing.T) {
	model := moveToField(testRun, telegramModel(testRun, &fakeSettingsStore{settings: configuredSettings()}, nil), FieldCooldown)
	model = press(testRun, model, "enter", "-")
	if rendered := model.View(); !strings.Contains(rendered, "digits only") {
		testRun.Fatalf("a rejected character needs a hint:\n%s", rendered)
	}
}

func TestPasteDropsBidiAndZeroWidthCharacters(testRun *testing.T) {
	model := moveToField(testRun, telegramModel(testRun, &fakeSettingsStore{settings: unconfiguredSettings()}, nil), FieldChatID)
	model = press(testRun, model, "enter")
	hostile := "a" + string(rune(0x202e)) + "b" + string(rune(0x202a)) + "c" + string(rune(0x2066)) + "d" + string(rune(0x2069)) + "e" + string(rune(0x200b)) + "f" + string(rune(0x200c)) + "g" + string(rune(0x200d)) + "h" + string(rune(0xfeff)) + "i"
	model, _ = applyWithCommand(testRun, model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(hostile), Paste: true})
	if model.settings.editor.text() != "abcdefghi" {
		testRun.Fatalf("buffer = %q", model.settings.editor.text())
	}
}
