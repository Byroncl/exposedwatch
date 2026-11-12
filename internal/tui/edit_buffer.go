package tui

import (
	"fmt"
	"unicode"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

// editBuffer holds the text being typed into a form field. It may hold a
// secret, so every fmt verb prints a placeholder; the content is only
// reachable through text.
type editBuffer []rune

// Format keeps the content out of fmt output, whatever the verb.
func (buffer editBuffer) Format(state fmt.State, _ rune) {
	_, _ = fmt.Fprint(state, editBufferPlaceholder)
}

// text returns the typed text.
func (buffer editBuffer) text() string {
	return string(buffer)
}

// length is the number of typed runes.
func (buffer editBuffer) length() int {
	return len(buffer)
}

// appended adds the accepted runes, up to limit runes in total, and reports
// whether a typed rune was rejected. Control characters (including line
// breaks from a paste), spaces, bidirectional overrides and zero-width
// characters are always dropped; when digitsOnly is set so is everything that
// is not a decimal digit.
func (buffer editBuffer) appended(typed []rune, limit int, digitsOnly bool) (editBuffer, bool) {
	result := append(editBuffer{}, buffer...)
	rejected := false
	for _, character := range typed {
		if len(result) >= limit {
			break
		}
		if domain.SanitizeLine(string(character)) == "" || (digitsOnly && !unicode.IsDigit(character)) {
			rejected = true
			continue
		}
		result = append(result, character)
	}
	return result, rejected
}

// withoutLast removes the last rune.
func (buffer editBuffer) withoutLast() editBuffer {
	if len(buffer) == 0 {
		return buffer
	}
	return append(editBuffer{}, buffer[:len(buffer)-1]...)
}
