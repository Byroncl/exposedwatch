// Package domain holds the pure business types of the exposure scanner.
package domain

import "net/netip"

// Protocol is a transport protocol.
type Protocol string

// Supported transport protocols.
const (
	TCP Protocol = "tcp"
	UDP Protocol = "udp"
)

// Source tells where a port was discovered.
type Source string

// Known port sources.
const (
	SourceHost     Source = "host"
	SourceDocker   Source = "docker"
	SourceExternal Source = "external"
)

// Exposure describes how far a bind address reaches.
// Values are ordered from least to most exposed.
type Exposure int

// Exposure levels.
const (
	Loopback Exposure = iota
	Private
	AllInterfaces
	Public
)

// String returns a stable lowercase name.
func (e Exposure) String() string {
	switch e {
	case Loopback:
		return "loopback"
	case Private:
		return "private"
	case AllInterfaces:
		return "all-interfaces"
	case Public:
		return "public"
	default:
		return "unknown"
	}
}

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
func (p Port) Exposure() Exposure {
	addr, err := netip.ParseAddr(p.BindAddress)
	if err != nil {
		return AllInterfaces
	}
	addr = addr.WithZone("").Unmap()
	switch {
	case addr.IsUnspecified():
		return AllInterfaces
	case addr.IsLoopback():
		return Loopback
	case addr.IsPrivate(), addr.IsLinkLocalUnicast():
		return Private
	default:
		return Public
	}
}
