package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/Byroncl/exposedwatch/internal/adapters/configfile"
)

// runNotifyTest executes the notify-test subcommand.
func runNotifyTest(arguments []string, stdout, stderr io.Writer) ExitCode {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return runNotifyTestWith(ctx, systemRuntime(), arguments, stdout, stderr)
}

// runNotifyTestWith sends one test alert. It needs the token and the chat
// id but not the "enabled" switch, so a setup can be checked before the
// watch command is turned on. It prints the masked token and where it came
// from, never the token itself.
func runNotifyTestWith(ctx context.Context, runtime commandRuntime, arguments []string, stdout, stderr io.Writer) ExitCode {
	options, err := parseNotifyTestOptions(arguments, stderr)
	if err != nil {
		return exitCodeForParseError(err, stderr)
	}
	loaded, err := runtime.loadConfiguration(options.ConfigPath)
	if err != nil {
		fmt.Fprintf(stderr, configFailedFormat+"\n", err)
		return ExitRuntimeError
	}
	settings := loaded.Effective.Telegram
	if settings.BotToken.IsEmpty() || settings.ChatID == "" {
		fmt.Fprintf(stderr, notifyNotConfiguredFormat, loaded.Path, configfile.EnvTelegramToken, configfile.EnvTelegramChatID)
		return ExitRuntimeError
	}
	fmt.Fprintf(stdout, notifySendingFormat, configfile.MaskToken(settings.BotToken), loaded.TokenSource, loaded.ChatSource)

	testAlert := newTestAlert(runtime.hostNameOrUnknown())
	if err := runtime.NewNotifier(settings.BotToken, settings.ChatID).Send(ctx, testAlert); err != nil {
		// The adapter already redacts; masking again is defence in depth.
		fmt.Fprintf(stderr, notifyFailedFormat, maskSecret(err.Error(), settings.BotToken))
		return ExitRuntimeError
	}
	fmt.Fprintln(stdout, notifySucceededText)
	return ExitOK
}
