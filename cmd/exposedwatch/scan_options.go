package main

import (
	"flag"
	"fmt"
	"io"
	"time"

	"github.com/Byroncl/exposedwatch/internal/adapters/nmapscan"
	"github.com/Byroncl/exposedwatch/internal/domain"
)

// ScanOptions are the validated flags of the scan command.
type ScanOptions struct {
	Format       OutputFormat
	MinSeverity  domain.Severity
	ProcRoot     string
	DockerSocket string
	NoDocker     bool
	// Firewall adds the firewall verdicts of the listening ports (information only).
	Firewall bool

	External         bool
	ExternalTargets  nmapscan.TargetList
	ConfirmOwnership bool
	ExternalPorts    nmapscan.PortSpec
	ExternalTimeout  time.Duration
	NmapPath         string
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
	flagSet.BoolVar(&options.Firewall, flagFirewall, false, usageFirewall)
	flagSet.BoolVar(&options.External, flagExternal, false, usageExternal)
	var rawTargets stringListFlag
	flagSet.Var(&rawTargets, flagExternalTarget, usageExternalTarget)
	flagSet.BoolVar(&options.ConfirmOwnership, flagConfirmOwnership, false, usageConfirmOwnership)
	rawPorts := flagSet.String(flagExternalPorts, "", usageExternalPorts)
	flagSet.DurationVar(&options.ExternalTimeout, flagExternalTimeout, nmapscan.DefaultTimeout, usageExternalTimeout)
	flagSet.StringVar(&options.NmapPath, flagNmapPath, defaultNmapPath, usageNmapPath)
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
	if err := validateExternalOptions(&options, rawTargets, *rawPorts); err != nil {
		return ScanOptions{}, err
	}
	return options, nil
}

// validateExternalOptions enforces the authorization guard of the external
// scan: other targets need --external and an explicit ownership confirmation,
// and every value is validated before any process can be started.
func validateExternalOptions(options *ScanOptions, rawTargets stringListFlag, rawPorts string) error {
	if len(rawTargets) > 0 {
		if !options.External {
			return usageError{externalTargetNeedsExternalMessage}
		}
		if !options.ConfirmOwnership {
			return usageError{externalTargetNeedsConfirmMessage}
		}
		targets, err := nmapscan.ParseTargets(rawTargets)
		if err != nil {
			return usageError{fmt.Sprintf(invalidExternalTargetFmt, err)}
		}
		options.ExternalTargets = targets
	}
	if !options.External {
		return nil
	}
	ports, err := nmapscan.ParsePortSpec(rawPorts)
	if err != nil {
		return usageError{fmt.Sprintf(invalidExternalPortsFmt, err)}
	}
	options.ExternalPorts = ports
	if err := nmapscan.ValidateTimeout(options.ExternalTimeout); err != nil {
		return usageError{fmt.Sprintf(invalidExternalTimeoutFmt, err)}
	}
	return nil
}
