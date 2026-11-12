package tui

// Sidebar widths, in cells, and the terminal widths that select them.
const (
	sidebarFullMinWidth    = 80
	sidebarCompactMinWidth = 50
	sidebarFullWidth       = 19
	sidebarCompactWidth    = 5
	// sidebarContentGap is the blank cells between the sidebar and the content area.
	sidebarContentGap = 1
	sidebarBorderSize = 2
)

// Sidebar text.
const (
	sidebarMarkerGap    = " "
	sidebarBadgeJoin    = " "
	sidebarBadgeCap     = 99
	sidebarBadgeCapText = "99+"
	sidebarBadgeOff     = "off"
	sidebarBadgeNew     = "NEW"
	sidebarCountFormat  = "%d"
	sidebarNoMarker     = " "
)

// Sidebar mode names.
const (
	sidebarModeNameFull    = "full"
	sidebarModeNameCompact = "compact"
	sidebarModeNameHidden  = "hidden"
)
