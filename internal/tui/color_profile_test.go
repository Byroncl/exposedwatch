package tui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// forceColors makes Lip Gloss emit the given color profile and theme for the
// duration of one test, then restores whatever was active before.
func forceColors(testRun *testing.T, profile termenv.Profile, darkBackground bool) {
	testRun.Helper()
	previousProfile := lipgloss.ColorProfile()
	previousDark := lipgloss.HasDarkBackground()
	lipgloss.SetColorProfile(profile)
	lipgloss.SetHasDarkBackground(darkBackground)
	testRun.Cleanup(func() {
		lipgloss.SetColorProfile(previousProfile)
		lipgloss.SetHasDarkBackground(previousDark)
	})
}
