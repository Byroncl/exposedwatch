package main

import (
	"context"
	"errors"

	"github.com/Byroncl/exposedwatch/internal/domain"
	"github.com/Byroncl/exposedwatch/internal/tui"
)

// telegramAlertTester implements tui.AlertTester with the same notifier and
// test alert as the notify-test command.
type telegramAlertTester struct {
	runtime commandRuntime
}

// newTelegramAlertTester builds the tester over the runtime's notifier factory.
func newTelegramAlertTester(runtime commandRuntime) tui.AlertTester {
	return telegramAlertTester{runtime: runtime}
}

// SendTest sends one test alert with the given settings, saved or not. It
// needs the token and the chat id but not the enabled switch, like notify-test.
func (tester telegramAlertTester) SendTest(ctx context.Context, settings tui.AlertSettings) error {
	if settings.Token.IsEmpty() || settings.ChatID == "" {
		return errors.New(testMessageNeedsCredentials)
	}
	ctx, cancel := context.WithTimeout(ctx, alertTestTimeout)
	defer cancel()
	notifier := tester.runtime.NewNotifier(settings.Token, domain.ChatID(settings.ChatID))
	if err := notifier.Send(ctx, newTestAlert(tester.runtime.hostNameOrUnknown())); err != nil {
		// The adapter already redacts; masking again is defence in depth.
		return sanitizedError(err, settings.Token)
	}
	return nil
}
