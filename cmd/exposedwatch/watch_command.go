package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Byroncl/exposedwatch/internal/adapters/configfile"
	"github.com/Byroncl/exposedwatch/internal/app"
	"github.com/Byroncl/exposedwatch/internal/domain"
	"github.com/Byroncl/exposedwatch/internal/ports"
)

// runWatch executes the watch subcommand until SIGINT or SIGTERM.
func runWatch(arguments []string, stdout, stderr io.Writer) ExitCode {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return runWatchWith(ctx, systemRuntime(), arguments, stdout, stderr)
}

// runWatchWith is runWatch with an injectable context and runtime.
func runWatchWith(ctx context.Context, runtime commandRuntime, arguments []string, _, stderr io.Writer) ExitCode {
	options, err := parseWatchOptions(arguments, stderr)
	if err != nil {
		return exitCodeForParseError(err, stderr)
	}
	loaded, err := runtime.loadConfiguration(options.ConfigPath)
	if err != nil {
		fmt.Fprintf(stderr, configFailedFormat+"\n", err)
		return ExitRuntimeError
	}
	policy, interval, err := resolveAlertSettings(loaded.Effective.Alerts, options)
	if err != nil {
		fmt.Fprintf(stderr, configFailedFormat+"\n", err)
		return ExitRuntimeError
	}

	token := loaded.Effective.Telegram.BotToken
	logger := watchLogger{writer: stderr, clock: runtime.Clock, secrets: []domain.BotToken{token}}
	for _, warning := range loaded.Warnings {
		logger.Logf(logConfigWarningFormat, warning)
	}
	reloader, acceptance, err := startBaseline(runtime, options, logger, stderr)
	if err != nil {
		return ExitRuntimeError
	}
	loop := &watchLoop{
		scan:      runtime.NewScanner(options.scanOptions()),
		evaluator: app.NewAlertEvaluator(policy, runtime.Clock, runtime.hostNameOrUnknown()),
		outbox: app.NewAlertOutbox(app.OutboxPolicy{
			InitialBackoff: interval, MaxBackoff: maxRetryBackoff, TimeToLive: alertTimeToLive,
		}, runtime.Clock),
		notifier: notifierFor(runtime, loaded.Effective.Telegram),
		logger:   logger,
		interval: interval,
		once:     options.Once,
		wait:     runtime.Wait,
	}
	if reloader != nil {
		loop.baseline = reloader
		loop.evaluator.UseBaseline(acceptance)
		logger.Logf(logBaselineStartFormat, options.BaselineFile, acceptance.EntryCount(), acceptance.SavedOn())
	}
	if loop.notifier == nil {
		logger.Logf(logNotificationsDisabled)
	} else {
		logger.Logf(logNotificationsEnabled)
	}
	logger.Logf(logWatchStartFormat, interval, policy.MinSeverity)
	return loop.run(ctx)
}

// notifierFor returns a notifier only when Telegram is enabled and both the
// token and the chat id are present.
func notifierFor(runtime commandRuntime, settings configfile.TelegramSettings) ports.Notifier {
	if !settings.Enabled || settings.BotToken.IsEmpty() || settings.ChatID == "" {
		return nil
	}
	return runtime.NewNotifier(settings.BotToken, settings.ChatID)
}

// resolveAlertSettings combines the config file with the command-line
// overrides; flags win.
func resolveAlertSettings(settings configfile.AlertSettings, options WatchOptions) (app.AlertPolicy, time.Duration, error) {
	if options.MinSeverity != "" {
		settings.MinSeverity = options.MinSeverity
	}
	minimum, err := settings.ParsedMinSeverity()
	if err != nil {
		return app.AlertPolicy{}, 0, err
	}
	interval := time.Duration(settings.IntervalSeconds) * time.Second
	if options.Interval != 0 {
		interval = options.Interval
	}
	cooldown := time.Duration(settings.CooldownMinutes) * time.Minute
	return app.AlertPolicy{MinSeverity: minimum, Cooldown: cooldown}, interval, nil
}
