package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Run starts the full-screen program and blocks until the user quits.
func Run(options Options) error {
	_, err := tea.NewProgram(New(options), tea.WithAltScreen()).Run()
	return err
}
