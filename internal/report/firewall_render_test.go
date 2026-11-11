package report

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/Byroncl/exposedwatch/internal/app"
	"github.com/Byroncl/exposedwatch/internal/domain"
)

const (
	regressionTableOutput = "SEVERITY  PORT  PROTOCOL  BIND ADDRESS  EXPOSURE        SOURCE  PROCESS/CONTAINER  TITLE\n" +
		"high      6379  tcp       0.0.0.0       all-interfaces  docker  cache              Database exposed\n\n" +
		"Recommendations\n  - [database-exposed] Bind to loopback.\n\n" +
		"1 port(s) scanned, 1 finding(s).\n\nSources that could not be read:\n  - host: proc unreadable\n"
	regressionJSONOutput = "{\n  \"findings\": [\n    {\n      \"port\": 6379,\n      \"protocol\": \"tcp\",\n      \"bind_address\": \"0.0.0.0\",\n" +
		"      \"exposure\": \"all-interfaces\",\n      \"source\": \"docker\",\n      \"container\": \"cache\",\n      \"severity\": \"high\",\n" +
		"      \"rule_id\": \"database-exposed\",\n      \"title\": \"Database exposed\",\n      \"recommendation\": \"Bind to loopback.\"\n    }\n  ],\n" +
		"  \"ports\": [\n    {\n      \"port\": 6379,\n      \"protocol\": \"tcp\",\n      \"bind_address\": \"0.0.0.0\",\n" +
		"      \"exposure\": \"all-interfaces\",\n      \"source\": \"docker\",\n      \"container\": \"cache\"\n    }\n  ],\n" +
		"  \"source_errors\": [\n    {\n      \"source\": \"host\",\n      \"error\": \"proc unreadable\"\n    }\n  ]\n}\n"
	informationalSentence = "Firewall verdicts are informational and never change severities."
)

func sampleFirewallReport() domain.FirewallReport {
	port := sampleReport().Ports[0]
	return domain.FirewallReport{
		Backend:   "iptables-save",
		Available: true,
		Notes:     []string{"IPv6 rules were not readable"},
		Findings: []domain.FirewallFinding{
			{Port: port, Verdict: domain.FirewallDockerPublished, Reason: "published by Docker; usually bypasses INPUT rules"},
		},
	}
}

func unavailableFirewallReport() domain.FirewallReport {
	port := sampleReport().Ports[0]
	return domain.FirewallReport{
		Backend: domain.BackendUnknown,
		Notes:   []string{"needs root to read firewall rules"},
		Findings: []domain.FirewallFinding{
			{Port: port, Verdict: domain.FirewallUnknown, Reason: "needs root to read firewall rules"},
		},
	}
}

func TestOutputWithoutFirewallIsUnchanged(testRun *testing.T) {
	var table bytes.Buffer
	if err := NewTableRenderer(ColorDisabled).Render(&table, sampleReport()); err != nil {
		testRun.Fatal(err)
	}
	if table.String() != regressionTableOutput {
		testRun.Errorf("table output changed:\n%q", table.String())
	}
	var document bytes.Buffer
	if err := NewJSONRenderer().Render(&document, sampleReport()); err != nil {
		testRun.Fatal(err)
	}
	if document.String() != regressionJSONOutput {
		testRun.Errorf("json output changed:\n%q", document.String())
	}
}

func TestTableShowsTheFirewallColumnAndFooter(testRun *testing.T) {
	var output bytes.Buffer
	if err := NewTableRenderer(ColorDisabled).RenderWithFirewall(&output, sampleReport(), sampleFirewallReport()); err != nil {
		testRun.Fatal(err)
	}
	text := output.String()
	for _, want := range []string{"FIREWALL", "docker ", "Firewall", "backend: iptables-save", "available: yes",
		"IPv6 rules were not readable", informationalSentence, "Database exposed", "host: proc unreadable"} {
		if !strings.Contains(text, want) {
			testRun.Errorf("output lacks %q:\n%s", want, text)
		}
	}
}

