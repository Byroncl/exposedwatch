package tui

// Module is one entry of the sidebar. Each module owns a set of inner tabs.
type Module int

// Modules, in sidebar order.
const (
	ModuleOverview Module = iota
	ModulePorts
	ModuleConnections
	ModuleFirewall
	ModuleAlerts
	ModuleSettings
	moduleCount
)

// innerTabSpec is one inner tab of a module: the view it shows and its title.
type innerTabSpec struct {
	view  TabID
	title string
}

// moduleSpec describes one module.
type moduleSpec struct {
	title        string
	compactLabel string
	innerTabs    []innerTabSpec
}

var moduleSpecs = [moduleCount]moduleSpec{
	ModuleOverview: {
		title: moduleTitleOverview, compactLabel: moduleCompactOverview,
		innerTabs: []innerTabSpec{{TabSummary, tabTitleSummary}},
	},
	ModulePorts: {
		title: moduleTitlePorts, compactLabel: moduleCompactPorts,
		innerTabs: []innerTabSpec{{TabFindings, tabTitleFindings}, {TabAllPorts, tabTitleAllPorts}, {TabChanges, tabTitleChanges}},
	},
	ModuleConnections: {
		title: moduleTitleConnections, compactLabel: moduleCompactConnections,
		innerTabs: []innerTabSpec{{TabConnections, tabTitleLive}, {TabTopRemote, tabTitleTopRemote}},
	},
	ModuleFirewall: {
		title: moduleTitleFirewall, compactLabel: moduleCompactFirewall,
		innerTabs: []innerTabSpec{{TabFirewall, tabTitleFirewallStatus}, {TabFirewallPorts, tabTitleFirewallPorts}},
	},
	ModuleAlerts: {
		title: moduleTitleAlerts, compactLabel: moduleCompactAlerts,
		innerTabs: []innerTabSpec{{TabAlertsStatus, tabTitleAlertsStatus}, {TabAlertsPreview, tabTitleAlertsPreview}},
	},
	ModuleSettings: {
		title: moduleTitleSettings, compactLabel: moduleCompactSettings,
		innerTabs: []innerTabSpec{{TabSettingsTelegram, tabTitleSettingsTelegram}, {TabSettingsDisplay, tabTitleSettingsDisplay}},
	},
}

// Next returns the following module, wrapping around.
func (module Module) Next() Module {
	return (module + 1) % moduleCount
}

// Previous returns the preceding module, wrapping around.
func (module Module) Previous() Module {
	return (module + moduleCount - 1) % moduleCount
}

// Title returns the display name of the module.
func (module Module) Title() string {
	return moduleSpecs[module].title
}

// String makes modules readable in messages.
func (module Module) String() string {
	return module.Title()
}

// compactLabel returns the short label of the compact sidebar.
func (module Module) compactLabel() string {
	return moduleSpecs[module].compactLabel
}

// innerTabs lists the inner tabs of the module in display order.
func (module Module) innerTabs() []innerTabSpec {
	return moduleSpecs[module].innerTabs
}
