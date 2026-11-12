package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/Byroncl/exposedwatch/internal/app"
	"github.com/Byroncl/exposedwatch/internal/domain"
)

// watchLogger writes single timestamped lines. As a second line of defence
// it masks any secret it was told about, even if a caller passed one in.
// Secrets are stored as domain.BotToken, and the logger itself prints a
// fixed placeholder, so formatting it never reveals them.
type watchLogger struct {
	writer  io.Writer
	clock   app.Clock
	secrets []domain.BotToken
}

// Logf writes one line. Control characters, line breaks, bidirectional
// overrides and zero-width characters are removed from the text.
func (logger watchLogger) Logf(format string, arguments ...any) {
	line := fmt.Sprintf(format, arguments...)
	for _, secret := range logger.secrets {
		line = maskSecret(line, secret)
	}
	fmt.Fprintf(logger.writer, logLineFormat, logger.clock.Now().UTC().Format(logTimeFormat), domain.SanitizeLine(line))
}

// String keeps the secrets out of fmt output.
func (logger watchLogger) String() string { return loggerPlaceholder }

// GoString keeps the secrets out of %#v output.
func (logger watchLogger) GoString() string { return loggerPlaceholder }

// Format keeps the secrets out of fmt output, whatever the verb.
func (logger watchLogger) Format(state fmt.State, _ rune) {
	_, _ = state.Write([]byte(loggerPlaceholder))
}

// maskSecret replaces every occurrence of the token in text.
func maskSecret(text string, token domain.BotToken) string {
	if token.IsEmpty() {
		return text
	}
	return strings.ReplaceAll(text, token.Reveal(), redactedLogText)
}

// alertLabel is the short, local-facts-only description used in log lines:
// the severity word first, then the port (when any) and the title.
func alertLabel(alert domain.Alert) string {
	description := alert.Title
	if alert.HasPort() {
		description = alert.PortLabel() + " " + alert.Title
	}
	return fmt.Sprintf(alertLabelFormat, strings.ToUpper(alert.Severity.String()), description)
}
