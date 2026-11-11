package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Run applies the theme, starts the full-screen program and blocks until the
// user quits.
func Run(options Options) error {
	ApplyTheme(options.Theme)
	_, err := tea.NewProgram(New(options), tea.WithAltScreen()).Run()
	return err
}
