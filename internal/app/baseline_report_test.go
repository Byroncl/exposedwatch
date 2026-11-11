package app

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

func TestBuildDriftReportAssessesNewPortsWithTheirFindings(testRun *testing.T) {
	telnet := hostPort(23, "0.0.0.0")
	quiet := hostPort(40000, "0.0.0.0")
	findings := []domain.Finding{
		{Port: telnet, Severity: domain.Low, RuleID: "a", Title: "minor"},
		{Port: telnet, Severity: domain.Critical, RuleID: "b", Title: "Telnet is exposed"},
	}
	drift := CompareToBaseline(baselineOf(), []domain.Port{quiet, telnet}, nil)
	driftReport := BuildDriftReport("/tmp/base.json", baselineOf(), drift, findings)
	if driftReport.BaselineFile != "/tmp/base.json" || !driftReport.CreatedAt.Equal(time.Date(2026, 11, 11, 10, 0, 0, 0, time.UTC)) {
		testRun.Fatalf("got %+v", driftReport)
	}
	if len(driftReport.New) != 2 {
		testRun.Fatalf("new: %+v", driftReport.New)
	}
	byPort := map[int]NewPortAssessment{}
	for _, assessment := range driftReport.New {
		byPort[assessment.Port.Number] = assessment
	}
	if byPort[23].Severity != domain.Critical || byPort[23].Title != "Telnet is exposed" {
		testRun.Fatalf("telnet: %+v", byPort[23])
	}
	if byPort[40000].Severity != domain.Info || byPort[40000].Title != "" {
		testRun.Fatalf("a port without findings counts as info: %+v", byPort[40000])
	}
}

func TestDriftReportHasNewAtOrAbove(testRun *testing.T) {
	driftReport := DriftReport{New: []NewPortAssessment{{Severity: domain.Info}, {Severity: domain.Medium}}}
	testCases := []struct {
		minimum domain.Severity
		want    bool
	}{
		{domain.Info, true}, {domain.Medium, true}, {domain.High, false}, {domain.Critical, false},
	}
	for _, testCase := range testCases {
		if got := driftReport.HasNewAtOrAbove(testCase.minimum); got != testCase.want {
			testRun.Errorf("minimum %s: got %v want %v", testCase.minimum, got, testCase.want)
		}
	}
	if (DriftReport{Closed: []domain.BaselineEntry{{}}}).HasNewAtOrAbove(domain.Info) {
		testRun.Error("closed entries never count")
	}
}

func TestBuildDriftReportNotesUnverifiedSources(testRun *testing.T) {
	baseline := baselineOf(entryOf(dockerPort(8080, "0.0.0.0")), entryOf(dockerPort(8081, "0.0.0.0")))
	drift := CompareToBaseline(baseline, nil, []domain.Source{domain.SourceDocker})
	driftReport := BuildDriftReport("f", baseline, drift, nil)
	if len(driftReport.Notes) != 1 || !strings.Contains(driftReport.Notes[0], "docker") || !strings.Contains(driftReport.Notes[0], "2") {
		testRun.Fatalf("notes: %v", driftReport.Notes)
	}
	if len(driftReport.UnverifiedSources) != 1 || driftReport.UnverifiedSources[0] != domain.SourceDocker {
		testRun.Fatalf("unverified: %v", driftReport.UnverifiedSources)
	}
	if driftReport.New == nil || driftReport.Closed == nil || driftReport.Notes == nil {
		testRun.Fatal("slices must be non-nil so JSON shows [] not null")
	}
}

func TestSnapshotBaselineKeepsOnlyHostAndDockerPorts(testRun *testing.T) {
	external := domain.Port{Number: 22, Protocol: domain.TCP, BindAddress: "192.0.2.1", Source: domain.SourceExternal}
	ports := []domain.Port{hostPort(80, "0.0.0.0"), hostPort(80, "0.0.0.0"), dockerPort(8080, "::"), external}
	createdAt := time.Date(2026, 11, 11, 10, 0, 0, 123, time.FixedZone("x", 3600))
	baseline := SnapshotBaseline(ports, createdAt, "v9")
	if baseline.Version != domain.BaselineVersion || baseline.ToolVersion != "v9" || baseline.CreatedAt.Location() != time.UTC {
		testRun.Fatalf("got %+v", baseline)
	}
	if baseline.CreatedAt.Nanosecond() != 0 {
		testRun.Fatal("created_at is stored with second precision")
	}
	if len(baseline.Entries) != 2 || baseline.Entries[0].Source != domain.SourceDocker {
		testRun.Fatalf("entries: %+v", baseline.Entries)
	}
	if err := baseline.Validate(); err != nil {
		testRun.Fatal(err)
	}
}

func TestBuildDriftReportNotesAFailedSourceWithoutBaselineEntries(testRun *testing.T) {
	drift := CompareToBaseline(baselineOf(), nil, []domain.Source{domain.SourceHost})
	driftReport := BuildDriftReport("f", baselineOf(), drift, nil)
	if len(driftReport.Notes) != 1 || !strings.Contains(driftReport.Notes[0], "host") || len(driftReport.UnverifiedSources) != 1 {
		testRun.Fatalf("got %+v", driftReport)
	}
}

func TestBuildDriftReportScalesLinearly(testRun *testing.T) {
	const portCount = 60000
	ports := make([]domain.Port, 0, portCount)
	findings := make([]domain.Finding, 0, portCount)
	for number := 0; number < portCount; number++ {
		port := hostPort(number%65000+1, fmt.Sprintf("10.%d.%d.%d", number/65025, number/255%255, number%255+1))
		ports = append(ports, port)
		findings = append(findings, domain.Finding{Port: port, Severity: domain.Low, Title: "t"})
	}
	started := time.Now()
	drift := CompareToBaseline(baselineOf(), ports, nil)
	driftReport := BuildDriftReport("f", baselineOf(), drift, findings)
	if elapsed := time.Since(started); elapsed > time.Second {
		testRun.Fatalf("took %v", elapsed)
	}
	if len(driftReport.New) != len(drift.New) || len(drift.New) < portCount/2 {
		testRun.Fatalf("new=%d", len(driftReport.New))
	}
}
