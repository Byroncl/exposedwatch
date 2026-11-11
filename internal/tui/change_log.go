package tui

import (
	"time"

	"github.com/Byroncl/exposedwatch/internal/app"
	"github.com/Byroncl/exposedwatch/internal/domain"
)

// ChangeLogEntry records one port that opened or closed during the session.
type ChangeLogEntry struct {
	At     time.Time
	Change app.Change
	// Severity is the highest severity the port had when last observed;
	// HasSeverity is false when it never produced a finding.
	Severity    domain.Severity
	HasSeverity bool
}
