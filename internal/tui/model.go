package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Byroncl/exposedwatch/internal/app"
	"github.com/Byroncl/exposedwatch/internal/domain"
)

// Options configure a Model.
type Options struct {
	Scanner Scanner
	// Connections feeds the Connections tab; nil disables it.
	Connections ConnectionSnapshotter
	// Firewall feeds the Firewall module; nil disables it and the firewall is never read.
	Firewall FirewallEvaluator
	Interval time.Duration
	Version  string
	// Theme selects the palette adaptation; the zero value means ThemeAuto.
	Theme Theme
	// Clock returns the current time; defaults to time.Now.
	Clock func() time.Time
}

// Model is the Bubble Tea model of the exposure browser.
type Model struct {
	scanner  Scanner
	interval time.Duration
	version  string
	clock    func() time.Time

	width  int
	height int

	report      app.Report
	hasReport   bool
	lastRefresh time.Time
	refreshing  bool
	refreshErr  error
	tracker     changeTracker

	connections           ConnectionSnapshotter
	connectionSnapshot    app.ConnectionSnapshot
	hasConnections        bool
	connectionsAt         time.Time
	connectionsRefreshing bool
	connectionsErr        error
	connectionTracker     connectionTracker
	// connectionSummary and connectionDisplayCount describe the connections
	// that pass the loopback and inbound-only toggles (and, for the count, the text filter).
	connectionSummary      app.ConnectionSummary
	connectionDisplayCount int
	hideLoopback           bool
	inboundOnly            bool
	connectionSortMode     ConnectionSortMode
	connectionFilterText   string

	// The firewall evaluation is independent from the scan: its own command,
	// message and throttle. firewallSignature is the host port set that was
	// last submitted, so a change of the ports can trigger a new evaluation.
	firewall            FirewallEvaluator
	firewallReport      domain.FirewallReport
	firewallIndex       map[firewallPortKey]domain.FirewallFinding
	hasFirewall         bool
	firewallAt          time.Time
	firewallRequestedAt time.Time
	firewallSignature   string
	firewallRefreshing  bool
	firewallErr         error
	firewallSortMode    FirewallSortMode
	firewallFilterText  string

	// activeModule and innerTabs (the remembered inner tab of every module)
	// choose activeTab, the view the content area shows.
	activeModule  Module
	innerTabs     [moduleCount]InnerTab
	activeTab     TabID
	sidebarHidden bool
	theme         Theme

	visibleRows  []displayRow
	cursor       int
	scrollOffset int

	sortMode    SortMode
	minSeverity domain.Severity
	filterText  string
	filterInput string
	viewState   ViewState

	styles styles
}

// New builds a model; Init starts the first scan.
func New(options Options) Model {
	clock := options.Clock
	if clock == nil {
		clock = time.Now
	}
	interval := options.Interval
	if interval < MinimumInterval {
		interval = DefaultInterval
	}
	theme := options.Theme
	if theme == "" {
		theme = ThemeAuto
	}
	return Model{
		scanner:      options.Scanner,
		theme:        theme,
		activeModule: ModulePorts,
		activeTab:    TabFindings,
		interval:     interval,
		version:      options.Version,
		clock:        clock,
		refreshing:   true,
		tracker:      newChangeTracker(),
		styles:       newStyles(),

		connections:           options.Connections,
		connectionsRefreshing: options.Connections != nil,
		connectionTracker:     newConnectionTracker(),
		hideLoopback:          true,

		firewall: options.Firewall,
	}
}

// Init starts the first scan and the auto-refresh timer.
func (model Model) Init() tea.Cmd {
	commands := []tea.Cmd{scanCommand(model.scanner, model.clock), tickCommand(model.interval)}
	if model.connectionsEnabled() {
		commands = append(commands, connectionsCommand(model.connections, model.clock))
	}
	return tea.Batch(commands...)
}
