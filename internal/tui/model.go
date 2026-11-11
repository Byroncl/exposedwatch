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
	Interval    time.Duration
	Version     string
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
	// that pass the loopback toggle (and, for the count, the text filter).
	connectionSummary      app.ConnectionSummary
	connectionDisplayCount int
	hideLoopback           bool
	connectionSortMode     ConnectionSortMode
	connectionFilterText   string

	activeTab    TabID
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
	return Model{
		scanner:    options.Scanner,
		interval:   interval,
		version:    options.Version,
		clock:      clock,
		refreshing: true,
		tracker:    newChangeTracker(),
		styles:     newStyles(),

		connections:           options.Connections,
		connectionsRefreshing: options.Connections != nil,
		connectionTracker:     newConnectionTracker(),
		hideLoopback:          true,
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
