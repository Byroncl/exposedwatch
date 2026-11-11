package domain

// ConnectionDirection tells who opened a connection relative to this host.
type ConnectionDirection int

// String returns a stable lowercase name.
func (direction ConnectionDirection) String() string {
	switch direction {
	case ConnectionInbound:
		return connectionDirectionNameInbound
	case ConnectionOutbound:
		return connectionDirectionNameOutbound
	default:
		return connectionDirectionNameUnknown
	}
}
