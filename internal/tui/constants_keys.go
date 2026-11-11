package tui

// Key names as reported by tea.KeyMsg.String().
const (
	keyQuit            = "q"
	keyInterrupt       = "ctrl+c"
	keyUp              = "up"
	keyUpVim           = "k"
	keyDown            = "down"
	keyDownVim         = "j"
	keyTop             = "g"
	keyBottom          = "G"
	keyRefresh         = "r"
	keySort            = "s"
	keyMinimumSeverity = "f"
	keySearch          = "/"
	keyHelp            = "?"
	keyEscape          = "esc"
	keyNextTab         = "tab"
	keyPreviousTab     = "shift+tab"
	keyFindingsTab     = "1"
	keyAllPortsTab     = "2"
	keyChangesTab      = "3"
	keyConnectionsTab  = "4"
	keyLoopback        = "l"
	keyInboundOnly     = "i"
)

// keyBinding documents one key for the help screen.
type keyBinding struct {
	Keys        string
	Description string
}

// keyBindings lists every key in help order.
var keyBindings = []keyBinding{
	{Keys: "tab / shift+tab", Description: "next / previous tab"},
	{Keys: "1 / 2 / 3 / 4", Description: "jump to Findings / All ports / Changes / Connections"},
	{Keys: "up/k down/j", Description: "move selection"},
	{Keys: "g / G", Description: "jump to top / bottom"},
	{Keys: "r", Description: "refresh now"},
	{Keys: "s", Description: "cycle sort (severity, port, process); Connections: direction, local port, remote, process, age"},
	{Keys: "f", Description: "cycle minimum severity filter; Findings and All ports only"},
	{Keys: "/", Description: "free-text filter (enter apply, esc cancel); Connections: remote address, port, process"},
	{Keys: "l", Description: "show or hide loopback-to-loopback connections; Connections"},
	{Keys: "i", Description: "show only inbound connections, or all; Connections"},
	{Keys: "?", Description: "toggle this help"},
	{Keys: "q / ctrl+c", Description: "quit"},
}

// footerHint is one footer entry: the tabs it applies to and how early it is
// dropped when the footer is too wide (higher rank drops first, zero never).
// Label, when set, builds the text from the model instead of Text.
type footerHint struct {
	Text  string
	Label func(model Model) string
	// NeedsConnections hides the hint on the Connections tab while it is disabled.
	NeedsConnections bool
	AppliesToTabs    [tabCount]bool
	DropRank         int
}

// footerHints lists footer entries in display order.
var footerHints = []footerHint{
	{Text: hintTabs, AppliesToTabs: [tabCount]bool{true, true, true, true}, DropRank: dropRankNever},
	{Text: hintMove, AppliesToTabs: [tabCount]bool{true, true, true, true}, DropRank: dropRankMove},
	{Text: hintEnds, AppliesToTabs: [tabCount]bool{false, false, true, false}, DropRank: dropRankEnds},
	{Text: hintSort, NeedsConnections: true, AppliesToTabs: [tabCount]bool{true, true, false, true}, DropRank: dropRankSort},
	{Text: hintMinimum, AppliesToTabs: [tabCount]bool{true, true, false, false}, DropRank: dropRankMinimum},
	{Text: hintFilter, NeedsConnections: true, AppliesToTabs: [tabCount]bool{true, true, false, true}, DropRank: dropRankFilter},
	{Label: inboundHintText, NeedsConnections: true, AppliesToTabs: [tabCount]bool{false, false, false, true}, DropRank: dropRankInbound},
	{Label: loopbackHintText, NeedsConnections: true, AppliesToTabs: [tabCount]bool{false, false, false, true}, DropRank: dropRankLoopback},
	{Text: hintRefresh, AppliesToTabs: [tabCount]bool{true, true, true, true}, DropRank: dropRankRefresh},
	{Text: hintHelp, AppliesToTabs: [tabCount]bool{true, true, true, true}, DropRank: dropRankNever},
	{Text: hintQuit, AppliesToTabs: [tabCount]bool{true, true, true, true}, DropRank: dropRankNever},
}
