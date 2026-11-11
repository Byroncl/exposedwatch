package tui

import (
	"errors"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Byroncl/exposedwatch/internal/app"
	"github.com/Byroncl/exposedwatch/internal/domain"
)

func TestDefaultSortIsSeverityDescendingThenPortAscending(testRun *testing.T) {
	model := newTestModel(testRun)
	want := []int{3306, 22, 443, 8080, 53}
	if got := visiblePorts(model); !equalInts(got, want) {
		testRun.Fatalf("order = %v, want %v", got, want)
	}
}

func TestSortCyclesThroughPortProcessAndBack(testRun *testing.T) {
	model := newTestModel(testRun)
	model = typeKeys(testRun, model, "s")
	if model.sortMode != SortByPort || !equalInts(visiblePorts(model), []int{22, 53, 443, 3306, 8080}) {
		testRun.Fatalf("port sort: mode=%v order=%v", model.sortMode, visiblePorts(model))
	}
	model = typeKeys(testRun, model, "s")
	// dnsmasq, mysqld, nginx, sshd, webapp
	if model.sortMode != SortByProcess || !equalInts(visiblePorts(model), []int{53, 3306, 443, 22, 8080}) {
		testRun.Fatalf("process sort: mode=%v order=%v", model.sortMode, visiblePorts(model))
	}
	model = typeKeys(testRun, model, "s")
	if model.sortMode != SortBySeverity {
		testRun.Fatalf("mode = %v, want severity", model.sortMode)
	}
}

func TestMinimumSeverityFilterCycles(testRun *testing.T) {
	model := newTestModel(testRun)
	model = typeKeys(testRun, model, "f")
	if model.minSeverity != domain.Low || len(model.visibleFindings) != 4 {
		testRun.Fatalf("low: min=%v rows=%d", model.minSeverity, len(model.visibleFindings))
	}
	model = typeKeys(testRun, model, "f", "f")
	if model.minSeverity != domain.High || !equalInts(visiblePorts(model), []int{3306, 22, 443}) {
		testRun.Fatalf("high: min=%v rows=%v", model.minSeverity, visiblePorts(model))
	}
	model = typeKeys(testRun, model, "f", "f")
	if model.minSeverity != domain.Info || len(model.visibleFindings) != 5 {
		testRun.Fatalf("wrapped: min=%v rows=%d", model.minSeverity, len(model.visibleFindings))
	}
}

func TestTextFilterAppliesOnEnterAndCancelsOnEscape(testRun *testing.T) {
	model := newTestModel(testRun)
	model = typeKeys(testRun, model, "/", "s", "s", "h")
	if model.viewState != ViewFilterInput {
		testRun.Fatalf("state = %v, want filter input", model.viewState)
	}
	model = typeKeys(testRun, model, "enter")
	if model.viewState != ViewTable || !equalInts(visiblePorts(model), []int{22}) {
		testRun.Fatalf("applied: state=%v rows=%v", model.viewState, visiblePorts(model))
	}
	model = typeKeys(testRun, model, "/", "x", "esc")
	if model.viewState != ViewTable || model.filterText != "ssh" || len(model.visibleFindings) != 1 {
		testRun.Fatalf("cancelled: state=%v filter=%q rows=%d", model.viewState, model.filterText, len(model.visibleFindings))
	}
}

func TestQuitKeyIsTypedWhileEditingFilter(testRun *testing.T) {
	model := newTestModel(testRun)
	model = typeKeys(testRun, model, "/")
	updated, command := model.Update(keyMessage("q"))
	if command != nil {
		testRun.Fatal("q must not quit while editing the filter")
	}
	if updated.(Model).filterInput != "q" {
		testRun.Fatalf("filterInput = %q", updated.(Model).filterInput)
	}
}

func TestQuitKeys(testRun *testing.T) {
	model := newTestModel(testRun)
	for _, key := range []string{"q", "ctrl+c"} {
		_, command := model.Update(keyMessage(key))
		if command == nil {
			testRun.Fatalf("%s returned no command", key)
		}
		if _, isQuit := command().(tea.QuitMsg); !isQuit {
			testRun.Fatalf("%s did not quit", key)
		}
	}
}

func TestCursorNavigationIsClamped(testRun *testing.T) {
	model := newTestModel(testRun)
	model = typeKeys(testRun, model, "k", "up")
	if model.cursor != 0 {
		testRun.Fatalf("cursor = %d at top", model.cursor)
	}
	model = typeKeys(testRun, model, "j", "down")
	if model.cursor != 2 {
		testRun.Fatalf("cursor = %d, want 2", model.cursor)
	}
	model = typeKeys(testRun, model, "G")
	if model.cursor != len(model.visibleFindings)-1 {
		testRun.Fatalf("cursor = %d at bottom", model.cursor)
	}
	model = typeKeys(testRun, model, "j")
	if model.cursor != len(model.visibleFindings)-1 {
		testRun.Fatalf("cursor moved past the end: %d", model.cursor)
	}
	model = typeKeys(testRun, model, "g")
	if model.cursor != 0 {
		testRun.Fatalf("cursor = %d after g", model.cursor)
	}
}

func TestRefreshKeyStartsScanAndTickSchedulesNext(testRun *testing.T) {
	model := newTestModel(testRun)
	updated, command := model.Update(keyMessage("r"))
	if command == nil || !updated.(Model).refreshing {
		testRun.Fatal("r must start a refresh")
	}
	if model.Init() == nil {
		testRun.Fatal("Init must schedule work")
	}
	_, tickCommand := model.Update(tickMessage{})
	if tickCommand == nil {
		testRun.Fatal("tick must schedule work")
	}
}

