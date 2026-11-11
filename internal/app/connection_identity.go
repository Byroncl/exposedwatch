package app

import "github.com/Byroncl/exposedwatch/internal/domain"

// ConnectionIdentity identifies a connection across snapshots: protocol and
// both endpoints. State, direction and process name are descriptive only.
type ConnectionIdentity struct {
	Protocol      domain.Protocol
	LocalAddress  string
	LocalPort     int
	RemoteAddress string
	RemotePort    int
}

// ConnectionIdentityOf returns the identity of a connection.
func ConnectionIdentityOf(connection domain.Connection) ConnectionIdentity {
	return ConnectionIdentity{
		Protocol:      connection.Protocol,
		LocalAddress:  connection.LocalAddress,
		LocalPort:     connection.LocalPort,
		RemoteAddress: connection.RemoteAddress,
		RemotePort:    connection.RemotePort,
	}
}

// ConnectionIdentityLess orders identities by local port, local address,
// remote address, remote port, protocol.
func ConnectionIdentityLess(left, right ConnectionIdentity) bool {
	if left.LocalPort != right.LocalPort {
		return left.LocalPort < right.LocalPort
	}
	if left.LocalAddress != right.LocalAddress {
		return left.LocalAddress < right.LocalAddress
	}
	if left.RemoteAddress != right.RemoteAddress {
		return left.RemoteAddress < right.RemoteAddress
	}
	if left.RemotePort != right.RemotePort {
		return left.RemotePort < right.RemotePort
	}
	return left.Protocol < right.Protocol
}
