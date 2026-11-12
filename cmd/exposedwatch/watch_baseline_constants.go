package main

// Messages of the watch baseline mode. None of them may contain file
// contents; only the start-up line shows the baseline path.
const (
	usageWatchBaseline = "alert only on drift: ports that are not in this baseline file (create it with \"exposedwatch baseline save\"); " +
		"there is no first-run summary and the file is reloaded every cycle; it must pass the same trust checks as scan --baseline; " +
		"accepting a port accepts all findings on it, whatever their severity"
	logBaselineStartFormat        = "baseline: %s (%d entries, saved %s); alerting only on ports that are not in it"
	logBaselineReloadFailedFormat = "baseline reload failed: %v; using the last good baseline"
	logBaselineReloadedFormat     = "baseline reloaded: %d entries (was %d), saved %s%s"
	logBaselineRecovered          = "baseline reload recovered; using the current file"
	logBaselineWarningFormat      = "baseline warning: %s"
)

// A reload that grows the accepted set by at least widenedEntryGrowth
// entries, or to more than widenedGrowthFactor times its previous size, is
// marked so that a quietly widened baseline stands out in the log.
const (
	widenedEntryGrowth  = 50
	widenedGrowthFactor = 2
	widenedMarker       = " WIDENED"
)
