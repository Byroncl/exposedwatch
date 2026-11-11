package tui

import (
	"strings"
	"testing"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

func TestModuleKeysCycleBothWaysAndWrap(testRun *testing.T) {
	model := newTestModel(testRun)
	forward := []Module{ModuleConnections, ModuleFirewall, ModuleAlerts, ModuleSettings, ModuleOverview, ModulePorts}
	for _, want := range forward {
		model = typeKeys(testRun, model, "tab")
		if model.activeModule != want {
			testRun.Fatalf("tab -> %v, want %v", model.activeModule, want)
		}
	}
	backward := []Module{ModuleOverview, ModuleSettings, ModuleAlerts, ModuleFirewall, ModuleConnections, ModulePorts}
	for _, want := range backward {
		model = typeKeys(testRun, model, "shift+tab")
		if model.activeModule != want {
			testRun.Fatalf("shift+tab -> %v, want %v", model.activeModule, want)
		}
	}
}

func TestInnerTabKeysWrapInsideEachModule(testRun *testing.T) {
	model := showModule(testRun, newTestModel(testRun), ModuleConnections)
	model = typeKeys(testRun, model, "]")
	if model.activeTab != TabTopRemote {
		testRun.Fatalf("] in connections -> %v", model.activeTab)
	}
	model = typeKeys(testRun, model, "]")
	if model.activeTab != TabConnections {
		testRun.Fatalf("] wraps -> %v", model.activeTab)
	}
	model = typeKeys(testRun, model, "[")
	if model.activeTab != TabTopRemote {
		testRun.Fatalf("[ wraps backwards -> %v", model.activeTab)
	}
	single := showModule(testRun, model, ModuleOverview)
	if got := typeKeys(testRun, single, "]", "[").activeTab; got != TabSummary {
		testRun.Fatalf("a single inner tab stays put, got %v", got)
	}
}

func TestNumberKeysOutsideTheModuleAreIgnored(testRun *testing.T) {
	model := typeKeys(testRun, newTestModel(testRun), "3")
	model = typeKeys(testRun, model, "4", "9", "0")
	if model.activeTab != TabChanges {
		testRun.Fatalf("keys beyond the module's tabs must be ignored, tab = %v", model.activeTab)
	}
	connections := typeKeys(testRun, showModule(testRun, model, ModuleConnections), "2", "3")
	if connections.activeTab != TabTopRemote {
		testRun.Fatalf("connections has two inner tabs, tab = %v", connections.activeTab)
	}
	overview := typeKeys(testRun, showModule(testRun, model, ModuleOverview), "2")
	if overview.activeTab != TabSummary {
		testRun.Fatalf("overview has one inner tab, tab = %v", overview.activeTab)
	}
}

func TestEachModuleRemembersItsLastInnerTab(testRun *testing.T) {
	model := newTestModel(testRun)
	model = typeKeys(testRun, model, "3", "tab", "2")
	if model.activeModule != ModuleConnections || model.activeTab != TabTopRemote {
		testRun.Fatalf("module %v tab %v", model.activeModule, model.activeTab)
	}
	model = typeKeys(testRun, model, "shift+tab")
	if model.activeTab != TabChanges {
		testRun.Fatalf("ports must come back on Changes, got %v", model.activeTab)
	}
	model = typeKeys(testRun, model, "tab")
	if model.activeTab != TabTopRemote {
		testRun.Fatalf("connections must come back on Top remote, got %v", model.activeTab)
	}
}

func TestSwitchingModuleResetsTheCursor(testRun *testing.T) {
	model := typeKeys(testRun, newTestModel(testRun), "j", "j", "tab", "shift+tab")
	if model.cursor != 0 || model.scrollOffset != 0 {
		testRun.Fatalf("cursor=%d scroll=%d", model.cursor, model.scrollOffset)
	}
}

func TestFilteringKeysAreIgnoredOutsideTheTablesThatUseThem(testRun *testing.T) {
	for _, target := range []moduleView{
		{ModuleOverview, 0}, {ModuleFirewall, 0}, {ModuleAlerts, 0}, {ModuleSettings, 0}, {ModuleConnections, 1},
	} {
		model := showView(testRun, newConnectionsModel(testRun, fullConnections()), target)
		model = typeKeys(testRun, model, "s", "f", "/", "j", "G")
		if model.sortMode != SortBySeverity || model.connectionSortMode != ConnectionSortByDirection || model.minSeverity != domain.Info ||
			model.viewState != ViewTable || model.cursor != 0 {
			testRun.Errorf("%v/%d: sort=%v/%v min=%v state=%v cursor=%d", target.module, target.inner,
				model.sortMode, model.connectionSortMode, model.minSeverity, model.viewState, model.cursor)
		}
	}
}

func TestConnectionTogglesActOnTopRemoteAndNotOnOtherModules(testRun *testing.T) {
	base := newConnectionsModel(testRun, fullConnections())
	topRemote := typeKeys(testRun, base, "2", "l", "i")
	if topRemote.hideLoopback || !topRemote.inboundOnly {
		testRun.Fatalf("l and i must act on Top remote: hideLoopback=%v inboundOnly=%v", topRemote.hideLoopback, topRemote.inboundOnly)
	}
	for _, module := range []Module{ModuleOverview, ModulePorts, ModuleFirewall, ModuleAlerts, ModuleSettings} {
		other := typeKeys(testRun, showModule(testRun, base, module), "l", "i")
		if !other.hideLoopback || other.inboundOnly {
			testRun.Errorf("%v: l and i must be ignored", module)
		}
	}
}

func TestFooterHintsOnlyListKeysThatApplyToTheView(testRun *testing.T) {
	connections := newConnectionsModel(testRun, fullConnections())
	cases := []struct {
		target  moduleView
		want    []string
		missing []string
	}{
		{moduleView{ModuleOverview, 0}, []string{"tab", "? help", "q quit", "r refresh"}, []string{"s sort", "f min-sev", "/ filter", "j/k", "inbound", "loopback", "g/G"}},
		{moduleView{ModulePorts, 0}, []string{"s sort", "f min-sev", "/ filter", "j/k"}, []string{"inbound", "loopback"}},
		{moduleView{ModulePorts, 2}, []string{"j/k", "g/G"}, []string{"s sort", "f min-sev", "/ filter"}},
		{moduleView{ModuleConnections, 0}, []string{"s sort", "/ filter", "i inbound", "l loopback"}, []string{"f min-sev"}},
		{moduleView{ModuleConnections, 1}, []string{"i inbound", "l loopback", "? help"}, []string{"s sort", "f min-sev", "/ filter", "j/k"}},
		{moduleView{ModuleFirewall, 0}, []string{"tab", "q quit"}, []string{"s sort", "i inbound", "j/k"}},
		{moduleView{ModuleSettings, 0}, []string{"tab", "q quit"}, []string{"s sort", "i inbound", "j/k"}},
	}
	for _, testCase := range cases {
		footer := strings.Join(showView(testRun, connections, testCase.target).footerHintTexts(testWidth), footerHintJoin)
		for _, want := range testCase.want {
			if !strings.Contains(footer, want) {
				testRun.Errorf("%v/%d footer lacks %q: %q", testCase.target.module, testCase.target.inner, want, footer)
			}
		}
		for _, missing := range testCase.missing {
			if strings.Contains(footer, missing) {
				testRun.Errorf("%v/%d footer must hide %q: %q", testCase.target.module, testCase.target.inner, missing, footer)
			}
		}
	}
}

func TestDisabledConnectionsHideTheirFooterHintsOnBothInnerTabs(testRun *testing.T) {
	disabled := newTestModel(testRun)
	for _, inner := range []InnerTab{0, 1} {
		footer := strings.Join(showView(testRun, disabled, moduleView{ModuleConnections, inner}).footerHintTexts(testWidth), footerHintJoin)
		for _, hidden := range []string{"inbound", "loopback", "s sort", "/ filter"} {
			if strings.Contains(footer, hidden) {
				testRun.Errorf("inner %d: disabled connections must hide %q: %q", inner, hidden, footer)
			}
		}
	}
}

func TestFooterDropsTheSidebarHintFirstAndKeepsTabsQuitAndHelp(testRun *testing.T) {
	model := newTestModel(testRun)
	wide := strings.Join(model.footerHintTexts(testWidth), footerHintJoin)
	if !strings.Contains(wide, "b sidebar") {
		testRun.Fatalf("wide footer lacks the sidebar hint: %q", wide)
	}
	for width := 1; width <= testWidth; width++ {
		joined := strings.Join(model.footerHintTexts(width), footerHintJoin)
		if !strings.Contains(joined, "q quit") || !strings.Contains(joined, "? help") {
			testRun.Fatalf("width %d loses quit or help: %q", width, joined)
		}
		if strings.Contains(joined, "b sidebar") && !strings.Contains(joined, "r refresh") {
			testRun.Fatalf("width %d keeps the sidebar hint but drops refresh: %q", width, joined)
		}
	}
}

func TestHelpListsModuleKeysAndTheSidebarToggle(testRun *testing.T) {
	rendered := typeKeys(testRun, newTestModel(testRun), "?").View()
	for _, expected := range []string{"next / previous module", "next / previous tab", "show or hide the sidebar"} {
		if !strings.Contains(rendered, expected) {
			testRun.Errorf("help lacks %q:\n%s", expected, rendered)
		}
	}
}
