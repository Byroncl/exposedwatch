package tui

// TabID selects which table the main panel shows.
type TabID int

// Tabs, in display order.
const (
	TabFindings TabID = iota
	TabAllPorts
	TabChanges
	TabConnections
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
	case TabConnections:
		return tabTitleConnections
	default:
		return tabTitleChanges
	}
}

// supportsMinimumSeverity tells whether the minimum severity filter applies.
func (tab TabID) supportsMinimumSeverity() bool {
	return tab == TabFindings || tab == TabAllPorts
}

// supportsSortAndSearch tells whether sort and free-text search apply.
func (tab TabID) supportsSortAndSearch() bool {
	return tab != TabChanges
}
