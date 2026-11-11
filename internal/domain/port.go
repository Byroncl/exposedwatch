package domain

import "net/netip"

// Port is a listening or reachable port.
type Port struct {
	Number      int
	Protocol    Protocol
	BindAddress string
	Process     string // optional
	Source      Source
	Container   string // optional
}

// Exposure classifies the bind address. Empty, wildcard and unparseable
// addresses are treated as all interfaces (fail-safe: assume the widest).
func (port Port) Exposure() Exposure {
	address, err := netip.ParseAddr(port.BindAddress)
	if err != nil {
		return AllInterfaces
	}
	address = address.WithZone("").Unmap()
	switch {
	case address.IsUnspecified():
		return AllInterfaces
	case address.IsLoopback():
		return Loopback
	case address.IsPrivate(), address.IsLinkLocalUnicast():
		return Private
	default:
		return Public
	}
}
