package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

func TestConnectionsModuleFollowsPortsAndHasLiveAsFirstTab(testRun *testing.T) {
	model := newTestModel(testRun)
	model = typeKeys(testRun, model, "tab")
	if model.activeModule != ModuleConnections || model.activeTab != TabConnections || TabConnections.Title() != "Connections" {
		testRun.Fatalf("module = %v tab = %v", model.activeModule, model.activeTab)
	}
	model = typeKeys(testRun, model, "shift+tab")
	if model.activeModule != ModulePorts || model.activeTab != TabFindings {
		testRun.Fatalf("shift+tab from connections -> %v / %v", model.activeModule, model.activeTab)
	}
	model = typeKeys(testRun, model, "shift+tab", "shift+tab")
	if model.activeModule != ModuleSettings {
		testRun.Fatalf("shift+tab wraps through the module list, got %v", model.activeModule)
	}
	if got := typeKeys(testRun, model, "tab", "tab", "tab").activeTab; got != TabConnections {
		testRun.Fatalf("tab from settings wraps to overview, ports, connections -> %v", got)
	}
}

func TestConnectionRowsHideLoopbackByDefaultAndToggleWithL(testRun *testing.T) {
	model := newConnectionsModel(testRun, fullConnections())
	if len(model.visibleRows) != 5 {
		testRun.Fatalf("rows = %v, want 5 without the loopback one", visibleLocalPorts(model))
	}
	if !model.hideLoopback || !strings.Contains(model.View(), "loopback:hidden") {
		testRun.Fatalf("loopback must be hidden by default and say so:\n%s", model.View())
	}
	model = typeKeys(testRun, model, "l")
	if len(model.visibleRows) != 6 || !strings.Contains(model.View(), "loopback:shown") {
		testRun.Fatalf("rows = %d after l:\n%s", len(model.visibleRows), model.View())
	}
	model = typeKeys(testRun, model, "l")
	if len(model.visibleRows) != 5 || !model.hideLoopback {
		testRun.Fatalf("second l must hide again, rows = %d", len(model.visibleRows))
	}
}

func TestLoopbackKeyIsIgnoredOnOtherTabs(testRun *testing.T) {
	model := typeKeys(testRun, newConnectionsModel(testRun, fullConnections()), "shift+tab", "l")
	if !model.hideLoopback {
		testRun.Fatal("l must only act on the Connections tab")
	}
}

func TestConnectionsDefaultOrderIsInboundThenLocalPortThenRemote(testRun *testing.T) {
	model := newConnectionsModel(testRun, fullConnections())
	if got := visibleLocalPorts(model); !equalInts(got, []int{22, 443, 9000, 4000, 50000}) {
		testRun.Fatalf("order = %v", got)
	}
	if rowForLocalPort(testRun, model, 22).connection.Direction != domain.ConnectionInbound {
		testRun.Fatal("first row must be inbound")
	}
}

func TestConnectionSortKeyCyclesEveryMode(testRun *testing.T) {
	model := newConnectionsModel(testRun, baselineConnections())
	model = observeConnections(testRun, model, fullConnections(), 60)
	cases := []struct {
		mode ConnectionSortMode
		want []int
	}{
		{ConnectionSortByLocalPort, []int{22, 443, 4000, 9000, 50000}},
		{ConnectionSortByRemoteAddress, []int{4000, 50000, 443, 22, 9000}},
		{ConnectionSortByProcess, []int{4000, 9000, 443, 50000, 22}},
		{ConnectionSortByAge, []int{9000, 50000, 22, 443, 4000}},
		{ConnectionSortByDirection, []int{22, 443, 9000, 4000, 50000}},
	}
	for _, testCase := range cases {
		model = typeKeys(testRun, model, "s")
		if model.connectionSortMode != testCase.mode || !equalInts(visibleLocalPorts(model), testCase.want) {
			testRun.Errorf("mode %v: order = %v, want %v", model.connectionSortMode, visibleLocalPorts(model), testCase.want)
		}
	}
	if model.sortMode != SortBySeverity {
		testRun.Fatal("the port sort mode must not change on the Connections tab")
	}
}

