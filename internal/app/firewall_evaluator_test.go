package app

import (
	"context"
	"strings"
	"testing"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

type evaluationCase struct {
	name            string
	ruleSet         domain.RuleSet
	port            domain.Port
	expected        domain.FirewallVerdict
	reasonFragments []string
}

func runEvaluationCases(testRun *testing.T, testCases []evaluationCase) {
	testRun.Helper()
	evaluator := NewFirewallEvaluator()
	for _, testCase := range testCases {
		testRun.Run(testCase.name, func(testRun *testing.T) {
			finding := evaluator.EvaluatePort(context.Background(), testCase.ruleSet, testCase.port)
			if finding.Verdict != testCase.expected {
				testRun.Fatalf("verdict = %v (%s), want %v", finding.Verdict, finding.Reason, testCase.expected)
			}
			for _, fragment := range testCase.reasonFragments {
				if !strings.Contains(finding.Reason, fragment) {
					testRun.Errorf("reason %q lacks %q", finding.Reason, fragment)
				}
			}
			if finding.Reason == "" {
				testRun.Error("every finding needs a reason")
			}
			if finding.Port != testCase.port {
				testRun.Error("finding must carry the evaluated port")
			}
		})
	}
}

func TestEvaluatorPolicyAndSimpleRules(testRun *testing.T) {
	runEvaluationCases(testRun, []evaluationCase{
		{
			name:            "no rules with policy ACCEPT is allowed",
			ruleSet:         ipv4RuleSet(domain.TargetAccept, nil),
			port:            tcpPort(testSSHPort, testAnyBind),
			expected:        domain.FirewallAllowed,
			reasonFragments: []string{"no firewall rules (policy ACCEPT)"},
		},
		{
			name:            "policy ACCEPT with unrelated rules is allowed by policy",
			ruleSet:         ipv4RuleSet(domain.TargetAccept, []domain.Rule{portRule(domain.TCP, testWebPort, domain.TargetDrop)}),
			port:            tcpPort(testSSHPort, testAnyBind),
			expected:        domain.FirewallAllowed,
			reasonFragments: []string{"INPUT policy ACCEPT"},
		},
		{
			name:            "policy DROP with no rule is blocked",
			ruleSet:         ipv4RuleSet(domain.TargetDrop, nil),
			port:            tcpPort(testSSHPort, testAnyBind),
			expected:        domain.FirewallBlocked,
			reasonFragments: []string{"INPUT policy DROP and no rule allows tcp/22"},
		},
		{
			name:            "explicit accept rule is allowed",
			ruleSet:         ipv4RuleSet(domain.TargetDrop, []domain.Rule{portRule(domain.TCP, testSSHPort, domain.TargetAccept)}),
			port:            tcpPort(testSSHPort, testAnyBind),
			expected:        domain.FirewallAllowed,
			reasonFragments: []string{"allowed by rule in INPUT:"},
		},
		{
			name:     "unlisted port is blocked by policy",
			ruleSet:  ipv4RuleSet(domain.TargetDrop, []domain.Rule{portRule(domain.TCP, testSSHPort, domain.TargetAccept)}),
			port:     tcpPort(testUnlistedPort, testAnyBind),
			expected: domain.FirewallBlocked,
		},
		{
			name:            "explicit drop rule is blocked",
			ruleSet:         ipv4RuleSet(domain.TargetAccept, []domain.Rule{portRule(domain.TCP, testSSHPort, domain.TargetDrop)}),
			port:            tcpPort(testSSHPort, testAnyBind),
			expected:        domain.FirewallBlocked,
			reasonFragments: []string{"blocked by rule in INPUT:"},
		},
		{
			name:     "reject rule is blocked",
			ruleSet:  ipv4RuleSet(domain.TargetAccept, []domain.Rule{portRule(domain.TCP, testSSHPort, domain.TargetReject)}),
			port:     tcpPort(testSSHPort, testAnyBind),
			expected: domain.FirewallBlocked,
		},
		{
			name:     "first terminal rule wins",
			ruleSet:  ipv4RuleSet(domain.TargetDrop, []domain.Rule{portRule(domain.TCP, testSSHPort, domain.TargetDrop), portRule(domain.TCP, testSSHPort, domain.TargetAccept)}),
			port:     tcpPort(testSSHPort, testAnyBind),
			expected: domain.FirewallBlocked,
		},
		{
			name:     "udp rule does not allow a tcp port",
			ruleSet:  ipv4RuleSet(domain.TargetDrop, []domain.Rule{portRule(domain.UDP, testSSHPort, domain.TargetAccept)}),
			port:     tcpPort(testSSHPort, testAnyBind),
			expected: domain.FirewallBlocked,
		},
		{
			name: "port range rule matches inside the range",
			ruleSet: ipv4RuleSet(domain.TargetDrop, []domain.Rule{{
				Protocol: domain.TCP, DestinationPorts: []domain.PortRange{{First: 8000, Last: 8010}}, Target: domain.TargetAccept,
			}}),
			port:     tcpPort(8005, testAnyBind),
			expected: domain.FirewallAllowed,
		},
		{
			name: "protocol-less accept covers every port",
			ruleSet: ipv4RuleSet(domain.TargetDrop, []domain.Rule{
				targetRule(domain.TargetAccept),
			}),
			port:     tcpPort(testUnlistedPort, testAnyBind),
			expected: domain.FirewallAllowed,
		},
		{
			name:     "icmp rule is not a match for tcp",
			ruleSet:  ipv4RuleSet(domain.TargetDrop, []domain.Rule{{Protocol: "icmp", Target: domain.TargetAccept}}),
			port:     tcpPort(testSSHPort, testAnyBind),
			expected: domain.FirewallBlocked,
		},
	})
}

func TestEvaluatorConnectionStateInterfaceAndAddressType(testRun *testing.T) {
	established := domain.Rule{
		ConnectionStates: []domain.ConntrackState{domain.StateRelated, domain.StateEstablished},
		Target:           domain.TargetAccept,
	}
	newState := domain.Rule{ConnectionStates: []domain.ConntrackState{domain.StateNew}, Target: domain.TargetAccept}
	loopback := domain.Rule{InInterface: "lo", Target: domain.TargetAccept}
	multicast := domain.Rule{AddressTypes: []domain.AddressType{domain.AddressTypeMulticast}, Target: domain.TargetAccept}
	local := domain.Rule{AddressTypes: []domain.AddressType{domain.AddressTypeLocal}, Target: domain.TargetAccept}
	unusualType := domain.Rule{AddressTypes: []domain.AddressType{"UNICAST"}, Target: domain.TargetAccept}
	runEvaluationCases(testRun, []evaluationCase{
		{name: "established accept does not match a new connection", ruleSet: ipv4RuleSet(domain.TargetDrop, []domain.Rule{established}), port: tcpPort(testSSHPort, testAnyBind), expected: domain.FirewallBlocked},
		{name: "new state accept matches", ruleSet: ipv4RuleSet(domain.TargetDrop, []domain.Rule{newState}), port: tcpPort(testSSHPort, testAnyBind), expected: domain.FirewallAllowed},
		{name: "loopback interface accept does not match", ruleSet: ipv4RuleSet(domain.TargetDrop, []domain.Rule{loopback}), port: tcpPort(testSSHPort, testAnyBind), expected: domain.FirewallBlocked},
		{name: "multicast type accept does not match", ruleSet: ipv4RuleSet(domain.TargetDrop, []domain.Rule{multicast}), port: tcpPort(testSSHPort, testAnyBind), expected: domain.FirewallBlocked},
		{name: "local type accept matches", ruleSet: ipv4RuleSet(domain.TargetDrop, []domain.Rule{local}), port: tcpPort(testSSHPort, testAnyBind), expected: domain.FirewallAllowed},
		{name: "unusual address type is uncertain", ruleSet: ipv4RuleSet(domain.TargetDrop, []domain.Rule{unusualType}), port: tcpPort(testSSHPort, testAnyBind), expected: domain.FirewallUnknown},
		{
			name: "specific interface restricts instead of allowing",
			ruleSet: ipv4RuleSet(domain.TargetDrop, []domain.Rule{
				{InInterface: "eth0", Protocol: domain.TCP, DestinationPorts: []domain.PortRange{{First: testSSHPort, Last: testSSHPort}}, Target: domain.TargetAccept},
			}),
			port: tcpPort(testSSHPort, testAnyBind), expected: domain.FirewallRestricted, reasonFragments: []string{"interface eth0"},
		},
	})
}

func TestEvaluatorTCPFlags(testRun *testing.T) {
	synOnly := &domain.TCPFlagsMatch{
		Mask: []domain.TCPFlag{domain.TCPFlagFin, domain.TCPFlagSyn, domain.TCPFlagRst, domain.TCPFlagAck},
		Set:  []domain.TCPFlag{domain.TCPFlagSyn},
	}
	nullScan := &domain.TCPFlagsMatch{
		Mask: []domain.TCPFlag{domain.TCPFlagFin, domain.TCPFlagSyn, domain.TCPFlagRst, domain.TCPFlagPsh, domain.TCPFlagAck, domain.TCPFlagUrg},
	}
	finOnly := &domain.TCPFlagsMatch{Mask: []domain.TCPFlag{domain.TCPFlagFin, domain.TCPFlagSyn}, Set: []domain.TCPFlag{domain.TCPFlagFin}}
	build := func(flags *domain.TCPFlagsMatch, target domain.TargetName) domain.RuleSet {
		return ipv4RuleSet(domain.TargetAccept, []domain.Rule{{Protocol: domain.TCP, TCPFlags: flags, Target: target}})
	}
	runEvaluationCases(testRun, []evaluationCase{
		{name: "syn-only drop matches a new connection", ruleSet: build(synOnly, domain.TargetDrop), port: tcpPort(testSSHPort, testAnyBind), expected: domain.FirewallBlocked},
		{name: "null scan drop does not match", ruleSet: build(nullScan, domain.TargetDrop), port: tcpPort(testSSHPort, testAnyBind), expected: domain.FirewallAllowed},
		{name: "fin drop does not match", ruleSet: build(finOnly, domain.TargetDrop), port: tcpPort(testSSHPort, testAnyBind), expected: domain.FirewallAllowed},
	})
}

func TestEvaluatorSourceRestrictions(testRun *testing.T) {
	restrictedAccept := domain.Rule{
		SourceCIDRs: prefixes(testPrivateNetwork), Protocol: domain.TCP,
		DestinationPorts: []domain.PortRange{{First: testDatabasePort, Last: testDatabasePort}}, Target: domain.TargetAccept,
	}
	restrictedDrop := domain.Rule{SourceCIDRs: prefixes("192.0.2.0/24"), Target: domain.TargetDrop}
	everywhere := domain.Rule{SourceCIDRs: prefixes("0.0.0.0/0"), Protocol: domain.TCP, DestinationPorts: []domain.PortRange{{First: testSSHPort, Last: testSSHPort}}, Target: domain.TargetAccept}
	loopbackSource := domain.Rule{SourceCIDRs: prefixes("127.0.0.0/8"), Target: domain.TargetAccept}
	restrictedReturn := domain.Rule{SourceCIDRs: prefixes("192.0.2.0/24"), Target: domain.TargetReturn}
	runEvaluationCases(testRun, []evaluationCase{
		{
			name: "accept for a source range only is restricted", ruleSet: ipv4RuleSet(domain.TargetDrop, []domain.Rule{restrictedAccept}),
			port: tcpPort(testDatabasePort, testAnyBind), expected: domain.FirewallRestricted, reasonFragments: []string{"restricted to source 10.0.0.0/8"},
		},
		{
			name: "restricted accept with policy ACCEPT stays allowed", ruleSet: ipv4RuleSet(domain.TargetAccept, []domain.Rule{restrictedAccept}),
			port: tcpPort(testDatabasePort, testAnyBind), expected: domain.FirewallAllowed,
		},
		{
			name: "restricted drop then general accept is restricted", ruleSet: ipv4RuleSet(domain.TargetAccept, []domain.Rule{restrictedDrop}),
			port: tcpPort(testSSHPort, testAnyBind), expected: domain.FirewallRestricted, reasonFragments: []string{"blocked for source 192.0.2.0/24"},
		},
		{
			name: "restricted drop with policy DROP is blocked", ruleSet: ipv4RuleSet(domain.TargetDrop, []domain.Rule{restrictedDrop}),
			port: tcpPort(testSSHPort, testAnyBind), expected: domain.FirewallBlocked,
		},
		{
			name: "source 0.0.0.0/0 is not a restriction", ruleSet: ipv4RuleSet(domain.TargetDrop, []domain.Rule{everywhere}),
			port: tcpPort(testSSHPort, testAnyBind), expected: domain.FirewallAllowed,
		},
		{
			name: "loopback-only source never matches an outside source", ruleSet: ipv4RuleSet(domain.TargetDrop, []domain.Rule{loopbackSource}),
			port: tcpPort(testSSHPort, testAnyBind), expected: domain.FirewallBlocked,
		},
		{
			name: "restricted RETURN is uncertain", ruleSet: ipv4RuleSet(domain.TargetDrop, []domain.Rule{restrictedReturn}),
			port: tcpPort(testSSHPort, testAnyBind), expected: domain.FirewallUnknown,
		},
	})
}

func TestEvaluatorDestinationAddresses(testRun *testing.T) {
	specificAccept := domain.Rule{DestinationCIDRs: prefixes("192.168.1.10/32"), Target: domain.TargetAccept}
	coveringAccept := domain.Rule{DestinationCIDRs: prefixes("192.168.1.0/24"), Target: domain.TargetAccept}
	everywhere := domain.Rule{DestinationCIDRs: prefixes("0.0.0.0/0"), Target: domain.TargetAccept}
	loopbackDestination := domain.Rule{DestinationCIDRs: prefixes("127.0.0.0/8"), Target: domain.TargetAccept}
	runEvaluationCases(testRun, []evaluationCase{
		{name: "equal destination matches", ruleSet: ipv4RuleSet(domain.TargetDrop, []domain.Rule{specificAccept}), port: tcpPort(testSSHPort, testSpecificBind), expected: domain.FirewallAllowed},
		{name: "covering destination matches", ruleSet: ipv4RuleSet(domain.TargetDrop, []domain.Rule{coveringAccept}), port: tcpPort(testSSHPort, testSpecificBind), expected: domain.FirewallAllowed},
		{name: "different destination does not match", ruleSet: ipv4RuleSet(domain.TargetDrop, []domain.Rule{specificAccept}), port: tcpPort(testSSHPort, testOtherBind), expected: domain.FirewallBlocked},
		{name: "wildcard bind with specific destination is uncertain", ruleSet: ipv4RuleSet(domain.TargetDrop, []domain.Rule{specificAccept}), port: tcpPort(testSSHPort, testAnyBind), expected: domain.FirewallUnknown},
		{name: "wildcard bind with destination 0.0.0.0/0 matches", ruleSet: ipv4RuleSet(domain.TargetDrop, []domain.Rule{everywhere}), port: tcpPort(testSSHPort, testAnyBind), expected: domain.FirewallAllowed},
		{name: "loopback destination never receives outside traffic", ruleSet: ipv4RuleSet(domain.TargetDrop, []domain.Rule{loopbackDestination}), port: tcpPort(testSSHPort, testAnyBind), expected: domain.FirewallBlocked},
	})
}

func TestEvaluatorUnsupportedAndUncertainRules(testRun *testing.T) {
	runEvaluationCases(testRun, []evaluationCase{
		{
			name: "unsupported match before accept is unknown",
			ruleSet: ipv4RuleSet(domain.TargetDrop, []domain.Rule{
				{Protocol: domain.TCP, DestinationPorts: []domain.PortRange{{First: testSSHPort, Last: testSSHPort}}, Unsupported: true, UnsupportedReason: "module: set", Target: domain.TargetAccept},
				portRule(domain.TCP, testSSHPort, domain.TargetAccept),
			}),
			port: tcpPort(testSSHPort, testAnyBind), expected: domain.FirewallUnknown,
			reasonFragments: []string{"contains rules the engine cannot evaluate (module: set)"},
		},
		{
			name: "unsupported drop before accept is unknown not blocked",
			ruleSet: ipv4RuleSet(domain.TargetAccept, []domain.Rule{
				unsupportedRule("module: recent", domain.TargetDrop),
			}),
			port: tcpPort(testSSHPort, testAnyBind), expected: domain.FirewallUnknown,
		},
		{
			name: "unsupported rule for another port is irrelevant",
			ruleSet: ipv4RuleSet(domain.TargetDrop, []domain.Rule{
				{Protocol: domain.TCP, DestinationPorts: []domain.PortRange{{First: testWebPort, Last: testWebPort}}, Unsupported: true, UnsupportedReason: "module: set", Target: domain.TargetAccept},
				portRule(domain.TCP, testSSHPort, domain.TargetAccept),
			}),
			port: tcpPort(testSSHPort, testAnyBind), expected: domain.FirewallAllowed,
		},
		{
			name:     "unsupported logging rule does not change the outcome",
			ruleSet:  ipv4RuleSet(domain.TargetDrop, []domain.Rule{unsupportedRule("module: limit", domain.TargetLog), portRule(domain.TCP, testSSHPort, domain.TargetAccept)}),
			port:     tcpPort(testSSHPort, testAnyBind),
			expected: domain.FirewallAllowed,
		},
		{
			name:     "unsupported counting rule without target does not change the outcome",
			ruleSet:  ipv4RuleSet(domain.TargetDrop, []domain.Rule{unsupportedRule("module: recent", ""), portRule(domain.TCP, testSSHPort, domain.TargetAccept)}),
			port:     tcpPort(testSSHPort, testAnyBind),
			expected: domain.FirewallAllowed,
		},
		{
			name:            "unknown target such as MARK is uncertain",
			ruleSet:         ipv4RuleSet(domain.TargetDrop, []domain.Rule{targetRule("MARK"), portRule(domain.TCP, testSSHPort, domain.TargetAccept)}),
			port:            tcpPort(testSSHPort, testAnyBind),
			expected:        domain.FirewallUnknown,
			reasonFragments: []string{"MARK"},
		},
		{
			name:     "NFQUEUE is uncertain",
			ruleSet:  ipv4RuleSet(domain.TargetDrop, []domain.Rule{targetRule(domain.TargetNfqueue)}),
			port:     tcpPort(testSSHPort, testAnyBind),
			expected: domain.FirewallUnknown,
		},
		{
			name:     "LOG target is not terminal",
			ruleSet:  ipv4RuleSet(domain.TargetDrop, []domain.Rule{targetRule(domain.TargetLog), portRule(domain.TCP, testSSHPort, domain.TargetAccept)}),
			port:     tcpPort(testSSHPort, testAnyBind),
			expected: domain.FirewallAllowed,
		},
		{
			name:     "jump to a missing chain is uncertain",
			ruleSet:  ipv4RuleSet(domain.TargetDrop, []domain.Rule{jumpRule("ghost-chain")}),
			port:     tcpPort(testSSHPort, testAnyBind),
			expected: domain.FirewallUnknown,
		},
		{
			name:     "built-in policy that is not ACCEPT or DROP is uncertain",
			ruleSet:  ipv4RuleSet("", nil),
			port:     tcpPort(testSSHPort, testAnyBind),
			expected: domain.FirewallUnknown,
		},
	})
}

func TestEvaluatorUserChains(testRun *testing.T) {
	acceptChain := domain.Chain{Name: testUserChain, Rules: []domain.Rule{portRule(domain.TCP, testSSHPort, domain.TargetAccept)}}
	returnChain := domain.Chain{Name: testUserChain, Rules: []domain.Rule{targetRule(domain.TargetReturn), targetRule(domain.TargetDrop)}}
	fallThroughChain := domain.Chain{Name: testUserChain, Rules: []domain.Rule{portRule(domain.TCP, testWebPort, domain.TargetAccept)}}
	restrictedChain := domain.Chain{Name: testUserChain, Rules: []domain.Rule{targetRule(domain.TargetAccept)}}
	restrictedJump := domain.Rule{SourceCIDRs: prefixes(testPrivateNetwork), JumpChain: testUserChain}
	cycleFirst := domain.Chain{Name: testUserChain, Rules: []domain.Rule{jumpRule(testSecondChain)}}
	cycleSecond := domain.Chain{Name: testSecondChain, Rules: []domain.Rule{jumpRule(testUserChain)}}
	selfLoop := domain.Chain{Name: testUserChain, Rules: []domain.Rule{jumpRule(testUserChain)}}
	gotoRule := domain.Rule{GotoChain: testUserChain}
	runEvaluationCases(testRun, []evaluationCase{
		{name: "accept inside a user chain", ruleSet: ipv4RuleSet(domain.TargetDrop, []domain.Rule{jumpRule(testUserChain)}, acceptChain), port: tcpPort(testSSHPort, testAnyBind), expected: domain.FirewallAllowed},
		{name: "RETURN goes back to the caller", ruleSet: ipv4RuleSet(domain.TargetAccept, []domain.Rule{jumpRule(testUserChain)}, returnChain), port: tcpPort(testSSHPort, testAnyBind), expected: domain.FirewallAllowed},
		{name: "chain without a match falls through to later rules", ruleSet: ipv4RuleSet(domain.TargetDrop, []domain.Rule{jumpRule(testUserChain), portRule(domain.TCP, testSSHPort, domain.TargetAccept)}, fallThroughChain), port: tcpPort(testSSHPort, testAnyBind), expected: domain.FirewallAllowed},
		{name: "accept in a chain reached from a restricted jump is restricted", ruleSet: ipv4RuleSet(domain.TargetDrop, []domain.Rule{restrictedJump}, restrictedChain), port: tcpPort(testSSHPort, testAnyBind), expected: domain.FirewallRestricted},
		{name: "goto never comes back to the calling chain", ruleSet: ipv4RuleSet(domain.TargetDrop, []domain.Rule{gotoRule, portRule(domain.TCP, testSSHPort, domain.TargetAccept)}, fallThroughChain), port: tcpPort(testSSHPort, testAnyBind), expected: domain.FirewallBlocked},
		{name: "two chain cycle is unknown", ruleSet: ipv4RuleSet(domain.TargetDrop, []domain.Rule{jumpRule(testUserChain)}, cycleFirst, cycleSecond), port: tcpPort(testSSHPort, testAnyBind), expected: domain.FirewallUnknown, reasonFragments: []string{"cycle"}},
		{name: "self loop is unknown", ruleSet: ipv4RuleSet(domain.TargetDrop, []domain.Rule{jumpRule(testUserChain)}, selfLoop), port: tcpPort(testSSHPort, testAnyBind), expected: domain.FirewallUnknown},
		{name: "chain nested beyond the depth limit is unknown", ruleSet: deepChainRuleSet(), port: tcpPort(testSSHPort, testAnyBind), expected: domain.FirewallUnknown, reasonFragments: []string{"too deep"}},
	})
}

// deepChainRuleSet nests user chains one level beyond the supported depth.
func deepChainRuleSet() domain.RuleSet {
	var chains []domain.Chain
	for level := 0; level <= maxChainDepth+1; level++ {
		name := domain.ChainName("deep-" + string(rune('a'+level)))
		next := domain.ChainName("deep-" + string(rune('a'+level+1)))
		rules := []domain.Rule{jumpRule(next)}
		if level == maxChainDepth+1 {
			rules = []domain.Rule{targetRule(domain.TargetAccept)}
		}
		chains = append(chains, domain.Chain{Name: name, Rules: rules})
	}
	return ipv4RuleSet(domain.TargetDrop, []domain.Rule{jumpRule("deep-a")}, chains...)
}

func TestEvaluatorPortSources(testRun *testing.T) {
	dockerPort := domain.Port{Number: 8080, Protocol: domain.TCP, BindAddress: testAnyBind, Source: domain.SourceDocker, Container: "web"}
	dockerLoopback := domain.Port{Number: 8081, Protocol: domain.TCP, BindAddress: testLoopbackBind, Source: domain.SourceDocker}
	runEvaluationCases(testRun, []evaluationCase{
		{name: "loopback bind is not exposed", ruleSet: ipv4RuleSet(domain.TargetDrop, nil), port: tcpPort(testSSHPort, testLoopbackBind), expected: domain.FirewallNotExposed},
		{name: "ipv6 loopback bind is not exposed", ruleSet: ipv4RuleSet(domain.TargetDrop, nil), port: tcpPort(testSSHPort, "::1"), expected: domain.FirewallNotExposed},
		{
			name: "docker published port bypasses INPUT", ruleSet: ipv4RuleSet(domain.TargetDrop, nil), port: dockerPort,
			expected: domain.FirewallDockerPublished, reasonFragments: []string{"published by Docker", "usually bypasses", "DOCKER-USER"},
		},
		{name: "docker port bound to loopback is not exposed", ruleSet: ipv4RuleSet(domain.TargetAccept, nil), port: dockerLoopback, expected: domain.FirewallNotExposed},
		{name: "host port next to docker ports uses INPUT", ruleSet: ipv4RuleSet(domain.TargetAccept, nil), port: tcpPort(testSSHPort, testAnyBind), expected: domain.FirewallAllowed},
	})
}

func TestEvaluatorAddressFamilies(testRun *testing.T) {
	allowSSH := []domain.Rule{portRule(domain.TCP, testSSHPort, domain.TargetAccept)}
	ipv4Open := filterFamily(domain.FamilyIPv4, domain.TargetDrop, allowSSH)
	ipv4Closed := filterFamily(domain.FamilyIPv4, domain.TargetDrop, nil)
	ipv6Open := filterFamily(domain.FamilyIPv6, domain.TargetDrop, allowSSH)
	ipv6Closed := filterFamily(domain.FamilyIPv6, domain.TargetDrop, nil)
	ipv6Missing := domain.FamilyRuleSet{Family: domain.FamilyIPv6}
	runEvaluationCases(testRun, []evaluationCase{
		{name: "dual-stack open on both is allowed", ruleSet: dualRuleSet(ipv4Open, ipv6Open), port: tcpPort(testSSHPort, testDualStackBind), expected: domain.FirewallAllowed},
		{name: "dual-stack open only on ipv4 is allowed", ruleSet: dualRuleSet(ipv4Open, ipv6Closed), port: tcpPort(testSSHPort, testDualStackBind), expected: domain.FirewallAllowed},
		{name: "dual-stack open only on ipv6 is allowed", ruleSet: dualRuleSet(ipv4Closed, ipv6Open), port: tcpPort(testSSHPort, testDualStackBind), expected: domain.FirewallAllowed},
		{name: "dual-stack closed on both is blocked", ruleSet: dualRuleSet(ipv4Closed, ipv6Closed), port: tcpPort(testSSHPort, testDualStackBind), expected: domain.FirewallBlocked},
		{
			name: "dual-stack blocked on ipv4 and ipv6 tool missing is unknown", ruleSet: dualRuleSet(ipv4Closed, ipv6Missing),
			port: tcpPort(testSSHPort, testDualStackBind), expected: domain.FirewallUnknown, reasonFragments: []string{"ip6tables"},
		},
		{name: "dual-stack allowed on ipv4 with ipv6 tool missing stays allowed", ruleSet: dualRuleSet(ipv4Open, ipv6Missing), port: tcpPort(testSSHPort, testDualStackBind), expected: domain.FirewallAllowed},
		{name: "specific ipv4 bind ignores ipv6 rules", ruleSet: dualRuleSet(ipv4Closed, ipv6Open), port: tcpPort(testSSHPort, testSpecificBind), expected: domain.FirewallBlocked},
		{name: "wildcard ipv4 bind ignores ipv6 rules", ruleSet: dualRuleSet(ipv4Closed, ipv6Open), port: tcpPort(testSSHPort, testAnyBind), expected: domain.FirewallBlocked},
		{name: "specific ipv6 bind uses only ipv6 rules", ruleSet: dualRuleSet(ipv4Open, ipv6Closed), port: tcpPort(testSSHPort, "2001:db8::10"), expected: domain.FirewallBlocked},
		{name: "specific ipv6 bind with tool missing is unknown", ruleSet: dualRuleSet(ipv4Open, ipv6Missing), port: tcpPort(testSSHPort, "2001:db8::10"), expected: domain.FirewallUnknown},
		{name: "ipv4-mapped bind uses ipv4 rules", ruleSet: dualRuleSet(ipv4Open, ipv6Closed), port: tcpPort(testSSHPort, "::ffff:192.168.1.10"), expected: domain.FirewallAllowed},
		{name: "empty bind is treated as both families", ruleSet: dualRuleSet(ipv4Closed, ipv6Missing), port: tcpPort(testSSHPort, ""), expected: domain.FirewallUnknown},
	})
}

func TestCombineVerdicts(testRun *testing.T) {
	allowed, restricted, blocked, unknown := domain.FirewallAllowed, domain.FirewallRestricted, domain.FirewallBlocked, domain.FirewallUnknown
	testCases := []struct {
		first, second, expected domain.FirewallVerdict
	}{
		{allowed, allowed, allowed},
		{allowed, restricted, allowed},
		{allowed, blocked, allowed},
		{allowed, unknown, allowed},
		{restricted, restricted, restricted},
		{restricted, blocked, restricted},
		{restricted, unknown, unknown},
		{blocked, blocked, blocked},
		{blocked, unknown, unknown},
		{unknown, unknown, unknown},
		{domain.FirewallNotExposed, blocked, unknown},
	}
	for _, testCase := range testCases {
		if combineVerdicts(testCase.first, testCase.second) != testCase.expected {
			testRun.Errorf("combine(%v,%v) = %v, want %v", testCase.first, testCase.second, combineVerdicts(testCase.first, testCase.second), testCase.expected)
		}
		if combineVerdicts(testCase.second, testCase.first) != testCase.expected {
			testRun.Errorf("combine must be symmetric for (%v,%v)", testCase.first, testCase.second)
		}
	}
}

func TestEvaluatorGlobalCaveats(testRun *testing.T) {
	rawDrop := domain.Table{Name: domain.TableRaw, Chains: []domain.Chain{{Name: domain.ChainPrerouting, Policy: domain.TargetAccept, Rules: []domain.Rule{{SourceCIDRs: prefixes("192.0.2.66/32"), Target: domain.TargetDrop}}}}}
	rawAccept := domain.Table{Name: domain.TableRaw, Chains: []domain.Chain{{Name: domain.ChainPrerouting, Policy: domain.TargetAccept, Rules: []domain.Rule{{Target: domain.TargetAccept}}}}}
	mangleDrop := domain.Table{Name: domain.TableMangle, Chains: []domain.Chain{{Name: domain.ChainInput, Policy: domain.TargetAccept, Rules: []domain.Rule{portRule(domain.TCP, testSSHPort, domain.TargetDrop)}}}}
	mangleForwardDrop := domain.Table{Name: domain.TableMangle, Chains: []domain.Chain{{Name: domain.ChainForward, Policy: domain.TargetAccept, Rules: []domain.Rule{targetRule(domain.TargetDrop)}}}}
	natRedirect := domain.Table{Name: domain.TableNat, Chains: []domain.Chain{{Name: domain.ChainPrerouting, Policy: domain.TargetAccept, Rules: []domain.Rule{{Protocol: domain.TCP, DestinationPorts: []domain.PortRange{{First: testWebPort, Last: testWebPort}}, Target: domain.TargetRedirect}}}}}
	dockerNat := domain.Table{Name: domain.TableNat, Chains: []domain.Chain{
		{Name: domain.ChainPrerouting, Policy: domain.TargetAccept, Rules: []domain.Rule{jumpRule("DOCKER")}},
		{Name: "DOCKER", Rules: []domain.Rule{{Protocol: domain.TCP, DestinationPorts: []domain.PortRange{{First: 8080, Last: 8080}}, Target: domain.TargetDnat}}},
	}}
	allowedBase := func(extra domain.Table) domain.RuleSet {
		ruleSet := ipv4RuleSet(domain.TargetAccept, nil)
		ruleSet.IPv4 = withTable(ruleSet.IPv4, extra)
		return ruleSet
	}
	blockedBase := func(extra domain.Table) domain.RuleSet {
		ruleSet := ipv4RuleSet(domain.TargetDrop, nil)
		ruleSet.IPv4 = withTable(ruleSet.IPv4, extra)
		return ruleSet
	}
	noFilter := domain.RuleSet{IPv4: domain.FamilyRuleSet{Family: domain.FamilyIPv4, Available: true, Tables: []domain.Table{rawAccept}}}
	noTables := domain.RuleSet{IPv4: domain.FamilyRuleSet{Family: domain.FamilyIPv4, Available: true}}
	noInputChain := domain.RuleSet{IPv4: domain.FamilyRuleSet{Family: domain.FamilyIPv4, Available: true, Tables: []domain.Table{{Name: domain.TableFilter}}}}
	port := tcpPort(testSSHPort, testAnyBind)
	runEvaluationCases(testRun, []evaluationCase{
		{name: "raw drop downgrades allowed", ruleSet: allowedBase(rawDrop), port: port, expected: domain.FirewallUnknown, reasonFragments: []string{"raw"}},
		{name: "raw drop downgrades blocked", ruleSet: blockedBase(rawDrop), port: port, expected: domain.FirewallUnknown},
		{name: "raw accept keeps allowed", ruleSet: allowedBase(rawAccept), port: port, expected: domain.FirewallAllowed},
		{name: "mangle INPUT drop downgrades allowed", ruleSet: allowedBase(mangleDrop), port: port, expected: domain.FirewallUnknown, reasonFragments: []string{"mangle"}},
		{name: "mangle FORWARD drop is irrelevant", ruleSet: allowedBase(mangleForwardDrop), port: port, expected: domain.FirewallAllowed},
		{name: "nat redirect outside docker downgrades blocked", ruleSet: blockedBase(natRedirect), port: port, expected: domain.FirewallUnknown, reasonFragments: []string{"nat"}},
		{name: "docker nat chain is not a caveat", ruleSet: allowedBase(dockerNat), port: port, expected: domain.FirewallAllowed},
		{name: "missing filter table is unknown", ruleSet: noFilter, port: port, expected: domain.FirewallUnknown, reasonFragments: []string{"no filter table"}},
		{name: "no tables at all is unknown", ruleSet: noTables, port: port, expected: domain.FirewallUnknown, reasonFragments: []string{"native nftables rules are not evaluated yet"}},
		{name: "missing INPUT chain is unknown", ruleSet: noInputChain, port: port, expected: domain.FirewallUnknown},
	})
}

func TestEvaluatorRestrictedIsNotDowngradedByCaveats(testRun *testing.T) {
	rawDrop := domain.Table{Name: domain.TableRaw, Chains: []domain.Chain{{Name: domain.ChainPrerouting, Rules: []domain.Rule{targetRule(domain.TargetDrop)}}}}
	ruleSet := ipv4RuleSet(domain.TargetDrop, []domain.Rule{{SourceCIDRs: prefixes(testPrivateNetwork), Target: domain.TargetAccept}})
	ruleSet.IPv4 = withTable(ruleSet.IPv4, rawDrop)
	runEvaluationCases(testRun, []evaluationCase{
		{name: "restricted stays restricted", ruleSet: ruleSet, port: tcpPort(testSSHPort, testAnyBind), expected: domain.FirewallRestricted},
	})
}
