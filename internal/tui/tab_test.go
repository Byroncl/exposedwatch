package tui

import "testing"

func TestModuleCyclesForwardAndBackward(testRun *testing.T) {
	order := []Module{ModuleOverview, ModulePorts, ModuleConnections, ModuleFirewall, ModuleAlerts, ModuleSettings}
	for index, module := range order {
		if module.Next() != order[(index+1)%len(order)] {
			testRun.Errorf("%v.Next() = %v", module, module.Next())
		}
		if module.Previous() != order[(index+len(order)-1)%len(order)] {
			testRun.Errorf("%v.Previous() = %v", module, module.Previous())
		}
	}
}

func TestInnerTabsOfEveryModule(testRun *testing.T) {
	want := map[Module][]TabID{
		ModuleOverview:    {TabSummary},
		ModulePorts:       {TabFindings, TabAllPorts, TabChanges},
		ModuleConnections: {TabConnections, TabTopRemote},
		ModuleFirewall:    {TabFirewall},
		ModuleAlerts:      {TabAlerts},
		ModuleSettings:    {TabSettings},
	}
	for module, tabs := range want {
		specs := module.innerTabs()
		if len(specs) != len(tabs) {
			testRun.Fatalf("%v has %d inner tabs, want %d", module, len(specs), len(tabs))
		}
		for index, spec := range specs {
			if spec.view != tabs[index] || spec.title == "" {
				testRun.Errorf("%v inner tab %d = %+v, want view %v and a title", module, index, spec, tabs[index])
			}
		}
	}
}

func TestTabTitles(testRun *testing.T) {
	want := map[TabID]string{TabFindings: "Findings", TabAllPorts: "All ports", TabChanges: "Changes", TabConnections: "Connections"}
	for tab, title := range want {
		if tab.Title() != title {
			testRun.Errorf("title of %d = %q, want %q", tab, tab.Title(), title)
		}
	}
}

func TestInnerTabKeyParsing(testRun *testing.T) {
	for key, want := range map[string]InnerTab{"1": 0, "2": 1, "9": 8} {
		got, isInnerTabKey := parseInnerTabKey(key)
		if !isInnerTabKey || got != want {
			testRun.Errorf("key %q = %v, %v", key, got, isInnerTabKey)
		}
	}
	for _, key := range []string{"0", "a", "10", "", "tab"} {
		if _, isInnerTabKey := parseInnerTabKey(key); isInnerTabKey {
			testRun.Errorf("key %q must not be an inner tab key", key)
		}
	}
}
