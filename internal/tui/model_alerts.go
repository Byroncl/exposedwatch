package tui

import "fmt"

// alertsReady tells whether the saved settings were read successfully.
func (model Model) alertsReady() bool {
	return model.settings.isAvailable() && model.settings.loaded && !model.settings.loadFailed
}

// alertsConfigured tells whether the saved settings hold a bot token and a
// chat id (from the file or the environment).
func (model Model) alertsConfigured() bool {
	return model.alertsReady() && model.settings.baseline.isConfigured()
}

// alertsActive tells whether `exposedwatch watch` would send alerts with the
// saved settings: enabled and configured. Unsaved edits do not count, because
// watch reads the saved file.
func (model Model) alertsActive() bool {
	return model.alertsConfigured() && model.settings.baseline.enabled
}

// alertsBadge is the sidebar note of the Alerts module: on when enabled and
// configured, off when disabled or not configured, n/a when the config could
// not be loaded, nothing until the first read finished.
func (model Model) alertsBadge() sidebarBadge {
	switch {
	case !model.settings.isAvailable():
		return sidebarBadge{text: sidebarBadgeOff}
	case !model.settings.loaded:
		return sidebarBadge{}
	case model.settings.loadFailed:
		return sidebarBadge{text: sidebarBadgeNotAvailable}
	case model.alertsActive():
		return sidebarBadge{text: sidebarBadgeOn}
	default:
		return sidebarBadge{text: sidebarBadgeOff}
	}
}

// overviewAlertsLine shows the alert switch and threshold on the Overview.
func (model Model) overviewAlertsLine() string {
	switch {
	case !model.settings.isAvailable():
		return columnSeparator + model.styles.dim.Render(alertsOverviewUnavailable)
	case !model.settings.loaded:
		return columnSeparator + model.styles.dim.Render(alertsOverviewReading)
	case model.settings.loadFailed:
		return columnSeparator + model.styles.warning.Render(alertsOverviewNotLoaded)
	}
	state := alertsWordOff
	if model.alertsActive() {
		state = alertsWordOn
	}
	return columnSeparator + model.styles.plain.Render(fmt.Sprintf(alertsOverviewFormat, state, model.settings.baseline.minSeverity))
}
