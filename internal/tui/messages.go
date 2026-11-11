package tui

import (
	"time"

	"github.com/Byroncl/exposedwatch/internal/app"
)

// RefreshResultMessage carries the outcome of one background scan.
type RefreshResultMessage struct {
	Report app.Report
	Err    error
	At     time.Time
}

// ConnectionsResultMessage carries the outcome of one background connection
// read. It is separate from RefreshResultMessage so a slow or failing
// connection read never blocks or breaks the port scan.
type ConnectionsResultMessage struct {
	Snapshot app.ConnectionSnapshot
	Err      error
	At       time.Time
}

// tickMessage fires when the auto-refresh interval elapsed.
type tickMessage struct{}
