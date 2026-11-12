package tui

import (
	"strings"
	"testing"

	"github.com/Byroncl/exposedwatch/internal/app"
	"github.com/Byroncl/exposedwatch/internal/domain"
)

func overviewModel(testRun *testing.T, base Model) Model {
	testRun.Helper()
	return showModule(testRun, base, ModuleOverview)
}

func sevenFindingReport() app.Report {
	findings := []domain.Finding{
		makeFinding(8080, domain.Low, "webapp"),
		makeFinding(22, domain.High, "sshd"),
		makeFinding(3306, domain.Critical, "mysqld"),
		makeFinding(53, domain.Info, "dnsmasq"),
		makeFinding(443, domain.High, "nginx"),
		makeFinding(9000, domain.Medium, "metrics"),
		makeFinding(25, domain.Info, "postfix"),
	}
	ports := make([]domain.Port, 0, len(findings))
	for _, finding := range findings {
		ports = append(ports, finding.Port)
	}
	return app.Report{Ports: ports, Findings: findings}
}

func TestOverviewListsTheFiveWorstFindingsBySeverityThenPort(testRun *testing.T) {
	model := overviewModel(testRun, refreshWith(testRun, newTestModel(testRun), sevenFindingReport(), 1))
	rendered := model.View()
	section := rendered[strings.Index(rendered, "Worst findings"):strings.Index(rendered, "Recent changes")]
	last := -1
	for _, title := range []string{"Title for mysqld", "Title for sshd", "Title for nginx", "Title for metrics", "Title for webapp"} {
		position := strings.Index(section, title)
		if position < 0 || position < last {
			testRun.Fatalf("%q missing or out of order in:\n%s", title, section)
		}
		last = position
	}
	for _, absent := range []string{"dnsmasq", "postfix"} {
		if strings.Contains(section, absent) {
			testRun.Errorf("only the top 5 are listed, found %q", absent)
		}
	}
	if !strings.Contains(section, "3306/tcp") || !strings.Contains(section, "CRITICAL") {
		testRun.Errorf("severity chip and port expected:\n%s", section)
	}
}

func TestOverviewShowsRiskChipsAndCounters(testRun *testing.T) {
	model := refreshWith(testRun, newConnectionsModel(testRun, fullConnections()), sampleReportWithQuietPort(), 1)
	rendered := overviewModel(testRun, model).View()
	for _, want := range []string{"CRITICAL", "RISK", "ports 6", "conn 5 (in 3 / out 2)", "+1 new", "-0 closed", "Summary"} {
		if !strings.Contains(rendered, want) {
			testRun.Errorf("overview lacks %q:\n%s", want, rendered)
		}
	}
}

func TestOverviewShowsTheLastFiveChangesNewestFirst(testRun *testing.T) {
	model := newTestModel(testRun)
	remaining := sampleReportWithQuietPort()
	model = refreshWith(testRun, model, remaining, 0)
	for index, number := range []int{631, 8080, 22, 53, 443, 3306} {
		remaining = reportWithout(remaining, number)
		model = refreshWith(testRun, model, remaining, index+1)
	}
	rendered := overviewModel(testRun, model).View()
	section := rendered[strings.Index(rendered, "Recent changes"):]
	if !strings.Contains(section, "3306/tcp") || strings.Contains(section, "631/tcp") {
		testRun.Fatalf("only the last five events are listed, newest first:\n%s", section)
	}
	if strings.Index(section, "3306/tcp") > strings.Index(section, "8080/tcp") {
		testRun.Fatalf("newest event must come first:\n%s", section)
	}
	if strings.Count(section, "CLOSED") != 5 {
		testRun.Errorf("want 5 closed events:\n%s", section)
	}
}

func TestOverviewEmptyStatesAreHonest(testRun *testing.T) {
	scanning := overviewModel(testRun, New(Options{Scanner: fakeScanner{}, Version: testVersion}))
	if rendered := scanning.View(); !strings.Contains(rendered, scanningMessage) {
		testRun.Errorf("before the first scan:\n%s", rendered)
	}
	empty := overviewModel(testRun, refreshWith(testRun, New(Options{Scanner: fakeScanner{}, Version: testVersion}), app.Report{}, 0))
	rendered := empty.View()
	for _, want := range []string{noFindingsMessage, noRecentChangesMessage, "SAFE", summaryConnectionsDisabledMessage} {
		if !strings.Contains(rendered, want) {
			testRun.Errorf("empty overview lacks %q:\n%s", want, rendered)
		}
	}
	if strings.Contains(rendered, "Top remote addresses") {
		testRun.Errorf("disabled connections hide the remote list:\n%s", rendered)
	}
}

func TestOverviewTopRemoteAddressesAreBoundedToThree(testRun *testing.T) {
	var connections []domain.Connection
	for index, address := range []string{"203.0.113.1", "203.0.113.2", "203.0.113.3", "203.0.113.4"} {
		for repeat := 0; repeat <= 4-index; repeat++ {
			connections = append(connections, testConnection(domain.ConnectionInbound, 22, address, 40000+10*index+repeat, "sshd"))
		}
	}
	model := overviewModel(testRun, newConnectionsModel(testRun, connections))
	rendered := model.View()
	section := rendered[strings.Index(rendered, "Top remote addresses"):]
	for _, want := range []string{"203.0.113.1", "203.0.113.2", "203.0.113.3"} {
		if !strings.Contains(section, want) {
			testRun.Errorf("top remote lacks %q:\n%s", want, section)
		}
	}
	if strings.Contains(section, "203.0.113.4") {
		testRun.Errorf("only three addresses are listed:\n%s", section)
	}
	quiet := overviewModel(testRun, newConnectionsModel(testRun, []domain.Connection{connectionCurlOutbound()}))
	if !strings.Contains(quiet.View(), noInboundConnectionsMessage) {
		testRun.Errorf("enabled but no inbound traffic:\n%s", quiet.View())
	}
}

func TestOverviewUsesASingleHeaderLineAndKeepsTheTitle(testRun *testing.T) {
	model := overviewModel(testRun, newTestModel(testRun))
	lines := strings.Split(model.View(), "\n")
	if !strings.Contains(lines[0], appTitle) || strings.Contains(lines[1], "RISK") || strings.Contains(lines[1], "CRITICAL 1") {
		testRun.Fatalf("overview header must be the title only:\n%s", strings.Join(lines[:3], "\n"))
	}
}
