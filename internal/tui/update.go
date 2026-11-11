package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Update handles one message. It never blocks: scans run as commands.
func (model Model) Update(incomingMessage tea.Msg) (tea.Model, tea.Cmd) {
	switch typedMessage := incomingMessage.(type) {
	case tea.WindowSizeMsg:
		model.width = typedMessage.Width
		model.height = typedMessage.Height
		model.ensureCursorVisible()
		return model, nil
	case tickMessage:
		return model.handleTick()
	case RefreshResultMessage:
		return model.handleRefreshResult(typedMessage), nil
	case tea.KeyMsg:
		return model.handleKey(typedMessage)
	default:
		return model, nil
	}
}

func (model Model) handleTick() (tea.Model, tea.Cmd) {
	if model.refreshing {
		return model, tickCommand(model.interval)
	}
	model.refreshing = true
	return model, tea.Batch(scanCommand(model.scanner, model.clock), tickCommand(model.interval))
}

// handleRefreshResult stores a scan outcome. A failed scan keeps the previous
// data and never touches the change tracker, so it cannot fake closed ports.
func (model Model) handleRefreshResult(result RefreshResultMessage) Model {
	model.refreshing = false
	model.lastRefresh = result.At
	model.refreshErr = result.Err
	if result.Err != nil {
		return model
	}
	model.report = result.Report
	model.hasReport = true
	model.tracker = model.tracker.observe(result.Report, result.At)
	model.rebuildVisibleRows()
	return model
}
