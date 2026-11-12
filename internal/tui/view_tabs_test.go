package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

const (
	fullTabStripWidth    = 110
	compactTabStripWidth = 20
)

func TestTabsAreVisuallySeparated(testRun *testing.T) {
	model := newTestModel(testRun)
	line := model.tabLine(fullTabStripWidth)
	wantDividers := len(ModulePorts.innerTabs()) - 1
	if gotDividers := strings.Count(line, tabDivider); gotDividers != wantDividers {
		testRun.Fatalf("want %d dividers, got %d in %q", wantDividers, gotDividers, line)
	}
}

func TestTabStripWidthDoesNotChangeWithTheActiveTab(testRun *testing.T) {
	model := newTestModel(testRun)
	baselineWidth := lipgloss.Width(model.tabLine(fullTabStripWidth))
	for _, key := range []string{"2", "3", "1"} {
		model = typeKeys(testRun, model, key)
		if gotWidth := lipgloss.Width(model.tabLine(fullTabStripWidth)); gotWidth != baselineWidth {
			testRun.Errorf("key %q: width %d, want %d (the strip must not jump)", key, gotWidth, baselineWidth)
		}
	}
}

func TestCompactTabsKeepTheirDividers(testRun *testing.T) {
	model := newTestModel(testRun)
	line := model.tabLine(compactTabStripWidth)
	if gotDividers := strings.Count(line, tabDivider); gotDividers != len(ModulePorts.innerTabs())-1 {
		testRun.Fatalf("compact line %q has %d dividers", line, gotDividers)
	}
	if lipgloss.Width(line) > compactTabStripWidth {
		testRun.Fatalf("compact line %q is wider than %d", line, compactTabStripWidth)
	}
}
