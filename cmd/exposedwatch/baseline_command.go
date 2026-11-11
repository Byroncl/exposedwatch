package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/Byroncl/exposedwatch/internal/adapters/baselinefile"
	"github.com/Byroncl/exposedwatch/internal/app"
	"github.com/Byroncl/exposedwatch/internal/domain"
	"github.com/Byroncl/exposedwatch/internal/ports"
)

// baselineDependencies are the parts of the baseline command that tests replace.
type baselineDependencies struct {
	environment baselinefile.Environment
	now         func() time.Time
}

// runBaseline dispatches the baseline subcommands.
func runBaseline(arguments []string, stdout, stderr io.Writer) ExitCode {
	return executeBaseline(arguments, stdout, stderr, baselineDependencies{
		environment: baselinefile.SystemEnvironment(),
		now:         time.Now,
	})
}

func executeBaseline(arguments []string, stdout, stderr io.Writer, dependencies baselineDependencies) ExitCode {
	if len(arguments) == 0 || arguments[0] != subcommandBaselineSave {
		fmt.Fprint(stderr, baselineUsageText)
		return ExitUsage
	}
	options, err := parseBaselineSaveOptions(arguments[1:], stderr)
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
	return saveBaseline(options, stdout, stderr, dependencies)
}

// saveBaseline scans the host (and Docker) and writes the baseline. It never
// changes the system: the only file it writes is the baseline itself.
func saveBaseline(options BaselineSaveOptions, stdout, stderr io.Writer, dependencies baselineDependencies) ExitCode {
	path, err := resolveBaselinePath(options, dependencies.environment)
	if err != nil {
		fmt.Fprintf(stderr, baselinePathFailedFormat, err)
		return ExitRuntimeError
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	scanOptions := ScanOptions{ProcRoot: options.ProcRoot, DockerSocket: options.DockerSocket, NoDocker: options.NoDocker}
	scanReport, err := app.NewScanService(buildSources(scanOptions)).Scan(ctx)
	if err != nil {
		fmt.Fprintf(stderr, scanFailedFormat, err)
		return ExitRuntimeError
	}
	if failed := failedSourceNames(scanReport); failed != "" {
		if !options.AllowPartial {
			fmt.Fprintf(stderr, baselinePartialFormat, failed)
			return ExitRuntimeError
		}
		fmt.Fprintf(stderr, baselinePartialWarningFormat, failed)
	}
	baseline := app.SnapshotBaseline(scanReport.Ports, dependencies.now(), version)
	store := baselinefile.NewStore(baselinefile.SystemTrust())
	if options.Force {
		store = store.Forcing()
	}
	if err := store.Save(path, baseline); err != nil {
		if errors.Is(err, baselinefile.ErrNotABaseline) {
			fmt.Fprintf(stderr, baselineRefuseOverwriteFormat, path)
			return ExitRuntimeError
		}
		fmt.Fprintf(stderr, baselineSaveFailedFormat, err)
		return ExitRuntimeError
	}
	fmt.Fprintf(stdout, baselineSavedFormat, path, len(baseline.Entries))
	return ExitOK
}

func resolveBaselinePath(options BaselineSaveOptions, environment baselinefile.Environment) (ports.BaselinePath, error) {
	if options.File != "" {
		return ports.BaselinePath(options.File), nil
	}
	return baselinefile.DefaultPath(environment)
}

// failedSourceNames lists the sources that could not be read, comma separated.
func failedSourceNames(scanReport app.Report) string {
	names := make([]string, 0, len(scanReport.SourceErrors))
	for _, sourceError := range scanReport.SourceErrors {
		names = append(names, string(sourceError.Source))
	}
	return strings.Join(names, sourceListSeparator)
}

// unreadSources lists the sources that could not be read during a scan.
func unreadSources(scanReport app.Report) []domain.Source {
	sources := make([]domain.Source, 0, len(scanReport.SourceErrors))
	for _, sourceError := range scanReport.SourceErrors {
		sources = append(sources, sourceError.Source)
	}
	return sources
}
