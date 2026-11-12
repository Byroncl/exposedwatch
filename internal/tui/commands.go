package tui

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Byroncl/exposedwatch/internal/domain"
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

// connectionsCommand reads the connections off the UI loop and reports them
// as a message of their own.
func connectionsCommand(source ConnectionSnapshotter, clock func() time.Time) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), connectionsTimeout)
		defer cancel()
		snapshot, err := source.Snapshot(ctx)
		return ConnectionsResultMessage{Snapshot: snapshot, Err: err, At: clock()}
	}
}

// firewallCommand evaluates the ports against the firewall off the UI loop and
// reports the outcome as a message of its own.
func firewallCommand(evaluator FirewallEvaluator, listening []domain.Port, clock func() time.Time) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), firewallTimeout)
		defer cancel()
		firewallReport, err := evaluator.Evaluate(ctx, listening)
		return FirewallResultMessage{Report: firewallReport, Err: err, At: clock()}
	}
}

// tickCommand schedules the next auto-refresh tick.
func tickCommand(interval time.Duration) tea.Cmd {
	return tea.Tick(interval, func(time.Time) tea.Msg { return tickMessage{} })
}
