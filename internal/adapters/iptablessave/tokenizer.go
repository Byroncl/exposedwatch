package iptablessave

import (
	"fmt"
	"strings"
)

// tokenizeRule splits a rule line into arguments. Double quotes group words
// and a backslash escapes the next character inside quotes. Quotes are kept in
// the returned token text only through their content (the quote characters are
// removed).
func tokenizeRule(line string) ([]string, error) {
	var tokens []string
	var current strings.Builder
	inQuotes := false
	hasToken := false
	escaped := false
	for _, character := range line {
		switch {
		case escaped:
			current.WriteRune(character)
			escaped = false
		case inQuotes && character == escapeCharacter:
			escaped = true
		case character == quoteCharacter:
			inQuotes = !inQuotes
			hasToken = true
		case !inQuotes && (character == ' ' || character == '\t'):
			if hasToken {
				tokens = append(tokens, current.String())
				current.Reset()
				hasToken = false
			}
		default:
			current.WriteRune(character)
			hasToken = true
		}
	}
	if inQuotes || escaped {
		return nil, fmt.Errorf("%s", unterminatedQuoteDetail)
	}
	if hasToken {
		tokens = append(tokens, current.String())
	}
	return tokens, nil
}
