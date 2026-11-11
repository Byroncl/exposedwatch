package tui

import (
	"strings"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

// maskToken is the only form in which a bot token is ever displayed: at most
// the public bot id before the colon, for example "123456789:****". It has
// the same semantics as the config file's masking.
func maskToken(token domain.BotToken) string {
	secret := token.Reveal()
	if secret == "" {
		return ""
	}
	botID, _, found := strings.Cut(secret, tokenSeparator)
	if !found {
		return maskedSuffix
	}
	return botID + tokenSeparator + maskedSuffix
}

// redactSecrets replaces every occurrence of a non-empty secret in text.
func redactSecrets(text string, secrets ...domain.BotToken) string {
	for _, secret := range secrets {
		if !secret.IsEmpty() {
			text = strings.ReplaceAll(text, secret.Reveal(), redactedText)
		}
	}
	return text
}
