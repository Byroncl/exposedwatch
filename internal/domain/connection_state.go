package domain

// ConnectionState is the lifecycle state of a connection.
type ConnectionState int

// String returns a stable lowercase name.
func (state ConnectionState) String() string {
	switch state {
	case ConnectionEstablished:
		return connectionStateNameEstablished
	default:
		return connectionStateNameUnknown
	}
}
