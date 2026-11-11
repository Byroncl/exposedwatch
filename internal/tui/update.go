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
	case ConnectionsResultMessage:
		return model.handleConnectionsResult(typedMessage), nil
	case tea.KeyMsg:
		return model.handleKey(typedMessage)
	default:
		return model, nil
	}
}

// handleTick schedules the next tick and starts whichever reads are idle. The
// scan and the connection read are independent: a slow one never holds back
// the other.
func (model Model) handleTick() (tea.Model, tea.Cmd) {
	commands := []tea.Cmd{tickCommand(model.interval)}
	commands = append(commands, model.startIdleReads()...)
	return model, tea.Batch(commands...)
}

// startIdleReads marks the idle reads as running and returns their commands.
func (model *Model) startIdleReads() []tea.Cmd {
	var commands []tea.Cmd
	if !model.refreshing {
		model.refreshing = true
		commands = append(commands, scanCommand(model.scanner, model.clock))
	}
	if model.connectionsEnabled() && !model.connectionsRefreshing {
		model.connectionsRefreshing = true
		commands = append(commands, connectionsCommand(model.connections, model.clock))
	}
	return commands
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
