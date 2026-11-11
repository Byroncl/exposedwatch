package tui

// Glyphs.
const (
	selectionMark        = "▌"
	noSelectionMark      = " "
	gaugeFilledGlyph     = "█"
	gaugeEmptyGlyph      = "░"
	emptyValue           = "-"
	headerJoin           = "  "
	processContainerJoin = " / "
)

// Header.
const (
	appTitle            = "exposedwatch"
	titleFormat         = " %s %s"
	updatedFormat       = "  updated %s"
	intervalFormat      = "  every %s"
	updatedPlaceholder  = "--:--:--"
	refreshingText      = "  refreshing..."
	timeLayout          = "15:04:05"
	portsCountFormat    = "ports %d"
	riskLabel           = " RISK "
	riskFormat          = " %s"
	sourceCountFormat   = "%s %d"
	openedCountFormat   = "+%d new"
	closedCountFormat   = "-%d closed"
	chipCountFormat     = "%s %d"
	sourceLabelExternal = "ext"
)

// Tabs.
const (
	activeTabFormat   = "[%d %s %d]"
	inactiveTabFormat = " %d %s %d "
	// Tabs without a count (Summary, Top remote, placeholders).
	activeTabPlainFormat   = "[%d %s]"
	inactiveTabPlainFormat = " %d %s "
	compactActiveTab       = "[%d]"
	compactInactiveTab     = " %d "
	tabDivider             = "│"
	tabGap                 = " "
)

// Table headings.
const (
	headingSeverity     = "SEVERITY"
	headingPort         = "PORT"
	headingExposure     = "EXPOSURE"
	headingBind         = "BIND"
	headingSource       = "SOURCE"
	headingProcess      = "PROCESS/CONTAINER"
	headingProcessShort = "PROCESS"
	headingFinding      = "FINDING"
	headingTime         = "TIME"
	headingKind         = "EVENT"
	headingNew          = ""
	newBadgeText        = "NEW"
	portFormat          = "%d/%s"
	kindOpenedLabel     = "OPENED"
	kindClosedLabel     = "CLOSED"
	timeLayoutShort     = "15:04:05"
)

// Table panel.
const (
	panelPositionFormat = " %s %d/%d "
	panelPositionEmpty  = " %s "
	panelStatusFormat   = "· sort:%s min:%s"
	panelFilterFormat   = " filter:%q"
	panelChangesNote    = "· newest first"
	panelDetailsTitle   = " Details "
	panelHelpTitle      = " Keys "
	scanningMessage     = "Scanning..."
	noMatchMessage      = "No rows match the current filters."
	noFindingsMessage   = "No findings."
	noPortsMessage      = "No listening ports found."
	noChangesMessage    = "No changes yet. The first scan is the baseline; later opened and closed ports appear here."
)

// Warnings.
const (
	sourceWarningFormat  = "! %s"
	refreshWarningFormat = "! refresh failed: %v"
	moreWarningsFormat   = "! ... and %d more warnings"
)

// Footer.
const (
	filterPromptFormat = "/%s%s  (enter apply, esc cancel)"
	filterCursor       = "_"
	footerHintJoin     = "  "
	hintModules        = "tab modules"
	hintInnerTabs      = "[/] tabs"
	hintSidebar        = "b sidebar"
	hintMove           = "j/k move"
	hintEnds           = "g/G top/bottom"
	hintSort           = "s sort"
	hintMinimum        = "f min-sev"
	hintFilter         = "/ filter"
	hintRefresh        = "r refresh"
	hintHelp           = "? help"
	hintQuit           = "q quit"
)

// Detail pane.
const (
	detailEmptyMessage     = "Select a row to see details."
	detailNoFindingMessage = "No finding for this port."
	detailRuleFormat       = "Rule: %s   Severity: %s   Source: %s"
	detailBindFormat       = "Bind: %s   Exposure: %s"
	detailProcessFormat    = "Process: %s   Container: %s"
	detailFirstSeenFormat  = "First seen: %s   State: %s"
	detailChangeFormat     = "When: %s   Event: %s   Last severity: %s"
	detailFixFormat        = "Fix: %s"
	detailStateNew         = "NEW"
	detailStateStable      = "stable"
	detailOpenedTitle      = "Port opened"
	detailClosedTitle      = "Port closed"
	detailSourceFormat     = "Source: %s"
	portBindFormat         = "%s:%d/%s"
	detailUnknownFirstSeen = "unknown"
)

// Help.
const (
	helpCloseHint         = "Press ? or esc to close."
	helpKeyDescriptionFmt = "%-16s %s"
	helpCountersNote      = "+N new / -M closed count ports opened and closed since this session started."
)
