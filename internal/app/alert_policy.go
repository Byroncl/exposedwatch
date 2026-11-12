package app

import (
	"time"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

// AlertPolicy decides which findings may alert and how often.
type AlertPolicy struct {
	// MinSeverity is the lowest severity that alerts (inclusive).
	MinSeverity domain.Severity
	// Cooldown is how long a finding that disappeared stays remembered: if it
	// reappears sooner than this it does not alert again (spam guard).
	Cooldown time.Duration
}
