package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

var layoutSizes = []tea.WindowSizeMsg{
	{Width: 120, Height: 40}, {Width: 110, Height: 28}, {Width: 80, Height: 24}, {Width: 70, Height: 20},
	{Width: 79, Height: 24}, {Width: 60, Height: 18}, {Width: 50, Height: 16}, {Width: 49, Height: 12},
	{Width: 40, Height: 10}, {Width: 30, Height: 6}, {Width: 20, Height: 5}, {Width: 10, Height: 2}, {Width: 1, Height: 1},
}

func TestEveryModuleViewAndHelpStayInsideTheTerminal(testRun *testing.T) {
	base := newConnectionsModel(testRun, fullConnections())
	base = refreshWith(testRun, base, reportWithout(sampleReportWithQuietPort(), 3306), 1)
	base = refreshWith(testRun, base, reportWith(sampleReportWithQuietPort(), freshPortNumber), 2)
	for _, size := range layoutSizes {
		sized := applyMessage(testRun, base, size)
		for _, target := range everyModuleView() {
			tabbed := showView(testRun, sized, target)
			assertRendersWithin(testRun, tabbed, size.Width, size.Height)
			assertRendersWithin(testRun, typeKeys(testRun, tabbed, "G"), size.Width, size.Height)
			assertRendersWithin(testRun, typeKeys(testRun, tabbed, "?"), size.Width, size.Height)
			assertRendersWithin(testRun, typeKeys(testRun, tabbed, "/"), size.Width, size.Height)
			assertRendersWithin(testRun, typeKeys(testRun, tabbed, "b"), size.Width, size.Height)
		}
	}
}

func TestWarningsAndEmptyStatesStayInsideTheTerminal(testRun *testing.T) {
	failing := applyMessage(testRun, New(Options{Scanner: fakeScanner{}, Version: testVersion}),
		RefreshResultMessage{Err: errDockerDown, At: testRefreshTime})
	for _, size := range layoutSizes {
		sized := applyMessage(testRun, failing, size)
		for _, target := range everyModuleView() {
			assertRendersWithin(testRun, showView(testRun, sized, target), size.Width, size.Height)
		}
	}
}

func TestConnectionWarningsAndTruncationStayInsideTheTerminal(testRun *testing.T) {
	failing := applyMessage(testRun, newConnectionsModel(testRun, fullConnections()),
		ConnectionsResultMessage{Err: errConnectionsDown, At: testRefreshTime})
	truncatedSnapshot := snapshotOf(fullConnections())
	truncatedSnapshot.Truncated = true
	truncated := applyMessage(testRun, newConnectionsModel(testRun, fullConnections()),
		ConnectionsResultMessage{Snapshot: truncatedSnapshot, At: testRefreshTime})
	for _, size := range layoutSizes {
		for _, model := range []Model{failing, truncated} {
			sized := applyMessage(testRun, model, size)
			assertRendersWithin(testRun, sized, size.Width, size.Height)
			assertRendersWithin(testRun, typeKeys(testRun, sized, "l", "G"), size.Width, size.Height)
			for _, target := range everyModuleView() {
				assertRendersWithin(testRun, showView(testRun, sized, target), size.Width, size.Height)
			}
		}
	}
}

func TestZeroSizeUsesFallbackAndDoesNotPanic(testRun *testing.T) {
	model := New(Options{Scanner: fakeScanner{}, Version: testVersion})
	assertRendersWithin(testRun, model, fallbackWidth, fallbackHeight)
}

