package tui

// TabID selects which table the main panel shows.
type TabID int

// Tabs, in display order.
const (
	TabFindings TabID = iota
	TabAllPorts
	TabChanges
	tabCount
)

// Next returns the following tab, wrapping around.
func (tab TabID) Next() TabID {
	return (tab + 1) % tabCount
}

// Previous returns the preceding tab, wrapping around.
func (tab TabID) Previous() TabID {
	return (tab + tabCount - 1) % tabCount
}

// Title returns the display name of the tab.
func (tab TabID) Title() string {
	switch tab {
	case TabFindings:
		return tabTitleFindings
	case TabAllPorts:
		return tabTitleAllPorts
	default:
		return tabTitleChanges
	}
}

// supportsFiltering tells whether sort, severity filter and search apply.
func (tab TabID) supportsFiltering() bool {
	return tab != TabChanges
}
