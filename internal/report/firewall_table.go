package report

import (
	"fmt"
	"strings"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

// verdictKey identifies the port a firewall finding belongs to.
type verdictKey struct {
	number      int
	protocol    domain.Protocol
	bindAddress string
	source      domain.Source
	container   string
}

func verdictKeyOf(port domain.Port) verdictKey {
	return verdictKey{number: port.Number, protocol: port.Protocol, bindAddress: port.BindAddress, source: port.Source, container: port.Container}
}

// verdictLookup finds the firewall verdict of a port.
type verdictLookup map[verdictKey]domain.FirewallVerdict

func newVerdictLookup(firewallReport *domain.FirewallReport) verdictLookup {
	lookup := verdictLookup{}
	if firewallReport == nil {
		return lookup
	}
	for _, finding := range firewallReport.Findings {
		lookup[verdictKeyOf(finding.Port)] = finding.Verdict
	}
	return lookup
}

// wordFor is the verdict word of the port, or a placeholder when the port was
// not evaluated (for example a port seen from outside).
func (lookup verdictLookup) wordFor(port domain.Port) string {
	verdict, found := lookup[verdictKeyOf(port)]
	if !found {
		return emptyCellText
	}
	return verdictWord(verdict)
}

// verdictWord is the short word shown in the table.
func verdictWord(verdict domain.FirewallVerdict) string {
	if verdict == domain.FirewallDockerPublished {
		return verdictWordDocker
	}
	return verdict.String()
}

// writeFirewallSection prints the backend, whether the rules could be read and
// the notes of the engine, and says that the verdicts are information only.
func (renderer *TableRenderer) writeFirewallSection(builder *strings.Builder, firewallReport domain.FirewallReport) {
	builder.WriteString(newline + paint(renderer.colorMode, ansiBold, firewallSectionTitle) + newline)
	fmt.Fprintf(builder, firewallBackendFormat, firewallReport.Backend)
	available := availableNo
	if firewallReport.Available {
		available = availableYes
	}
	fmt.Fprintf(builder, firewallAvailableFormat, available)
	for _, note := range firewallReport.Notes {
		fmt.Fprintf(builder, firewallNoteFormat, note)
	}
	builder.WriteString(firewallInformationText)
}
