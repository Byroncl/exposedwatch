package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/Byroncl/exposedwatch/internal/adapters/baselinefile"
	"github.com/Byroncl/exposedwatch/internal/app"
	"github.com/Byroncl/exposedwatch/internal/domain"
	"github.com/Byroncl/exposedwatch/internal/ports"
)

// baselineInput is a trusted baseline together with the file it came from.
type baselineInput struct {
	file     string
	baseline domain.Baseline
}

// loadBaselineInput reads the baseline through the trust checks and prints
// the non-fatal warnings (for example a group-writable directory).
func loadBaselineInput(file string, stderr io.Writer) (baselineInput, error) {
	loaded, err := baselinefile.NewStore(baselinefile.SystemTrust()).Load(ports.BaselinePath(file))
	if err != nil {
		return baselineInput{}, err
	}
	for _, warning := range loaded.Warnings {
		fmt.Fprintf(stderr, baselineWarningFormat, warning)
	}
	return baselineInput{file: file, baseline: loaded.Baseline}, nil
}

// compare builds the drift report of a scan. A source that failed during the
// scan is unknown, so its baseline entries are not reported as closed.
func (input baselineInput) compare(scanReport app.Report) app.DriftReport {
	drift := app.CompareToBaseline(input.baseline, scanReport.Ports, unreadSources(scanReport))
	return app.BuildDriftReport(input.file, input.baseline, drift, scanReport.Findings)
}

// baselineExitCode picks the exit code of a scan compared to a baseline.
// Accepted ports are expected, so code 3 is never used. New ports at or above
// the minimum severity give 4 and win over everything else; otherwise a
// source that could not be read gives 1 (a silent partial scan must not look
// clean); otherwise 0. Closed ports never matter.
func baselineExitCode(driftReport app.DriftReport, minimum domain.Severity, stderr io.Writer) ExitCode {
	if len(driftReport.UnverifiedSources) > 0 {
		names := make([]string, 0, len(driftReport.UnverifiedSources))
		for _, source := range driftReport.UnverifiedSources {
			names = append(names, string(source))
		}
		fmt.Fprintf(stderr, unverifiedSourcesFormat, strings.Join(names, sourceListSeparator))
	}
	switch {
	case driftReport.HasNewAtOrAbove(minimum):
		return ExitDrift
	case len(driftReport.UnverifiedSources) > 0:
		return ExitRuntimeError
	default:
		return ExitOK
	}
}
