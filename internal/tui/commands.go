package tui

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// scanCommand runs one scan off the UI loop and reports it as a message.
func scanCommand(scanner Scanner, clock func() time.Time) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), scanTimeout)
		defer cancel()
		scanReport, err := scanner.Scan(ctx)
		return RefreshResultMessage{Report: scanReport, Err: err, At: clock()}
	}
}

// tickCommand schedules the next auto-refresh tick.
func tickCommand(interval time.Duration) tea.Cmd {
	return tea.Tick(interval, func(time.Time) tea.Msg { return tickMessage{} })
}
