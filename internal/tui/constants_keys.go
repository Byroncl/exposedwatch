package tui

// Key names as reported by tea.KeyMsg.String().
const (
	keyQuit             = "q"
	keyInterrupt        = "ctrl+c"
	keyUp               = "up"
	keyUpVim            = "k"
	keyDown             = "down"
	keyDownVim          = "j"
	keyTop              = "g"
	keyBottom           = "G"
	keyRefresh          = "r"
	keySort             = "s"
	keyMinimumSeverity  = "f"
	keySearch           = "/"
	keyHelp             = "?"
	keyEscape           = "esc"
	keyEnter            = "enter"
	keyNextModule       = "tab"
	keyPreviousModule   = "shift+tab"
	keyNextInnerTab     = "]"
	keyPreviousInnerTab = "["
	keySidebar          = "b"
	keyLoopback         = "l"
	keyInboundOnly      = "i"
)

// keyBinding documents one key for the help screen.
type keyBinding struct {
	Keys        string
	Description string
}

// keyBindings lists every key in help order.
var keyBindings = []keyBinding{
	{Keys: "tab / shift+tab", Description: "next / previous module"},
	{Keys: "] / [", Description: "next / previous tab of the module"},
	{Keys: "1-9", Description: "jump to that tab of the module (Ports: Findings, All ports, Changes; Connections: Live, Top remote; Firewall: Status, Ports; Alerts: Status, Preview; Settings: Telegram, Display)"},
	{Keys: "b", Description: "show or hide the sidebar"},
	{Keys: "up/k down/j", Description: "move selection"},
	{Keys: "g / G", Description: "jump to top / bottom"},
	{Keys: "enter", Description: "Settings > Telegram: toggle, cycle, start editing a value, or run [ Save ] / [ Send test message ]"},
	{Keys: "while editing", Description: "Settings > Telegram: type or paste the value; enter saves the value, esc cancels, backspace deletes; no other key acts (ctrl+c still quits)"},
	{Keys: "r", Description: "refresh now (also re-reads the firewall rules)"},
	{Keys: "s", Description: "cycle sort (severity, port, process); Connections: direction, local port, remote, process, age; Firewall ports: verdict, port, process"},
	{Keys: "f", Description: "cycle minimum severity filter; Findings and All ports only"},
	{Keys: "/", Description: "free-text filter (enter apply, esc cancel); Connections: remote address, port, process; Firewall ports: port, process, verdict, reason"},
	{Keys: "l", Description: "show or hide loopback-to-loopback connections; Connections (Live, Top remote)"},
	{Keys: "i", Description: "show only inbound connections, or all; Connections (Live, Top remote)"},
	{Keys: "?", Description: "toggle this help"},
	{Keys: "q / ctrl+c", Description: "quit"},
}
