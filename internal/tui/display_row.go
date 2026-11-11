package tui

import (
	"time"

	"github.com/Byroncl/exposedwatch/internal/app"
	"github.com/Byroncl/exposedwatch/internal/domain"
)

// rowKey identifies a row across refreshes so the selection can follow it.
type rowKey struct {
	ruleID     string
	identity   app.PortIdentity
	container  string
	kind       app.ChangeKind
	changedAt  time.Time
	connection app.ConnectionIdentity
}

// displayRow is one table row of any tab.
type displayRow struct {
	key  rowKey
	port domain.Port

	// severity is meaningful only when hasSeverity is true.
	severity    domain.Severity
	hasSeverity bool

	title          string
	ruleID         string
	recommendation string

	isNew        bool
	firstSeen    time.Time
	hasFirstSeen bool
	isChange     bool
	changeKind   app.ChangeKind
	changedAt    time.Time

	// connection and connectionAge are meaningful only when isConnection is true.
	isConnection  bool
	connection    domain.Connection
	connectionAge time.Duration
}

// severityRank orders rows without a finding below every severity.
func (row displayRow) severityRank() int {
	if !row.hasSeverity {
		return -1
	}
	return int(row.severity)
}

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
