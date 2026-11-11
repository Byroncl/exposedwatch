package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func sizedModel(testRun *testing.T, model Model, width, height int) Model {
	testRun.Helper()
	return applyMessage(testRun, model, tea.WindowSizeMsg{Width: width, Height: height})
}

func TestSidebarModeFollowsTheWidth(testRun *testing.T) {
	cases := []struct {
		width int
		want  SidebarMode
		cells int
	}{
		{120, SidebarFull, 19}, {80, SidebarFull, 19}, {79, SidebarCompact, 5},
		{50, SidebarCompact, 5}, {49, SidebarHidden, 0}, {30, SidebarHidden, 0},
	}
	for _, testCase := range cases {
		frame := sizedModel(testRun, newTestModel(testRun), testCase.width, 30).computeLayout()
		if frame.sidebar != testCase.want || frame.sidebarWidth != testCase.cells {
			testRun.Errorf("width %d: mode %v width %d, want %v / %d", testCase.width, frame.sidebar, frame.sidebarWidth, testCase.want, testCase.cells)
		}
		wantContent := testCase.width
		if testCase.cells > 0 {
			wantContent = testCase.width - testCase.cells - sidebarContentGap
		}
		if frame.contentWidth != wantContent {
			testRun.Errorf("width %d: content width %d, want %d", testCase.width, frame.contentWidth, wantContent)
		}
	}
}

func TestSidebarKeyHidesAndShowsOnTopOfTheResponsiveMode(testRun *testing.T) {
	model := sizedModel(testRun, newTestModel(testRun), 120, 30)
	hidden := typeKeys(testRun, model, "b")
	if frame := hidden.computeLayout(); frame.sidebar != SidebarHidden || frame.contentWidth != 120 {
		testRun.Fatalf("b must hide the sidebar: %+v", frame)
	}
	if frame := typeKeys(testRun, hidden, "b").computeLayout(); frame.sidebar != SidebarFull {
		testRun.Fatalf("second b must show it again: %+v", frame)
	}
	compact := typeKeys(testRun, sizedModel(testRun, model, 60, 30), "b", "b")
	if frame := compact.computeLayout(); frame.sidebar != SidebarCompact {
		testRun.Fatalf("compact must come back after b b: %+v", frame)
	}
	narrow := typeKeys(testRun, sizedModel(testRun, model, 40, 30), "b")
	if frame := narrow.computeLayout(); frame.sidebar != SidebarHidden {
		testRun.Fatalf("b cannot show the sidebar below the minimum width: %+v", frame)
	}
	if typeKeys(testRun, narrow, "b").sidebarHidden == narrow.sidebarHidden {
		testRun.Fatal("b must keep toggling the preference")
	}
}

func TestSidebarMarksTheSelectedModuleWithAPlainTextCue(testRun *testing.T) {
	model := sizedModel(testRun, newTestModel(testRun), 120, 30)
	for _, module := range []Module{ModulePorts, ModuleConnections, ModuleSettings} {
		shown := showModule(testRun, model, module)
		rendered := strings.Join(shown.sidebarPanel(shown.computeLayout()), "\n")
		if !strings.Contains(rendered, "▌ "+module.Title()) {
			testRun.Errorf("%v: marker missing:\n%s", module, rendered)
		}
		if strings.Count(rendered, "▌") != 1 {
			testRun.Errorf("%v: exactly one module is marked:\n%s", module, rendered)
		}
	}
}

func TestCompactSidebarUsesShortLabels(testRun *testing.T) {
	model := sizedModel(testRun, newTestModel(testRun), 60, 30)
	frame := model.computeLayout()
	lines := model.sidebarPanel(frame)
	rendered := strings.Join(lines, "\n")
	if !strings.Contains(rendered, "▌P") || !strings.Contains(rendered, " O") {
		testRun.Fatalf("compact sidebar:\n%s", rendered)
	}
	for _, line := range lines {
		if lipgloss.Width(line) != frame.sidebarWidth {
			testRun.Errorf("line %q is %d cells, want %d", line, lipgloss.Width(line), frame.sidebarWidth)
		}
	}
}

func TestSidebarBadges(testRun *testing.T) {
	model := newConnectionsModel(testRun, fullConnections())
	model = refreshWith(testRun, model, reportWith(sampleReport(), freshPortNumber), 1)
	model = sizedModel(testRun, model, 120, 30)
	rendered := strings.Join(model.sidebarPanel(model.computeLayout()), "\n")
	for _, want := range []string{"Ports", "6 NEW", "Connections", " 5", "Alerts", "off"} {
		if !strings.Contains(rendered, want) {
			testRun.Errorf("sidebar lacks %q:\n%s", want, rendered)
		}
	}
	quiet := sizedModel(testRun, newTestModel(testRun), 120, 30)
	plain := strings.Join(quiet.sidebarPanel(quiet.computeLayout()), "\n")
	if strings.Contains(plain, "NEW") {
		testRun.Errorf("no new port, no NEW marker:\n%s", plain)
	}
	if !strings.Contains(plain, "off") {
		testRun.Errorf("disabled connections and alerts show off:\n%s", plain)
	}
}

func TestSidebarAndContentNeverTouch(testRun *testing.T) {
	model := previewModel(testRun)
	borders := "╭│╰╮╯"
	for _, size := range []tea.WindowSizeMsg{{Width: 120, Height: 40}, {Width: 110, Height: 28}, {Width: 80, Height: 24}, {Width: 70, Height: 20}, {Width: 50, Height: 16}} {
		for _, target := range everyModuleView() {
			sized := showView(testRun, sizedModel(testRun, model, size.Width, size.Height), target)
			frame := sized.computeLayout()
			lines := strings.Split(sized.View(), "\n")
			for index, line := range lines {
				runes := []rune(line)
				if frame.sidebarWidth == 0 || len(runes) <= frame.sidebarWidth || !strings.ContainsRune(borders, runes[0]) {
					continue
				}
				if !strings.ContainsRune(borders, runes[frame.sidebarWidth-1]) {
					testRun.Errorf("%dx%d %v/%d line %d: sidebar does not end with a border: %q", size.Width, size.Height, target.module, target.inner, index, line)
				}
				if runes[frame.sidebarWidth] != ' ' {
					testRun.Errorf("%dx%d %v/%d line %d: no gap between sidebar and content: %q", size.Width, size.Height, target.module, target.inner, index, line)
				}
			}
		}
	}
}
