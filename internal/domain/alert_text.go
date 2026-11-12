package domain

import (
	"strings"
	"unicode"
)

// SanitizeLine makes untrusted text safe for a single-line field of a
// message or a log line: bidirectional overrides and zero-width characters
// are removed, line breaks and tabs become spaces, every other control
// character is removed, and the ends are trimmed.
func SanitizeLine(text string) string {
	return strings.TrimSpace(strings.Map(sanitizeRune, text))
}

// SanitizeBlock is SanitizeLine for multi-line text: newlines survive,
// everything else is treated as in SanitizeLine.
func SanitizeBlock(text string) string {
	return strings.TrimSpace(strings.Map(func(character rune) rune {
		if character == '\n' {
			return character
		}
		if character == '\r' {
			return -1
		}
		return sanitizeRune(character)
	}, text))
}

func sanitizeRune(character rune) rune {
	switch {
	case isInvisibleFormatting(character):
		return -1
	case character == '\n' || character == '\r' || character == '\t':
		return ' '
	case unicode.IsControl(character):
		return -1
	default:
		return character
	}
}

// isInvisibleFormatting matches bidirectional overrides and isolates and
// zero-width characters, which can reorder or hide text.
func isInvisibleFormatting(character rune) bool {
	switch {
	case character >= bidiEmbeddingStart && character <= bidiEmbeddingEnd:
		return true
	case character >= bidiIsolateStart && character <= bidiIsolateEnd:
		return true
	case character >= zeroWidthStart && character <= zeroWidthEnd:
		return true
	default:
		return character == byteOrderMark
	}
}
