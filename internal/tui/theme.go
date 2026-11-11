package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Theme selects how the palette adapts to the terminal background.
type Theme string

// Themes.
const (
	// ThemeAuto detects the terminal background.
	ThemeAuto Theme = themeNameAuto
	// ThemeDark forces the palette for dark backgrounds.
	ThemeDark Theme = themeNameDark
	// ThemeLight forces the palette for light backgrounds.
	ThemeLight Theme = themeNameLight
)

// ParseTheme converts a case-insensitive theme name into a Theme.
func ParseTheme(name string) (Theme, error) {
	switch Theme(strings.ToLower(strings.TrimSpace(name))) {
	case ThemeAuto:
		return ThemeAuto, nil
	case ThemeDark:
		return ThemeDark, nil
	case ThemeLight:
		return ThemeLight, nil
	default:
		return ThemeAuto, fmt.Errorf(unknownThemeText, name)
	}
}

// ApplyTheme forces the background detection of Lip Gloss for the dark and
// light themes and leaves automatic detection untouched for ThemeAuto.
func ApplyTheme(theme Theme) {
	switch theme {
	case ThemeDark:
		lipgloss.SetHasDarkBackground(true)
	case ThemeLight:
		lipgloss.SetHasDarkBackground(false)
	case ThemeAuto:
	}
}
