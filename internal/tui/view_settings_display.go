package tui

import "fmt"

// settingsDisplayPanel renders the Display inner tab of the Settings module:
// the current display preferences, which cannot be edited here.
func (model Model) settingsDisplayPanel(frame layout) []string {
	lines := []string{model.styles.heading.Render(settingsHeading)}
	lines = append(lines, model.settingsValueLines()...)
	lines = firstLines(lines, max(frame.panelLines-panelBorderLineCount, 0))
	return model.styles.renderPanel(" "+tabTitleSettingsDisplay+" ", lines, frame.contentWidth)
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
		fmt.Sprintf(settingsFirewallFormat, model.firewallSettingText()),
	}
}
