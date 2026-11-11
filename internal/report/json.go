package report

import (
	"encoding/json"
	"io"

	"github.com/Byroncl/exposedwatch/internal/app"
	"github.com/Byroncl/exposedwatch/internal/domain"
)

// JSONRenderer prints the report as indented JSON.
type JSONRenderer struct{}

// NewJSONRenderer returns a JSON renderer.
func NewJSONRenderer() *JSONRenderer {
	return &JSONRenderer{}
}

// jsonPort is the stable JSON shape of a port.
type jsonPort struct {
	Port        int    `json:"port"`
	Protocol    string `json:"protocol"`
	BindAddress string `json:"bind_address"`
	Exposure    string `json:"exposure"`
	Source      string `json:"source"`
	Process     string `json:"process,omitempty"`
	Container   string `json:"container,omitempty"`
}

// jsonFinding is the stable JSON shape of a finding.
type jsonFinding struct {
	jsonPort
	Severity       string `json:"severity"`
	RuleID         string `json:"rule_id"`
	Title          string `json:"title"`
	Recommendation string `json:"recommendation"`
}

// jsonSourceError is the stable JSON shape of a failed source.
type jsonSourceError struct {
	Source string `json:"source"`
	Error  string `json:"error"`
}

// jsonReport is the stable top-level JSON document.
type jsonReport struct {
	Findings     []jsonFinding     `json:"findings"`
	Ports        []jsonPort        `json:"ports"`
	SourceErrors []jsonSourceError `json:"source_errors"`
	// Firewall is only present when the report was rendered with firewall verdicts.
	Firewall *jsonFirewall `json:"firewall,omitempty"`
}

// Render writes the report as JSON.
func (renderer *JSONRenderer) Render(output io.Writer, scanReport app.Report) error {
	return renderer.render(output, scanReport, nil)
}

// RenderWithFirewall writes the report as JSON with an extra top-level
// firewall object.
func (renderer *JSONRenderer) RenderWithFirewall(output io.Writer, scanReport app.Report, firewallReport domain.FirewallReport) error {
	return renderer.render(output, scanReport, toJSONFirewall(firewallReport))
}

func (renderer *JSONRenderer) render(output io.Writer, scanReport app.Report, firewall *jsonFirewall) error {
	document := jsonReport{
		Findings:     make([]jsonFinding, 0, len(scanReport.Findings)),
		Ports:        make([]jsonPort, 0, len(scanReport.Ports)),
		SourceErrors: make([]jsonSourceError, 0, len(scanReport.SourceErrors)),
		Firewall:     firewall,
	}
	for _, finding := range scanReport.Findings {
		document.Findings = append(document.Findings, jsonFinding{
			jsonPort:       toJSONPort(finding.Port),
			Severity:       finding.Severity.String(),
			RuleID:         finding.RuleID,
			Title:          finding.Title,
			Recommendation: finding.Recommendation,
		})
	}
	for _, port := range scanReport.Ports {
		document.Ports = append(document.Ports, toJSONPort(port))
	}
	for _, sourceError := range scanReport.SourceErrors {
		document.SourceErrors = append(document.SourceErrors, jsonSourceError{
			Source: string(sourceError.Source), Error: sourceError.Err.Error(),
		})
	}
	encoder := json.NewEncoder(output)
	encoder.SetIndent(jsonPrefix, jsonIndent)
	return encoder.Encode(document)
}

func toJSONPort(port domain.Port) jsonPort {
	return jsonPort{
		Port:        port.Number,
		Protocol:    string(port.Protocol),
		BindAddress: port.BindAddress,
		Exposure:    port.Exposure().String(),
		Source:      string(port.Source),
		Process:     port.Process,
		Container:   port.Container,
	}
}
