package tui

// SortMode selects how table rows are ordered.
type SortMode int

// Sort modes, in cycling order.
const (
	SortBySeverity SortMode = iota
	SortByPort
	SortByProcess
	sortModeCount
)

// Next returns the following mode, wrapping around.
func (mode SortMode) Next() SortMode {
	return (mode + 1) % sortModeCount
}

// String returns the display name of the mode.
func (mode SortMode) String() string {
	switch mode {
	case SortBySeverity:
		return sortModeNameSeverity
	case SortByPort:
		return sortModeNamePort
	case SortByProcess:
		return sortModeNameProcess
	default:
		return sortModeNameUnknown
	}
}
