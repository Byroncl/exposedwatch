package domain

// Stable lowercase names for connection directions.
const (
	connectionDirectionNameInbound  = "inbound"
	connectionDirectionNameOutbound = "outbound"
	connectionDirectionNameUnknown  = "unknown"
)

// Stable lowercase names for connection states.
const (
	connectionStateNameEstablished = "established"
	connectionStateNameUnknown     = "unknown"
)

// Connection directions.
const (
	// ConnectionInbound is a connection to a port this host listens on.
	ConnectionInbound ConnectionDirection = iota
	// ConnectionOutbound is a connection this host opened to a remote peer.
	ConnectionOutbound
)

// Connection states.
const (
	// ConnectionEstablished is a fully open TCP connection.
	ConnectionEstablished ConnectionState = iota
)
