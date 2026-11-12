package domain

// FirewallFinding is the firewall verdict for one port.
type FirewallFinding struct {
	Port    Port
	Verdict FirewallVerdict
	Reason  string
}

// FirewallReport collects the findings of one firewall evaluation.
type FirewallReport struct {
	// Backend names the firewall tool the rules were read from.
	Backend string
	// Findings are ordered by port, protocol and bind address.
	Findings []FirewallFinding
	// Notes are host-level caveats worth showing next to the findings.
	Notes []string
	// Available is false when the rules could not be read at all.
	Available bool
}