func TestNewBadgeBaselineThenNewInboundThenExpiry(testRun *testing.T) {
	model := newConnectionsModel(testRun, baselineConnections())
	for _, row := range model.visibleRows {
		if row.isNew {
			testRun.Fatalf("baseline flagged %v", row.connection)
		}
	}
	model = observeConnections(testRun, model, fullConnections(), 2)
	if !rowForLocalPort(testRun, model, 9000).isNew || !strings.Contains(model.View(), newBadgeText) {
		testRun.Fatalf("new inbound connection must carry NEW:\n%s", model.View())
	}
	if rowForLocalPort(testRun, model, 50000).isNew || rowForLocalPort(testRun, model, 22).isNew {
		testRun.Fatal("outbound and baseline connections must not carry NEW")
	}
	for snapshot := 1; snapshot < connectionNewBadgeRefreshCount; snapshot++ {
		model = observeConnections(testRun, model, fullConnections(), 2+snapshot)
	}
	if !rowForLocalPort(testRun, model, 9000).isNew {
		testRun.Fatal("NEW badge must last for the configured number of snapshots")
	}
	model = observeConnections(testRun, model, fullConnections(), 2+connectionNewBadgeRefreshCount)
	if rowForLocalPort(testRun, model, 9000).isNew {
		testRun.Fatal("NEW badge must expire")
	}
}

func TestConnectionAgeComesFromFirstSeenInThisSession(testRun *testing.T) {
	model := newConnectionsModel(testRun, baselineConnections())
	model = observeConnections(testRun, model, fullConnections(), 125)
	if age := rowForLocalPort(testRun, model, 22).connectionAge; age != 125*time.Second {
		testRun.Fatalf("persisting connection age = %v", age)
	}
	if age := rowForLocalPort(testRun, model, 9000).connectionAge; age != 0 {
		testRun.Fatalf("new connection age = %v", age)
	}
	if !strings.Contains(model.View(), "2m") || !strings.Contains(model.View(), "0s") {
		testRun.Fatalf("ages missing from the view:\n%s", model.View())
	}
}

func TestClosedConnectionsDisappearWithoutErrors(testRun *testing.T) {
	model := newConnectionsModel(testRun, fullConnections())
	model = observeConnections(testRun, model, baselineConnections(), 2)
	if got := visibleLocalPorts(model); !equalInts(got, []int{22, 443, 4000}) {
		testRun.Fatalf("rows = %v", got)
	}
	if model.connectionsErr != nil || strings.Contains(model.View(), "! ") {
		testRun.Fatalf("closed connections must not produce warnings:\n%s", model.View())
	}
}

func TestConnectionReadFailureKeepsDataShowsWarningAndPortsStillRefresh(testRun *testing.T) {
	model := newConnectionsModel(testRun, fullConnections())
	model = applyMessage(testRun, model, ConnectionsResultMessage{Err: errConnectionsDown, At: testRefreshTime.Add(time.Second)})
	if len(model.visibleRows) != 5 {
		testRun.Fatalf("previous rows must survive a failure, got %d", len(model.visibleRows))
	}
	if rendered := model.View(); !strings.Contains(rendered, "connections refresh failed") || !strings.Contains(rendered, "cannot read the connection tables") {
		testRun.Fatalf("warning missing:\n%s", rendered)
	}
	refreshed := applyMessage(testRun, model, RefreshResultMessage{Report: reportWith(sampleReport(), freshPortNumber), At: testRefreshTime.Add(2 * time.Second)})
	if len(refreshed.report.Ports) != len(sampleReport().Ports)+1 || refreshed.refreshErr != nil {
		testRun.Fatal("the port scan must refresh independently of the connection failure")
	}
	recovered := observeConnections(testRun, refreshed, fullConnections(), 3)
	if strings.Contains(recovered.View(), "connections refresh failed") {
		testRun.Fatal("warning must clear after a successful read")
	}
	for _, row := range recovered.visibleRows {
		if row.isNew {
			testRun.Fatalf("a failure followed by the same data must not fake NEW: %v", row.connection)
		}
	}
}

