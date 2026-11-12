package main

import (
	"github.com/Byroncl/exposedwatch/internal/adapters/nmapscan"
	"github.com/Byroncl/exposedwatch/internal/tui"
)

// ExitCode is the process exit status.
type ExitCode int

// Exit codes. Scheme: 0 clean, 1 runtime failure, 2 bad usage, 3 findings
// at or above --min-severity were found (lets cron/CI gate on it). With
// "scan --baseline" code 3 is never used: 4 means new ports not present in
// the baseline were found; a source that failed during the scan gives 1 unless 4 applies.
const (
	ExitOK           ExitCode = 0
	ExitRuntimeError ExitCode = 1
	ExitUsage        ExitCode = 2
	ExitFindings     ExitCode = 3
	ExitDrift        ExitCode = 4
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
	commandScan       = "scan"
	commandTui        = "tui"
	commandHelp       = "help"
	commandWatch      = "watch"
	commandNotifyTest = "notify-test"

	// Top-level flags handled before the tui takes over a leading flag.
	flagPrefix      = "-"
	flagHelpShort   = "-h"
	flagHelpLong    = "--help"
	flagVersionLong = "--version"
	commandVersion  = "version"
)

// Flag names.
const (
	flagFormat        = "format"
	flagMinSeverity   = "min-severity"
	flagProcRoot      = "proc-root"
	flagDockerSocket  = "docker-socket"
	flagNoDocker      = "no-docker"
	flagNoConnections = "no-connections"
	flagNoFirewall    = "no-firewall"
	flagFirewall      = "firewall"
	flagInterval      = "interval"
	flagTheme         = "theme"

	flagExternal         = "external"
	flagExternalTarget   = "external-target"
	flagConfirmOwnership = "confirm-ownership"
	flagExternalPorts    = "external-ports"
	flagExternalTimeout  = "external-timeout"
	flagNmapPath         = "nmap-path"
)

// Flag defaults.
const (
	defaultFormat      = FormatTable
	defaultMinSeverity = "info"
	defaultTheme       = string(tui.ThemeAuto)
	defaultProcRoot    = "/proc"
	defaultDockerSock  = "/var/run/docker.sock"
)

// Flag descriptions.
const (
	usageFormat        = "output format: table or json"
	usageMinSeverity   = "minimum severity to display and to trigger exit code 3 (exit code 4 with --baseline): info, low, medium, high or critical"
	usageProcRoot      = "path of the proc filesystem"
	usageDockerSocket  = "path of the Docker Engine unix socket"
	usageNoDocker      = "skip Docker container discovery"
	usageNoConnections = "disable the Connections tab (no connection reads)"
	usageNoFirewall    = "disable the Firewall module (no firewall reads)"
	usageFirewall      = "add the firewall verdict of every listening port (reads iptables-save, needs root; information only, never changes severities or exit codes)"
	usageInterval      = "auto-refresh interval of the tui, minimum 1s"
	usageTheme         = "color theme of the tui: auto (detect the terminal background), dark or light"

	usageExternal = "also scan this machine from the network side with nmap (TCP connect scan only, UDP is not scanned); " +
		"by default only this host's own non-loopback IPv4 addresses are scanned; requires nmap to be installed"
	usageExternalTarget = "additional nmap target: IP, CIDR (IPv4 /24 or narrower) or hostname; comma separated or repeatable; " +
		"requires --external and --confirm-ownership"
	usageConfirmOwnership = "confirm that you own, or have permission to scan, every --external-target; " +
		"scanning systems you do not own or have no permission to scan may be illegal"
	usageExternalPorts   = "ports for the external scan, e.g. 22,80,443,8000-8100 (default: nmap's top 1000 TCP ports)"
	usageExternalTimeout = "time limit of the external scan, between 5s and 30m"
	usageNmapPath        = "path or name of the nmap executable"
)

// Flag defaults for the external scan.
const defaultNmapPath = nmapscan.BinaryName

// Environment variables.
const envNoColor = "NO_COLOR"

// Messages.
const (
	usageText = `Usage: exposedwatch [command] [flags]

Run without a command to open the interactive view (same as "tui").

Commands:
  tui         interactive full-screen view that refreshes automatically (default)
  scan        one-shot report for scripts and cron (table or JSON);
              add --external to also scan from the network side with nmap,
              add --firewall to show the host firewall verdict of each port,
              add --baseline FILE to report only drift (for cron)
  baseline    "baseline save" records the ports exposed now as an accepted baseline
  watch       headless loop that sends Telegram alerts for new risky ports (--baseline FILE: only drift)
              (this is the only feature that sends data to a third party)
  notify-test send one test message to the configured Telegram chat
  version     print the version
  help        show this help

Run "exposedwatch tui -h", "exposedwatch scan -h", "exposedwatch baseline save -h",
"exposedwatch watch -h" or "exposedwatch notify-test -h" for flags.
The tui accepts --interval, --theme auto|dark|light, --no-connections, --no-firewall
and --config (the file edited by Settings > Telegram).
Exit codes: 0 ok, 1 runtime error, 2 usage error, 3 findings found,
4 drift found (only with scan --baseline, which never exits with 3; with it,
1 also means a source could not be read, unless new ports give 4).
`
	versionFormat        = "exposedwatch %s"
	scanFlagSetName      = "scan"
	tuiFlagSetName       = "tui"
	nonTerminalMessage   = "the tui command needs an interactive terminal (stdout is not a terminal); use \"exposedwatch scan\" for non-interactive output"
	invalidThemeFmt      = "invalid --theme: %v"
	invalidIntervalFmt   = "invalid --interval %v: must be at least %v"
	tuiFailedFormat      = "tui failed: %v"
	invalidFormatMessage = "invalid --format %q: want table or json"
	invalidSeverityFmt   = "invalid --min-severity: %v"
	unexpectedArgsFmt    = "unexpected arguments: %v"
	scanFailedFormat     = "scan failed: %v"
	renderFailedFormat   = "rendering failed: %v"

	firewallReadFailedFormat = "firewall could not be read: %v"

	externalTargetNeedsExternalMessage = "--external-target requires --external"
	externalTargetNeedsConfirmMessage  = "--external-target requires --confirm-ownership: confirm that you own, " +
		"or have permission to scan, the targets (scanning systems you do not own may be illegal)"
	invalidExternalTargetFmt  = "invalid --external-target: %v"
	invalidExternalPortsFmt   = "invalid --external-ports: %v"
	invalidExternalTimeoutFmt = "invalid --external-timeout: %v"
	listSeparator             = ","
)
