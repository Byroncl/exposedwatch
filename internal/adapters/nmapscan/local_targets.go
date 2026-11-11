package nmapscan

import (
	"fmt"
	"net"
	"net/netip"
)

// AddressLister returns the local interface addresses. It matches
// net.InterfaceAddrs so tests can inject a fake.
type AddressLister func() ([]net.Addr, error)

// LocalTargets returns this host's own non-loopback, non-link-local IPv4
// addresses: the safe default scan targets.
func LocalTargets(lister AddressLister) (TargetList, error) {
	addresses, err := lister()
	if err != nil {
		return nil, fmt.Errorf(listAddressesFormat, err)
	}
	seen := map[netip.Addr]bool{}
	var targets TargetList
	for _, networkAddress := range addresses {
		address, ok := ipv4Of(networkAddress)
		if !ok || seen[address] || !isScannableLocalAddress(address) {
			continue
		}
		seen[address] = true
		targets = append(targets, Target(address.String()))
	}
	if len(targets) == 0 {
		return nil, ErrNoTargets
	}
	return targets, nil
}

func ipv4Of(networkAddress net.Addr) (netip.Addr, bool) {
	var ipAddress net.IP
	switch typed := networkAddress.(type) {
	case *net.IPNet:
		ipAddress = typed.IP
	case *net.IPAddr:
		ipAddress = typed.IP
	default:
		return netip.Addr{}, false
	}
	ipv4 := ipAddress.To4()
	if ipv4 == nil {
		return netip.Addr{}, false
	}
	address, ok := netip.AddrFromSlice(ipv4)
	return address, ok
}

func isScannableLocalAddress(address netip.Addr) bool {
	return !address.IsLoopback() && !address.IsLinkLocalUnicast() &&
		!address.IsUnspecified() && !address.IsMulticast()
}
