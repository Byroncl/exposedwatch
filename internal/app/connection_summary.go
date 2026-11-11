package app

import (
	"sort"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

// PortConnectionCount is how many inbound connections one local port has.
type PortConnectionCount struct {
	LocalPort int
	Count     int
}

// RemoteAddressCount is how many inbound connections one remote IP has.
type RemoteAddressCount struct {
	Address string
	Count   int
}

// ConnectionSummary aggregates a connection snapshot. Everything is derived
// in memory; remote addresses are never stored or resolved.
type ConnectionSummary struct {
	Total    int
	Inbound  int
	Outbound int
	// PerLocalPort lists inbound connection counts, highest count first,
	// then lowest port.
	PerLocalPort []PortConnectionCount
	// TopRemoteAddresses lists the busiest inbound remote IPs, highest count
	// first, then lowest address, bounded by maxTopRemoteAddresses.
	TopRemoteAddresses []RemoteAddressCount
}

// SummarizeConnections counts connections by direction, local port and
// remote address. Ordering is deterministic.
func SummarizeConnections(connections []domain.Connection) ConnectionSummary {
	summary := ConnectionSummary{Total: len(connections)}
	countsByPort := map[int]int{}
	countsByRemote := map[string]int{}
	for _, connection := range connections {
		if connection.Direction != domain.ConnectionInbound {
			summary.Outbound++
			continue
		}
		summary.Inbound++
		countsByPort[connection.LocalPort]++
		countsByRemote[connection.RemoteAddress]++
	}
	for localPort, count := range countsByPort {
		summary.PerLocalPort = append(summary.PerLocalPort, PortConnectionCount{LocalPort: localPort, Count: count})
	}
	sort.Slice(summary.PerLocalPort, func(left, right int) bool {
		leftEntry, rightEntry := summary.PerLocalPort[left], summary.PerLocalPort[right]
		if leftEntry.Count != rightEntry.Count {
			return leftEntry.Count > rightEntry.Count
		}
		return leftEntry.LocalPort < rightEntry.LocalPort
	})
	for address, count := range countsByRemote {
		summary.TopRemoteAddresses = append(summary.TopRemoteAddresses, RemoteAddressCount{Address: address, Count: count})
	}
	sort.Slice(summary.TopRemoteAddresses, func(left, right int) bool {
		leftEntry, rightEntry := summary.TopRemoteAddresses[left], summary.TopRemoteAddresses[right]
		if leftEntry.Count != rightEntry.Count {
			return leftEntry.Count > rightEntry.Count
		}
		return leftEntry.Address < rightEntry.Address
	})
	if len(summary.TopRemoteAddresses) > maxTopRemoteAddresses {
		summary.TopRemoteAddresses = summary.TopRemoteAddresses[:maxTopRemoteAddresses]
	}
	return summary
}
