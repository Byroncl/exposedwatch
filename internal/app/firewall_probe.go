package app

import (
	"net/netip"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

// connectionProbe describes the hypothetical connection the engine evaluates:
// a new inbound connection from a non-loopback source to one listening socket.
type connectionProbe struct {
	port     int
	protocol domain.Protocol
	// bind is the address the socket is bound to; hasBind is false for a
	// wildcard bind.
	bind    netip.Addr
	hasBind bool
}

// parseBind splits a bind address into its address and wildcard flag. Empty
// and unparseable addresses are wildcard IPv6 binds so both families are
// evaluated (fail-safe).
func parseBind(bindAddress string) (address netip.Addr, specific bool, families []domain.AddressFamily) {
	parsed, err := netip.ParseAddr(bindAddress)
	if err != nil {
		return netip.Addr{}, false, bothFamilies()
	}
	parsed = parsed.WithZone("").Unmap()
	switch {
	case parsed.IsUnspecified() && parsed.Is4():
		return netip.Addr{}, false, []domain.AddressFamily{domain.FamilyIPv4}
	case parsed.IsUnspecified():
		return netip.Addr{}, false, bothFamilies()
	case parsed.Is4():
		return parsed, true, []domain.AddressFamily{domain.FamilyIPv4}
	default:
		return parsed, true, []domain.AddressFamily{domain.FamilyIPv6}
	}
}

func bothFamilies() []domain.AddressFamily {
	return []domain.AddressFamily{domain.FamilyIPv4, domain.FamilyIPv6}
}
