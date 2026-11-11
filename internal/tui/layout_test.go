package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

var layoutSizes = []tea.WindowSizeMsg{
	{Width: 120, Height: 40}, {Width: 110, Height: 28}, {Width: 80, Height: 24}, {Width: 70, Height: 20},
	{Width: 40, Height: 10}, {Width: 30, Height: 6}, {Width: 20, Height: 5}, {Width: 10, Height: 2}, {Width: 1, Height: 1},
}

func TestEveryTabAndHelpStayInsideTheTerminal(testRun *testing.T) {
	base := newTestModel(testRun)
	base = refreshWith(testRun, base, reportWithout(sampleReportWithQuietPort(), 3306), 1)
	base = refreshWith(testRun, base, reportWith(sampleReportWithQuietPort(), freshPortNumber), 2)
	for _, size := range layoutSizes {
		sized := applyMessage(testRun, base, size)
		for _, key := range []string{"1", "2", "3"} {
			tabbed := typeKeys(testRun, sized, key)
			assertRendersWithin(testRun, tabbed, size.Width, size.Height)
			assertRendersWithin(testRun, typeKeys(testRun, tabbed, "G"), size.Width, size.Height)
			assertRendersWithin(testRun, typeKeys(testRun, tabbed, "?"), size.Width, size.Height)
			assertRendersWithin(testRun, typeKeys(testRun, tabbed, "/"), size.Width, size.Height)
		}
	}
}

func TestWarningsAndEmptyStatesStayInsideTheTerminal(testRun *testing.T) {
	failing := applyMessage(testRun, New(Options{Scanner: fakeScanner{}, Version: testVersion}),
		RefreshResultMessage{Err: errDockerDown, At: testRefreshTime})
	for _, size := range layoutSizes {
		sized := applyMessage(testRun, failing, size)
		for _, key := range []string{"1", "2", "3"} {
			assertRendersWithin(testRun, typeKeys(testRun, sized, key), size.Width, size.Height)
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
	}{
		{"findings", 110, 28, nil},
		{"all ports", 110, 28, []string{"2"}},
		{"changes", 110, 28, []string{"3"}},
		{"findings narrow", 70, 20, nil},
	}
	for _, preview := range previews {
		sized := applyMessage(testRun, model, tea.WindowSizeMsg{Width: preview.width, Height: preview.height})
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

// TestHeadersStartAboveTheirCellContent checks, for every tab and width, that
// each header label starts exactly above the first visible character of its cells.
func TestHeadersStartAboveTheirCellContent(testRun *testing.T) {
	model := previewModel(testRun)
	for _, width := range []int{70, 90, 110, 140} {
		for _, key := range []string{"1", "2", "3"} {
			sized := applyMessage(testRun, model, tea.WindowSizeMsg{Width: width, Height: 40})
			sized = typeKeys(testRun, sized, key)
			frame := sized.computeLayout()
			lines := sized.tableContent(frame)
			spec := tableSpecFor(sized.activeTab)
			shown, _ := spec.layoutColumns(panelContentWidth(frame.width))
			offset := selectionMarkWidth + len(columnSeparator)
			for _, candidate := range shown {
				assertColumnAligned(testRun, lines, offset, candidate.width, candidate.heading, width, key)
				offset += candidate.width + len(columnSeparator)
			}
			assertColumnAligned(testRun, lines, offset, frame.width, spec.flexHeading, width, key)
		}
	}
}

func assertColumnAligned(testRun *testing.T, lines []string, offset, columnWidth int, heading string, width int, key string) {
	testRun.Helper()
	headingStart := firstVisibleOffset(lines[0], offset, columnWidth)
	if heading == "" || headingStart < 0 {
		return
	}
	for _, line := range lines[1:] {
		if rowStart := firstVisibleOffset(line, offset, columnWidth); rowStart >= 0 && rowStart != headingStart {
			testRun.Errorf("width %d tab %s column %q: header at %d, row at %d: %q", width, key, heading, headingStart, rowStart, line)
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
