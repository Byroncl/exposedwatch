package app

import "github.com/Byroncl/exposedwatch/internal/domain"

// PortIdentity identifies a port across snapshots: protocol, port number,
// bind address and discovery source. Process and container names are
// descriptive only and never part of the identity.
type PortIdentity struct {
	Protocol    domain.Protocol
	Number      int
	BindAddress string
	Source      domain.Source
}

// IdentityOf returns the identity of a port.
func IdentityOf(port domain.Port) PortIdentity {
	return PortIdentity{
		Protocol:    port.Protocol,
		Number:      port.Number,
		BindAddress: port.BindAddress,
		Source:      port.Source,
	}
}
