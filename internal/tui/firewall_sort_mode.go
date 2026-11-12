package tui

// FirewallSortMode selects how the rows of the Firewall ports table are ordered.
type FirewallSortMode int

// Firewall sort modes, in cycling order.
const (
	FirewallSortByVerdict FirewallSortMode = iota
	FirewallSortByPort
	FirewallSortByProcess
	firewallSortModeCount
)

// Next returns the following mode, wrapping around.
func (mode FirewallSortMode) Next() FirewallSortMode {
	return (mode + 1) % firewallSortModeCount
}

// String returns the display name of the mode.
func (mode FirewallSortMode) String() string {
	switch mode {
	case FirewallSortByVerdict:
		return firewallSortNameVerdict
	case FirewallSortByPort:
		return firewallSortNamePort
	case FirewallSortByProcess:
		return firewallSortNameProcess
	default:
		return firewallSortNameUnknown
	}
}
