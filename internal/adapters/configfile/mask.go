package configfile

import (
	"strings"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

// MaskToken returns a display form of the token that reveals at most the
// public bot id before the colon, for example "123456789:****". An empty
// token stays empty.
func MaskToken(token domain.BotToken) string {
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
