package tui

// topRemoteView shows the two connection summary panels at full width. Like
// the bottom panels of the Live tab they follow the loopback and inbound-only
// toggles. The panels are as tall as their rows (bounded by the available
// height); the rest of the area stays empty.
func (model Model) topRemoteView(frame layout) []string {
	available := max(frame.panelLines-panelBorderLineCount, 0)
	if !model.connectionsEnabled() {
		return model.summaryPanel(panelTopRemoteTitle, []string{model.styles.dim.Render(connectionsDisabledMessage)}, frame.contentWidth, min(1, available))
	}
	needed := max(len(model.connectionSummary.TopRemoteAddresses), len(model.connectionSummary.PerLocalPort), 1)
	stackedAvailable := max(frame.panelLines/2-panelBorderLineCount, 0)
	return model.connectionPanels(frame.contentWidth, min(needed, available), min(needed, stackedAvailable))
}
