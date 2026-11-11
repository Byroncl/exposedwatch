package main

import (
	"flag"
	"fmt"
	"io"
	"time"

	"github.com/Byroncl/exposedwatch/internal/tui"
)

// TuiOptions are the validated flags of the tui command.
type TuiOptions struct {
	ScanOptions
	Interval time.Duration
}

// parseTuiOptions parses and validates tui flags.
func parseTuiOptions(arguments []string, errorOutput io.Writer) (TuiOptions, error) {
	flagSet := flag.NewFlagSet(tuiFlagSetName, flag.ContinueOnError)
	flagSet.SetOutput(errorOutput)
	options := TuiOptions{}
	flagSet.StringVar(&options.ProcRoot, flagProcRoot, defaultProcRoot, usageProcRoot)
	flagSet.StringVar(&options.DockerSocket, flagDockerSocket, defaultDockerSock, usageDockerSocket)
	flagSet.BoolVar(&options.NoDocker, flagNoDocker, false, usageNoDocker)
	flagSet.DurationVar(&options.Interval, flagInterval, tui.DefaultInterval, usageInterval)
	if err := flagSet.Parse(arguments); err != nil {
		return TuiOptions{}, err
	}
	if flagSet.NArg() > 0 {
		return TuiOptions{}, usageError{fmt.Sprintf(unexpectedArgsFmt, flagSet.Args())}
	}
	if options.Interval < tui.MinimumInterval {
		return TuiOptions{}, usageError{fmt.Sprintf(invalidIntervalFmt, options.Interval, tui.MinimumInterval)}
	}
	return options, nil
}