// TestLayoutPreview prints plain-text previews of the realistic fixture.
// Run: go test ./internal/tui -run TestLayoutPreview -v
func TestLayoutPreview(testRun *testing.T) {
	model := previewModel(testRun)
	previews := []struct {
		name   string
		width  int
		height int
		keys   []string
		// needsRoot shows the firewall as it looks when the rules cannot be read.
		needsRoot bool
		// build replaces the realistic fixture, for the settings previews.
		build func(testRun *testing.T) Model
	}{
		{"overview", 110, 28, []string{"shift+tab"}, false, nil},
		{"ports / findings", 110, 28, nil, false, nil},
		{"ports / all ports", 110, 28, []string{"2"}, false, nil},
		{"ports / changes", 110, 28, []string{"3"}, false, nil},
		{"connections / live", 110, 28, []string{"tab"}, false, nil},
		{"connections / top remote", 110, 28, []string{"tab", "2"}, false, nil},
		{"firewall / status", 110, 28, []string{"tab", "tab"}, false, nil},
		{"firewall / ports", 110, 28, []string{"tab", "tab", "2"}, false, nil},
		{"firewall / ports compact sidebar", 70, 20, []string{"tab", "tab", "2"}, false, nil},
		{"firewall / status unavailable (needs root)", 110, 28, []string{"tab", "tab"}, true, nil},
		{"ports / findings compact sidebar", 70, 20, nil, false, nil},
		{"ports / findings compact sidebar (lowest width)", 50, 16, nil, false, nil},
		{"ports / findings hidden sidebar", 49, 16, nil, false, nil},
		{"ports / findings hidden sidebar", 40, 12, nil, false, nil},
		{"connections / live compact sidebar", 70, 20, []string{"tab"}, false, nil},
		{"settings / telegram (token set, unsaved)", 110, 28, []string{"shift+tab", "shift+tab", "enter"}, false, previewSettingsModel},
		{"settings / telegram editing the chat id", 110, 28, []string{"shift+tab", "shift+tab", "j", "j", "enter", "4", "2"}, false, previewSettingsModel},
		{"settings / telegram environment-sourced token", 110, 28, []string{"shift+tab", "shift+tab"}, false, previewEnvironmentSettingsModel},
		{"alerts / status", 110, 28, []string{"shift+tab", "shift+tab", "shift+tab"}, false, previewSettingsModel},
		{"alerts / preview", 110, 28, []string{"shift+tab", "shift+tab", "shift+tab", "2"}, false, previewSettingsModel},
		{"settings / telegram compact sidebar", 70, 20, []string{"shift+tab", "shift+tab"}, false, previewSettingsModel},
	}
	for _, preview := range previews {
		base := model
		if preview.build != nil {
			base = preview.build(testRun)
		}
		if preview.needsRoot {
			base = applyMessage(testRun, model, FirewallResultMessage{Report: previewUnavailableFirewallReport(), At: timeAfter(3)})
		}
		sized := applyMessage(testRun, base, tea.WindowSizeMsg{Width: preview.width, Height: preview.height})
		sized = typeKeys(testRun, sized, preview.keys...)
		assertRendersWithin(testRun, sized, preview.width, preview.height)
		if testing.Verbose() {
			fmt.Printf("\n=== %s (%dx%d) ===\n%s\n", preview.name, preview.width, preview.height, sized.View())
		}
	}
}

func TestColumnsDropByPriorityAndKeepTheTitleWide(testRun *testing.T) {
	spec := tableSpecFor(TabFindings)
	previousCount := -1
	for _, contentWidth := range []int{40, 60, 80, 100, 140} {
		shown, flexWidth := spec.layoutColumns(contentWidth)
		if flexWidth < minTitleWidth && len(shown) > 3 {
			testRun.Errorf("width %d: flexible column only %d cells with %d columns", contentWidth, flexWidth, len(shown))
		}
		if len(shown) < previousCount {
			testRun.Errorf("width %d: %d columns, fewer than at a narrower width", contentWidth, len(shown))
		}
		previousCount = len(shown)
	}
	narrow, _ := spec.layoutColumns(10)
	if len(narrow) != 3 {
		testRun.Fatalf("mandatory columns = %d, want 3", len(narrow))
	}
}

func TestDetailPanelIsHiddenWhenTheTerminalIsTooShort(testRun *testing.T) {
	model := applyMessage(testRun, newTestModel(testRun), tea.WindowSizeMsg{Width: 80, Height: 18})
	if model.computeLayout().showDetail || strings.Contains(model.View(), "Details") {
		testRun.Fatal("detail panel must be hidden on short terminals")
	}
	tall := applyMessage(testRun, model, tea.WindowSizeMsg{Width: 80, Height: 30})
	if !tall.computeLayout().showDetail || !strings.Contains(tall.View(), "Details") {
		testRun.Fatal("detail panel must be visible on tall terminals")
	}
}

