package main

import (
	"flag"
	"fmt"
	"io"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

// ScanOptions are the validated flags of the scan command.
type ScanOptions struct {
	Format       OutputFormat
	MinSeverity  domain.Severity
	ProcRoot     string
	DockerSocket string
	NoDocker     bool
}

// parseScanOptions parses and validates scan flags. Flag errors are written
// to errorOutput by the flag package; validation errors are returned.
func parseScanOptions(arguments []string, errorOutput io.Writer) (ScanOptions, error) {
	flagSet := flag.NewFlagSet(scanFlagSetName, flag.ContinueOnError)
	flagSet.SetOutput(errorOutput)
	formatText := flagSet.String(flagFormat, string(defaultFormat), usageFormat)
	severityText := flagSet.String(flagMinSeverity, defaultMinSeverity, usageMinSeverity)
	options := ScanOptions{}
	flagSet.StringVar(&options.ProcRoot, flagProcRoot, defaultProcRoot, usageProcRoot)
	flagSet.StringVar(&options.DockerSocket, flagDockerSocket, defaultDockerSock, usageDockerSocket)
	flagSet.BoolVar(&options.NoDocker, flagNoDocker, false, usageNoDocker)
	if err := flagSet.Parse(arguments); err != nil {
		return ScanOptions{}, err
	}
	if flagSet.NArg() > 0 {
		return ScanOptions{}, usageError{fmt.Sprintf(unexpectedArgsFmt, flagSet.Args())}
	}
	options.Format = OutputFormat(*formatText)
	if options.Format != FormatTable && options.Format != FormatJSON {
		return ScanOptions{}, usageError{fmt.Sprintf(invalidFormatMessage, *formatText)}
	}
	minimum, err := domain.ParseSeverity(*severityText)
	if err != nil {
		return ScanOptions{}, usageError{fmt.Sprintf(invalidSeverityFmt, err)}
	}
	options.MinSeverity = minimum
	return options, nil
}