func TestConnectionFailureDoesNotBlockTheScanOrTheNextConnectionRead(testRun *testing.T) {
	model := newConnectionsModel(testRun, fullConnections())
	model.connectionsRefreshing = true
	model = applyMessage(testRun, model, ConnectionsResultMessage{Err: errConnectionsDown, At: testRefreshTime})
	if model.connectionsRefreshing || model.refreshing {
		testRun.Fatalf("connectionsRefreshing=%v refreshing=%v", model.connectionsRefreshing, model.refreshing)
	}
}

func TestTruncatedSnapshotShowsNotice(testRun *testing.T) {
	model := newConnectionsModel(testRun, fullConnections())
	truncated := snapshotOf(fullConnections())
	truncated.Truncated = true
	model = applyMessage(testRun, model, ConnectionsResultMessage{Snapshot: truncated, At: testRefreshTime.Add(time.Second)})
	if !strings.Contains(model.View(), "showing the first 6 connections") || len(model.visibleRows) != 5 {
		testRun.Fatalf("notice missing:\n%s", model.View())
	}
}

func TestDisabledConnectionsShowAMessageAndIgnoreKeys(testRun *testing.T) {
	model := New(Options{Scanner: fakeScanner{}, Interval: testInterval, Version: testVersion})
	model = applyMessage(testRun, model, tea.WindowSizeMsg{Width: testWidth, Height: testHeight})
	model = typeKeys(testRun, showModule(testRun, model, ModuleConnections), "l", "s", "r", "/", "esc")
	rendered := model.View()
	if !strings.Contains(rendered, "Connections are disabled") || strings.Contains(rendered, "Top remote addresses") {
		testRun.Fatalf("disabled message missing:\n%s", rendered)
	}
	if model.connectionsRefreshing {
		testRun.Fatal("a disabled model must never start a connection read")
	}
	if strings.Contains(strings.Split(rendered, "\n")[2], "conn ") {
		testRun.Fatal("header must not show a connection counter when disabled")
	}
}

func TestSearchFiltersRemoteAddressPortAndProcess(testRun *testing.T) {
	model := newConnectionsModel(testRun, fullConnections())
	cases := map[string][]int{
		"203.0.113": {22, 9000},
		"PSQL":      {50000},
		"5432":      {50000},
		"nomatch":   nil,
	}
	for text, want := range cases {
		filtered := typeKeys(testRun, model, "/")
		filtered = typeKeys(testRun, filtered, strings.Split(text, "")...)
		filtered = typeKeys(testRun, filtered, "enter")
		if got := visibleLocalPorts(filtered); !equalInts(got, want) {
			testRun.Errorf("filter %q: rows = %v, want %v", text, got, want)
		}
		if filtered.filterText != "" || filtered.connectionFilterText != text {
			testRun.Errorf("filter %q must stay private to the Connections tab", text)
		}
	}
}

func TestMinimumSeverityKeyIsIgnoredOnConnections(testRun *testing.T) {
	model := typeKeys(testRun, newConnectionsModel(testRun, fullConnections()), "f")
	if model.minSeverity != domain.Info || len(model.visibleRows) != 5 {
		testRun.Fatalf("min=%v rows=%d", model.minSeverity, len(model.visibleRows))
	}
}

