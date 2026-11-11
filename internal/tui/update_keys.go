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
		if !model.refreshing {
			model.refreshing = true
			return model, scanCommand(model.scanner, model.clock)
		}
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
	case ActionCycleSort, ActionCycleMinimumSeverity, ActionStartSearch:
		model.handleFilteringAction(action)
	}
	return model, nil
}

// handleFilteringAction applies sort, severity and search keys. They only
// exist on the tabs whose rows can be filtered; the Changes tab ignores them.
func (model *Model) handleFilteringAction(action Action) {
	if !model.activeTab.supportsFiltering() {
		return
	}
	switch action {
	case ActionCycleSort:
		model.sortMode = model.sortMode.Next()
		model.rebuildVisibleRows()
	case ActionCycleMinimumSeverity:
		model.minSeverity = nextMinimumSeverity(model.minSeverity)
		model.rebuildVisibleRows()
	case ActionStartSearch:
		model.viewState = ViewFilterInput
		model.filterInput = model.filterText
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
		model.filterText = strings.TrimSpace(model.filterInput)
		model.viewState = ViewTable
		model.rebuildVisibleRows()
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
