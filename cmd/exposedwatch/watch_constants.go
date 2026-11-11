package main

import "time"

// Flag names of the watch and notify-test commands.
const (
	flagConfig = "config"
	flagOnce   = "once"
)

// Flag set names.
const (
	watchFlagSetName      = "watch"
	notifyTestFlagSetName = "notify-test"
)

// Flag descriptions.
const (
	usageConfig = "path of the config file (default: /etc/exposedwatch/config.json for root, " +
		"otherwise $XDG_CONFIG_HOME/exposedwatch/config.json or ~/.config/exposedwatch/config.json)"
	usageWatchInterval    = "time between scans, at least 10s (default: alerts.interval_seconds of the config file)"
	usageWatchMinSeverity = "minimum severity that alerts: info, low, medium, high or critical (default: alerts.min_severity of the config file)"
	usageOnce             = "scan once, send the alerts and exit (for cron); exit 0 ok, 1 runtime error, 2 usage error"
)

// Watch loop limits.
const (
	// maxRetryBackoff caps the doubling wait between delivery attempts of
	// one alert; the first wait is the scan interval.
	maxRetryBackoff = 15 * time.Minute
	// alertTimeToLive is how long an undelivered alert is retried before it
	// is dropped (its finding can then alert again).
	alertTimeToLive = 24 * time.Hour
	// shutdownGracePeriod is how long an in-flight send may continue after
	// SIGINT or SIGTERM.
	shutdownGracePeriod = 5 * time.Second
	// minWatchInterval is the shortest accepted --interval.
	minWatchInterval = 10 * time.Second
	unknownHostName  = "unknown-host"
)

// Log and console messages. None of them may contain a token or a URL.
const (
	logTimeFormat              = time.RFC3339
	logLineFormat              = "%s %s\n"
	redactedLogText            = "***"
	logWatchStartFormat        = "watching every %s, alerting at or above %s"
	logNotificationsDisabled   = "notifications disabled: Telegram is not enabled or the bot token / chat id is missing; alerts are only logged here"
	logNotificationsEnabled    = "notifications enabled: Telegram"
	logConfigWarningFormat     = "config warning: %s"
	logSourceUnavailableFormat = "source %s unavailable: %v"
	logScanFailedFormat        = "scan failed: %v"
	logAlertFormat             = "alert: %s %s"
	logAlertSentCountFormat    = "alerts sent: %d"
	logDeliveryFailedFormat    = "alert delivery failed (attempt %d), will retry in %s: %v"
	logDeliveryDroppedFormat   = "giving up on alert %s, it will alert again if it is still open: %v"
	logQueueOverflowFormat     = "alert queue full: dropped the %d oldest alerts"
	loggerPlaceholder          = "main.watchLogger{***}"
	logShutdownText            = "shutting down"
	alertLabelFormat           = "%s %s"
	configFailedFormat         = "config error: %v"
	invalidWatchIntervalFmt    = "invalid --interval %v: must be at least %v"
	notifyNotConfiguredFormat  = "Telegram is not configured: set bot_token and chat_id in %s or the %s and %s environment variables\n"
	notifySendingFormat        = "Sending a test message to Telegram (bot token %s from %s, chat id from %s)...\n"
	notifySucceededText        = "OK: the test message was delivered"
	notifyFailedFormat         = "FAILED: the test message was not delivered: %v\n"
	notifyTestAlertTitleFormat = "exposedwatch test message from %s"
	notifyBodyText             = "If you can read this, alerts are working."
)
