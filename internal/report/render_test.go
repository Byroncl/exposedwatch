package report

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/Byroncl/exposedwatch/internal/app"
	"github.com/Byroncl/exposedwatch/internal/domain"
)

func sampleReport() app.Report {
	port := domain.Port{Number: 6379, Protocol: domain.TCP, BindAddress: "0.0.0.0", Source: domain.SourceDocker, Container: "cache"}
	return app.Report{
		Ports: []domain.Port{port},
		Findings: []domain.Finding{{
			Port: port, Severity: domain.High, RuleID: "database-exposed",
			Title: "Database exposed", Recommendation: "Bind to loopback.",
		}},
		SourceErrors: []app.SourceError{{Source: domain.SourceHost, Err: errors.New("proc unreadable")}},
	}
}

func TestTableRendererContent(testRun *testing.T) {
	var output bytes.Buffer
	if err := NewTableRenderer(ColorDisabled).Render(&output, sampleReport()); err != nil {
		testRun.Fatal(err)
	}
	text := output.String()
	for _, want := range []string{"SEVERITY", "PORT", "PROTOCOL", "BIND ADDRESS", "EXPOSURE", "SOURCE", "PROCESS/CONTAINER", "TITLE",
		"high", "6379", "tcp", "0.0.0.0", "all-interfaces", "docker", "cache", "Database exposed",
		"Recommendations", "Bind to loopback.", "host: proc unreadable"} {
		if !strings.Contains(text, want) {
			testRun.Errorf("output missing %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, "\x1b[") {
		testRun.Errorf("color disabled but ANSI present")
	}
}

func TestTableRendererColor(testRun *testing.T) {
	var output bytes.Buffer
	if err := NewTableRenderer(ColorEnabled).Render(&output, sampleReport()); err != nil {
		testRun.Fatal(err)
	}
	if !strings.Contains(output.String(), "\x1b[") {
		testRun.Errorf("color enabled but no ANSI present")
	}
}

func TestTableRendererNoFindings(testRun *testing.T) {
	var output bytes.Buffer
	if err := NewTableRenderer(ColorDisabled).Render(&output, app.Report{}); err != nil {
		testRun.Fatal(err)
	}
	if !strings.Contains(output.String(), "No findings") {
		testRun.Errorf("missing empty message: %s", output.String())
	}
}

func TestJSONRendererShape(testRun *testing.T) {
	var output bytes.Buffer
	if err := NewJSONRenderer().Render(&output, sampleReport()); err != nil {
		testRun.Fatal(err)
	}
	var decoded struct {
		Findings []struct {
			Severity       string `json:"severity"`
			Port           int    `json:"port"`
			Protocol       string `json:"protocol"`
			BindAddress    string `json:"bind_address"`
			Exposure       string `json:"exposure"`
			Source         string `json:"source"`
			Container      string `json:"container"`
			RuleID         string `json:"rule_id"`
			Title          string `json:"title"`
			Recommendation string `json:"recommendation"`
		} `json:"findings"`
		Ports        []map[string]any    `json:"ports"`
		SourceErrors []map[string]string `json:"source_errors"`
	}
	if err := json.Unmarshal(output.Bytes(), &decoded); err != nil {
		testRun.Fatalf("invalid json: %v\n%s", err, output.String())
	}
	if len(decoded.Findings) != 1 || decoded.Findings[0].Severity != "high" || decoded.Findings[0].Port != 6379 ||
		decoded.Findings[0].Exposure != "all-interfaces" || decoded.Findings[0].RuleID != "database-exposed" ||
		decoded.Findings[0].Container != "cache" {
		testRun.Fatalf("findings = %+v", decoded.Findings)
	}
	if len(decoded.Ports) != 1 || len(decoded.SourceErrors) != 1 || decoded.SourceErrors[0]["source"] != "host" {
		testRun.Fatalf("decoded = %+v", decoded)
	}
}

func TestJSONRendererEmptyUsesArrays(testRun *testing.T) {
	var output bytes.Buffer
	if err := NewJSONRenderer().Render(&output, app.Report{}); err != nil {
		testRun.Fatal(err)
	}
	for _, want := range []string{`"findings": []`, `"ports": []`, `"source_errors": []`} {
		if !strings.Contains(output.String(), want) {
			testRun.Errorf("missing %s in %s", want, output.String())
		}
	}
}

func TestDetectColorSupport(testRun *testing.T) {
	var buffer bytes.Buffer
	if DetectColorMode(&buffer, "") != ColorDisabled {
		testRun.Errorf("non-file writer must disable color")
	}
	if DetectColorMode(&buffer, "1") != ColorDisabled {
		testRun.Errorf("NO_COLOR must disable color")
	}
}
