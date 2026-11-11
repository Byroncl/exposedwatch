package tui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestParseThemeAcceptsKnownNamesCaseInsensitively(testRun *testing.T) {
	cases := map[string]Theme{"auto": ThemeAuto, "dark": ThemeDark, "LIGHT": ThemeLight, " Dark ": ThemeDark}
	for name, want := range cases {
		got, err := ParseTheme(name)
		if err != nil || got != want {
			testRun.Errorf("ParseTheme(%q) = %v, %v; want %v", name, got, err, want)
		}
	}
}

func TestParseThemeRejectsUnknownNames(testRun *testing.T) {
	for _, name := range []string{"", "blue", "darkish"} {
		if _, err := ParseTheme(name); err == nil {
			testRun.Errorf("ParseTheme(%q) must fail", name)
		}
	}
}

func TestApplyThemeForcesBackgroundDetection(testRun *testing.T) {
	original := lipgloss.HasDarkBackground()
	testRun.Cleanup(func() { lipgloss.SetHasDarkBackground(original) })
	ApplyTheme(ThemeLight)
	if lipgloss.HasDarkBackground() {
		testRun.Error("light theme must force a light background")
	}
	ApplyTheme(ThemeDark)
	if !lipgloss.HasDarkBackground() {
		testRun.Error("dark theme must force a dark background")
	}
	lipgloss.SetHasDarkBackground(false)
	ApplyTheme(ThemeAuto)
	if lipgloss.HasDarkBackground() {
		testRun.Error("auto theme must leave detection untouched")
	}
}
