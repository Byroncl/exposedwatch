package report

import "time"

// Drift table texts.
const (
	driftSectionTitleFormat = "Drift since baseline %s (saved %s)"
	driftNoDriftText        = "  No drift: every exposed port is in the baseline.\n"
	driftNewTitleFormat     = "  NEW (%d): not in the baseline\n"
	driftClosedTitleFormat  = "  CLOSED (informational, %d): in the baseline but no longer exposed\n"
	driftNoteFormat         = "  - %s\n"
	driftRowIndent          = "    "
	driftCreatedAtLayout    = time.RFC3339
)
