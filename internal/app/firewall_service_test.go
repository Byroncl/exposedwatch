package app

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/Byroncl/exposedwatch/internal/domain"
	"github.com/Byroncl/exposedwatch/internal/ports"
)

type fakeRuleSource struct {
	ruleSet domain.RuleSet
	err     error
}

func (source fakeRuleSource) LoadRuleSet(context.Context) (domain.RuleSet, error) {
	return source.ruleSet, source.err
}

func TestFirewallServiceUnavailableToolsAreNotFatal(testRun *testing.T) {
	testCases := []struct {
		name           string
		err            error
		reasonFragment string
	}{
		{"no tool", fmt.Errorf("wrapped: %w", ports.ErrFirewallUnavailable), "no iptables tool found"},
		{"needs root", fmt.Errorf("wrapped: %w", ports.ErrFirewallNeedsRoot), "needs root to read firewall rules"},
	}
	listening := []domain.Port{
		tcpPort(testSSHPort, testAnyBind),
		tcpPort(testWebPort, testLoopbackBind),
		{Number: 8080, Protocol: domain.TCP, BindAddress: testAnyBind, Source: domain.SourceDocker},
	}
	for _, testCase := range testCases {
		testRun.Run(testCase.name, func(testRun *testing.T) {
			report, err := NewFirewallService(fakeRuleSource{err: testCase.err}).Evaluate(context.Background(), listening)
			if err != nil || report.Available {
				testRun.Fatalf("report=%+v err=%v", report, err)
			}
			if len(report.Findings) != 3 || len(report.Notes) != 1 {
				testRun.Fatalf("findings=%d notes=%v", len(report.Findings), report.Notes)
			}
			byPort := map[int]domain.FirewallFinding{}
			for _, finding := range report.Findings {
				byPort[finding.Port.Number] = finding
			}
			if byPort[testSSHPort].Verdict != domain.FirewallUnknown || !strings.Contains(byPort[testSSHPort].Reason, testCase.reasonFragment) {
				testRun.Errorf("exposed host port = %+v", byPort[testSSHPort])
			}
			if byPort[testWebPort].Verdict != domain.FirewallNotExposed || byPort[8080].Verdict != domain.FirewallDockerPublished {
				testRun.Errorf("rule-independent verdicts lost: %+v", report.Findings)
			}
		})
	}
}

func TestFirewallServiceReturnsOtherErrors(testRun *testing.T) {
	boom := errors.New("kernel exploded")
	_, err := NewFirewallService(fakeRuleSource{err: boom}).Evaluate(context.Background(), nil)
	if !errors.Is(err, boom) {
		testRun.Fatalf("err=%v", err)
	}
}

func TestFirewallServiceWithoutSource(testRun *testing.T) {
	_, err := NewFirewallService(nil).Evaluate(context.Background(), nil)
	if !errors.Is(err, ErrNoFirewallSource) {
		testRun.Fatalf("err=%v", err)
	}
}

func TestFirewallServiceOrdersFindingsAndSkipsExternalPorts(testRun *testing.T) {
	ruleSet := ipv4RuleSet(domain.TargetAccept, nil)
	listening := []domain.Port{
		tcpPort(443, testSpecificBind),
		{Number: 22, Protocol: domain.UDP, BindAddress: testAnyBind, Source: domain.SourceHost},
		tcpPort(22, testOtherBind),
		{Number: 22, Protocol: domain.TCP, BindAddress: testAnyBind, Source: domain.SourceExternal},
		tcpPort(22, testAnyBind),
		tcpPort(80, testAnyBind),
	}
	report, err := NewFirewallService(fakeRuleSource{ruleSet: ruleSet}).Evaluate(context.Background(), listening)
	if err != nil || !report.Available {
		testRun.Fatalf("report=%+v err=%v", report, err)
	}
	var order []string
	for _, finding := range report.Findings {
		order = append(order, fmt.Sprintf("%d/%s/%s", finding.Port.Number, finding.Port.Protocol, finding.Port.BindAddress))
	}
	expected := []string{"22/tcp/0.0.0.0", "22/tcp/192.168.1.20", "22/udp/0.0.0.0", "80/tcp/0.0.0.0", "443/tcp/192.168.1.10"}
	if strings.Join(order, " ") != strings.Join(expected, " ") {
		testRun.Fatalf("order = %v, want %v", order, expected)
	}
	if report.Backend != "iptables v1.8.11" {
		testRun.Fatalf("backend = %q", report.Backend)
	}
}

func TestFirewallServiceNativeNftablesHostIsUnknownWithNote(testRun *testing.T) {
	ruleSet := domain.RuleSet{
		IPv4: domain.FamilyRuleSet{Family: domain.FamilyIPv4, Available: true},
		IPv6: domain.FamilyRuleSet{Family: domain.FamilyIPv6, Available: true},
	}
	report, err := NewFirewallService(fakeRuleSource{ruleSet: ruleSet}).Evaluate(context.Background(), []domain.Port{tcpPort(testSSHPort, testAnyBind), tcpPort(testWebPort, testDualStackBind)})
	if err != nil || !report.Available {
		testRun.Fatalf("report=%+v err=%v", report, err)
	}
	if report.Backend != domain.BackendUnknown {
		testRun.Fatalf("backend = %q", report.Backend)
	}
	if len(report.Notes) != 1 || report.Notes[0] != "no iptables rules found; native nftables rules are not evaluated yet" {
		testRun.Fatalf("notes = %v", report.Notes)
	}
	for _, finding := range report.Findings {
		if finding.Verdict != domain.FirewallUnknown {
			testRun.Errorf("finding = %+v", finding)
		}
	}
}

func TestFirewallServiceKeepsRuleSetNotes(testRun *testing.T) {
	ruleSet := ipv4RuleSet(domain.TargetAccept, nil)
	ruleSet.Notes = []string{"ip6tables-save not found; IPv6 rules are not evaluated"}
	report, _ := NewFirewallService(fakeRuleSource{ruleSet: ruleSet}).Evaluate(context.Background(), nil)
	if len(report.Notes) != 1 || !strings.Contains(report.Notes[0], "ip6tables-save") {
		testRun.Fatalf("notes = %v", report.Notes)
	}
}
