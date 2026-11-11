package tui

import "time"

// Connection reads.
const (
	// connectionsTimeout bounds one connection read so a stuck /proc walk
	// cannot keep the next read from starting.
	connectionsTimeout = 10 * time.Second
	// connectionNewBadgeRefreshCount is how many consecutive successful
	// connection snapshots, the one that detected it included, a new inbound
	// connection keeps its NEW badge.
	connectionNewBadgeRefreshCount = newBadgeRefreshCount
)

// Connection sort mode names.
const (
	connectionSortNameDirection = "direction"
	connectionSortNameLocalPort = "local-port"
	connectionSortNameRemote    = "remote"
	connectionSortNameProcess   = "process"
	connectionSortNameAge       = "age"
	connectionSortNameUnknown   = "unknown"
)

// Connection age units, shown as the largest whole unit.
const (
	ageFormatSeconds = "%ds"
	ageFormatMinutes = "%dm"
	ageFormatHours   = "%dh"
	ageFormatDays    = "%dd"
	hoursPerDay      = 24
)

// Connection table.
const (
	headingDirection     = "DIR"
	headingLocal         = "LOCAL"
	headingRemote        = "REMOTE"
	headingState         = "STATE"
	headingAge           = "AGE"
	directionInboundTag  = "IN"
	directionOutboundTag = "OUT"
)

// Connection table widths, in cells.
const (
	directionColumnWidth      = 5
	localColumnWidth          = 6
	remoteColumnWidth         = 24
	stateColumnWidth          = 11
	ageColumnWidth            = 5
	connectionMinProcessWidth = 12
)

// Connection table panel and messages.
const (
	connectionStatusFormat     = "· sort:%s loopback:%s"
	loopbackStateHidden        = "hidden"
	loopbackStateShown         = "shown"
	connectionsDisabledMessage = "Connections are disabled (started with --no-connections)."
	connectionsReadingMessage  = "Reading connections..."
	connectionsUnavailableText = "Connections could not be read yet; see the warning above."
	loopbackHiddenMessage      = "Loopback-to-loopback connections are hidden; press l to show them."
	noConnectionsMessage       = "No established connections."
	connectionsWarningFormat   = "! connections refresh failed: %v"
	connectionsTruncatedFormat = "! connections truncated: showing the first %d connections"
	loopbackFooterFormat       = "l loopback:%s"
	headerConnectionCounterFmt = "conn %d (in %d / out %d)"
)

// Connection summary panels.
const (
	panelTopRemoteTitle           = " Top remote addresses (inbound) "
	panelInboundPortsTitle        = " Inbound by local port "
	noInboundConnectionsMessage   = "No inbound connections."
	inboundPortFormat             = "%d/tcp"
	countFormat                   = "%d"
	moreEntriesFormat             = "... and %d more"
	connectionSideBySideMinWidth  = 60
	connectionStackedContentLines = 3
	// connectionPanelGap is the blank cells between the two side-by-side panels.
	connectionPanelGap = 1
	pairGapWidth       = 1
)
