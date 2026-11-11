package app

import (
	"fmt"
	"time"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

// NewPortAssessment is a port that is not in the baseline, with the severity
// of its most severe finding. A port without any finding counts as info.
type NewPortAssessment struct {
	Port     domain.Port
	Severity domain.Severity
	Title    string
}

// DriftReport is what the renderers show for a scan compared to a baseline.
// Its slices are never nil, so JSON shows [] instead of null.
type DriftReport struct {
	BaselineFile string
	CreatedAt    time.Time
	New          []NewPortAssessment
	Closed       []domain.BaselineEntry
	Notes        []string
	// UnverifiedSources are the sources that failed during the scan.
	UnverifiedSources []domain.Source
}

// BuildDriftReport joins a drift with the findings of the same scan. Findings
// are matched to new ports by identity (protocol, port, bind address, source).
func BuildDriftReport(baselineFile string, baseline domain.Baseline, drift Drift, findings []domain.Finding) DriftReport {
	driftReport := DriftReport{
		BaselineFile: baselineFile,
		CreatedAt:    baseline.CreatedAt,
		New:          make([]NewPortAssessment, 0, len(drift.New)),
		Closed:       append(make([]domain.BaselineEntry, 0, len(drift.Closed)), drift.Closed...),
		Notes:        make([]string, 0, len(drift.Unverified)),
	}
	strongest := strongestFindings(findings)
	for _, port := range drift.New {
		driftReport.New = append(driftReport.New, assessNewPort(port, strongest))
	}
	for _, unverified := range drift.Unverified {
		driftReport.UnverifiedSources = append(driftReport.UnverifiedSources, unverified.Source)
		driftReport.Notes = append(driftReport.Notes, unverifiedNote(unverified))
	}
	return driftReport
}

// HasNewAtOrAbove reports whether a new port reaches the severity threshold.
// Closed entries never count.
func (driftReport DriftReport) HasNewAtOrAbove(threshold domain.Severity) bool {
	for _, assessment := range driftReport.New {
		if assessment.Severity >= threshold {
			return true
		}
	}
	return false
}

// strongestFindings indexes the most severe finding of every port identity,
// so assessing many new ports stays linear.
func strongestFindings(findings []domain.Finding) map[domain.BaselineEntry]domain.Finding {
	strongest := make(map[domain.BaselineEntry]domain.Finding, len(findings))
	for _, finding := range findings {
		identity := domain.BaselineEntryOf(finding.Port)
		current, found := strongest[identity]
		if !found || finding.Severity > current.Severity {
			strongest[identity] = finding
		}
	}
	return strongest
}

func assessNewPort(port domain.Port, strongest map[domain.BaselineEntry]domain.Finding) NewPortAssessment {
	finding, found := strongest[domain.BaselineEntryOf(port)]
	if !found {
		return NewPortAssessment{Port: port, Severity: domain.Info}
	}
	return NewPortAssessment{Port: port, Severity: finding.Severity, Title: finding.Title}
}

func unverifiedNote(unverified UnverifiedSource) string {
	if unverified.Entries == 0 {
		return fmt.Sprintf(unverifiedNoEntriesNoteFormat, unverified.Source)
	}
	return fmt.Sprintf(unverifiedSourceNoteFormat, unverified.Source, unverified.Entries)
}
