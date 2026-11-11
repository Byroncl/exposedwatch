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
	"github.com/Byroncl/exposedwatch/internal/adapters/procnet"
	"github.com/Byroncl/exposedwatch/internal/app"
	"github.com/Byroncl/exposedwatch/internal/domain"
	"github.com/Byroncl/exposedwatch/internal/report"
)

// runScan executes the scan subcommand and selects the exit code.
func runScan(arguments []string, stdout, stderr io.Writer) ExitCode {
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

	if err := buildRenderer(options.Format, stdout).Render(stdout, scanReport); err != nil {
		fmt.Fprintf(stderr, renderFailedFormat, err)
		return ExitRuntimeError
	}
	if scanReport.HasFindingsAtOrAbove(options.MinSeverity) {
		return ExitFindings
	}
	return ExitOK
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
	return sources
}

func buildRenderer(format OutputFormat, stdout io.Writer) report.Renderer {
	if format == FormatJSON {
		return report.NewJSONRenderer()
	}
	return report.NewTableRenderer(report.DetectColorMode(stdout, os.Getenv(envNoColor)))
}
