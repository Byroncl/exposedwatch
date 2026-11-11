package tui

// SidebarMode tells how much of the sidebar the terminal shows.
type SidebarMode int

// Sidebar modes.
const (
	SidebarHidden SidebarMode = iota
	SidebarCompact
	SidebarFull
)

// sidebarModeFor picks the sidebar mode from the terminal width. The user's
// hide preference wins over the width.
func sidebarModeFor(width int, hiddenByUser bool) SidebarMode {
	switch {
	case hiddenByUser:
		return SidebarHidden
	case width >= sidebarFullMinWidth:
		return SidebarFull
	case width >= sidebarCompactMinWidth:
		return SidebarCompact
	default:
		return SidebarHidden
	}
}

// cellWidth is the outer width of the sidebar, border included.
func (mode SidebarMode) cellWidth() int {
	switch mode {
	case SidebarFull:
		return sidebarFullWidth
	case SidebarCompact:
		return sidebarCompactWidth
	default:
		return 0
	}
}

// String names the mode.
func (mode SidebarMode) String() string {
	switch mode {
	case SidebarFull:
		return sidebarModeNameFull
	case SidebarCompact:
		return sidebarModeNameCompact
	default:
		return sidebarModeNameHidden
	}
}
