package tui

import (
	"sort"
	"strconv"
	"strings"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

// processOrContainer is the label for the process/container column.
func processOrContainer(port domain.Port) string {
	switch {
	case port.Process != "" && port.Container != "":
		return port.Process + processContainerJoin + port.Container
	case port.Container != "":
		return port.Container
	case port.Process != "":
		return port.Process
	default:
		return emptyValue
	}
}

// matchesText reports whether any displayed field contains the text.
func matchesText(finding domain.Finding, text string) bool {
	if text == "" {
		return true
	}
	port := finding.Port
	haystack := strings.ToLower(strings.Join([]string{
		strconv.Itoa(port.Number), string(port.Protocol), port.BindAddress, port.Exposure().String(),
		string(port.Source), port.Process, port.Container,
		finding.Severity.String(), finding.RuleID, finding.Title,
	}, " "))
	return strings.Contains(haystack, strings.ToLower(text))
}

// sortFindings orders findings in place according to the mode.
func sortFindings(findings []domain.Finding, mode SortMode) {
	sort.SliceStable(findings, func(left, right int) bool {
		return findingLess(findings[left], findings[right], mode)
	})
}

func findingLess(left, right domain.Finding, mode SortMode) bool {
	switch mode {
	case SortByPort:
		return lessByPortThenSeverity(left, right)
	case SortByProcess:
		leftName := strings.ToLower(processOrContainer(left.Port))
		rightName := strings.ToLower(processOrContainer(right.Port))
		if leftName != rightName {
			return leftName < rightName
		}
		return lessByPortThenSeverity(left, right)
	default:
		if left.Severity != right.Severity {
			return left.Severity > right.Severity
		}
		return left.Port.Number < right.Port.Number
	}
}

func lessByPortThenSeverity(left, right domain.Finding) bool {
	if left.Port.Number != right.Port.Number {
		return left.Port.Number < right.Port.Number
	}
	return left.Severity > right.Severity
}

// sameFinding identifies a finding across refreshes.
func sameFinding(left, right domain.Finding) bool {
	return left.RuleID == right.RuleID &&
		left.Port.Number == right.Port.Number &&
		left.Port.Protocol == right.Port.Protocol &&
		left.Port.BindAddress == right.Port.BindAddress &&
		left.Port.Source == right.Port.Source &&
		left.Port.Container == right.Port.Container
}

// countBySeverity tallies findings per severity.
func countBySeverity(findings []domain.Finding) map[domain.Severity]int {
	counts := make(map[domain.Severity]int)
	for _, finding := range findings {
		counts[finding.Severity]++
	}
	return counts
}
