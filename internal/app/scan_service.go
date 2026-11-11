// Package app orchestrates port sources and the risk engine into scan reports.
package app

import (
	"context"
	"errors"

	"github.com/Byroncl/exposedwatch/internal/domain"
	"github.com/Byroncl/exposedwatch/internal/ports"
	"github.com/Byroncl/exposedwatch/internal/risk"
)

// NamedSource pairs a port source with the name used when reporting errors.
type NamedSource struct {
	Name   domain.Source
	Source ports.PortSource
}

// ScanService merges several port sources and evaluates their exposure.
type ScanService struct {
	sources []NamedSource
}

// NewScanService builds a service over the given sources.
func NewScanService(sources []NamedSource) *ScanService {
	return &ScanService{sources: sources}
}

// Scan queries every source. A failing source is recorded in the report and
// does not stop the scan; the scan fails only when every source failed.
func (service *ScanService) Scan(ctx context.Context) (Report, error) {
	if len(service.sources) == 0 {
		return Report{}, ErrNoSources
	}
	var report Report
	for _, namedSource := range service.sources {
		sourcePorts, err := namedSource.Source.ListPorts(ctx)
		if err != nil {
			if ctxErr := ctx.Err(); ctxErr != nil {
				return report, errors.Join(ctxErr, err)
			}
			report.SourceErrors = append(report.SourceErrors, SourceError{Source: namedSource.Name, Err: err})
			continue
		}
		report.Ports = append(report.Ports, sourcePorts...)
	}
	if len(report.SourceErrors) == len(service.sources) {
		return report, errors.Join(append([]error{ErrAllSourcesFailed}, sourceErrorsAsErrors(report.SourceErrors)...)...)
	}
	report.Findings = risk.Evaluate(report.Ports)
	return report, nil
}

func sourceErrorsAsErrors(sourceErrors []SourceError) []error {
	collected := make([]error, 0, len(sourceErrors))
	for _, sourceError := range sourceErrors {
		collected = append(collected, sourceError)
	}
	return collected
}
