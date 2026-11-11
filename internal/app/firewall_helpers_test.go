package app

import (
	"net/netip"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

const (
	testSSHPort        = 22
	testWebPort        = 80
	testDatabasePort   = 5432
	testUnlistedPort   = 9999
	testAnyBind        = "0.0.0.0"
	testDualStackBind  = "::"
	testLoopbackBind   = "127.0.0.1"
	testSpecificBind   = "192.168.1.10"
	testOtherBind      = "192.168.1.20"
	testPrivateNetwork = "10.0.0.0/8"
	testUserChain      = domain.ChainName("user-chain")
	testSecondChain    = domain.ChainName("second-chain")
)

func tcpPort(number int, bind string) domain.Port {
	return domain.Port{Number: number, Protocol: domain.TCP, BindAddress: bind, Source: domain.SourceHost}
}

func portRule(protocol domain.Protocol, number int, target domain.TargetName) domain.Rule {
	return domain.Rule{
		Text:             "-A INPUT -p " + string(protocol) + " --dport",
		Protocol:         protocol,
		DestinationPorts: []domain.PortRange{{First: number, Last: number}},
		Target:           target,
	}
}

func targetRule(target domain.TargetName) domain.Rule {
	return domain.Rule{Text: "-A INPUT -j " + string(target), Target: target}
}

func jumpRule(chain domain.ChainName) domain.Rule {
	return domain.Rule{Text: "-A INPUT -j " + string(chain), JumpChain: chain}
}

func unsupportedRule(reason string, target domain.TargetName) domain.Rule {
	return domain.Rule{Text: "-A INPUT unsupported", Unsupported: true, UnsupportedReason: reason, Target: target}
}

func filterFamily(family domain.AddressFamily, policy domain.TargetName, inputRules []domain.Rule, extraChains ...domain.Chain) domain.FamilyRuleSet {
	chains := []domain.Chain{{Name: domain.ChainInput, Policy: policy, Rules: inputRules}}
	chains = append(chains, extraChains...)
	return domain.FamilyRuleSet{
		Family:    family,
		Available: true,
		Tables:    []domain.Table{{Name: domain.TableFilter, Chains: chains}},
	}
}

// ipv4RuleSet builds a rule set whose IPv6 half is unavailable.
func ipv4RuleSet(policy domain.TargetName, inputRules []domain.Rule, extraChains ...domain.Chain) domain.RuleSet {
	return domain.RuleSet{
		Backend: "iptables v1.8.11",
		IPv4:    filterFamily(domain.FamilyIPv4, policy, inputRules, extraChains...),
		IPv6:    domain.FamilyRuleSet{Family: domain.FamilyIPv6},
	}
}

func dualRuleSet(ipv4 domain.FamilyRuleSet, ipv6 domain.FamilyRuleSet) domain.RuleSet {
	return domain.RuleSet{Backend: "iptables v1.8.11", IPv4: ipv4, IPv6: ipv6}
}

func prefixes(texts ...string) []netip.Prefix {
	var result []netip.Prefix
	for _, text := range texts {
		result = append(result, netip.MustParsePrefix(text))
	}
	return result
}

func withTable(family domain.FamilyRuleSet, table domain.Table) domain.FamilyRuleSet {
	family.Tables = append(family.Tables, table)
	return family
}
