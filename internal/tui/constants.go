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

// Layout.
const (
	headerLineCount      = 2
	footerLineCount      = 1
	tableHeaderLineCount = 1
	detailPaneHeight     = 6
	minTableRowsToShow   = 3
	maxWarningLines      = 3
	columnSeparator      = " "
	minTitleColumnWidth  = 10
	scrollOffsetFloor    = 0
)

// Table column widths and the terminal width from which each column shows.
const (
	severityColumnWidth = 8
	portColumnWidth     = 6
	protocolColumnWidth = 5
	bindColumnWidth     = 16
	exposureColumnWidth = 14
	sourceColumnWidth   = 7
	processColumnWidth  = 20

	showProtocolFromWidth = 50
	showProcessFromWidth  = 70
	showBindFromWidth     = 85
	showExposureFromWidth = 100
	showSourceFromWidth   = 110
)

// Table headings.
const (
	headingSeverity = "SEVERITY"
	headingPort     = "PORT"
	headingProtocol = "PROTO"
	headingBind     = "BIND"
	headingExposure = "EXPOSURE"
	headingSource   = "SOURCE"
	headingProcess  = "PROCESS/CONTAINER"
	headingTitle    = "TITLE"
)

// User-facing strings.
const (
	appTitle              = "exposedwatch"
	titleFormat           = " %s %s"
	updatedFormat         = "  updated %s"
	updatedPlaceholder    = "--:--:--"
	refreshingIndicator   = "  refreshing..."
	timeLayout            = "15:04:05"
	portsCountFormat      = " ports %d"
	severityCountFormat   = "%s %d"
	countSeparator        = "  "
	scanningMessage       = "Scanning..."
	noMatchMessage        = "No findings match the current filters."
	noFindingsMessage     = "No findings."
	emptyValue            = "-"
	sourceWarningFormat   = "! %s"
	refreshWarningFormat  = "! refresh failed: %v"
	moreWarningsFormat    = "! ... and %d more warnings"
	processContainerJoin  = " / "
	filterPromptFormat    = "/%s%s  (enter apply, esc cancel)"
	filterCursor          = "_"
	statusFormat          = "  [sort:%s min:%s filter:%q]"
	statusFormatNoFilter  = "  [sort:%s min:%s]"
	detailRuleFormat      = "Rule: %s   Severity: %s   Source: %s"
	detailWhereFormat     = "Where: %s   Exposure: %s   Process/Container: %s"
	detailFixFormat       = "Fix: %s"
	detailEmptyMessage    = "Select a finding to see details."
	detailSeparatorGlyph  = "-"
	helpTitle             = "Keys"
	helpCloseHint         = "Press ? or esc to close."
	helpKeyDescriptionFmt = "%-12s %s"
	sortModeNameSeverity  = "severity"
	sortModeNamePort      = "port"
	sortModeNameProcess   = "process"
	sortModeNameUnknown   = "unknown"
)
