package domain

import "net/netip"

// PortRange is an inclusive range of ports. A single port has First == Last.
type PortRange struct {
	First int
	Last  int
}

// Contains reports whether port lies inside the range.
func (portRange PortRange) Contains(port int) bool {
	return port >= portRange.First && port <= portRange.Last
}

// TCPFlagsMatch is the --tcp-flags match: of the flags in Mask, exactly the
// flags in Set must be set.
type TCPFlagsMatch struct {
	Mask []TCPFlag
	Set  []TCPFlag
}

// Rule is one parsed iptables rule. Matches the engine cannot evaluate set
// Unsupported instead of being dropped, so that no rule is ever silently
// ignored.
type Rule struct {
	// Text is the original rule line, for explanations.
	Text string
	// Protocol is empty when the rule matches any protocol.
	Protocol         Protocol
	DestinationPorts []PortRange
	SourceCIDRs      []netip.Prefix
	DestinationCIDRs []netip.Prefix
	InInterface      InterfaceName
	OutInterface     InterfaceName
	// ConnectionStates is empty when the rule matches any state.
	ConnectionStates []ConntrackState
	AddressTypes     []AddressType
	TCPFlags         *TCPFlagsMatch
	// Target is empty for rules without -j (counting rules). A jump to a user
	// chain is expressed through JumpChain instead.
	Target            TargetName
	JumpChain         ChainName
	GotoChain         ChainName
	Unsupported       bool
	UnsupportedReason string
}
