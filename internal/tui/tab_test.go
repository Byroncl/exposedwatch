package tui

import "testing"

func TestTabCyclesForwardAndBackward(testRun *testing.T) {
	if TabFindings.Next() != TabAllPorts || TabAllPorts.Next() != TabChanges || TabChanges.Next() != TabConnections || TabConnections.Next() != TabFindings {
		testRun.Fatal("Next must cycle findings, all ports, changes, connections")
	}
	if TabFindings.Previous() != TabConnections || TabConnections.Previous() != TabChanges || TabChanges.Previous() != TabAllPorts || TabAllPorts.Previous() != TabFindings {
		testRun.Fatal("Previous must cycle in reverse")
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
