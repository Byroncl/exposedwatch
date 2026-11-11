package tui

import (
	"testing"

	"github.com/Byroncl/exposedwatch/internal/app"
	"github.com/Byroncl/exposedwatch/internal/domain"
	"github.com/Byroncl/exposedwatch/internal/risk"
)

const previewNewPortNumber = 6379

func previewPort(number int, protocol domain.Protocol, bindAddress, process, container string, source domain.Source) domain.Port {
	return domain.Port{Number: number, Protocol: protocol, BindAddress: bindAddress, Process: process, Container: container, Source: source}
}

func previewPorts() []domain.Port {
	return []domain.Port{
		previewPort(22, domain.TCP, "0.0.0.0", "sshd", "", domain.SourceHost),
		previewPort(23, domain.TCP, "0.0.0.0", "telnetd", "", domain.SourceHost),
		previewPort(80, domain.TCP, "0.0.0.0", "nginx", "", domain.SourceHost),
		previewPort(443, domain.TCP, "0.0.0.0", "nginx", "", domain.SourceHost),
		previewPort(631, domain.TCP, "127.0.0.1", "cupsd", "", domain.SourceHost),
		previewPort(5432, domain.TCP, "10.0.0.5", "postgres", "", domain.SourceHost),
		previewPort(2375, domain.TCP, "0.0.0.0", "dockerd", "", domain.SourceHost),
		previewPort(6379, domain.TCP, "0.0.0.0", "redis-server", "cache", domain.SourceDocker),
		previewPort(8080, domain.TCP, "0.0.0.0", "", "web-app", domain.SourceDocker),
		previewPort(9000, domain.TCP, "172.17.0.1", "", "metrics", domain.SourceDocker),
		previewPort(53, domain.UDP, "127.0.0.53", "systemd-resolve", "", domain.SourceHost),
		previewPort(8443, domain.TCP, "203.0.113.10", "", "", domain.SourceExternal),
	}
}

func previewReport(ports []domain.Port) app.Report {
	return app.Report{Ports: ports, Findings: risk.Evaluate(ports)}
}

func previewConnection(direction domain.ConnectionDirection, localPort int, remoteAddress string, remotePort int, process string) domain.Connection {
	return testConnection(direction, localPort, remoteAddress, remotePort, process)
}

// previewConnectionsBaseline is the first connection snapshot of the preview.
func previewConnectionsBaseline() []domain.Connection {
	return []domain.Connection{
		previewConnection(domain.ConnectionInbound, 22, "203.0.113.7", 51000, "sshd"),
		previewConnection(domain.ConnectionInbound, 22, "203.0.113.8", 51044, "sshd"),
		previewConnection(domain.ConnectionInbound, 443, "198.51.100.23", 40001, "nginx"),
		previewConnection(domain.ConnectionInbound, 443, "198.51.100.24", 40002, "nginx"),
		previewConnection(domain.ConnectionInbound, 443, "203.0.113.7", 40003, "nginx"),
		previewConnection(domain.ConnectionInbound, 5432, "10.0.0.9", 52100, "postgres"),
		previewConnection(domain.ConnectionOutbound, 44122, "192.0.2.10", 443, "curl"),
		previewConnection(domain.ConnectionOutbound, 44190, "192.0.2.53", 5432, ""),
		previewConnection(domain.ConnectionInbound, 5432, "127.0.0.1", 50210, "postgres"),
	}
}

// previewConnections adds one new inbound connection to the baseline.
func previewConnections() []domain.Connection {
	return append(previewConnectionsBaseline(),
		previewConnection(domain.ConnectionInbound, 8080, "198.51.100.77", 41000, "webapp"))
}

// previewModel builds a model with a baseline, one NEW port and two closed ports.
func previewModel(testRun *testing.T) Model {
	testRun.Helper()
	var baseline []domain.Port
	for _, port := range previewPorts() {
		if port.Number != previewNewPortNumber {
			baseline = append(baseline, port)
		}
	}
	closedFirst := previewPort(3306, domain.TCP, "0.0.0.0", "mysqld", "", domain.SourceHost)
	closedSecond := previewPort(3389, domain.TCP, "0.0.0.0", "xrdp", "", domain.SourceHost)
	model := New(Options{Scanner: fakeScanner{}, Connections: fakeConnectionSource{}, Firewall: &fakeFirewall{}, Interval: testInterval, Version: testVersion})
	model = refreshWith(testRun, model, previewReport(append(append([]domain.Port{}, baseline...), closedFirst, closedSecond)), 0)
	model = refreshWith(testRun, model, previewReport(append(previewPorts(), closedSecond)), 1)
	model = refreshWith(testRun, model, previewReport(previewPorts()), 2)
	model = observeConnections(testRun, model, previewConnectionsBaseline(), 0)
	model = observeConnections(testRun, model, previewConnections(), 130)
	return applyMessage(testRun, model, FirewallResultMessage{Report: previewFirewallReport(), At: timeAfter(2)})
}

// previewFirewallReport evaluates the preview ports with one verdict of every kind.
func previewFirewallReport() domain.FirewallReport {
	verdicts := map[int]struct {
		verdict domain.FirewallVerdict
		reason  string
	}{
		22:   {domain.FirewallAllowed, "INPUT accepts new tcp/22 from any source"},
		23:   {domain.FirewallBlocked, "INPUT drops new tcp/23: no accept rule and the default policy is DROP"},
		80:   {domain.FirewallUnknown, "a rule uses a match the engine does not evaluate"},
		443:  {domain.FirewallAllowed, "INPUT accepts new tcp/443 from any source"},
		631:  {domain.FirewallNotExposed, "listens on loopback only, so it is not reachable from outside"},
		2375: {domain.FirewallBlocked, "INPUT rejects new tcp/2375 from any source"},
		5432: {domain.FirewallRestricted, "INPUT accepts new tcp/5432 only from 10.0.0.0/8"},
		6379: {domain.FirewallDockerPublished, "published by Docker; usually bypasses INPUT and ufw rules"},
		8080: {domain.FirewallDockerPublished, "published by Docker; usually bypasses INPUT and ufw rules"},
		9000: {domain.FirewallDockerPublished, "published by Docker; usually bypasses INPUT and ufw rules"},
		53:   {domain.FirewallNotExposed, "listens on loopback only, so it is not reachable from outside"},
	}
	var findings []domain.FirewallFinding
	for _, port := range previewPorts() {
		if entry, evaluated := verdicts[port.Number]; evaluated && port.Source != domain.SourceExternal {
			findings = append(findings, firewallFinding(port, entry.verdict, entry.reason))
		}
	}
	return domain.FirewallReport{Backend: "iptables-save", Available: true, Findings: findings, Notes: []string{"IPv6 rules were not readable"}}
}

// previewUnavailableFirewallReport is what a normal user gets: no rules, every exposed port unknown.
func previewUnavailableFirewallReport() domain.FirewallReport {
	var findings []domain.FirewallFinding
	for _, port := range previewPorts() {
		if port.Source != domain.SourceExternal {
			findings = append(findings, firewallFinding(port, domain.FirewallUnknown, "needs root to read firewall rules"))
		}
	}
	return domain.FirewallReport{Backend: domain.BackendUnknown, Findings: findings, Notes: []string{"needs root to read firewall rules"}}
}
