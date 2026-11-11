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

// runTui executes the tui subcommand.
func runTui(arguments []string, stdout, stderr io.Writer) ExitCode {
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
	if !isTerminal(stdout) {
		fmt.Fprintln(stderr, nonTerminalMessage)
		return ExitUsage
	}
	err = tui.Run(tui.Options{
		Scanner:     app.NewScanService(buildSources(options.ScanOptions)),
		Connections: buildConnectionSource(options),
		Interval:    options.Interval,
		Version:     version,
		Theme:       options.Theme,
	})
	if err != nil {
		fmt.Fprintf(stderr, tuiFailedFormat, err)
		return ExitRuntimeError
	}
	return ExitOK
}

// buildConnectionSource builds the connection service over the same proc
// tree as the port scan, or returns nil when connections are disabled.
func buildConnectionSource(options TuiOptions) tui.ConnectionSnapshotter {
	if options.NoConnections {
		return nil
	}
	return app.NewConnectionService(procnet.New(options.ProcRoot))
}
