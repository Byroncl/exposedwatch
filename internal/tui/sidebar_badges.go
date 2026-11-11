package tui

import "fmt"

// sidebarBadge is the small right-aligned note of a sidebar item: a count or a
// word, optionally followed by a highlighted flag.
type sidebarBadge struct {
	text string
	flag string
}

// width is the badge width in cells.
func (badge sidebarBadge) width() int {
	if badge.text == "" && badge.flag == "" {
		return 0
	}
	if badge.flag == "" {
		return len([]rune(badge.text))
	}
	if badge.text == "" {
		return len([]rune(badge.flag))
	}
	return len([]rune(badge.text)) + len(sidebarBadgeJoin) + len([]rune(badge.flag))
}

// sidebarBadgeFor builds the badge of a module from the data already in the model.
func (model Model) sidebarBadgeFor(module Module) sidebarBadge {
	switch module {
	case ModulePorts:
		badge := sidebarBadge{text: cappedCount(len(model.report.Findings))}
		if len(model.tracker.newAge) > 0 {
			badge.flag = sidebarBadgeNew
		}
		return badge
	case ModuleConnections:
		if !model.connectionsEnabled() {
			return sidebarBadge{text: sidebarBadgeOff}
		}
		if !model.hasConnections {
			return sidebarBadge{}
		}
		return sidebarBadge{text: cappedCount(model.connectionSummary.Total)}
	case ModuleAlerts:
		return sidebarBadge{text: sidebarBadgeOff}
	default:
		return sidebarBadge{}
	}
}

// cappedCount formats a count, showing "99+" beyond the cap.
func cappedCount(count int) string {
	if count > sidebarBadgeCap {
		return sidebarBadgeCapText
	}
	return fmt.Sprintf(sidebarCountFormat, count)
}