func TestSummaryPanelsListTopRemotesAndInboundPorts(testRun *testing.T) {
	model := newConnectionsModel(testRun, fullConnections())
	rendered := model.View()
	for _, expected := range []string{"Top remote addresses", "Inbound by local port", testSshPeer, testWebPeer, "22/tcp", "443/tcp", "9000/tcp"} {
		if !strings.Contains(rendered, expected) {
			testRun.Errorf("panels lack %q:\n%s", expected, rendered)
		}
	}
	if strings.Contains(rendered, "5432/tcp") || strings.Contains(bottomHalf(rendered), testLoopbackAddress) {
		testRun.Fatalf("hidden loopback traffic must not reach the summaries:\n%s", rendered)
	}
	shown := typeKeys(testRun, model, "l").View()
	if !strings.Contains(bottomHalf(shown), testLoopbackAddress) || !strings.Contains(shown, "5432/tcp") {
		testRun.Fatalf("shown loopback traffic must reach the summaries:\n%s", shown)
	}
}

// bottomHalf returns the lines of the bottom panels (below the table).
func bottomHalf(rendered string) string {
	lines := strings.Split(rendered, "\n")
	return strings.Join(lines[len(lines)/2:], "\n")
}

func TestSummaryPanelsStackOnNarrowTallTerminalsAndHideWhenShort(testRun *testing.T) {
	model := newConnectionsModel(testRun, fullConnections())
	stacked := applyMessage(testRun, model, tea.WindowSizeMsg{Width: 50, Height: 40})
	if rendered := stacked.View(); !strings.Contains(rendered, "Top remote addresses") || !strings.Contains(rendered, "Inbound by local port") {
		testRun.Fatalf("narrow tall terminal must stack both panels:\n%s", rendered)
	}
	short := applyMessage(testRun, model, tea.WindowSizeMsg{Width: 120, Height: 18})
	if strings.Contains(short.View(), "Top remote addresses") {
		testRun.Fatal("panels must be hidden on short terminals")
	}
}

func TestHeaderShowsConnectionCounterAndDropsItWhenNarrow(testRun *testing.T) {
	model := newConnectionsModel(testRun, fullConnections())
	header := strings.Split(model.View(), "\n")[2]
	if !strings.Contains(header, "conn 5 (in 3 / out 2)") {
		testRun.Fatalf("header = %q", header)
	}
	if shown := typeKeys(testRun, model, "l"); !strings.Contains(strings.Split(shown.View(), "\n")[2], "conn 6 (in 4 / out 2)") {
		testRun.Fatalf("the counter must follow the loopback toggle: %q", strings.Split(shown.View(), "\n")[2])
	}
	narrow := applyMessage(testRun, model, tea.WindowSizeMsg{Width: 60, Height: 40})
	if strings.Contains(strings.Split(narrow.View(), "\n")[2], "conn ") {
		testRun.Fatal("the counter must be dropped when it does not fit")
	}
}

func TestTabStripCountsDisplayedConnections(testRun *testing.T) {
	model := newConnectionsModel(testRun, fullConnections())
	if !strings.Contains(model.View(), "[1 Live 5]") {
		testRun.Fatalf("tab strip:\n%s", model.View())
	}
	other := typeKeys(testRun, model, "2")
	if !strings.Contains(other.View(), " 1 Live 5 ") {
		testRun.Fatalf("count must stay visible from other tabs:\n%s", other.View())
	}
	if shown := typeKeys(testRun, model, "l"); !strings.Contains(shown.View(), "[1 Live 6]") {
		testRun.Fatal("count must follow the loopback toggle")
	}
}

func TestConnectionsFooterAndHelpDescribeTheNewKeys(testRun *testing.T) {
	model := newConnectionsModel(testRun, fullConnections())
	footer := model.footerLine(testWidth)
	for _, expected := range []string{"l loopback:hidden", "s sort", "/ filter"} {
		if !strings.Contains(footer, expected) {
			testRun.Errorf("footer lacks %q: %q", expected, footer)
		}
	}
	if strings.Contains(footer, "min-sev") {
		testRun.Error("f is ignored on this tab, so its hint must not show")
	}
	if shown := typeKeys(testRun, model, "l").footerLine(testWidth); !strings.Contains(shown, "l loopback:shown") {
		testRun.Errorf("footer after l: %q", shown)
	}
	help := typeKeys(testRun, model, "?").View()
	for _, expected := range []string{"1-9", "loopback"} {
		if !strings.Contains(help, expected) {
			testRun.Errorf("help lacks %q:\n%s", expected, help)
		}
	}
}

