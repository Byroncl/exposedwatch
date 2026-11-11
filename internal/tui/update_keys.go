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
	case ActionMoveUp, ActionMoveDown, ActionJumpTop, ActionJumpBottom:
		model.handleCursorAction(action)
	case ActionRefresh:
		return model, tea.Batch(model.startIdleReads(true)...)
	case ActionToggleHelp:
		model.viewState = ViewHelp
	case ActionNextModule:
		model.openModule(model.activeModule.Next())
	case ActionPreviousModule:
		model.openModule(model.activeModule.Previous())
	case ActionNextInnerTab:
		model.cycleInnerTab(1)
	case ActionPreviousInnerTab:
		model.cycleInnerTab(-1)
	case ActionJumpInnerTab:
		if inner, isInnerTabKey := parseInnerTabKey(keyPress.String()); isInnerTabKey {
			model.openInnerTab(inner)
		}
	case ActionToggleSidebar:
		model.toggleSidebar()
	case ActionToggleLoopback:
		model.toggleLoopback()
	case ActionToggleInboundOnly:
		model.toggleInboundOnly()
	case ActionCycleSort, ActionCycleMinimumSeverity, ActionStartSearch:
		model.handleFilteringAction(action)
	}
	return model, nil
}

// handleCursorAction moves the selection. Only the table views have one.
func (model *Model) handleCursorAction(action Action) {
	if !model.activeTab.isTable() {
		return
	}
	switch action {
	case ActionMoveUp:
		model.moveCursor(-1)
	case ActionMoveDown:
		model.moveCursor(1)
	case ActionJumpTop:
		model.setCursor(0)
	case ActionJumpBottom:
		model.setCursor(len(model.visibleRows) - 1)
	}
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
		model.cycleActiveSort()
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

// cycleActiveSort moves the active tab to its next sort mode. Every table
// keeps its own mode.
func (model *Model) cycleActiveSort() {
	switch model.activeTab {
	case TabConnections:
		model.connectionSortMode = model.connectionSortMode.Next()
	case TabFirewallPorts:
		model.firewallSortMode = model.firewallSortMode.Next()
	default:
		model.sortMode = model.sortMode.Next()
	}
}

// activeFilterText is the free-text filter of the active tab. The Connections
// and Firewall ports tabs keep their own, so a port filter never hides
// connections or verdicts.
func (model Model) activeFilterText() string {
	switch model.activeTab {
	case TabConnections:
		return model.connectionFilterText
	case TabFirewallPorts:
		return model.firewallFilterText
	default:
		return model.filterText
	}
}

func (model *Model) setActiveFilterText(text string) {
	switch model.activeTab {
	case TabConnections:
		model.connectionFilterText = text
	case TabFirewallPorts:
		model.firewallFilterText = text
	default:
		model.filterText = text
	}
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
