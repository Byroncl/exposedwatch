package main

import (
	"context"
	"os"
	"time"

	"github.com/Byroncl/exposedwatch/internal/adapters/baselinefile"
	"github.com/Byroncl/exposedwatch/internal/adapters/configfile"
	"github.com/Byroncl/exposedwatch/internal/adapters/telegram"
	"github.com/Byroncl/exposedwatch/internal/app"
	"github.com/Byroncl/exposedwatch/internal/domain"
	"github.com/Byroncl/exposedwatch/internal/ports"
)

// scanFunc runs one scan.
type scanFunc func(ctx context.Context) (app.Report, error)

// commandRuntime holds everything the watch and notify-test commands take
// from the outside world, so tests can replace it. main wiring stays thin.
type commandRuntime struct {
	Environment configfile.Environment
	HostName    func() (string, error)
	NewNotifier func(token domain.BotToken, chatID domain.ChatID) ports.Notifier
	NewScanner  func(options ScanOptions) scanFunc
	Clock       app.Clock
	// LoadBaseline reads a trusted baseline file for watch --baseline.
	LoadBaseline baselineLoader
	// Wait pauses between scans and reports false when the context ended.
	Wait func(ctx context.Context, duration time.Duration) bool
}

// systemRuntime is the production runtime.
func systemRuntime() commandRuntime {
	return commandRuntime{
		Environment: configfile.SystemEnvironment(),
		HostName:    os.Hostname,
		NewNotifier: func(token domain.BotToken, chatID domain.ChatID) ports.Notifier {
			return telegram.New(telegram.Options{Token: token, ChatID: chatID})
		},
		NewScanner: func(options ScanOptions) scanFunc {
			return app.NewScanService(buildSources(options)).Scan
		},
		Clock: app.SystemClock{},
		LoadBaseline: func(path ports.BaselinePath) (ports.BaselineLoad, error) {
			return baselinefile.NewStore(baselinefile.SystemTrust()).Load(path)
		},
		Wait: waitForNextScan,
	}
}

func waitForNextScan(ctx context.Context, duration time.Duration) bool {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

// hostNameOrUnknown never fails: a missing host name only degrades the text.
func (runtime commandRuntime) hostNameOrUnknown() string {
	name, err := runtime.HostName()
	if err != nil || name == "" {
		return unknownHostName
	}
	return name
}

// loadConfiguration resolves the config path and loads the file with the
// environment overrides applied.
func (runtime commandRuntime) loadConfiguration(configPath string) (configfile.Loaded, error) {
	if configPath == "" {
		resolved, err := configfile.DefaultPath(runtime.Environment)
		if err != nil {
			return configfile.Loaded{}, err
		}
		configPath = resolved
	}
	return configfile.Load(configPath, runtime.Environment.Getenv)
}
