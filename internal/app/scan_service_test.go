package app

import (
	"context"
	"errors"
	"testing"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

type fakePortSource struct {
	ports []domain.Port
	err   error
}

func (source fakePortSource) ListPorts(_ context.Context) ([]domain.Port, error) {
	return source.ports, source.err
}

var (
	exposedHostPort   = domain.Port{Number: 22, Protocol: domain.TCP, BindAddress: "0.0.0.0", Source: domain.SourceHost}
	exposedDockerPort = domain.Port{Number: 6379, Protocol: domain.TCP, BindAddress: "0.0.0.0", Source: domain.SourceDocker, Container: "cache"}
	errDockerDown     = errors.New("docker down")
	errProcDown       = errors.New("proc down")
)

func TestScanMergesSourcesAndEvaluates(testRun *testing.T) {
	service := NewScanService([]NamedSource{
		{Name: domain.SourceHost, Source: fakePortSource{ports: []domain.Port{exposedHostPort}}},
		{Name: domain.SourceDocker, Source: fakePortSource{ports: []domain.Port{exposedDockerPort}}},
	})
	report, err := service.Scan(context.Background())
	if err != nil {
		testRun.Fatalf("unexpected error: %v", err)
	}
	if len(report.Ports) != 2 || len(report.Findings) != 2 || len(report.SourceErrors) != 0 {
		testRun.Fatalf("report = %+v", report)
	}
}

func TestScanDegradesGracefullyOnPartialFailure(testRun *testing.T) {
	service := NewScanService([]NamedSource{
		{Name: domain.SourceHost, Source: fakePortSource{ports: []domain.Port{exposedHostPort}}},
		{Name: domain.SourceDocker, Source: fakePortSource{err: errDockerDown}},
	})
	report, err := service.Scan(context.Background())
	if err != nil {
		testRun.Fatalf("partial failure must not fail the scan: %v", err)
	}
	if len(report.Ports) != 1 {
		testRun.Fatalf("ports = %d, want 1", len(report.Ports))
	}
	if len(report.SourceErrors) != 1 || report.SourceErrors[0].Source != domain.SourceDocker || !errors.Is(report.SourceErrors[0].Err, errDockerDown) {
		testRun.Fatalf("source errors = %+v", report.SourceErrors)
	}
}

func TestScanFailsWhenEverySourceFails(testRun *testing.T) {
	service := NewScanService([]NamedSource{
		{Name: domain.SourceHost, Source: fakePortSource{err: errProcDown}},
		{Name: domain.SourceDocker, Source: fakePortSource{err: errDockerDown}},
	})
	report, err := service.Scan(context.Background())
	if !errors.Is(err, ErrAllSourcesFailed) {
		testRun.Fatalf("error = %v, want ErrAllSourcesFailed", err)
	}
	if len(report.SourceErrors) != 2 {
		testRun.Fatalf("source errors = %d, want 2", len(report.SourceErrors))
	}
}

func TestScanWithoutSourcesFails(testRun *testing.T) {
	_, err := NewScanService(nil).Scan(context.Background())
	if !errors.Is(err, ErrNoSources) {
		testRun.Fatalf("error = %v, want ErrNoSources", err)
	}
}

func TestReportFilterByMinimumSeverity(testRun *testing.T) {
	report := Report{Findings: []domain.Finding{
		{Severity: domain.Critical}, {Severity: domain.Medium}, {Severity: domain.Low},
	}}
	filtered := report.FilterBySeverity(domain.Medium)
	if len(filtered.Findings) != 2 {
		testRun.Fatalf("findings = %d, want 2", len(filtered.Findings))
	}
	if report.HighestSeverity() != domain.Critical {
		testRun.Fatalf("highest = %v", report.HighestSeverity())
	}
	if !filtered.HasFindingsAtOrAbove(domain.Critical) || filtered.HasFindingsAtOrAbove(domain.Critical+1) {
		testRun.Fatalf("HasFindingsAtOrAbove misbehaves")
	}
}
