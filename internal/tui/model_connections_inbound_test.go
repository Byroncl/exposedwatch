package tui

import (
	"strings"
	"testing"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

func TestInboundOnlyIsOffByDefaultAndTogglesWithI(testRun *testing.T) {
	model := newConnectionsModel(testRun, fullConnections())
	if model.inboundOnly || len(model.visibleRows) != 5 || !strings.Contains(model.View(), "inbound:all") {
		testRun.Fatalf("inbound-only must start off and say inbound:all:\n%s", model.View())
	}
	model = typeKeys(testRun, model, "i")
	if !model.inboundOnly || len(model.visibleRows) != 3 || !strings.Contains(model.View(), "inbound:only") {
		testRun.Fatalf("rows = %v after i:\n%s", visibleLocalPorts(model), model.View())
	}
	for _, row := range model.visibleRows {
		if row.connection.Direction != domain.ConnectionInbound {
			testRun.Fatalf("outbound row leaked: %v", row.connection)
		}
	}
	model = typeKeys(testRun, model, "i")
	if model.inboundOnly || len(model.visibleRows) != 5 {
		testRun.Fatalf("second i must show everything again, rows = %d", len(model.visibleRows))
	}
}

func TestInboundOnlyCombinesWithLoopbackToggle(testRun *testing.T) {
	model := typeKeys(testRun, newConnectionsModel(testRun, fullConnections()), "i", "l")
	if len(model.visibleRows) != 4 {
		testRun.Fatalf("inbound + loopback shown = %v, want 4 inbound rows", visibleLocalPorts(model))
	}
}

func TestInboundOnlyCombinesWithTextFilterAndSortModes(testRun *testing.T) {
	model := typeKeys(testRun, newConnectionsModel(testRun, fullConnections()), "i", "/")
	model = typeKeys(testRun, model, "n", "g", "i", "n", "x", "enter")
	if ports := visibleLocalPorts(model); len(ports) != 1 || ports[0] != 443 {
		testRun.Fatalf("text filter inside inbound-only = %v, want [443]", ports)
	}
	model = typeKeys(testRun, newConnectionsModel(testRun, fullConnections()), "i")
	for range connectionSortModeCount {
		model = typeKeys(testRun, model, "s")
		if len(model.visibleRows) != 3 {
			testRun.Fatalf("sort %v changed the visible set: %v", model.connectionSortMode, visibleLocalPorts(model))
		}
	}
}

func TestInboundOnlyCountersFollowTheVisibleSet(testRun *testing.T) {
	model := typeKeys(testRun, newConnectionsModel(testRun, fullConnections()), "i")
	if header := strings.Split(model.View(), "\n")[2]; !strings.Contains(header, "conn 3 (in 3 / out 0)") {
		testRun.Fatalf("header counter must follow the visible set: %q", header)
	}
	if !strings.Contains(model.View(), "[4 Connections 3]") {
		testRun.Fatalf("tab count must follow the visible set:\n%s", model.View())
	}
	rendered := model.View()
	if !strings.Contains(rendered, "9000/tcp") || !strings.Contains(rendered, testSshPeer) {
		testRun.Fatalf("panels must keep the inbound data:\n%s", rendered)
	}
}

func TestInboundOnlySelectionSurvivesAndClamps(testRun *testing.T) {
	model := typeKeys(testRun, newConnectionsModel(testRun, fullConnections()), "j")
	selected := model.visibleRows[model.cursor].connection.LocalPort
	model = typeKeys(testRun, model, "i")
	if got := model.visibleRows[model.cursor].connection.LocalPort; got != selected {
		testRun.Fatalf("selected inbound row moved from %d to %d", selected, got)
	}
	model = typeKeys(testRun, newConnectionsModel(testRun, fullConnections()), "G", "i")
	if model.cursor < 0 || model.cursor >= len(model.visibleRows) {
		testRun.Fatalf("cursor %d outside %d rows", model.cursor, len(model.visibleRows))
	}
}

func TestInboundKeyIsIgnoredOnOtherTabsAndWhenDisabled(testRun *testing.T) {
	model := typeKeys(testRun, newConnectionsModel(testRun, fullConnections()), "1", "i")
	if model.inboundOnly {
		testRun.Fatal("i must only act on the Connections tab")
	}
	disabled := New(Options{Scanner: fakeScanner{}, Interval: testInterval, Version: testVersion})
	disabled = typeKeys(testRun, disabled, "4", "i")
	if disabled.inboundOnly {
		testRun.Fatal("i must be ignored while connections are disabled")
	}
}

func TestInboundOnlyFooterHelpAndEmptyMessage(testRun *testing.T) {
	model := newConnectionsModel(testRun, fullConnections())
	if footer := model.footerLine(testWidth); !strings.Contains(footer, "i inbound:all") {
		testRun.Errorf("footer lacks the inbound hint: %q", footer)
	}
	only := typeKeys(testRun, model, "i")
	if footer := only.footerLine(testWidth); !strings.Contains(footer, "i inbound:only") {
		testRun.Errorf("footer after i: %q", footer)
	}
	if help := typeKeys(testRun, model, "?").View(); !strings.Contains(help, "inbound") {
		testRun.Errorf("help lacks the inbound key:\n%s", help)
	}
	outboundOnly := newConnectionsModel(testRun, []domain.Connection{connectionCurlOutbound()})
	outboundOnly = typeKeys(testRun, outboundOnly, "i")
	if !strings.Contains(outboundOnly.View(), "press i to show all") {
		testRun.Fatalf("empty inbound-only table must explain itself:\n%s", outboundOnly.View())
	}
}

func TestInboundOnlyStaysInsideEverySize(testRun *testing.T) {
	base := typeKeys(testRun, newConnectionsModel(testRun, fullConnections()), "i")
	for _, size := range layoutSizes {
		sized := applyMessage(testRun, base, size)
		assertRendersWithin(testRun, sized, size.Width, size.Height)
		assertRendersWithin(testRun, typeKeys(testRun, sized, "G"), size.Width, size.Height)
		assertRendersWithin(testRun, typeKeys(testRun, sized, "?"), size.Width, size.Height)
	}
}

func TestLoopbackHintIsDroppedBeforeInboundAndKeepsQuitAndHelp(testRun *testing.T) {
	model := newConnectionsModel(testRun, fullConnections())
	wide := strings.Join(model.footerHintTexts(testWidth), footerHintJoin)
	if !strings.Contains(wide, "i inbound:") || !strings.Contains(wide, "l loopback:") {
		testRun.Fatalf("wide footer = %q", wide)
	}
	for width := 1; width <= testWidth; width++ {
		joined := strings.Join(model.footerHintTexts(width), footerHintJoin)
		if strings.Contains(joined, "l loopback:") && !strings.Contains(joined, "i inbound:") {
			testRun.Fatalf("width %d keeps the loopback hint but drops inbound: %q", width, joined)
		}
	}
}
