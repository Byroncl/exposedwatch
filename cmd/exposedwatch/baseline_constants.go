package main

// Baseline command and flag names.
const (
	commandBaseline        = "baseline"
	subcommandBaselineSave = "save"
	baselineFlagSetName    = "baseline save"
	flagBaseline           = "baseline"
	flagFile               = "file"
	flagAllowPartial       = "allow-partial"
	flagForce              = "force"
)

// Baseline flag descriptions.
const (
	usageBaseline = "compare the scan with this baseline file and report drift: new ports that are not in it " +
		"(closed ones are informational); exit code 4 when a new port reaches --min-severity, " +
		"exit code 3 is not used; create the file with \"exposedwatch baseline save\""
	usageBaselineFile = "baseline file to write (default: /var/lib/exposedwatch/baseline.json as root, " +
		"otherwise $XDG_STATE_HOME or ~/.local/state under exposedwatch/baseline.json)"
	usageAllowPartial = "save the baseline even when a source (for example Docker) could not be read; " +
		"the ports of that source are then missing from the baseline and will count as new"
)

const usageForce = "DANGEROUS: replace the --file target even when it is not a valid baseline " +
	"(never follows a symbolic link and never replaces a directory)"

// Baseline messages.
const (
	baselineUsageText = `Usage: exposedwatch baseline save [--file PATH] [--proc-root PATH] [--docker-socket PATH] [--no-docker] [--allow-partial] [--force]

Writes the ports that are exposed right now (host and Docker) as a baseline
file. Compare later scans with "exposedwatch scan --baseline PATH".
`
	baselineSavedFormat           = "Baseline saved to %s (%d ports)\n"
	baselineWarningFormat         = "warning: %s\n"
	baselinePartialFormat         = "baseline not saved: %s could not be read, so its ports would be missing from the baseline and later count as new; fix the problem or pass --allow-partial\n"
	baselinePartialWarningFormat  = "warning: saving a partial baseline because %s could not be read; its ports will count as new\n"
	baselineRefuseOverwriteFormat = "refusing to overwrite %s: it is not a baseline file (use --force to replace it anyway)\n"
	unverifiedSourcesFormat       = "unverified sources: %s\n"
	baselineSaveFailedFormat      = "baseline save failed: %v\n"
	baselineLoadFailedFormat      = "baseline could not be used: %v\nCreate or refresh it with \"exposedwatch baseline save\"\n"
	baselinePathFailedFormat      = "cannot choose a baseline path: %v\nPass --file explicitly.\n"
	emptyFlagValueFormat          = "--%s must not be empty"
	sourceListSeparator           = ", "
)