func TestEmptyConnectionTableMessages(testRun *testing.T) {
	model := New(Options{Scanner: fakeScanner{}, Connections: fakeConnectionSource{}, Interval: testInterval, Version: testVersion})
	model = applyMessage(testRun, model, tea.WindowSizeMsg{Width: testWidth, Height: testHeight})
	model = showModule(testRun, model, ModuleConnections)
	if !strings.Contains(model.View(), "Reading connections") {
		testRun.Fatalf("before the first read:\n%s", model.View())
	}
	model = observeConnections(testRun, model, []domain.Connection{connectionLoopbackInbound()}, 0)
	if !strings.Contains(model.View(), "Loopback-to-loopback") {
		testRun.Fatalf("hidden loopback must be explained:\n%s", model.View())
	}
}

func TestRefreshKeyAndTickStartIndependentConnectionReads(testRun *testing.T) {
	model := newConnectionsModel(testRun, fullConnections())
	updated, command := model.Update(keyMessage("r"))
	if command == nil || !updated.(Model).connectionsRefreshing || !updated.(Model).refreshing {
		testRun.Fatal("r must start the scan and the connection read")
	}
	busy := model
	busy.refreshing = true
	tickResult, tickBatch := busy.Update(tickMessage{})
	if tickBatch == nil || !tickResult.(Model).connectionsRefreshing {
		testRun.Fatal("a busy scan must not stop the connection read on tick")
	}
}

func TestConnectionsCommandDeliversResultAndFailure(testRun *testing.T) {
	clock := func() time.Time { return testRefreshTime }
	command := connectionsCommand(fakeConnectionSource{snapshot: snapshotOf(fullConnections())}, clock)
	result, isResult := command().(ConnectionsResultMessage)
	if !isResult || len(result.Snapshot.Connections) != 6 || result.Err != nil || !result.At.Equal(testRefreshTime) {
		testRun.Fatalf("result = %+v", result)
	}
	failed := connectionsCommand(fakeConnectionSource{err: errConnectionsDown}, clock)().(ConnectionsResultMessage)
	if failed.Err == nil {
		testRun.Fatal("failure must be reported in the message")
	}
}

func TestSelectionSurvivesConnectionRefresh(testRun *testing.T) {
	model := typeKeys(testRun, newConnectionsModel(testRun, fullConnections()), "j", "j")
	selected := model.visibleRows[model.cursor].connection.LocalPort
	model = observeConnections(testRun, model, append([]domain.Connection{
		testConnection(domain.ConnectionInbound, 21, testWebPeer, 1000, "ftpd"),
	}, fullConnections()...), 2)
	if got := model.visibleRows[model.cursor].connection.LocalPort; got != selected {
		testRun.Fatalf("selected local port = %d, want %d", got, selected)
	}
}

func TestSideBySidePanelsAreSeparatedByAGap(testRun *testing.T) {
	model := newConnectionsModel(testRun, fullConnections())
	for _, size := range []tea.WindowSizeMsg{{Width: 110, Height: 28}, {Width: 80, Height: 24}, {Width: 79, Height: 30}} {
		sized := applyMessage(testRun, model, size)
		rendered := sized.View()
		for _, glued := range []string{"╮╭", "╯╰", "││"} {
			if strings.Contains(rendered, glued) {
				testRun.Errorf("%dx%d: panels are glued together (%q):\n%s", size.Width, size.Height, glued, rendered)
			}
		}
		if !strings.Contains(rendered, "╮ ╭") || !strings.Contains(rendered, "╯ ╰") {
			testRun.Errorf("%dx%d: no gap between the panel borders:\n%s", size.Width, size.Height, rendered)
		}
		assertRendersWithin(testRun, sized, size.Width, size.Height)
	}
}
