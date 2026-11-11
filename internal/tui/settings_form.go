package tui

import (
	"fmt"
	"strconv"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

// formValues are the values of the editable rows. The token is a
// domain.BotToken, and Format keeps the struct out of fmt output because
// fmt cannot call methods of unexported fields.
type formValues struct {
	enabled         bool
	token           domain.BotToken
	chatID          string
	minSeverity     domain.Severity
	cooldownMinutes int
	intervalSeconds int
}

// Format prints a placeholder whatever the verb.
func (values formValues) Format(state fmt.State, _ rune) {
	_, _ = fmt.Fprint(state, formValuesPlaceholder)
}

// formValuesFrom converts what the store loaded.
func formValuesFrom(settings AlertSettings) formValues {
	severity, err := domain.ParseSeverity(settings.MinSeverity)
	if err != nil {
		severity = fallbackMinSeverity
	}
	return formValues{
		enabled:         settings.Enabled,
		token:           settings.Token,
		chatID:          settings.ChatID,
		minSeverity:     severity,
		cooldownMinutes: settings.CooldownMinutes,
		intervalSeconds: settings.IntervalSeconds,
	}
}

// alertSettings overlays the values on the metadata of the loaded settings.
func (values formValues) alertSettings(loaded AlertSettings) AlertSettings {
	loaded.Enabled = values.enabled
	loaded.Token = values.token
	loaded.ChatID = values.chatID
	loaded.MinSeverity = values.minSeverity.String()
	loaded.CooldownMinutes = values.cooldownMinutes
	loaded.IntervalSeconds = values.intervalSeconds
	return loaded
}

// isConfigured tells whether a bot token and a chat id are present.
func (values formValues) isConfigured() bool {
	return !values.token.IsEmpty() && values.chatID != ""
}

// numberText is the current text of a number row.
func (values formValues) numberText(field FormField) string {
	if field == FieldCooldown {
		return strconv.Itoa(values.cooldownMinutes)
	}
	return strconv.Itoa(values.intervalSeconds)
}

// withNumber stores a number row; ok is false when the text is not a whole number.
func (values formValues) withNumber(field FormField, text string) (formValues, bool) {
	number, err := strconv.Atoi(text)
	if err != nil {
		return values, false
	}
	if field == FieldCooldown {
		values.cooldownMinutes = number
	} else {
		values.intervalSeconds = number
	}
	return values, true
}

// withChatID stores the chat id row.
func (values formValues) withChatID(chatID string) formValues {
	values.chatID = chatID
	return values
}

// withCycledSeverity moves to the next severity.
func (values formValues) withCycledSeverity() formValues {
	values.minSeverity = nextMinimumSeverity(values.minSeverity)
	return values
}

// withToggledEnabled flips the enabled switch.
func (values formValues) withToggledEnabled() formValues {
	values.enabled = !values.enabled
	return values
}

// withToken stores the token row.
func (values formValues) withToken(token domain.BotToken) formValues {
	values.token = token
	return values
}
