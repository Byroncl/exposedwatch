package report

import "github.com/Byroncl/exposedwatch/internal/domain"

// jsonFirewallPort is the stable JSON shape of one firewall verdict.
type jsonFirewallPort struct {
	Port        int    `json:"port"`
	Protocol    string `json:"protocol"`
	BindAddress string `json:"bind_address"`
	Source      string `json:"source"`
	Verdict     string `json:"verdict"`
	Reason      string `json:"reason"`
}

// jsonFirewall is the stable JSON shape of the firewall evaluation.
type jsonFirewall struct {
	Available bool               `json:"available"`
	Backend   string             `json:"backend"`
	Notes     []string           `json:"notes"`
	Ports     []jsonFirewallPort `json:"ports"`
}

func toJSONFirewall(firewallReport domain.FirewallReport) *jsonFirewall {
	firewall := &jsonFirewall{
		Available: firewallReport.Available,
		Backend:   firewallReport.Backend,
		Notes:     make([]string, 0, len(firewallReport.Notes)),
		Ports:     make([]jsonFirewallPort, 0, len(firewallReport.Findings)),
	}
	firewall.Notes = append(firewall.Notes, firewallReport.Notes...)
	for _, finding := range firewallReport.Findings {
		firewall.Ports = append(firewall.Ports, jsonFirewallPort{
			Port:        finding.Port.Number,
			Protocol:    string(finding.Port.Protocol),
			BindAddress: finding.Port.BindAddress,
			Source:      string(finding.Port.Source),
			Verdict:     finding.Verdict.String(),
			Reason:      finding.Reason,
		})
	}
	return firewall
}
