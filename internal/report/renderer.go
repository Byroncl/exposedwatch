// Package report renders scan reports for humans and machines.
package report

import (
	"io"

	"github.com/Byroncl/exposedwatch/internal/app"
	"github.com/Byroncl/exposedwatch/internal/domain"
)

// Renderer writes a report to an output stream.
type Renderer interface {
	Render(output io.Writer, scanReport app.Report) error
}

// FirewallRenderer is a Renderer that can also show the firewall verdicts of
// a report. Verdicts are information only: they never change a finding.
type FirewallRenderer interface {
	Renderer
	RenderWithFirewall(output io.Writer, scanReport app.Report, firewallReport domain.FirewallReport) error
}
