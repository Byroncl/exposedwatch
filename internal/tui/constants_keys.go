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
)

// keyBinding documents one key for the help screen and footer.
type keyBinding struct {
	Keys        string
	Description string
	FooterHint  string
}

// keyBindings lists every key in help order.
var keyBindings = []keyBinding{
	{Keys: "up/k down/j", Description: "move selection", FooterHint: ""},
	{Keys: "g / G", Description: "jump to top / bottom", FooterHint: ""},
	{Keys: "r", Description: "refresh now", FooterHint: "r refresh"},
	{Keys: "s", Description: "cycle sort (severity, port, process)", FooterHint: "s sort"},
	{Keys: "f", Description: "cycle minimum severity filter", FooterHint: "f min-sev"},
	{Keys: "/", Description: "free-text filter (enter apply, esc cancel)", FooterHint: "/ filter"},
	{Keys: "?", Description: "toggle this help", FooterHint: "? help"},
	{Keys: "q / ctrl+c", Description: "quit", FooterHint: "q quit"},
}
