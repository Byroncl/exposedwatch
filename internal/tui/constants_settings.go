package tui

import (
	"time"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

// Form behavior.
const (
	// maxEditLength bounds what one field accepts, also for a paste.
	maxEditLength = 256
	// maxNumberLength bounds the digits of a number row.
	maxNumberLength = 9
	// fallbackMinSeverity is shown when the loaded severity cannot be parsed.
	fallbackMinSeverity = domain.High
	// maxFailureRunes bounds a failure reason kept in the model.
	maxFailureRunes = 300
)

// Secret handling.
const (
	tokenSeparator           = ":"
	maskedSuffix             = "****"
	redactedText             = "***"
	secretBullet             = "•"
	editBufferPlaceholder    = "tui.editBuffer{***}"
	formValuesPlaceholder    = "tui.formValues{***}"
	settingsStatePlaceholder = "tui.settingsState{***}"
	modelPlaceholder         = "tui.Model{***}"
)

// Row labels.
const (
	fieldLabelEnabled     = "Enabled"
	fieldLabelToken       = "Bot token"
	fieldLabelChatID      = "Chat ID"
	fieldLabelMinSeverity = "Minimum severity"
	fieldLabelCooldown    = "Cooldown (minutes)"
	fieldLabelInterval    = "Check interval (seconds)"
	fieldLabelSave        = "[ Save ]"
	fieldLabelTest        = "[ Send test message ]"
)

// Form layout, in cells.
const (
	formLabelWidth   = 24
	formEnabledText  = "[x] on"
	formDisabledText = "[ ] off"
	formNotSetText   = "(not set)"
	formEditCursor   = "_"
)

// Form text.
const (
	panelTelegramTitle   = " Telegram "
	unsavedChangesText   = "unsaved changes"
	configPathFormat     = "Config file: %s"
	fromEnvironmentFmt   = "from environment (%s), read-only"
	editHintFormat       = "editing %s: enter save value, esc cancel"
	settingsNotAvailable = "Telegram settings are not available: no config store was provided."
	settingsLoadingText  = "Reading the config file..."
	loadFailedFormat     = "The config file could not be loaded: %s"
	loadFailedHint       = "The form is read-only until the file is fixed (or pass --config); then restart the tui."
	warningLineFormat    = "! %s"
	moreRowsFormat       = "... and %d more"
)

// Status messages.
const (
	statusSavingText        = "Saving..."
	statusSavedFormat       = "Saved to %s (restart watch to apply)"
	statusSaveFailedFormat  = "Not saved: %s"
	statusUnavailableText   = "Settings cannot be changed: the config file was not loaded."
	statusNotWholeNumber    = "Enter a whole number."
	statusEnvironmentFormat = "%s is set by the environment variable %s; unset it to edit it here."
	statusNeedsCredentials  = "A test message needs a bot token and a chat ID."
	statusAlreadySending    = "A test message is already being sent."
	statusNoTester          = "Test messages are not available."
	statusNoChanges         = "No changes to save"
	statusWaitFormat        = "Wait %ds before sending another test message"
	statusDigitsOnly        = "digits only"
	refreshFailedFormat     = "could not re-read the config file after saving: %s"
)

// Test message lines.
const (
	testSendingText  = "Test message: Sending..."
	testSentFormat   = "Test message: Sent (%s)"
	testFailedFormat = "Test message: Failed (%s): %s"
	testIdleText     = "Test message: not sent in this session"
)

// Timing.
const (
	// testMessageCooldown is the least time between two test messages, so a
	// held enter key cannot flood Telegram.
	testMessageCooldown = 5 * time.Second
	// alertTestGuardTimeout bounds one test message; the tester applies a
	// shorter timeout of its own to the network call.
	alertTestGuardTimeout = 30 * time.Second
)
