package tui

import tea "github.com/charmbracelet/bubbletea"

// Action is a user intention resolved from a key press.
type Action int

// Actions.
const (
	ActionNone Action = iota
	ActionQuit
	ActionMoveUp
	ActionMoveDown
	ActionJumpTop
	ActionJumpBottom
	ActionRefresh
	ActionCycleSort
	ActionCycleMinimumSeverity
	ActionStartSearch
	ActionToggleHelp
	ActionNextTab
	ActionPreviousTab
	ActionShowFindings
	ActionShowAllPorts
	ActionShowChanges
)

// actionForKey maps a key press to an action when browsing the table.
func actionForKey(keyPress tea.KeyMsg) Action {
	switch keyPress.String() {
	case keyQuit, keyInterrupt:
		return ActionQuit
	case keyUp, keyUpVim:
		return ActionMoveUp
	case keyDown, keyDownVim:
		return ActionMoveDown
	case keyTop:
		return ActionJumpTop
	case keyBottom:
		return ActionJumpBottom
	case keyRefresh:
		return ActionRefresh
	case keySort:
		return ActionCycleSort
	case keyMinimumSeverity:
		return ActionCycleMinimumSeverity
	case keySearch:
		return ActionStartSearch
	case keyHelp:
		return ActionToggleHelp
	case keyNextTab:
		return ActionNextTab
	case keyPreviousTab:
		return ActionPreviousTab
	case keyFindingsTab:
		return ActionShowFindings
	case keyAllPortsTab:
		return ActionShowAllPorts
	case keyChangesTab:
		return ActionShowChanges
	default:
		return ActionNone
	}
}
