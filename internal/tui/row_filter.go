package tui

import (
	"sort"
	"strconv"
	"strings"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

// meetsMinimum reports whether the row passes the minimum severity filter.
// Rows without a finding only pass while the filter is at its lowest level.
func (row displayRow) meetsMinimum(minimum domain.Severity) bool {
	if !row.hasSeverity {
		return minimum <= domain.Info
	}
	return row.severity >= minimum
}

// matchesText reports whether any displayed field contains the text.
func (row displayRow) matchesText(text string) bool {
	if text == "" {
		return true
	}
	port := row.port
	fields := []string{
		strconv.Itoa(port.Number), string(port.Protocol), port.BindAddress, port.Exposure().String(),
		string(port.Source), port.Process, port.Container, row.ruleID, row.title,
	}
	if row.hasSeverity {
		fields = append(fields, row.severity.String())
	}
	if row.isNew {
		fields = append(fields, newBadgeText)
	}
	return strings.Contains(strings.ToLower(strings.Join(fields, " ")), strings.ToLower(text))
}

// sortRows orders rows in place according to the mode.
func sortRows(rows []displayRow, mode SortMode) {
	sort.SliceStable(rows, func(left, right int) bool {
		return rowLess(rows[left], rows[right], mode)
	})
}

func rowLess(left, right displayRow, mode SortMode) bool {
	switch mode {
	case SortByPort:
		return lessByPortThenSeverity(left, right)
	case SortByProcess:
		leftName := strings.ToLower(processOrContainer(left.port))
		rightName := strings.ToLower(processOrContainer(right.port))
		if leftName != rightName {
			return leftName < rightName
		}
		return lessByPortThenSeverity(left, right)
	default:
		if left.severityRank() != right.severityRank() {
			return left.severityRank() > right.severityRank()
		}
		return left.port.Number < right.port.Number
	}
}

func lessByPortThenSeverity(left, right displayRow) bool {
	if left.port.Number != right.port.Number {
		return left.port.Number < right.port.Number
	}
	return left.severityRank() > right.severityRank()
}
