package tui

// Alerts words and badges.
const (
	sidebarBadgeOn = "on"
	alertsWordOn   = "on"
	alertsWordOff  = "off"
)

// Overview line.
const (
	alertsOverviewFormat      = "Alerts: %s (min %s)"
	alertsOverviewUnavailable = "Alerts: off (settings unavailable)"
	alertsOverviewReading     = "Alerts: reading..."
	alertsOverviewNotLoaded   = "Alerts: n/a (config not loaded)"
)

// Panel titles.
const (
	panelAlertsStatusTitle  = " Alerts "
	panelAlertsPreviewTitle = " Alerts preview "
)

// Status tab.
const (
	alertsLabelFormat       = "%-17s %s"
	alertsLabelTelegram     = "Telegram:"
	alertsLabelAlerts       = "Alerts:"
	alertsLabelMinimum      = "Minimum severity:"
	alertsLabelCooldown     = "Cooldown:"
	alertsLabelInterval     = "Check interval:"
	alertsLabelFile         = "Config file:"
	alertsLabelLastTest     = "Last test:"
	alertsConfiguredFormat  = "configured (token %s, chat ID set)"
	alertsNotConfiguredText = "not configured: set a bot token and a chat ID in Settings > Telegram"
	alertsEnabledText       = "enabled"
	alertsDisabledText      = "disabled"
	alertsCooldownFormat    = "%d minutes"
	alertsIntervalFormat    = "%d seconds"
	alertsNoTestText        = "none this session"
	alertsTestSentFormat    = "sent at %s"
	alertsTestFailedFormat  = "failed at %s (%s)"
	alertsTestSendingText   = "sending..."
	alertsHonestNote        = "The TUI does not send alerts itself: `exposedwatch watch` does, reading the saved config file when it starts."
	alertsNotAvailableText  = "Telegram alerts are not available: no config store was provided."
	alertsLoadFailedFormat  = "The config file could not be loaded: %s"
)

// Preview tab.
const (
	previewWouldAlertFormat  = "Would alert now (%d)"
	previewWouldIfFormat     = "Would alert if enabled (%d)"
	previewNoFindingsFormat  = "No findings at or above %s."
	previewDisabledText      = "Alerts are disabled; nothing would be sent."
	previewNotConfiguredText = "Not configured: set a bot token and a chat ID in Settings > Telegram. Nothing would be sent."
	previewSemanticsText     = "watch sends one summary at start, then only new findings."
	previewRowFormat         = "%s %s %s %s"
	previewMoreFormat        = "... and %d more"
)
