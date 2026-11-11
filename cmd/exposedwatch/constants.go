package main

// ExitCode is the process exit status.
type ExitCode int

// Exit codes. Scheme: 0 clean, 1 runtime failure, 2 bad usage, 3 findings
// at or above --min-severity were found (lets cron/CI gate on it).
const (
	ExitOK           ExitCode = 0
	ExitRuntimeError ExitCode = 1
	ExitUsage        ExitCode = 2
	ExitFindings     ExitCode = 3
)

// OutputFormat selects how a report is rendered.
type OutputFormat string

// Supported output formats.
const (
	FormatTable OutputFormat = "table"
	FormatJSON  OutputFormat = "json"
)

// Subcommand names.
const (
	commandScan    = "scan"
	commandVersion = "version"
)

// Flag names.
const (
	flagFormat       = "format"
	flagMinSeverity  = "min-severity"
	flagProcRoot     = "proc-root"
	flagDockerSocket = "docker-socket"
	flagNoDocker     = "no-docker"
)

// Flag defaults.
const (
	defaultFormat      = FormatTable
	defaultMinSeverity = "info"
	defaultProcRoot    = "/proc"
	defaultDockerSock  = "/var/run/docker.sock"
)

// Flag descriptions.
const (
	usageFormat       = "output format: table or json"
	usageMinSeverity  = "minimum severity to display and to trigger exit code 3: info, low, medium, high or critical"
	usageProcRoot     = "path of the proc filesystem"
	usageDockerSocket = "path of the Docker Engine unix socket"
	usageNoDocker     = "skip Docker container discovery"
)

// Environment variables.
const envNoColor = "NO_COLOR"

// Messages.
const (
	usageText = `Usage: exposedwatch <command> [flags]

Commands:
  scan      list listening ports and flag risky exposure
  version   print the version

Run "exposedwatch scan -h" for scan flags.
Exit codes: 0 ok, 1 runtime error, 2 usage error, 3 findings found.
`
	versionFormat        = "exposedwatch %s"
	scanFlagSetName      = "scan"
	invalidFormatMessage = "invalid --format %q: want table or json"
	invalidSeverityFmt   = "invalid --min-severity: %v"
	unexpectedArgsFmt    = "unexpected arguments: %v"
	scanFailedFormat     = "scan failed: %v"
	renderFailedFormat   = "rendering failed: %v"
)
