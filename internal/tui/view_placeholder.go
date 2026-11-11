package tui

import "fmt"

// placeholderPanel renders the honest "coming soon" screen of the modules that
// are not implemented yet. Settings also lists the current values, read-only.
func (model Model) placeholderPanel(frame layout) []string {
	innerWidth := panelContentWidth(frame.contentWidth)
	lines := []string{model.styles.warning.Bold(true).Render(placeholderHeading)}
	for _, line := range wrapText(placeholderText(model.activeModule), innerWidth) {
		lines = append(lines, model.styles.dim.Render(line))
	}
	if model.activeModule == ModuleSettings {
		lines = append(lines, "", model.styles.heading.Render(settingsHeading))
		lines = append(lines, model.settingsValueLines()...)
	}
	lines = firstLines(lines, max(frame.panelLines-panelBorderLineCount, 0))
	return model.styles.renderPanel(" "+model.activeModule.Title()+" ", lines, frame.contentWidth)
}

func placeholderText(module Module) string {
	switch module {
	case ModuleFirewall:
		return placeholderFirewallText
	case ModuleAlerts:
		return placeholderAlertsText
	default:
		return placeholderSettingsText
	}
}

// settingsValueLines lists the current preferences; none of them can be edited here.
func (model Model) settingsValueLines() []string {
	connections := settingsConnectionsOn
	if !model.connectionsEnabled() {
		connections = settingsConnectionsOff
	}
	return []string{
		fmt.Sprintf(settingsThemeFormat, model.theme),
		fmt.Sprintf(settingsIntervalFormat, model.interval),
		fmt.Sprintf(settingsConnectionsFormat, connections),
	}
}
