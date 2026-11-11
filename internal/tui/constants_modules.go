package tui

// Module titles, shown in the sidebar and as panel titles.
const (
	moduleTitleOverview    = "Overview"
	moduleTitlePorts       = "Ports"
	moduleTitleConnections = "Connections"
	moduleTitleFirewall    = "Firewall"
	moduleTitleAlerts      = "Alerts"
	moduleTitleSettings    = "Settings"
)

// Module labels of the compact sidebar.
const (
	moduleCompactOverview    = "O"
	moduleCompactPorts       = "P"
	moduleCompactConnections = "C"
	moduleCompactFirewall    = "F"
	moduleCompactAlerts      = "A"
	moduleCompactSettings    = "S"
)

// Inner tab titles that are not module names.
const (
	tabTitleSummary   = "Summary"
	tabTitleLive      = "Live"
	tabTitleTopRemote = "Top remote"
	// Firewall inner tabs.
	tabTitleFirewallStatus = "Status"
	tabTitleFirewallPorts  = "Ports"
)

// Placeholder modules.
const (
	placeholderHeading        = "Coming soon"
	placeholderAlertsText     = "Will send Telegram notifications when a new exposed port or a serious finding appears."
	placeholderSettingsText   = "Will hold the notification and display preferences."
	settingsHeading           = "Current values (read-only)"
	settingsThemeFormat       = "Theme:            %s"
	settingsIntervalFormat    = "Refresh interval: %s"
	settingsConnectionsFormat = "Connections:      %s"
	settingsConnectionsOn     = "enabled"
	settingsConnectionsOff    = "disabled (started with --no-connections)"
)
