package tui

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Byroncl/exposedwatch/internal/app"
	"github.com/Byroncl/exposedwatch/internal/domain"
)

const (
	testVersion       = "v9.9.9"
	testInterval      = 2 * time.Second
	testWidth         = 120
	testHeight        = 40
	testRecommendText = "Bind the service to 127.0.0.1 instead"
)

var testRefreshTime = time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

type fakeScanner struct {
	report app.Report
	err    error
}

func (scanner fakeScanner) Scan(_ context.Context) (app.Report, error) {
	return scanner.report, scanner.err
}

func makeFinding(port int, severity domain.Severity, process string) domain.Finding {
	return domain.Finding{
		Port: domain.Port{
			Number: port, Protocol: domain.TCP, BindAddress: "0.0.0.0",
			Process: process, Source: domain.SourceHost,
		},
		Severity:       severity,
		RuleID:         "RULE-" + process,
		Title:          "Title for " + process,
		Recommendation: testRecommendText + " (" + process + ")",
	}
}

func sampleReport() app.Report {
	findings := []domain.Finding{
		makeFinding(8080, domain.Low, "webapp"),
		makeFinding(22, domain.High, "sshd"),
		makeFinding(3306, domain.Critical, "mysqld"),
		makeFinding(53, domain.Info, "dnsmasq"),
		makeFinding(443, domain.High, "nginx"),
	}
	ports := make([]domain.Port, 0, len(findings))
	for _, finding := range findings {
		ports = append(ports, finding.Port)
	}
	return app.Report{Ports: ports, Findings: findings}
}

func newTestModel(testRun *testing.T) Model {
	testRun.Helper()
	model := New(Options{Scanner: fakeScanner{}, Interval: testInterval, Version: testVersion})
	model = applyMessage(testRun, model, tea.WindowSizeMsg{Width: testWidth, Height: testHeight})
	return applyMessage(testRun, model, RefreshResultMessage{Report: sampleReport(), At: testRefreshTime})
}

func applyMessage(testRun *testing.T, model Model, incomingMessage tea.Msg) Model {
	testRun.Helper()
	updated, _ := model.Update(incomingMessage)
	castModel, isModel := updated.(Model)
	if !isModel {
		testRun.Fatalf("Update returned %T, want Model", updated)
	}
	return castModel
}

func typeKeys(testRun *testing.T, model Model, keys ...string) Model {
	testRun.Helper()
	for _, key := range keys {
		model = applyMessage(testRun, model, keyMessage(key))
	}
	return model
}

func keyMessage(key string) tea.KeyMsg {
	switch key {
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "tab":
		return tea.KeyMsg{Type: tea.KeyTab}
	case "shift+tab":
		return tea.KeyMsg{Type: tea.KeyShiftTab}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "backspace":
		return tea.KeyMsg{Type: tea.KeyBackspace}
	case "ctrl+c":
		return tea.KeyMsg{Type: tea.KeyCtrlC}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
	}
}

func visiblePorts(model Model) []int {
	numbers := make([]int, 0, len(model.visibleRows))
	for _, row := range model.visibleRows {
		numbers = append(numbers, row.port.Number)
	}
	return numbers
}

func equalInts(left, right []int) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func assertRendersWithin(testRun *testing.T, model Model, width, height int) {
	testRun.Helper()
	rendered := model.View()
	lines := strings.Split(rendered, "\n")
	if len(lines) > height {
		testRun.Errorf("%dx%d: %d lines exceed height", width, height, len(lines))
	}
	for _, line := range lines {
		if lipgloss.Width(line) > width {
			testRun.Errorf("%dx%d: line width %d exceeds width: %q", width, height, lipgloss.Width(line), line)
		}
	}
}

// quietPort is a loopback port: it is listed but produces no finding.
func quietPort() domain.Port {
	return domain.Port{Number: 631, Protocol: domain.TCP, BindAddress: "127.0.0.1", Process: "cupsd", Source: domain.SourceHost}
}

func sampleReportWithQuietPort() app.Report {
	report := sampleReport()
	report.Ports = append(report.Ports, quietPort())
	return report
}

func refreshWith(testRun *testing.T, model Model, report app.Report, offsetMinutes int) Model {
	testRun.Helper()
	return applyMessage(testRun, model, RefreshResultMessage{Report: report, At: timeAfter(offsetMinutes)})
}

// reportWithout drops the ports (and their findings) with the given numbers.
func reportWithout(report app.Report, numbers ...int) app.Report {
	dropped := make(map[int]bool, len(numbers))
	for _, number := range numbers {
		dropped[number] = true
	}
	var reduced app.Report
	for _, port := range report.Ports {
		if !dropped[port.Number] {
			reduced.Ports = append(reduced.Ports, port)
		}
	}
	for _, finding := range report.Findings {
		if !dropped[finding.Port.Number] {
			reduced.Findings = append(reduced.Findings, finding)
		}
	}
	return reduced
}

// reportWith adds a medium-severity finding for a new host port.
func reportWith(report app.Report, number int) app.Report {
	finding := makeFinding(number, domain.Medium, "fresh")
	report.Ports = append(append([]domain.Port{}, report.Ports...), finding.Port)
	report.Findings = append(append([]domain.Finding{}, report.Findings...), finding)
	return report
}

var errDockerDown = errors.New("cannot connect to the docker daemon")

// moduleView names one module and one of its inner tabs.
type moduleView struct {
	module Module
	inner  InnerTab
}

// everyModuleView lists every inner tab of every module.
func everyModuleView() []moduleView {
	var views []moduleView
	for module := ModuleOverview; module < moduleCount; module++ {
		for inner := range module.innerTabs() {
			views = append(views, moduleView{module: module, inner: InnerTab(inner)})
		}
	}
	return views
}

func innerTabKey(inner InnerTab) string {
	return strconv.Itoa(int(inner) + 1)
}

// showModule cycles forward with tab until the module is active.
func showModule(testRun *testing.T, model Model, module Module) Model {
	testRun.Helper()
	for attempt := 0; attempt < int(moduleCount) && model.activeModule != module; attempt++ {
		model = typeKeys(testRun, model, "tab")
	}
	if model.activeModule != module {
		testRun.Fatalf("cannot reach module %v", module)
	}
	return model
}

// showView opens one inner tab of one module.
func showView(testRun *testing.T, model Model, target moduleView) Model {
	testRun.Helper()
	model = showModule(testRun, model, target.module)
	return typeKeys(testRun, model, innerTabKey(target.inner))
}