func TestFooterDropsLeastImportantHintsFirst(testRun *testing.T) {
	model := newTestModel(testRun)
	texts := model.footerHintTexts(40)
	joined := strings.Join(texts, footerHintJoin)
	if len(joined) > 40 || !strings.Contains(joined, "q quit") || !strings.Contains(joined, "? help") {
		testRun.Fatalf("footer = %q", joined)
	}
}

// TestHeadersStartAboveTheirCellContent checks, for every table-bearing inner
// tab and width, that each header label starts exactly above the first visible
// character of its cells.
func TestHeadersStartAboveTheirCellContent(testRun *testing.T) {
	model := previewModel(testRun)
	for _, width := range []int{50, 70, 90, 110, 140} {
		for _, target := range everyModuleView() {
			sized := applyMessage(testRun, model, tea.WindowSizeMsg{Width: width, Height: 40})
			sized = showView(testRun, sized, target)
			if !sized.activeTab.isTable() {
				continue
			}
			frame := sized.computeLayout()
			lines := sized.tableContent(frame)
			spec := tableSpecFor(sized.activeTab)
			shown, _ := spec.layoutColumns(panelContentWidth(frame.contentWidth))
			offset := selectionMarkWidth + len(columnSeparator)
			label := fmt.Sprintf("%v/%d", target.module, target.inner)
			for _, candidate := range shown {
				assertColumnAligned(testRun, lines, offset, candidate.width, candidate.heading, width, label)
				offset += candidate.width + len(columnSeparator)
			}
			assertColumnAligned(testRun, lines, offset, frame.contentWidth, spec.flexHeading, width, label)
		}
	}
}

func assertColumnAligned(testRun *testing.T, lines []string, offset, columnWidth int, heading string, width int, label string) {
	testRun.Helper()
	headingStart := firstVisibleOffset(lines[0], offset, columnWidth)
	if heading == "" || headingStart < 0 {
		return
	}
	for _, line := range lines[1:] {
		if rowStart := firstVisibleOffset(line, offset, columnWidth); rowStart >= 0 && rowStart != headingStart {
			testRun.Errorf("width %d view %s column %q: header at %d, row at %d: %q", width, label, heading, headingStart, rowStart, line)
			return
		}
	}
}

// firstVisibleOffset returns the offset of the first non-space cell in the
// column window of a line, or -1 when the window is blank.
func firstVisibleOffset(line string, offset, columnWidth int) int {
	runes := []rune(line)
	for position := offset; position < offset+columnWidth && position < len(runes); position++ {
		if runes[position] != ' ' {
			return position
		}
	}
	return -1
}

// TestPortTablesKeepTheProcessColumnAtEveryCommonWidth guards the column drop
// order: source first, then bind, then exposure, process last.
func TestPortTablesKeepTheProcessColumnAtEveryCommonWidth(testRun *testing.T) {
	model := previewModel(testRun)
	for _, width := range []int{140, 120, 110, 100} {
		for _, target := range []moduleView{{ModulePorts, 0}, {ModulePorts, 1}} {
			sized := showView(testRun, applyMessage(testRun, model, tea.WindowSizeMsg{Width: width, Height: 40}), target)
			frame := sized.computeLayout()
			if frame.sidebar != SidebarFull {
				testRun.Fatalf("width %d: expected the full sidebar", width)
			}
			spec := tableSpecFor(sized.activeTab)
			shown, flexWidth := spec.layoutColumns(panelContentWidth(frame.contentWidth))
			heading := sized.tableContent(frame)[0]
			if !strings.Contains(heading, headingProcessShort) {
				testRun.Errorf("width %d %v: header lacks PROCESS: %q", width, target, heading)
			}
			if flexWidth < minTitleWidth {
				testRun.Errorf("width %d %v: title column %d < %d", width, target, flexWidth, minTitleWidth)
			}
			for _, candidate := range shown {
				if candidate.heading == headingSource && !strings.Contains(heading, headingProcessShort) {
					testRun.Errorf("width %d: SOURCE kept without PROCESS", width)
				}
			}
		}
	}
}
