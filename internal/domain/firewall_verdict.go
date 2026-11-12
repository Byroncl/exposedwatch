package domain

// FirewallVerdict says whether a new inbound connection to a port would be
// let through by the host firewall. It is information only.
type FirewallVerdict int

// String returns a stable lowercase name.
func (verdict FirewallVerdict) String() string {
	switch verdict {
	case FirewallAllowed:
		return firewallNameAllowed
	case FirewallRestricted:
		return firewallNameRestricted
	case FirewallBlocked:
		return firewallNameBlocked
	case FirewallNotExposed:
		return firewallNameNotExposed
	case FirewallDockerPublished:
		return firewallNameDockerPublished
	default:
		return firewallNameUnknown
	}
}
