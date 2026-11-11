package app

// Alert evaluation limits.
const (
	// MaxSummaryTitles is how many finding titles the first-run summary lists.
	MaxSummaryTitles = 5
	// MaxPendingAlerts bounds the delivery queue; the oldest alerts are
	// dropped first so a long outage cannot grow memory without limit.
	MaxPendingAlerts = 200
)

// Outbox behaviour.
const (
	backoffMultiplier   = 2
	alertExpiredMessage = "alert could not be delivered before it expired"
)

// Alert texts.
const (
	summaryTitleFormat   = "%d %s at or above %s already present"
	sourceSummaryFormat  = "%d existing %s from %s now visible"
	summaryNounSingular  = "finding"
	summaryNounPlural    = "findings"
	summaryItemFormat    = "- %s"
	summaryMoreFormat    = "and %d more"
	summaryLineSeparator = "\n"
	singularFindingCount = 1
)
