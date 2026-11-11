package tui

// ConnectionSortMode selects how the Connections table rows are ordered.
type ConnectionSortMode int

// Connection sort modes, in cycling order.
const (
	ConnectionSortByDirection ConnectionSortMode = iota
	ConnectionSortByLocalPort
	ConnectionSortByRemoteAddress
	ConnectionSortByProcess
	ConnectionSortByAge
	connectionSortModeCount
)

// Next returns the following mode, wrapping around.
func (mode ConnectionSortMode) Next() ConnectionSortMode {
	return (mode + 1) % connectionSortModeCount
}

// String returns the display name of the mode.
func (mode ConnectionSortMode) String() string {
	switch mode {
	case ConnectionSortByDirection:
		return connectionSortNameDirection
	case ConnectionSortByLocalPort:
		return connectionSortNameLocalPort
	case ConnectionSortByRemoteAddress:
		return connectionSortNameRemote
	case ConnectionSortByProcess:
		return connectionSortNameProcess
	case ConnectionSortByAge:
		return connectionSortNameAge
	default:
		return connectionSortNameUnknown
	}
}
