package telegram

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

var descriptionPattern = regexp.MustCompile(descriptionCharset)

// tokenRedactor removes a bot token, in the spellings it can take inside an
// error or an echoed response, from text. It also judges whether remote text
// is safe to show. Printing it never reveals the token.
type tokenRedactor struct {
	secret  domain.BotToken
	pattern *regexp.Regexp
}

func newTokenRedactor(token domain.BotToken) tokenRedactor {
	secret := token.Reveal()
	if secret == "" {
		return tokenRedactor{}
	}
	spellings := tokenSpellings(secret)
	// Longest first, so "bot<token>" wins over the bare token.
	var alternatives []string
	for _, spelling := range spellings {
		alternatives = append(alternatives, regexp.QuoteMeta(botSegmentLabel+spelling))
	}
	for _, spelling := range spellings {
		alternatives = append(alternatives, regexp.QuoteMeta(spelling))
	}
	sort.SliceStable(alternatives, func(first, second int) bool {
		return len(alternatives[first]) > len(alternatives[second])
	})
	return tokenRedactor{
		secret:  token,
		pattern: regexp.MustCompile(caseInsensitiveFlag + strings.Join(alternatives, alternationSeparator)),
	}
}

// tokenSpellings lists the encodings of the token that are redacted.
func tokenSpellings(secret string) []string {
	spellings := []string{
		secret,
		url.PathEscape(secret),
		url.QueryEscape(secret),
		base64.StdEncoding.EncodeToString([]byte(secret)),
		base64.RawStdEncoding.EncodeToString([]byte(secret)),
		base64.URLEncoding.EncodeToString([]byte(secret)),
		base64.RawURLEncoding.EncodeToString([]byte(secret)),
		hex.EncodeToString([]byte(secret)),
	}
	if encoded, err := json.Marshal(secret); err == nil && len(encoded) > 2 {
		spellings = append(spellings, string(encoded[1:len(encoded)-1]))
	}
	return spellings
}

// Redact replaces the token and the bot path segment with "***",
// ignoring case.
func (redactor tokenRedactor) Redact(text string) string {
	if redactor.pattern == nil {
		return text
	}
	return redactor.pattern.ReplaceAllString(text, redactionText)
}

// cleanReason redacts, flattens and bounds LOCAL error text (for example a
// dial error) before it is shown to the operator.
func (redactor tokenRedactor) cleanReason(text string) string {
	flattened := domain.SanitizeLine(redactor.Redact(text))
	if utf8.RuneCountInString(flattened) > maxDescriptionRunes {
		return string([]rune(flattened)[:maxDescriptionRunes]) + truncationSuffix
	}
	return flattened
}

// safeDescription returns the remote description only when it looks like a
// short ordinary Telegram message and contains no part of the token in any
// common encoding. Anything else is not echoed at all.
func (redactor tokenRedactor) safeDescription(description string) (string, bool) {
	if description == "" || utf8.RuneCountInString(description) > maxDescriptionRunes ||
		!descriptionPattern.MatchString(description) || redactor.containsTokenMaterial(description) {
		return "", false
	}
	return description, true
}

// containsTokenMaterial reports whether the text, or a percent-, base64- or
// hex-decoded form of it, contains the token or any run of
// tokenFragmentLength consecutive characters of it, ignoring case.
func (redactor tokenRedactor) containsTokenMaterial(text string) bool {
	secret := strings.ToLower(redactor.secret.Reveal())
	if secret == "" {
		return false
	}
	fragments := tokenFragments(secret)
	for _, candidate := range decodedForms(text) {
		lowered := strings.ToLower(candidate)
		for _, fragment := range fragments {
			if strings.Contains(lowered, fragment) {
				return true
			}
		}
	}
	return false
}

func tokenFragments(secret string) []string {
	if len(secret) <= tokenFragmentLength {
		return []string{secret}
	}
	fragments := make([]string, 0, len(secret)-tokenFragmentLength+1)
	for start := 0; start+tokenFragmentLength <= len(secret); start++ {
		fragments = append(fragments, secret[start:start+tokenFragmentLength])
	}
	return fragments
}

// decodedForms returns the text and every decoding attempt of it.
func decodedForms(text string) []string {
	forms := []string{text}
	if unescaped, err := url.PathUnescape(text); err == nil {
		forms = append(forms, unescaped)
	}
	for _, field := range strings.FieldsFunc(text, isNotEncodedCharacter) {
		if decoded, err := hex.DecodeString(field); err == nil {
			forms = append(forms, string(decoded))
		}
		for _, encoding := range []*base64.Encoding{
			base64.StdEncoding, base64.RawStdEncoding, base64.URLEncoding, base64.RawURLEncoding,
		} {
			if decoded, err := encoding.DecodeString(field); err == nil {
				forms = append(forms, string(decoded))
			}
		}
	}
	return forms
}

func isNotEncodedCharacter(character rune) bool {
	return !strings.ContainsRune(encodedAlphabet, character)
}

// Format keeps the token out of fmt output, whatever the verb.
func (redactor tokenRedactor) Format(state fmt.State, _ rune) {
	_, _ = state.Write([]byte(redactorPlaceholder))
}

// String keeps the token out of fmt output.
func (redactor tokenRedactor) String() string { return redactorPlaceholder }

// GoString keeps the token out of %#v output.
func (redactor tokenRedactor) GoString() string { return redactorPlaceholder }
