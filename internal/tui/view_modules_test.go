package tui

import (
	"strings"
	"testing"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

func manyInboundConnections(count int) []domain.Connection {
	connections := make([]domain.Connection, 0, count)
	for index := 0; index < count; index++ {
		connections = append(connections, testConnection(domain.ConnectionInbound, 2000+index, "203.0.113.7", 40000+index, "svc"))
	}
	return connections
}

func TestTopRemoteShowsBothSummaryPanelsWithMoreRowsThanTheBottomPanels(testRun *testing.T) {
	model := showView(testRun, newConnectionsModel(testRun, manyInboundConnections(12)), moduleView{ModuleConnections, 1})
	rendered := model.View()
	for _, want := range []string{"Top remote addresses (inbound)", "Inbound by local port", "2011/tcp", "203.0.113.7"} {
		if !strings.Contains(rendered, want) {
			testRun.Errorf("top remote lacks %q:\n%s", want, rendered)
		}
	}
	if strings.Contains(rendered, "... and") {
		testRun.Errorf("a tall terminal needs no collapsed rows:\n%s", rendered)
	}
	if strings.Contains(rendered, "Details") {
		testRun.Error("top remote has no details panel")
	}
}

func TestTopRemoteFollowsTheLoopbackAndInboundToggles(testRun *testing.T) {
	model := showView(testRun, newConnectionsModel(testRun, fullConnections()), moduleView{ModuleConnections, 1})
	if strings.Contains(model.View(), "127.0.0.1") {
		testRun.Fatal("loopback is hidden by default")
	}
	shown := typeKeys(testRun, model, "l")
	if !strings.Contains(shown.View(), "127.0.0.1") || !strings.Contains(shown.View(), "5432/tcp") {
		testRun.Fatalf("l must reveal loopback traffic:\n%s", shown.View())
	}
	only := typeKeys(testRun, model, "i")
	if !only.inboundOnly || !strings.Contains(only.footerLine(testWidth), "i inbound:only") {
		testRun.Fatalf("i must switch to inbound only: %q", only.footerLine(testWidth))
	}
	if !strings.Contains(only.View(), "203.0.113.7") {
		testRun.Fatalf("inbound addresses stay listed:\n%s", only.View())
	}
}

func TestTopRemoteStacksTheTwoPanelsWhenNarrow(testRun *testing.T) {
	model := sizedModel(testRun, showView(testRun, newConnectionsModel(testRun, fullConnections()), moduleView{ModuleConnections, 1}), 60, 30)
	if frame := model.computeLayout(); frame.contentWidth >= connectionSideBySideMinWidth {
		testRun.Fatalf("fixture is not narrow: %d", frame.contentWidth)
	}
	rendered := model.View()
	if !strings.Contains(rendered, "Top remote addresses") || !strings.Contains(rendered, "Inbound by local port") {
		testRun.Fatalf("stacked panels missing:\n%s", rendered)
	}
	assertRendersWithin(testRun, model, 60, 30)
}

func TestDisabledConnectionsShowTheDisabledMessageOnBothInnerTabs(testRun *testing.T) {
	for _, inner := range []InnerTab{0, 1} {
		model := showView(testRun, newTestModel(testRun), moduleView{ModuleConnections, inner})
		if rendered := model.View(); !strings.Contains(rendered, "Connections are disabled") || strings.Contains(rendered, "Inbound by local port") {
			testRun.Errorf("inner %d:\n%s", inner, rendered)
		}
	}
}

func TestPlaceholderModulesAreHonest(testRun *testing.T) {
	cases := []struct {
		module Module
		want   []string
	}{
		{ModuleFirewall, []string{"Coming soon", "ufw", "nftables", "allowed"}},
		{ModuleAlerts, []string{"Coming soon", "Telegram"}},
		{ModuleSettings, []string{"Coming soon", "notification", "display"}},
	}
	for _, testCase := range cases {
		rendered := showModule(testRun, newTestModel(testRun), testCase.module).View()
		for _, want := range testCase.want {
			if !strings.Contains(rendered, want) {
				testRun.Errorf("%v lacks %q:\n%s", testCase.module, want, rendered)
			}
		}
	}
}

func TestSettingsShowsCurrentValuesReadOnly(testRun *testing.T) {
	model := New(Options{Scanner: fakeScanner{}, Connections: fakeConnectionSource{}, Interval: testInterval, Version: testVersion, Theme: ThemeLight})
	model = sizedModel(testRun, model, testWidth, testHeight)
	rendered := showModule(testRun, model, ModuleSettings).View()
	for _, want := range []string{"Theme", "light", "Refresh interval", "2s", "Connections", "enabled"} {
		if !strings.Contains(rendered, want) {
			testRun.Errorf("settings lacks %q:\n%s", want, rendered)
		}
	}
	disabled := showModule(testRun, newTestModel(testRun), ModuleSettings).View()
	if !strings.Contains(disabled, "disabled") || !strings.Contains(disabled, "auto") {
		testRun.Errorf("settings must reflect disabled connections and the default theme:\n%s", disabled)
	}
}

func TestFilterPromptAndHelpStayInsideEverySizeForEveryModule(testRun *testing.T) {
	base := newConnectionsModel(testRun, fullConnections())
	for _, size := range layoutSizes {
		sized := sizedModel(testRun, base, size.Width, size.Height)
		for _, module := range []Module{ModulePorts, ModuleConnections} {
			prompt := typeKeys(testRun, showModule(testRun, sized, module), "/", "a", "b")
			assertRendersWithin(testRun, prompt, size.Width, size.Height)
		}
		for module := ModuleOverview; module < moduleCount; module++ {
			assertRendersWithin(testRun, typeKeys(testRun, showModule(testRun, sized, module), "?"), size.Width, size.Height)
		}
	}
}

func TestTopRemotePanelsAreSizedToTheirRows(testRun *testing.T) {
	model := showView(testRun, newConnectionsModel(testRun, manyInboundConnections(5)), moduleView{ModuleConnections, 1})
	lines := strings.Split(model.View(), "\n")
	panelLines := 0
	for _, line := range lines {
		if strings.Contains(line, "╭") && strings.Contains(line, "Top remote addresses") {
			panelLines = 1
		} else if panelLines > 0 && strings.Contains(line, "╰") {
			panelLines++
			break
		} else if panelLines > 0 {
			panelLines++
		}
	}
	// five rows (the per-port panel is the tallest) plus two border lines
	if panelLines != 5+panelBorderLineCount {
		testRun.Fatalf("panel height = %d, want %d:\n%s", panelLines, 5+panelBorderLineCount, model.View())
	}
}

func TestPlaceholderBoxIsSizedToItsText(testRun *testing.T) {
	model := showModule(testRun, newTestModel(testRun), ModuleFirewall)
	rendered := model.View()
	if strings.Count(rendered, "│ │") > 0 && len(strings.Split(rendered, "\n")) != testHeight {
		testRun.Fatal("frame must keep its height")
	}
	frame := model.computeLayout()
	box := model.placeholderPanel(frame)
	if len(box) >= frame.panelLines {
		testRun.Fatalf("placeholder box has %d lines in an area of %d", len(box), frame.panelLines)
	}
}
