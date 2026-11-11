package tui

import (
	"strings"
	"testing"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

func newQuietPortModel(testRun *testing.T) Model {
	testRun.Helper()
	return refreshWith(testRun, newTestModel(testRun), sampleReportWithQuietPort(), 0)
}

func TestNumberKeysJumpToTabs(testRun *testing.T) {
	model := newTestModel(testRun)
	for key, want := range map[string]TabID{"2": TabAllPorts, "3": TabChanges, "4": TabConnections, "1": TabFindings} {
		model = typeKeys(testRun, model, key)
		if model.activeTab != want {
			testRun.Fatalf("key %s: tab = %v, want %v", key, model.activeTab, want)
		}
	}
}

func TestTabKeysCycleBothWays(testRun *testing.T) {
	model := newTestModel(testRun)
	model = typeKeys(testRun, model, "tab")
	if model.activeTab != TabAllPorts {
		testRun.Fatalf("tab -> %v", model.activeTab)
	}
	model = typeKeys(testRun, model, "tab", "tab", "tab")
	if model.activeTab != TabFindings {
		testRun.Fatalf("wrapped -> %v", model.activeTab)
	}
	model = typeKeys(testRun, model, "shift+tab", "shift+tab")
	if model.activeTab != TabChanges {
		testRun.Fatalf("shift+tab -> %v", model.activeTab)
	}
}

func TestAllPortsTabListsPortsWithoutFindings(testRun *testing.T) {
	model := typeKeys(testRun, newQuietPortModel(testRun), "2")
	if len(model.visibleRows) != 6 {
		testRun.Fatalf("rows = %d, want 6", len(model.visibleRows))
	}
	quiet := findRow(testRun, model, 631)
	if quiet.hasSeverity {
		testRun.Fatal("a port without finding must have no severity")
	}
	if !strings.Contains(model.View(), "cupsd") {
		testRun.Fatalf("quiet port missing from view:\n%s", model.View())
	}
	findings := typeKeys(testRun, model, "1")
	if len(findings.visibleRows) != 5 {
		testRun.Fatalf("findings rows = %d, want 5", len(findings.visibleRows))
	}
}

func TestSortAndFilterApplyToAllPortsTab(testRun *testing.T) {
	model := typeKeys(testRun, newQuietPortModel(testRun), "2")
	if got := visiblePorts(model); !equalInts(got, []int{3306, 22, 443, 8080, 53, 631}) {
		testRun.Fatalf("default order = %v (no-finding ports last)", got)
	}
	model = typeKeys(testRun, model, "s")
	if got := visiblePorts(model); !equalInts(got, []int{22, 53, 443, 631, 3306, 8080}) {
		testRun.Fatalf("port order = %v", got)
	}
	model = typeKeys(testRun, model, "/", "c", "u", "p", "s", "enter")
	if got := visiblePorts(model); !equalInts(got, []int{631}) {
		testRun.Fatalf("text filter = %v", got)
	}
	model = typeKeys(testRun, model, "/", "esc")
	model = typeKeys(testRun, model, "f")
	if model.minSeverity != domain.Low {
		testRun.Fatalf("min severity = %v", model.minSeverity)
	}
	if len(model.visibleRows) != 0 {
		testRun.Fatalf("quiet port must be hidden above info, got %v", visiblePorts(model))
	}
}

func TestSortFilterAndSearchKeysAreIgnoredOnChangesTab(testRun *testing.T) {
	model := typeKeys(testRun, newTestModel(testRun), "3")
	model = typeKeys(testRun, model, "s", "f", "/")
	if model.sortMode != SortBySeverity || model.minSeverity != domain.Info || model.viewState != ViewTable {
		testRun.Fatalf("sort=%v min=%v state=%v", model.sortMode, model.minSeverity, model.viewState)
	}
}

func TestChangesTabKeepsNavigationKeys(testRun *testing.T) {
	model := newTestModel(testRun)
	model = refreshWith(testRun, model, reportWithout(sampleReport(), 3306, 53, 22), 1)
	model = typeKeys(testRun, model, "3", "j", "j")
	if model.cursor != 2 {
		testRun.Fatalf("cursor = %d", model.cursor)
	}
	model = typeKeys(testRun, model, "g")
	if model.cursor != 0 {
		testRun.Fatalf("cursor = %d after g", model.cursor)
	}
}

func TestSwitchingTabsResetsTheCursor(testRun *testing.T) {
	model := typeKeys(testRun, newTestModel(testRun), "j", "j", "2")
	if model.cursor != 0 || model.scrollOffset != 0 {
		testRun.Fatalf("cursor=%d scroll=%d", model.cursor, model.scrollOffset)
	}
}

func TestActiveTabIsVisiblyMarked(testRun *testing.T) {
	model := newTestModel(testRun)
	if !strings.Contains(model.View(), "[1 Findings 5]") {
		testRun.Fatalf("active findings tab not marked:\n%s", model.View())
	}
	model = typeKeys(testRun, model, "2")
	if !strings.Contains(model.View(), "[2 All ports 5]") || strings.Contains(model.View(), "[1 Findings") {
		testRun.Fatalf("active all-ports tab not marked:\n%s", model.View())
	}
}

func TestFooterHintsDifferPerTab(testRun *testing.T) {
	model := newTestModel(testRun)
	if !strings.Contains(model.View(), "s sort") {
		testRun.Error("findings footer must offer sort")
	}
	changes := typeKeys(testRun, model, "3")
	if strings.Contains(changes.View(), "s sort") || !strings.Contains(changes.View(), "tab") {
		testRun.Errorf("changes footer wrong:\n%s", changes.View())
	}
}

func TestHelpListsTheNewKeys(testRun *testing.T) {
	model := typeKeys(testRun, newTestModel(testRun), "?")
	rendered := model.View()
	for _, expected := range []string{"tab", "shift+tab", "1 / 2 / 3"} {
		if !strings.Contains(rendered, expected) {
			testRun.Errorf("help lacks %q:\n%s", expected, rendered)
		}
	}
}
