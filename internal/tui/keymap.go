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
	ActionNextModule
	ActionPreviousModule
	ActionNextInnerTab
	ActionPreviousInnerTab
	ActionJumpInnerTab
	ActionToggleSidebar
	ActionToggleLoopback
	ActionToggleInboundOnly
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
	case keyNextModule:
		return ActionNextModule
	case keyPreviousModule:
		return ActionPreviousModule
	case keyNextInnerTab:
		return ActionNextInnerTab
	case keyPreviousInnerTab:
		return ActionPreviousInnerTab
	case keySidebar:
		return ActionToggleSidebar
	case keyLoopback:
		return ActionToggleLoopback
	case keyInboundOnly:
		return ActionToggleInboundOnly
	default:
		if _, isInnerTabKey := parseInnerTabKey(keyPress.String()); isInnerTabKey {
			return ActionJumpInnerTab
		}
		return ActionNone
	}
}
