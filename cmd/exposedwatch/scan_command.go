package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"

	"github.com/Byroncl/exposedwatch/internal/adapters/dockerapi"
	"github.com/Byroncl/exposedwatch/internal/adapters/nmapscan"
	"github.com/Byroncl/exposedwatch/internal/adapters/procnet"
	"github.com/Byroncl/exposedwatch/internal/app"
	"github.com/Byroncl/exposedwatch/internal/domain"
	"github.com/Byroncl/exposedwatch/internal/report"
)

// scanDependencies are the parts of the scan command that tests replace.
type scanDependencies struct {
	// firewall builds the firewall evaluator; it is only called with --firewall.
	firewall func() firewallEvaluator
}

// runScan executes the scan subcommand and selects the exit code.
func runScan(arguments []string, stdout, stderr io.Writer) ExitCode {
	return executeScan(arguments, stdout, stderr, scanDependencies{
		firewall: func() firewallEvaluator { return buildFirewallService() },
	})
}

func executeScan(arguments []string, stdout, stderr io.Writer, dependencies scanDependencies) ExitCode {
	options, err := parseScanOptions(arguments, stderr)
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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	scanReport, err := app.NewScanService(buildSources(options)).Scan(ctx)
	if err != nil {
		fmt.Fprintf(stderr, scanFailedFormat, err)
		return ExitRuntimeError
	}
	scanReport = scanReport.FilterBySeverity(options.MinSeverity)

	if err := renderScan(ctx, options, stdout, scanReport, dependencies); err != nil {
		fmt.Fprintf(stderr, renderFailedFormat, err)
		return ExitRuntimeError
	}
	if scanReport.HasFindingsAtOrAbove(options.MinSeverity) {
		return ExitFindings
	}
	return ExitOK
}

// renderScan writes the report. With --firewall it also evaluates the listening
// ports against the host firewall; that is information only and never changes
// the report, its findings or the exit code.
func renderScan(ctx context.Context, options ScanOptions, stdout io.Writer, scanReport app.Report, dependencies scanDependencies) error {
	renderer := buildRenderer(options.Format, stdout)
	if !options.Firewall {
		return renderer.Render(stdout, scanReport)
	}
	firewallReport := evaluateFirewall(ctx, dependencies.firewall(), scanReport.Ports)
	return renderer.RenderWithFirewall(stdout, scanReport, firewallReport)
}

func buildSources(options ScanOptions) []app.NamedSource {
	sources := []app.NamedSource{
		{Name: domain.SourceHost, Source: procnet.New(options.ProcRoot)},
	}
	if !options.NoDocker {
		sources = append(sources, app.NamedSource{
			Name: domain.SourceDocker, Source: dockerapi.New(dockerapi.SocketPath(options.DockerSocket)),
		})
	}
	if options.External {
		sources = append(sources, app.NamedSource{
			Name: domain.SourceExternal,
			Source: nmapscan.New(nmapscan.Options{
				NmapPath: options.NmapPath,
				Targets:  options.ExternalTargets,
				Ports:    options.ExternalPorts,
				Timeout:  options.ExternalTimeout,
			}),
		})
	}
	return sources
}

func buildRenderer(format OutputFormat, stdout io.Writer) report.FirewallRenderer {
	if format == FormatJSON {
		return report.NewJSONRenderer()
	}
	return report.NewTableRenderer(report.DetectColorMode(stdout, os.Getenv(envNoColor)))
}
