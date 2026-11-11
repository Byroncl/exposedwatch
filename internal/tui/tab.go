package tui

// TabID identifies one view of the content area: the screen an inner tab shows.
type TabID int

// Views, one per inner tab of the modules.
const (
	TabFindings TabID = iota
	TabAllPorts
	TabChanges
	TabConnections
	TabSummary
	TabTopRemote
	TabFirewall
	TabAlerts
	TabSettings
)

// Title returns the display name of the view.
func (tab TabID) Title() string {
	switch tab {
	case TabFindings:
		return tabTitleFindings
	case TabAllPorts:
		return tabTitleAllPorts
	case TabConnections:
		return tabTitleConnections
	case TabSummary:
		return tabTitleSummary
	case TabTopRemote:
		return tabTitleTopRemote
	case TabFirewall:
		return moduleTitleFirewall
	case TabAlerts:
		return moduleTitleAlerts
	case TabSettings:
		return moduleTitleSettings
	default:
		return tabTitleChanges
	}
}

// isTable tells whether the view is a table of rows with a cursor.
func (tab TabID) isTable() bool {
	switch tab {
	case TabFindings, TabAllPorts, TabChanges, TabConnections:
		return true
	default:
		return false
	}
}

// usesConnections tells whether the view shows connection data and follows the
// loopback and inbound-only toggles.
func (tab TabID) usesConnections() bool {
	return tab == TabConnections || tab == TabTopRemote
}

// supportsMinimumSeverity tells whether the minimum severity filter applies.
func (tab TabID) supportsMinimumSeverity() bool {
	return tab == TabFindings || tab == TabAllPorts
}

// supportsSortAndSearch tells whether sort and free-text search apply.
func (tab TabID) supportsSortAndSearch() bool {
	return tab == TabFindings || tab == TabAllPorts || tab == TabConnections
}
