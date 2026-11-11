package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

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
	action := actionForKey(keyPress)
	switch action {
	case ActionQuit:
		return model, tea.Quit
	case ActionMoveUp:
		model.moveCursor(-1)
	case ActionMoveDown:
		model.moveCursor(1)
	case ActionJumpTop:
		model.setCursor(0)
	case ActionJumpBottom:
		model.setCursor(len(model.visibleRows) - 1)
	case ActionRefresh:
		return model, tea.Batch(model.startIdleReads()...)
	case ActionToggleHelp:
		model.viewState = ViewHelp
	case ActionNextTab:
		model.switchTab(model.activeTab.Next())
	case ActionPreviousTab:
		model.switchTab(model.activeTab.Previous())
	case ActionShowFindings:
		model.switchTab(TabFindings)
	case ActionShowAllPorts:
		model.switchTab(TabAllPorts)
	case ActionShowChanges:
		model.switchTab(TabChanges)
	case ActionShowConnections:
		model.switchTab(TabConnections)
	case ActionToggleLoopback:
		model.toggleLoopback()
	case ActionCycleSort, ActionCycleMinimumSeverity, ActionStartSearch:
		model.handleFilteringAction(action)
	}
	return model, nil
}

// handleFilteringAction applies sort, severity and search keys. Each key only
// acts on the tabs that support it: the Changes tab ignores all three and the
// Connections tab has no severity filter.
func (model *Model) handleFilteringAction(action Action) {
	switch action {
	case ActionCycleSort:
		if !model.acceptsSortAndSearch() {
			return
		}
		if model.activeTab == TabConnections {
			model.connectionSortMode = model.connectionSortMode.Next()
		} else {
			model.sortMode = model.sortMode.Next()
		}
		model.rebuildVisibleRows()
	case ActionCycleMinimumSeverity:
		if !model.activeTab.supportsMinimumSeverity() {
			return
		}
		model.minSeverity = nextMinimumSeverity(model.minSeverity)
		model.rebuildVisibleRows()
	case ActionStartSearch:
		if !model.acceptsSortAndSearch() {
			return
		}
		model.viewState = ViewFilterInput
		model.filterInput = model.activeFilterText()
	}
}

// activeFilterText is the free-text filter of the active tab. The Connections
// tab keeps its own, so a port filter never hides connections.
func (model Model) activeFilterText() string {
	if model.activeTab == TabConnections {
		return model.connectionFilterText
	}
	return model.filterText
}

func (model *Model) setActiveFilterText(text string) {
	if model.activeTab == TabConnections {
		model.connectionFilterText = text
		return
	}
	model.filterText = text
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
		model.setActiveFilterText(strings.TrimSpace(model.filterInput))
		model.viewState = ViewTable
		model.rebuildVisibleRows()
	case tea.KeyEsc:
		model.filterInput = model.activeFilterText()
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
