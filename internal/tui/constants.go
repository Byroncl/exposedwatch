// Package tui implements the interactive full-screen terminal interface.
package tui

import "time"

// Timing.
const (
	// MinimumInterval is the shortest allowed auto-refresh interval.
	MinimumInterval = time.Second
	// DefaultInterval is the auto-refresh interval used when none is given.
	DefaultInterval = 2 * time.Second
	scanTimeout     = 15 * time.Second
)

// Fallback terminal size used until the first tea.WindowSizeMsg arrives.
const (
	fallbackWidth  = 80
	fallbackHeight = 24
)

// Change tracking.
const (
	// newBadgeRefreshCount is how many consecutive successful snapshots, the
	// one that detected the port included, an opened port keeps its NEW badge.
	newBadgeRefreshCount = 5
	// maxChangeLogEntries bounds the change log; the oldest entries are dropped.
	maxChangeLogEntries = 200
)

// Sort mode names.
const (
	sortModeNameSeverity = "severity"
	sortModeNamePort     = "port"
	sortModeNameProcess  = "process"
	sortModeNameUnknown  = "unknown"
)

// Theme names accepted by --theme.
const (
	themeNameAuto    = "auto"
	themeNameDark    = "dark"
	themeNameLight   = "light"
	unknownThemeText = "unknown theme %q: want auto, dark or light"
)

// Tab titles.
const (
	tabTitleFindings    = "Findings"
	tabTitleAllPorts    = "All ports"
	tabTitleChanges     = "Changes"
	tabTitleConnections = "Connections"
)

// Risk level words shown next to the gauge.
const (
	riskWordSafe     = "SAFE"
	riskWordLow      = "LOW"
	riskWordMedium   = "MEDIUM"
	riskWordHigh     = "HIGH"
	riskWordCritical = "CRITICAL"
)
