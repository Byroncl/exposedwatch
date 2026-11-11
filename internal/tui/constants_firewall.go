package tui

import "time"

// Firewall reads.
const (
	// firewallTimeout bounds one firewall evaluation so a stuck iptables-save
	// cannot keep the next one from starting.
	firewallTimeout = 20 * time.Second
	// firewallRefreshInterval is the least time between two evaluations that
	// are not caused by a key press or a change of the listening ports. It is
	// independent from the scan interval: iptables-save is not free.
	firewallRefreshInterval = 30 * time.Second
)

// Verdict words shown in the tables, the chips and the details.
const (
	firewallWordAllowed      = "allowed"
	firewallWordRestricted   = "restricted"
	firewallWordBlocked      = "blocked"
	firewallWordUnknown      = "unknown"
	firewallWordNotExposed   = "not-exposed"
	firewallWordDockerShort  = "docker"
	firewallLegendSeparator  = ": "
	firewallChipCountFormat  = "%s %d"
	firewallOverviewFormat   = "Firewall: %d allowed, %d restricted, %d blocked, %d unknown"
	firewallOverviewMissing  = "Firewall: not available (%s)"
	firewallOverviewDisabled = "Firewall: checks disabled"
	firewallOverviewReading  = "Firewall: reading rules..."
	firewallOverviewNoReason = "rules could not be read"
)

// Separators of the port set signature.
const (
	firewallSignatureJoin = "|"
	firewallEntryJoin     = ";"
)

// Firewall sort mode names.
const (
	firewallSortNameVerdict = "verdict"
	firewallSortNamePort    = "port"
	firewallSortNameProcess = "process"
	firewallSortNameUnknown = "unknown"
)

// Firewall table headings and widths.
const (
	headingVerdict       = "VERDICT"
	headingProtocol      = "PROTO"
	headingReason        = "REASON"
	headingFirewallShort = "FW"

	verdictColumnWidth     = 13
	portNumberColumnWidth  = 6
	protocolColumnWidth    = 5
	firewallColumnWidth    = 11
	minReasonWidth         = 24
	firewallPortsTitleText = "Firewall ports"
)

// Firewall screens.
const (
	panelFirewallTitle     = " Firewall "
	panelFirewallStatusFmt = "· sort:%s"
	firewallDisabledText   = "Firewall checks are disabled (started with --no-firewall)."
	firewallReadingText    = "Reading firewall rules..."
	firewallFailedText     = "The firewall could not be evaluated yet; see the warning above."
	noFirewallPortsText    = "No host or Docker ports to evaluate."
	firewallWarningFormat  = "! firewall refresh failed: %v"
	firewallBackendFormat  = "Backend: %s   Available: %s"
	firewallFreshFormat    = "Last evaluated: %s (rechecked every %s, on r and when the ports change)"
	firewallAvailableYes   = "yes"
	firewallAvailableNo    = "no (the rules could not be read)"
	firewallVerdictsLabel  = "Verdicts: "
	firewallNotesHeading   = "Notes"
	firewallNoteFormat     = "- %s"
	firewallLegendHeading  = "What the verdicts mean"
	firewallInformational  = "Verdicts are informational and never change severities."
	firewallEstimateNote   = "They estimate a NEW inbound connection from an external source."
)

// Firewall legend: what each verdict honestly means.
const (
	legendAllowed    = "a new inbound connection from outside would be accepted."
	legendRestricted = "accepted only from some source addresses."
	legendBlocked    = "dropped or rejected by the rules that were read."
	legendUnknown    = "could not be determined (rules unreadable or not understood)."
	legendNotExposed = "listens on loopback only, so it is not reachable from outside."
	legendDocker     = "published by Docker, which usually bypasses INPUT and ufw rules; never reported as blocked."
)

// Firewall details and sidebar.
const (
	detailFirewallFormat       = "Firewall: %s"
	detailFirewallReasonFormat = "Firewall: %s - %s"
	detailReasonFormat         = "Reason: %s"
	sidebarBadgeNotAvailable   = "n/a"
	settingsFirewallFormat     = "Firewall:         %s"
	settingsFirewallOn         = "enabled"
	settingsFirewallOff        = "disabled (started with --no-firewall)"
)