func TestTableShowsAnUnavailableFirewallHonestly(testRun *testing.T) {
	var output bytes.Buffer
	if err := NewTableRenderer(ColorDisabled).RenderWithFirewall(&output, sampleReport(), unavailableFirewallReport()); err != nil {
		testRun.Fatal(err)
	}
	text := output.String()
	for _, want := range []string{"unknown", "available: no", "needs root to read firewall rules", informationalSentence} {
		if !strings.Contains(text, want) {
			testRun.Errorf("output lacks %q:\n%s", want, text)
		}
	}
}

func TestTableShowsAPlaceholderForPortsWithoutAVerdict(testRun *testing.T) {
	var output bytes.Buffer
	empty := domain.FirewallReport{Backend: "iptables-save", Available: true}
	if err := NewTableRenderer(ColorDisabled).RenderWithFirewall(&output, sampleReport(), empty); err != nil {
		testRun.Fatal(err)
	}
	firstRow := strings.Split(output.String(), "\n")[1]
	if !strings.Contains(firstRow, "  -  ") {
		testRun.Errorf("a port without a verdict must show a dash: %q", firstRow)
	}
}

func TestTableWithFirewallAndNoFindingsStillExplainsTheFirewall(testRun *testing.T) {
	var output bytes.Buffer
	if err := NewTableRenderer(ColorDisabled).RenderWithFirewall(&output, app.Report{}, unavailableFirewallReport()); err != nil {
		testRun.Fatal(err)
	}
	if !strings.Contains(output.String(), "No findings") || !strings.Contains(output.String(), informationalSentence) {
		testRun.Errorf("output = %s", output.String())
	}
}

func TestJSONFirewallObjectShape(testRun *testing.T) {
	var output bytes.Buffer
	if err := NewJSONRenderer().RenderWithFirewall(&output, sampleReport(), sampleFirewallReport()); err != nil {
		testRun.Fatal(err)
	}
	var decoded struct {
		Firewall struct {
			Available bool     `json:"available"`
			Backend   string   `json:"backend"`
			Notes     []string `json:"notes"`
			Ports     []struct {
				Port        int    `json:"port"`
				Protocol    string `json:"protocol"`
				BindAddress string `json:"bind_address"`
				Source      string `json:"source"`
				Verdict     string `json:"verdict"`
				Reason      string `json:"reason"`
			} `json:"ports"`
		} `json:"firewall"`
	}
	if err := json.Unmarshal(output.Bytes(), &decoded); err != nil {
		testRun.Fatalf("invalid json: %v\n%s", err, output.String())
	}
	firewall := decoded.Firewall
	if !firewall.Available || firewall.Backend != "iptables-save" || len(firewall.Notes) != 1 || len(firewall.Ports) != 1 {
		testRun.Fatalf("firewall = %+v", firewall)
	}
	entry := firewall.Ports[0]
	if entry.Port != 6379 || entry.Protocol != "tcp" || entry.BindAddress != "0.0.0.0" || entry.Source != "docker" ||
		entry.Verdict != "docker-published" || entry.Reason == "" {
		testRun.Fatalf("entry = %+v", entry)
	}
}

func TestJSONFirewallUsesEmptyArraysNotNull(testRun *testing.T) {
	var output bytes.Buffer
	empty := domain.FirewallReport{Backend: domain.BackendUnknown}
	if err := NewJSONRenderer().RenderWithFirewall(&output, app.Report{}, empty); err != nil {
		testRun.Fatal(err)
	}
	for _, want := range []string{`"available": false`, `"notes": []`, `"ports": []`} {
		if !strings.Contains(output.String(), want) {
			testRun.Errorf("missing %s in %s", want, output.String())
		}
	}
	if strings.Contains(output.String(), "null") {
		testRun.Errorf("no null allowed: %s", output.String())
	}
}
