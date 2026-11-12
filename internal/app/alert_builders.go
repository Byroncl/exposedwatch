package app

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

// newFindingAlert describes one finding. Only local facts are copied.
func newFindingAlert(hostName string, finding domain.Finding) domain.Alert {
	process := finding.Port.Process
	if process == "" {
		process = finding.Port.Container
	}
	return domain.Alert{
		Kind:     domain.AlertKindNewFinding,
		Severity: finding.Severity,
		Title:    finding.Title,
		Body:     finding.Recommendation,
		Host:     hostName,
		Port:     finding.Port.Number,
		Protocol: finding.Port.Protocol,
		Exposure: finding.Port.Exposure(),
		Process:  process,
		RuleID:   finding.RuleID,
	}
}

// startupSummaryAlert condenses the findings that already existed at start-up.
func startupSummaryAlert(hostName string, minimum domain.Severity, findings []domain.Finding) domain.Alert {
	return summaryAlert(hostName, findings, func(count int, noun string) string {
		return fmt.Sprintf(summaryTitleFormat, count, noun, minimum)
	})
}

// sourceSummaryAlert condenses the findings of a source that was not
// readable at start-up and became visible later.
func sourceSummaryAlert(hostName string, source domain.Source, findings []domain.Finding) domain.Alert {
	return summaryAlert(hostName, findings, func(count int, noun string) string {
		return fmt.Sprintf(sourceSummaryFormat, count, noun, source)
	})
}

func summaryAlert(hostName string, findings []domain.Finding, title func(count int, noun string) string) domain.Alert {
	ordered := append([]domain.Finding(nil), findings...)
	sort.SliceStable(ordered, func(first, second int) bool {
		if ordered[first].Severity != ordered[second].Severity {
			return ordered[first].Severity > ordered[second].Severity
		}
		return ordered[first].Port.Number < ordered[second].Port.Number
	})
	noun := summaryNounPlural
	if len(ordered) == singularFindingCount {
		noun = summaryNounSingular
	}
	lines := make([]string, 0, MaxSummaryTitles+1)
	for index, finding := range ordered {
		if index == MaxSummaryTitles {
			lines = append(lines, fmt.Sprintf(summaryMoreFormat, len(ordered)-MaxSummaryTitles))
			break
		}
		lines = append(lines, fmt.Sprintf(summaryItemFormat, finding.Title))
	}
	return domain.Alert{
		Kind:     domain.AlertKindSummary,
		Severity: ordered[0].Severity,
		Title:    title(len(ordered), noun),
		Body:     strings.Join(lines, summaryLineSeparator),
		Host:     hostName,
	}
}
