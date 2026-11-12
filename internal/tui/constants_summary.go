package tui

// Overview limits.
const (
	worstFindingsLimit = 5
	recentChangesLimit = 5
	topRemoteLimit     = 3
)

// Overview risk gauge, in cells.
const (
	overviewGaugeMaxWidth = 30
	overviewGaugeMinWidth = 10
	overviewGaugeMargin   = 2
)

// Overview headings and messages.
const (
	panelOverviewTitle                = " Overview "
	headingWorstFindings              = "Worst findings"
	headingRecentChanges              = "Recent changes"
	headingTopRemote                  = "Top remote addresses (inbound)"
	noRecentChangesMessage            = "No changes yet."
	summaryConnectionsDisabledMessage = "connections disabled"
	summaryConnectionsReadingMessage  = "connections reading..."
	findingRowFormat                  = "%s %s %s"
	changeRowFormat                   = "%s %s %s %s"
)

// Overview column widths, in cells.
const (
	findingChipWidth = 10
	changeKindWidth  = 6
	overviewIndent   = "  "
)
