package tui

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestSendTestMessageLifecycleShowsSendingThenSent(testRun *testing.T) {
	tester := &fakeAlertTester{}
	model := moveToField(testRun, telegramModel(testRun, &fakeSettingsStore{settings: configuredSettings()}, tester), FieldTest)
	sending, command := pressWithoutRunning(testRun, model, "enter")
	if command == nil || !strings.Contains(sending.View(), "Sending...") {
		testRun.Fatalf("pressing enter shows Sending... and returns a command:\n%s", sending.View())
	}
	done := settle(testRun, sending, command)
	if len(tester.calls) != 1 || !strings.Contains(done.View(), "Sent") || strings.Contains(done.View(), "Sending...") {
		testRun.Fatalf("calls=%d\n%s", len(tester.calls), done.View())
	}
	if done.settings.testMessage.phase != testPhaseSucceeded {
		testRun.Fatalf("phase = %v", done.settings.testMessage.phase)
	}
}

func TestSendTestMessageFailureShowsTheRedactedError(testRun *testing.T) {
	tester := &fakeAlertTester{err: errors.New("telegram rejected the request (401 Unauthorized)")}
	model := press(testRun, moveToField(testRun, telegramModel(testRun, &fakeSettingsStore{settings: configuredSettings()}, tester), FieldTest), "enter")
	if rendered := model.View(); !strings.Contains(rendered, "Failed") || !strings.Contains(rendered, "401 Unauthorized") {
		testRun.Fatalf("failure feedback:\n%s", rendered)
	}
}

func TestOnlyOneTestMessageIsInFlight(testRun *testing.T) {
	tester := &fakeAlertTester{}
	model := moveToField(testRun, telegramModel(testRun, &fakeSettingsStore{settings: configuredSettings()}, tester), FieldTest)
	model, first := pressWithoutRunning(testRun, model, "enter")
	_, second := pressWithoutRunning(testRun, model, "enter")
	if first == nil || second != nil {
		testRun.Fatalf("first=%v second=%v: a second test must be ignored while one is in flight", first != nil, second != nil)
	}
}

func TestTestMessageUsesTheUnsavedFormValues(testRun *testing.T) {
	tester := &fakeAlertTester{}
	model := moveToField(testRun, telegramModel(testRun, &fakeSettingsStore{settings: configuredSettings()}, tester), FieldChatID)
	model = press(testRun, model, "enter", "backspace", "7", "enter")
	model = press(testRun, moveToField(testRun, model, FieldTest), "enter")
	if len(tester.calls) != 1 || tester.calls[0].ChatID != testChatIdentifier[:len(testChatIdentifier)-1]+"7" {
		testRun.Fatalf("calls = %+v", tester.calls)
	}
	if tester.calls[0].Token.Reveal() != testSecretToken {
		testRun.Fatal("the test uses the form token")
	}
	if !strings.Contains(model.View(), "unsaved changes") {
		testRun.Fatal("sending a test must not save")
	}
}

func TestTestMessageRequiresATokenAndAChatIdentifier(testRun *testing.T) {
	tester := &fakeAlertTester{}
	model := moveToField(testRun, telegramModel(testRun, &fakeSettingsStore{settings: unconfiguredSettings()}, tester), FieldTest)
	model, command := pressWithoutRunning(testRun, model, "enter")
	if command != nil || len(tester.calls) != 0 {
		testRun.Fatal("nothing may be sent without credentials")
	}
	if rendered := strings.Join(strings.Fields(model.View()), " "); !strings.Contains(rendered, "needs a bot token and a chat ID") {
		testRun.Fatalf("hint missing:\n%s", rendered)
	}
}

func TestTestMessageWithoutATesterIsUnavailable(testRun *testing.T) {
	model := moveToField(testRun, telegramModel(testRun, &fakeSettingsStore{settings: configuredSettings()}, nil), FieldTest)
	model, command := pressWithoutRunning(testRun, model, "enter")
	if command != nil || !strings.Contains(model.View(), "not available") {
		testRun.Fatal("without a tester the action explains itself")
	}
}

func TestTestMessageRunsInItsOwnCommandAndNeverBlocksUpdate(testRun *testing.T) {
	gate := make(chan struct{})
	tester := blockingTester{gate: gate}
	model := New(Options{Scanner: fakeScanner{}, Interval: testInterval, Version: testVersion,
		Settings: &fakeSettingsStore{settings: configuredSettings()}, AlertTester: tester})
	model = applyMessage(testRun, model, settingsLoadCommand(&fakeSettingsStore{settings: configuredSettings()})())
	model = showView(testRun, model, moduleView{ModuleSettings, 0})
	model = moveToField(testRun, model, FieldTest)
	finished := make(chan struct{})
	go func() {
		defer close(finished)
		_, command := applyWithCommand(testRun, model, keyMessage("enter"))
		if command == nil {
			testRun.Error("no command")
		}
	}()
	select {
	case <-finished:
	case <-time.After(time.Second):
		testRun.Fatal("Update blocked on the network call")
	}
	close(gate)
}

type blockingTester struct{ gate chan struct{} }

func (tester blockingTester) SendTest(_ context.Context, _ AlertSettings) error {
	<-tester.gate
	return nil
}

func TestTestFailureNeverEchoesTheToken(testRun *testing.T) {
	tester := &fakeAlertTester{err: errors.New("Post https://api.telegram.org/bot" + testSecretToken + "/sendMessage: timeout")}
	model := press(testRun, moveToField(testRun, telegramModel(testRun, &fakeSettingsStore{settings: configuredSettings()}, tester), FieldTest), "enter")
	if fragment, leaked := containsSecretFragment(model.View()); leaked {
		testRun.Fatalf("test failure leaked %q:\n%s", fragment, model.View())
	}
}

// clockedTelegramModel is a loaded Telegram form whose clock the test moves.
func clockedTelegramModel(testRun *testing.T, tester *fakeAlertTester, now *time.Time) Model {
	testRun.Helper()
	store := &fakeSettingsStore{settings: configuredSettings()}
	model := New(Options{Scanner: fakeScanner{}, Interval: testInterval, Version: testVersion, Settings: store, AlertTester: tester,
		Clock: func() time.Time { return *now }})
	model = applyMessage(testRun, model, tea.WindowSizeMsg{Width: testWidth, Height: testHeight})
	model = applyMessage(testRun, model, settingsLoadCommand(store)())
	return moveToField(testRun, showView(testRun, model, moduleView{ModuleSettings, 0}), FieldTest)
}

func TestTestMessagesAreRateLimitedAfterEachResult(testRun *testing.T) {
	now := testRefreshTime
	tester := &fakeAlertTester{}
	model := clockedTelegramModel(testRun, tester, &now)
	model = press(testRun, model, "enter")
	for range 50 {
		model = press(testRun, model, "enter")
	}
	if len(tester.calls) != 1 {
		testRun.Fatalf("held enter fired %d sends", len(tester.calls))
	}
	if rendered := model.View(); !strings.Contains(rendered, "Wait 5s before sending another test message") {
		testRun.Fatalf("wait hint missing:\n%s", rendered)
	}
	now = now.Add(2 * time.Second)
	if rendered := press(testRun, model, "enter").View(); !strings.Contains(rendered, "Wait 3s before") {
		testRun.Fatalf("remaining time missing:\n%s", rendered)
	}
	now = now.Add(3 * time.Second)
	model = press(testRun, model, "enter")
	if len(tester.calls) != 2 {
		testRun.Fatalf("after the cooldown a new send is allowed, calls %d", len(tester.calls))
	}
}
