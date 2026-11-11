package domain

import "net/netip"

// Connection is one open connection between this host and a peer.
type Connection struct {
	Protocol      Protocol
	LocalAddress  string
	LocalPort     int
	RemoteAddress string
	RemotePort    int
	State         ConnectionState
	Direction     ConnectionDirection
	ProcessName   string // optional
}

// IsLoopbackToLoopback reports whether both endpoints are loopback addresses,
// which is local traffic that never crosses the network. Unparseable
// addresses are not considered loopback.
func (connection Connection) IsLoopbackToLoopback() bool {
	return isLoopbackAddress(connection.LocalAddress) && isLoopbackAddress(connection.RemoteAddress)
}

func isLoopbackAddress(rawAddress string) bool {
	address, err := netip.ParseAddr(rawAddress)
	if err != nil {
		return false
	}
	return address.WithZone("").Unmap().IsLoopback()
}

// NormalizeAddress rewrites IPv4-mapped IPv6 addresses (::ffff:a.b.c.d) as
// plain IPv4 so the same peer always has the same textual form. Anything
// that is not a mapped address is returned unchanged.
func NormalizeAddress(rawAddress string) string {
	address, err := netip.ParseAddr(rawAddress)
	if err != nil || !address.Is4In6() {
		return rawAddress
	}
	return address.Unmap().String()
}
