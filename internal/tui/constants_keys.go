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
)

// keyBinding documents one key for the help screen.
type keyBinding struct {
	Keys        string
	Description string
}

// keyBindings lists every key in help order.
var keyBindings = []keyBinding{
	{Keys: "tab / shift+tab", Description: "next / previous tab"},
	{Keys: "1 / 2 / 3", Description: "jump to Findings / All ports / Changes"},
	{Keys: "up/k down/j", Description: "move selection"},
	{Keys: "g / G", Description: "jump to top / bottom"},
	{Keys: "r", Description: "refresh now"},
	{Keys: "s", Description: "cycle sort (severity, port, process); Findings and All ports"},
	{Keys: "f", Description: "cycle minimum severity filter; Findings and All ports"},
	{Keys: "/", Description: "free-text filter (enter apply, esc cancel); Findings and All ports"},
	{Keys: "?", Description: "toggle this help"},
	{Keys: "q / ctrl+c", Description: "quit"},
}

// footerHint is one footer entry: the tabs it applies to and how early it is
// dropped when the footer is too wide (higher rank drops first, zero never).
type footerHint struct {
	Text          string
	AppliesToTabs [tabCount]bool
	DropRank      int
}

// footerHints lists footer entries in display order.
var footerHints = []footerHint{
	{Text: hintTabs, AppliesToTabs: [tabCount]bool{true, true, true}, DropRank: dropRankNever},
	{Text: hintMove, AppliesToTabs: [tabCount]bool{true, true, true}, DropRank: dropRankMove},
	{Text: hintEnds, AppliesToTabs: [tabCount]bool{false, false, true}, DropRank: dropRankEnds},
	{Text: hintSort, AppliesToTabs: [tabCount]bool{true, true, false}, DropRank: dropRankSort},
	{Text: hintMinimum, AppliesToTabs: [tabCount]bool{true, true, false}, DropRank: dropRankMinimum},
	{Text: hintFilter, AppliesToTabs: [tabCount]bool{true, true, false}, DropRank: dropRankFilter},
	{Text: hintRefresh, AppliesToTabs: [tabCount]bool{true, true, true}, DropRank: dropRankRefresh},
	{Text: hintHelp, AppliesToTabs: [tabCount]bool{true, true, true}, DropRank: dropRankNever},
	{Text: hintQuit, AppliesToTabs: [tabCount]bool{true, true, true}, DropRank: dropRankNever},
}
