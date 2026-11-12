package telegram

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

// FormatAlert builds the plain-text message of an alert. It only uses local
// facts carried by the alert, never markup, and never exceeds Telegram's
// message limit. Single-line fields are stripped of line breaks, control
// characters, bidirectional overrides and zero-width characters, so a hostile
// process name cannot forge extra lines or reorder text.
func FormatAlert(alert domain.Alert) string {
	headerLine := fmt.Sprintf(headerFormat, strings.ToUpper(alert.Severity.String()), domain.SanitizeLine(alert.Title))
	lines := []string{headerLine}
	if host := domain.SanitizeLine(alert.Host); host != "" {
		lines = append(lines, fmt.Sprintf(fieldHostFormat, host))
	}
	if alert.HasPort() {
		lines = append(lines, fmt.Sprintf(fieldPortFormat, domain.SanitizeLine(alert.PortLabel())))
	}
	if alert.Kind == domain.AlertKindNewFinding {
		lines = append(lines, fmt.Sprintf(fieldExposureFormat, domain.SanitizeLine(alert.Exposure.String())))
	}
	if process := domain.SanitizeLine(alert.Process); process != "" {
		lines = append(lines, fmt.Sprintf(fieldProcessFormat, process))
	}
	if rule := domain.SanitizeLine(alert.RuleID); rule != "" {
		lines = append(lines, fmt.Sprintf(fieldRuleFormat, rule))
	}
	text := strings.Join(lines, lineSeparator)
	if alert.Body != "" {
		if alert.Kind == domain.AlertKindNewFinding {
			text += lineSeparator + fmt.Sprintf(fieldFixFormat, domain.SanitizeLine(alert.Body))
		} else {
			text += blankLineSeparator + domain.SanitizeBlock(alert.Body)
		}
	}
	return boundMessage(text)
}

// boundMessage cuts the text to MaxMessageRunes, ending with an ellipsis.
func boundMessage(text string) string {
	if utf8.RuneCountInString(text) <= MaxMessageRunes {
		return text
	}
	kept := []rune(text)[:MaxMessageRunes-utf8.RuneCountInString(truncationSuffix)]
	return string(kept) + truncationSuffix
}
