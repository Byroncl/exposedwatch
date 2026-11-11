package tui

import (
	"context"
	"errors"
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
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "ctrl+c":
		return tea.KeyMsg{Type: tea.KeyCtrlC}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
	}
}

func visiblePorts(model Model) []int {
	numbers := make([]int, 0, len(model.visibleFindings))
	for _, finding := range model.visibleFindings {
		numbers = append(numbers, finding.Port.Number)
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

var errDockerDown = errors.New("cannot connect to the docker daemon")
