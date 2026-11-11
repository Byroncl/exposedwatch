package tui

// Vertical budget of one frame, in lines.
const (
	headerLineCount      = 3
	tabLineCount         = 1
	footerLineCount      = 1
	panelBorderLineCount = 2
	tableHeadingLines    = 1
	detailContentLines   = 7
	minTableRowsToShow   = 5
	maxWarningLines      = 3
	scrollOffsetFloor    = 0
)

// Horizontal budget of one frame, in cells.
const (
	panelBorderWidth   = 2
	panelPaddingWidth  = 2
	minPanelWidth      = 6
	panelTitleMargin   = 4
	columnSeparator    = " "
	selectionMarkWidth = 1
	minTitleWidth      = 32
	riskGaugeWidth     = 10
	chipSidePadding    = 1
)

// Table column widths.
const (
	severityColumnWidth = 10
	newColumnWidth      = 3
	portColumnWidth     = 9
	exposureColumnWidth = 14
	bindColumnWidth     = 15
	processColumnWidth  = 14
	sourceColumnWidth   = 8
	timeColumnWidth     = 8
	kindColumnWidth     = 8
)

// Column priorities: lower numbers are kept first when the width is short.
const (
	priorityMandatory = iota
	priorityExposure
	priorityProcess
	priorityBind
	prioritySource
	priorityKind
	priorityConnectionAge
	priorityConnectionState
	priorityLocalBind
)

// Footer hint drop ranks: higher ranks disappear first on narrow terminals.
const (
	dropRankNever = iota
	dropRankInbound
	dropRankLoopback
	dropRankMove
	dropRankRefresh
	dropRankSort
	dropRankFilter
	dropRankMinimum
	dropRankEnds
)
