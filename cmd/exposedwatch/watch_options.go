package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"time"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

// WatchOptions are the validated flags of the watch command.
type WatchOptions struct {
	ConfigPath string
	// Interval and MinSeverity override the config file when set.
	Interval    time.Duration
	MinSeverity string
	Once        bool
	ProcRoot    string
	DockerSock  string
	NoDocker    bool
}

// parseWatchOptions parses and validates watch flags.
func parseWatchOptions(arguments []string, errorOutput io.Writer) (WatchOptions, error) {
	flagSet := flag.NewFlagSet(watchFlagSetName, flag.ContinueOnError)
	flagSet.SetOutput(errorOutput)
	options := WatchOptions{}
	flagSet.StringVar(&options.ConfigPath, flagConfig, "", usageConfig)
	flagSet.DurationVar(&options.Interval, flagInterval, 0, usageWatchInterval)
	flagSet.StringVar(&options.MinSeverity, flagMinSeverity, "", usageWatchMinSeverity)
	flagSet.BoolVar(&options.Once, flagOnce, false, usageOnce)
	flagSet.StringVar(&options.ProcRoot, flagProcRoot, defaultProcRoot, usageProcRoot)
	flagSet.StringVar(&options.DockerSock, flagDockerSocket, defaultDockerSock, usageDockerSocket)
	flagSet.BoolVar(&options.NoDocker, flagNoDocker, false, usageNoDocker)
	if err := flagSet.Parse(arguments); err != nil {
		return WatchOptions{}, err
	}
	if flagSet.NArg() > 0 {
		return WatchOptions{}, usageError{fmt.Sprintf(unexpectedArgsFmt, flagSet.Args())}
	}
	if options.Interval != 0 && options.Interval < minWatchInterval {
		return WatchOptions{}, usageError{fmt.Sprintf(invalidWatchIntervalFmt, options.Interval, minWatchInterval)}
	}
	if options.MinSeverity != "" {
		if _, err := domain.ParseSeverity(options.MinSeverity); err != nil {
			return WatchOptions{}, usageError{fmt.Sprintf(invalidSeverityFmt, err)}
		}
	}
	return options, nil
}

// scanOptions converts the watch flags into the scan source settings.
func (options WatchOptions) scanOptions() ScanOptions {
	return ScanOptions{ProcRoot: options.ProcRoot, DockerSocket: options.DockerSock, NoDocker: options.NoDocker}
}

// exitCodeForParseError maps a flag parsing failure to an exit code and
// prints validation errors (the flag package prints its own).
func exitCodeForParseError(err error, stderr io.Writer) ExitCode {
	if errors.Is(err, flag.ErrHelp) {
		return ExitOK
	}
	var validationError usageError
	if errors.As(err, &validationError) {
		fmt.Fprintln(stderr, validationError)
	}
	return ExitUsage
}
