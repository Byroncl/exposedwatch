package main

import (
	"flag"
	"fmt"
	"io"
)

// NotifyTestOptions are the validated flags of the notify-test command.
type NotifyTestOptions struct {
	ConfigPath string
}

// parseNotifyTestOptions parses notify-test flags.
func parseNotifyTestOptions(arguments []string, errorOutput io.Writer) (NotifyTestOptions, error) {
	flagSet := flag.NewFlagSet(notifyTestFlagSetName, flag.ContinueOnError)
	flagSet.SetOutput(errorOutput)
	options := NotifyTestOptions{}
	flagSet.StringVar(&options.ConfigPath, flagConfig, "", usageConfig)
	if err := flagSet.Parse(arguments); err != nil {
		return NotifyTestOptions{}, err
	}
	if flagSet.NArg() > 0 {
		return NotifyTestOptions{}, usageError{fmt.Sprintf(unexpectedArgsFmt, flagSet.Args())}
	}
	return options, nil
}
