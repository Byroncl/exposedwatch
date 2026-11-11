package app

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Byroncl/exposedwatch/internal/domain"
	"github.com/Byroncl/exposedwatch/internal/ports"
)

func natTable(chains ...domain.Chain) domain.Table {
	return domain.Table{Name: domain.TableNat, Chains: chains}
}

func tableWithRule(name domain.TableName, chain domain.ChainName, rule domain.Rule) domain.Table {
	return domain.Table{Name: name, Chains: []domain.Chain{{Name: chain, Policy: domain.TargetAccept, Rules: []domain.Rule{rule}}}}
}

func TestNatCaveatOnlyExemptsExactDockerChains(testRun *testing.T) {
	redirect := domain.Rule{Protocol: domain.TCP, DestinationPorts: []domain.PortRange{{First: 22, Last: 22}}, Target: domain.TargetRedirect}
	dnat := domain.Rule{Protocol: domain.TCP, DestinationPorts: []domain.PortRange{{First: 8080, Last: 8080}}, Target: domain.TargetDnat}
	build := func(chainName domain.ChainName, rule domain.Rule) domain.RuleSet {
		ruleSet := ipv4RuleSet(domain.TargetAccept, nil)
		ruleSet.IPv4 = withTable(ruleSet.IPv4, natTable(
			domain.Chain{Name: domain.ChainPrerouting, Policy: domain.TargetAccept, Rules: []domain.Rule{jumpRule(chainName)}},
			domain.Chain{Name: chainName, Rules: []domain.Rule{rule}},
		))
		return ruleSet
	}
	port := tcpPort(testSSHPort, testAnyBind)
	runEvaluationCases(testRun, []evaluationCase{
		{name: "lookalike chain hides a redirect", ruleSet: build("DOCKERISH", redirect), port: port, expected: domain.FirewallUnknown},
		{name: "lookalike chain hides a dnat", ruleSet: build("DOCKERISH", dnat), port: port, expected: domain.FirewallUnknown},
		{name: "exact DOCKER chain dnat is exempt", ruleSet: build("DOCKER", dnat), port: port, expected: domain.FirewallAllowed},
		{name: "exact DOCKER-INGRESS dnat is exempt", ruleSet: build("DOCKER-INGRESS", dnat), port: port, expected: domain.FirewallAllowed},
		{name: "redirect inside DOCKER is not exempt", ruleSet: build("DOCKER", redirect), port: port, expected: domain.FirewallUnknown},
	})
}

func TestPreFilterCaveatCoversConntrackAlteringTargets(testRun *testing.T) {
	for _, target := range []domain.TargetName{domain.TargetCt, domain.TargetNotrack, domain.TargetSynproxy, domain.TargetSecmark, domain.TargetConnsecmark} {
		for _, table := range []domain.TableName{domain.TableRaw, domain.TableMangle} {
			rule := domain.Rule{Protocol: domain.TCP, DestinationPorts: []domain.PortRange{{First: testSSHPort, Last: testSSHPort}}, Target: target}
			allowed := ipv4RuleSet(domain.TargetDrop, []domain.Rule{portRule(domain.TCP, testSSHPort, domain.TargetAccept)})
			allowed.IPv4 = withTable(allowed.IPv4, tableWithRule(table, domain.ChainPrerouting, rule))
			blocked := ipv4RuleSet(domain.TargetDrop, nil)
			blocked.IPv4 = withTable(blocked.IPv4, tableWithRule(table, domain.ChainPrerouting, rule))
			runEvaluationCases(testRun, []evaluationCase{
				{name: fmt.Sprintf("%s in %s makes allowed unknown", target, table), ruleSet: allowed, port: tcpPort(testSSHPort, testAnyBind), expected: domain.FirewallUnknown},
				{name: fmt.Sprintf("%s in %s makes blocked unknown", target, table), ruleSet: blocked, port: tcpPort(testSSHPort, testAnyBind), expected: domain.FirewallUnknown},
			})
		}
	}
}

func TestWalkStepBudgetStopsFanOut(testRun *testing.T) {
	const levels = 14
	const fanOut = 4
	var chains []domain.Chain
	for level := 0; level < levels; level++ {
		name := domain.ChainName(fmt.Sprintf("level-%d", level))
		next := domain.ChainName(fmt.Sprintf("level-%d", level+1))
		var rules []domain.Rule
		for branch := 0; branch < fanOut; branch++ {
			rules = append(rules, jumpRule(next))
		}
		chains = append(chains, domain.Chain{Name: name, Rules: rules})
	}
	chains = append(chains, domain.Chain{Name: domain.ChainName(fmt.Sprintf("level-%d", levels))})
	ruleSet := ipv4RuleSet(domain.TargetDrop, []domain.Rule{jumpRule("level-0")}, chains...)
	started := time.Now()
	finding := NewFirewallEvaluator().EvaluatePort(context.Background(), ruleSet, tcpPort(testSSHPort, testAnyBind))
	if time.Since(started) > 2*time.Second {
		testRun.Fatalf("walk took %v", time.Since(started))
	}
	if finding.Verdict != domain.FirewallUnknown || !strings.Contains(finding.Reason, "too many rules") {
		testRun.Fatalf("finding = %+v", finding)
	}
}

func TestWalkHonoursCancellation(testRun *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	rules := make([]domain.Rule, walkCancellationCheckInterval*2)
	for index := range rules {
		rules[index] = portRule(domain.TCP, testWebPort, domain.TargetDrop)
	}
	ruleSet := ipv4RuleSet(domain.TargetAccept, rules)
	finding := NewFirewallEvaluator().EvaluatePort(ctx, ruleSet, tcpPort(testSSHPort, testAnyBind))
	if finding.Verdict != domain.FirewallUnknown || !strings.Contains(finding.Reason, "cancelled") {
		testRun.Fatalf("finding = %+v", finding)
	}
}

