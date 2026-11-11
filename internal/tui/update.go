package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Byroncl/exposedwatch/internal/domain"
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

func (model Model) handleRefreshResult(result RefreshResultMessage) Model {
	model.refreshing = false
	model.lastRefresh = result.At
	model.refreshErr = result.Err
	if result.Err != nil {
		return model
	}
	model.report = result.Report
	model.hasReport = true
	model.rebuildVisibleFindings()
	return model
}

func (model Model) handleKey(keyPress tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch model.viewState {
	case ViewFilterInput:
		return model.handleFilterInputKey(keyPress)
	case ViewHelp:
		return model.handleHelpKey(keyPress)
	default:
		return model.handleTableKey(keyPress)
	}
}

func (model Model) handleTableKey(keyPress tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch actionForKey(keyPress) {
	case ActionQuit:
		return model, tea.Quit
	case ActionMoveUp:
		model.moveCursor(-1)
	case ActionMoveDown:
		model.moveCursor(1)
	case ActionJumpTop:
		model.setCursor(0)
	case ActionJumpBottom:
		model.setCursor(len(model.visibleFindings) - 1)
	case ActionRefresh:
		if !model.refreshing {
			model.refreshing = true
			return model, scanCommand(model.scanner, model.clock)
		}
	case ActionCycleSort:
		model.sortMode = model.sortMode.Next()
		model.rebuildVisibleFindings()
	case ActionCycleMinimumSeverity:
		model.minSeverity = nextMinimumSeverity(model.minSeverity)
		model.rebuildVisibleFindings()
	case ActionStartSearch:
		model.viewState = ViewFilterInput
		model.filterInput = model.filterText
	case ActionToggleHelp:
		model.viewState = ViewHelp
	}
	return model, nil
}

func (model Model) handleHelpKey(keyPress tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch keyPress.String() {
	case keyQuit, keyInterrupt:
		return model, tea.Quit
	case keyHelp, keyEscape:
		model.viewState = ViewTable
	}
	return model, nil
}

func (model Model) handleFilterInputKey(keyPress tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch keyPress.Type {
	case tea.KeyCtrlC:
		return model, tea.Quit
	case tea.KeyEnter:
		model.filterText = strings.TrimSpace(model.filterInput)
		model.viewState = ViewTable
		model.rebuildVisibleFindings()
	case tea.KeyEsc:
		model.filterInput = model.filterText
		model.viewState = ViewTable
	case tea.KeyBackspace:
		runes := []rune(model.filterInput)
		if len(runes) > 0 {
			model.filterInput = string(runes[:len(runes)-1])
		}
	case tea.KeySpace:
		model.filterInput += " "
	case tea.KeyRunes:
		model.filterInput += string(keyPress.Runes)
	}
	return model, nil
}

// nextMinimumSeverity cycles info, low, medium, high, critical, info.
func nextMinimumSeverity(current domain.Severity) domain.Severity {
	if current >= domain.Critical {
		return domain.Info
	}
	return current + 1
}

// rebuildVisibleFindings applies filters and sort, keeping the selection.
func (model *Model) rebuildVisibleFindings() {
	var selected *domain.Finding
	if model.cursor >= 0 && model.cursor < len(model.visibleFindings) {
		current := model.visibleFindings[model.cursor]
		selected = &current
	}
	visible := make([]domain.Finding, 0, len(model.report.Findings))
	for _, finding := range model.report.Findings {
		if finding.Severity >= model.minSeverity && matchesText(finding, model.filterText) {
			visible = append(visible, finding)
		}
	}
	sortFindings(visible, model.sortMode)
	model.visibleFindings = visible
	model.cursor = 0
	if selected != nil {
		for index, finding := range visible {
			if sameFinding(finding, *selected) {
				model.cursor = index
				break
			}
		}
	}
	model.ensureCursorVisible()
}

func (model *Model) moveCursor(delta int) {
	model.setCursor(model.cursor + delta)
}

func (model *Model) setCursor(position int) {
	last := len(model.visibleFindings) - 1
	if position > last {
		position = last
	}
	if position < 0 {
		position = 0
	}
	model.cursor = position
	model.ensureCursorVisible()
}

// ensureCursorVisible scrolls so the cursor row is inside the table window.
func (model *Model) ensureCursorVisible() {
	capacity := model.computeLayout().tableRows
	if capacity < 1 {
		model.scrollOffset = scrollOffsetFloor
		return
	}
	if model.cursor < model.scrollOffset {
		model.scrollOffset = model.cursor
	}
	if model.cursor >= model.scrollOffset+capacity {
		model.scrollOffset = model.cursor - capacity + 1
	}
	if model.scrollOffset < scrollOffsetFloor {
		model.scrollOffset = scrollOffsetFloor
	}
}