func TestScanCommandDeliversResult(testRun *testing.T) {
	want := sampleReport()
	command := scanCommand(fakeScanner{report: want}, func() time.Time { return testRefreshTime })
	result, isResult := command().(RefreshResultMessage)
	if !isResult || len(result.Report.Findings) != len(want.Findings) || result.Err != nil || !result.At.Equal(testRefreshTime) {
		testRun.Fatalf("result = %+v", result)
	}
	failing := scanCommand(fakeScanner{err: errDockerDown}, func() time.Time { return testRefreshTime })
	if failed := failing().(RefreshResultMessage); !errors.Is(failed.Err, errDockerDown) {
		testRun.Fatalf("err = %v", failed.Err)
	}
}

func TestRefreshSuccessUpdatesViewCountsAndTime(testRun *testing.T) {
	model := newTestModel(testRun)
	rendered := model.View()
	for _, expected := range []string{testVersion, "03:04:05", "critical 1", "high 2", "medium 0", "low 1", "info 1", "ports 5"} {
		if !strings.Contains(rendered, expected) {
			testRun.Errorf("view missing %q", expected)
		}
	}
	if model.refreshing {
		testRun.Error("refreshing must be cleared")
	}
}

func TestPartialFailureShowsWarningAndKeepsRows(testRun *testing.T) {
	model := newTestModel(testRun)
	partial := sampleReport()
	partial.SourceErrors = []app.SourceError{{Source: domain.SourceDocker, Err: errDockerDown}}
	model = applyMessage(testRun, model, RefreshResultMessage{Report: partial, At: testRefreshTime})
	rendered := model.View()
	if !strings.Contains(rendered, "docker") || !strings.Contains(rendered, "cannot connect") {
		testRun.Fatalf("warning missing:\n%s", rendered)
	}
	if len(model.visibleFindings) != 5 {
		testRun.Fatalf("rows = %d", len(model.visibleFindings))
	}
}

func TestFailedRefreshKeepsPreviousDataAndShowsError(testRun *testing.T) {
	model := newTestModel(testRun)
	model = applyMessage(testRun, model, RefreshResultMessage{Err: errDockerDown, At: testRefreshTime})
	if len(model.visibleFindings) != 5 || !strings.Contains(model.View(), "cannot connect") {
		testRun.Fatalf("rows=%d view=%s", len(model.visibleFindings), model.View())
	}
	model = applyMessage(testRun, model, RefreshResultMessage{Report: sampleReport(), At: testRefreshTime})
	if strings.Contains(model.View(), "cannot connect") {
		testRun.Fatal("error must clear after a successful refresh")
	}
}

func TestSelectionSurvivesRefresh(testRun *testing.T) {
	model := newTestModel(testRun)
	model = typeKeys(testRun, model, "j", "j") // port 443
	reordered := sampleReport()
	reordered.Findings = append([]domain.Finding{makeFinding(1, domain.Critical, "newcomer")}, reordered.Findings...)
	model = applyMessage(testRun, model, RefreshResultMessage{Report: reordered, At: testRefreshTime})
	if got := model.visibleFindings[model.cursor].Port.Number; got != 443 {
		testRun.Fatalf("selected port = %d, want 443", got)
	}
}

func TestDetailPaneShowsSelectedFinding(testRun *testing.T) {
	model := newTestModel(testRun)
	rendered := model.View()
	for _, expected := range []string{"Title for mysqld", "RULE-mysqld", testRecommendText} {
		if !strings.Contains(rendered, expected) {
			testRun.Errorf("detail missing %q", expected)
		}
	}
}

func TestEmptyStates(testRun *testing.T) {
	fresh := New(Options{Scanner: fakeScanner{}, Interval: testInterval, Version: testVersion})
	fresh = applyMessage(testRun, fresh, tea.WindowSizeMsg{Width: testWidth, Height: testHeight})
	if !strings.Contains(fresh.View(), "Scanning") {
		testRun.Error("want scanning placeholder before first result")
	}
	model := newTestModel(testRun)
	model = typeKeys(testRun, model, "/", "z", "z", "z", "enter")
	if !strings.Contains(model.View(), "No findings match") {
		testRun.Error("want empty filter message")
	}
	assertRendersWithin(testRun, model, testWidth, testHeight)
}

func TestHelpToggle(testRun *testing.T) {
	model := newTestModel(testRun)
	model = typeKeys(testRun, model, "?")
	if model.viewState != ViewHelp || !strings.Contains(model.View(), "cycle sort") {
		testRun.Fatalf("help not shown: state=%v", model.viewState)
	}
	model = typeKeys(testRun, model, "?")
	if model.viewState != ViewTable {
		testRun.Fatal("help not hidden")
	}
}

func TestResizeKeepsRenderingWithinBounds(testRun *testing.T) {
	model := newTestModel(testRun)
	for _, size := range []tea.WindowSizeMsg{
		{Width: 120, Height: 40}, {Width: 80, Height: 24}, {Width: 40, Height: 10},
		{Width: 20, Height: 5}, {Width: 10, Height: 2}, {Width: 1, Height: 1},
	} {
		resized := applyMessage(testRun, model, size)
		assertRendersWithin(testRun, resized, size.Width, size.Height)
		helpView := typeKeys(testRun, resized, "?")
		assertRendersWithin(testRun, helpView, size.Width, size.Height)
	}
}

func TestFooterShowsKeyHints(testRun *testing.T) {
	rendered := newTestModel(testRun).View()
	for _, hint := range []string{"quit", "refresh", "help"} {
		if !strings.Contains(rendered, hint) {
			testRun.Errorf("footer missing %q", hint)
		}
	}
}