func TestReasonsNeverLeakRuleTextOrComments(testRun *testing.T) {
	secret := "SECRET-COMMENT-TEXT"
	rule := portRule(domain.TCP, testSSHPort, domain.TargetAccept)
	rule.Text = `-A INPUT -p tcp --dport 22 -m comment --comment "` + secret + `" -j ACCEPT`
	rule.SourceCIDRs = prefixes(testPrivateNetwork)
	dropRule := portRule(domain.TCP, testWebPort, domain.TargetDrop)
	dropRule.Text = rule.Text
	ruleSet := ipv4RuleSet(domain.TargetAccept, []domain.Rule{rule, dropRule})
	evaluator := NewFirewallEvaluator()
	for _, port := range []domain.Port{tcpPort(testSSHPort, testAnyBind), tcpPort(testWebPort, testAnyBind)} {
		finding := evaluator.EvaluatePort(context.Background(), ruleSet, port)
		if strings.Contains(finding.Reason, secret) || strings.Contains(finding.Reason, "-A INPUT") {
			testRun.Errorf("reason leaks raw rule text: %q", finding.Reason)
		}
	}
	allowed := evaluator.EvaluatePort(context.Background(), ipv4RuleSet(domain.TargetDrop, []domain.Rule{portRule(domain.TCP, testSSHPort, domain.TargetAccept)}), tcpPort(testSSHPort, testAnyBind))
	if allowed.Reason != "allowed by rule in INPUT: tcp dport 22 -> ACCEPT" {
		testRun.Fatalf("reason = %q", allowed.Reason)
	}
}

func TestLoopbackInterfacePatterns(testRun *testing.T) {
	for _, name := range []domain.InterfaceName{"lo", "lo+"} {
		ruleSet := ipv4RuleSet(domain.TargetDrop, []domain.Rule{{InInterface: name, Target: domain.TargetAccept}})
		runEvaluationCases(testRun, []evaluationCase{
			{name: "interface " + string(name) + " is loopback", ruleSet: ruleSet, port: tcpPort(testSSHPort, testAnyBind), expected: domain.FirewallBlocked},
		})
	}
}

func TestNftablesStateDowngradesUnconfirmedVerdicts(testRun *testing.T) {
	restrictedRules := []domain.Rule{{SourceCIDRs: prefixes(testPrivateNetwork), Target: domain.TargetAccept}}
	build := func(state domain.NftablesState, policy domain.TargetName, rules []domain.Rule) domain.RuleSet {
		ruleSet := ipv4RuleSet(policy, rules)
		ruleSet.Nftables = state
		return ruleSet
	}
	port := tcpPort(testSSHPort, testAnyBind)
	runEvaluationCases(testRun, []evaluationCase{
		{name: "native tables make allowed unknown", ruleSet: build(domain.NftablesNative, domain.TargetAccept, nil), port: port, expected: domain.FirewallUnknown, reasonFragments: []string{"native nftables rules present and not evaluated yet"}},
		{name: "native tables make restricted unknown", ruleSet: build(domain.NftablesNative, domain.TargetDrop, restrictedRules), port: port, expected: domain.FirewallUnknown},
		{name: "native tables keep an iptables block", ruleSet: build(domain.NftablesNative, domain.TargetDrop, nil), port: port, expected: domain.FirewallBlocked},
		{name: "failed probe makes allowed unknown", ruleSet: build(domain.NftablesProbeFailed, domain.TargetAccept, nil), port: port, expected: domain.FirewallUnknown, reasonFragments: []string{"could not check nftables"}},
		{name: "failed probe keeps an iptables block", ruleSet: build(domain.NftablesProbeFailed, domain.TargetDrop, nil), port: port, expected: domain.FirewallBlocked},
		{name: "no native tables keeps allowed", ruleSet: build(domain.NftablesNone, domain.TargetAccept, nil), port: port, expected: domain.FirewallAllowed},
		{name: "unchecked behaves as before", ruleSet: build(domain.NftablesNotChecked, domain.TargetAccept, nil), port: port, expected: domain.FirewallAllowed},
	})
}

func TestFirewallServiceTruncatedOrInvalidOutputIsUnavailable(testRun *testing.T) {
	testCases := map[error]string{
		ports.ErrFirewallOutputTruncated: "truncated",
		ports.ErrFirewallOutputInvalid:   "could not be understood",
	}
	for sentinel, fragment := range testCases {
		report, err := NewFirewallService(fakeRuleSource{err: fmt.Errorf("wrapped: %w", sentinel)}).Evaluate(context.Background(), []domain.Port{tcpPort(testSSHPort, testAnyBind)})
		if err != nil || report.Available || len(report.Findings) != 1 {
			testRun.Fatalf("report=%+v err=%v", report, err)
		}
		if report.Findings[0].Verdict != domain.FirewallUnknown || !strings.Contains(report.Findings[0].Reason, fragment) {
			testRun.Errorf("finding = %+v", report.Findings[0])
		}
	}
}

func TestFirewallServiceStopsOnCancelledContext(testRun *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := NewFirewallService(fakeRuleSource{ruleSet: ipv4RuleSet(domain.TargetAccept, nil)}).Evaluate(ctx, []domain.Port{tcpPort(testSSHPort, testAnyBind)})
	if err == nil {
		testRun.Fatal("a cancelled context must be reported")
	}
}
