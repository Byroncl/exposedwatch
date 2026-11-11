package main

import (
	"errors"
	"flag"
	"fmt"
	"io"

	"github.com/Byroncl/exposedwatch/internal/adapters/procnet"
	"github.com/Byroncl/exposedwatch/internal/app"
	"github.com/Byroncl/exposedwatch/internal/tui"
)

// tuiLauncher starts the full-screen program; tui.Run in production.
type tuiLauncher func(options tui.Options) error

// runTui executes the tui subcommand.
func runTui(arguments []string, stdout, stderr io.Writer) ExitCode {
	return runTuiWith(systemRuntime(), tui.Run, isTerminal, arguments, stdout, stderr)
}

// runTuiWith is runTui with injectable runtime, launcher and terminal check.
func runTuiWith(runtime commandRuntime, launch tuiLauncher, terminal func(io.Writer) bool, arguments []string, stdout, stderr io.Writer) ExitCode {
	options, err := parseTuiOptions(arguments, stderr)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return ExitOK
		}
		var validationError usageError
		if errors.As(err, &validationError) {
			fmt.Fprintln(stderr, validationError)
		}
		return ExitUsage
	}
	if !terminal(stdout) {
		fmt.Fprintln(stderr, nonTerminalMessage)
		return ExitUsage
	}
	if err := launch(buildTuiOptions(options, runtime)); err != nil {
		fmt.Fprintf(stderr, tuiFailedFormat, err)
		return ExitRuntimeError
	}
	return ExitOK
}

// buildTuiOptions wires the scanners and the alert settings into the tui.
func buildTuiOptions(options TuiOptions, runtime commandRuntime) tui.Options {
	return tui.Options{
		Scanner:     app.NewScanService(buildSources(options.ScanOptions)),
		Connections: buildConnectionSource(options),
		Firewall:    buildFirewallEvaluator(options),
		Settings:    newConfigSettingsStore(runtime, options.ConfigPath),
		AlertTester: newTelegramAlertTester(runtime),
		Interval:    options.Interval,
		Version:     version,
		Theme:       options.Theme,
	}
}

// buildConnectionSource builds the connection service over the same proc
// tree as the port scan, or returns nil when connections are disabled.
func buildConnectionSource(options TuiOptions) tui.ConnectionSnapshotter {
	if options.NoConnections {
		return nil
	}
	return app.NewConnectionService(procnet.New(options.ProcRoot))
}
