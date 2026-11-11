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

// tickMessage fires when the auto-refresh interval elapsed.
type tickMessage struct{}
