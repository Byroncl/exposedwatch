package tui

// tabSet is a set of views. A nil set means every view.
type tabSet []TabID

func (set tabSet) contains(tab TabID) bool {
	if set == nil {
		return true
	}
	for _, member := range set {
		if member == tab {
			return true
		}
	}
	return false
}

// Views that share a footer hint.
var (
	tableTabs      = tabSet{TabFindings, TabAllPorts, TabChanges, TabConnections, TabFirewallPorts}
	portTableTabs  = tabSet{TabFindings, TabAllPorts}
	searchableTabs = tabSet{TabFindings, TabAllPorts, TabConnections, TabFirewallPorts}
	connectionTabs = tabSet{TabConnections, TabTopRemote}
	changesTabs    = tabSet{TabChanges}
)

// footerHint is one footer entry: the views it applies to and how early it is
// dropped when the footer is too wide (higher rank drops first, zero never).
// Label, when set, builds the text from the model instead of Text.
type footerHint struct {
	Text  string
	Label func(model Model) string
	// NeedsConnections hides the hint on the connection views while they are disabled.
	NeedsConnections bool
	// NeedsFirewall hides the hint on the Firewall ports view while the firewall is disabled.
	NeedsFirewall bool
	AppliesTo     tabSet
	DropRank      int
}

// footerHints lists footer entries in display order.
var footerHints = []footerHint{
	{Text: hintModules, DropRank: dropRankNever},
	{Text: hintInnerTabs, DropRank: dropRankInnerTabs},
	{Text: hintMove, AppliesTo: tableTabs, DropRank: dropRankMove},
	{Text: hintEnds, AppliesTo: changesTabs, DropRank: dropRankEnds},
	{Text: hintSort, NeedsConnections: true, NeedsFirewall: true, AppliesTo: searchableTabs, DropRank: dropRankSort},
	{Text: hintMinimum, AppliesTo: portTableTabs, DropRank: dropRankMinimum},
	{Text: hintFilter, NeedsConnections: true, NeedsFirewall: true, AppliesTo: searchableTabs, DropRank: dropRankFilter},
	{Label: inboundHintText, NeedsConnections: true, AppliesTo: connectionTabs, DropRank: dropRankInbound},
	{Label: loopbackHintText, NeedsConnections: true, AppliesTo: connectionTabs, DropRank: dropRankLoopback},
	{Text: hintRefresh, DropRank: dropRankRefresh},
	{Text: hintSidebar, DropRank: dropRankSidebar},
	{Text: hintHelp, DropRank: dropRankNever},
	{Text: hintQuit, DropRank: dropRankNever},
}

// shownOn tells whether the hint is listed for the model's active view.
func (hint footerHint) shownOn(model Model) bool {
	if !hint.AppliesTo.contains(model.activeTab) {
		return false
	}
	if hint.NeedsConnections && model.activeTab.usesConnections() && !model.connectionsEnabled() {
		return false
	}
	return !(hint.NeedsFirewall && model.activeTab.usesFirewall() && !model.firewallEnabled())
}
