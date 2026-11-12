package main

import (
	"flag"
	"fmt"
	"io"
)

// BaselineSaveOptions are the validated flags of "baseline save".
type BaselineSaveOptions struct {
	// File is empty when the default path must be used.
	File         string
	ProcRoot     string
	DockerSocket string
	NoDocker     bool
	AllowPartial bool
	Force        bool
}

// parseBaselineSaveOptions parses and validates the flags of "baseline save".
func parseBaselineSaveOptions(arguments []string, errorOutput io.Writer) (BaselineSaveOptions, error) {
	flagSet := flag.NewFlagSet(baselineFlagSetName, flag.ContinueOnError)
	flagSet.SetOutput(errorOutput)
	options := BaselineSaveOptions{}
	flagSet.StringVar(&options.File, flagFile, "", usageBaselineFile)
	flagSet.StringVar(&options.ProcRoot, flagProcRoot, defaultProcRoot, usageProcRoot)
	flagSet.StringVar(&options.DockerSocket, flagDockerSocket, defaultDockerSock, usageDockerSocket)
	flagSet.BoolVar(&options.NoDocker, flagNoDocker, false, usageNoDocker)
	flagSet.BoolVar(&options.Force, flagForce, false, usageForce)
	flagSet.BoolVar(&options.AllowPartial, flagAllowPartial, false, usageAllowPartial)
	if err := flagSet.Parse(arguments); err != nil {
		return BaselineSaveOptions{}, err
	}
	if flagSet.NArg() > 0 {
		return BaselineSaveOptions{}, usageError{fmt.Sprintf(unexpectedArgsFmt, flagSet.Args())}
	}
	if err := rejectEmptyFlag(flagSet, flagFile, options.File); err != nil {
		return BaselineSaveOptions{}, err
	}
	return options, nil
}

// rejectEmptyFlag turns an explicitly empty value, such as --file "", into a
// usage error instead of silently falling back to a default.
func rejectEmptyFlag(flagSet *flag.FlagSet, name string, value string) error {
	explicit := false
	flagSet.Visit(func(visited *flag.Flag) {
		if visited.Name == name {
			explicit = true
		}
	})
	if explicit && value == "" {
		return usageError{fmt.Sprintf(emptyFlagValueFormat, name)}
	}
	return nil
}
