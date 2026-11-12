package app

import (
	"fmt"
	"slices"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

// preFilterCaveat looks for rules outside the filter table that can drop,
// divert or rewrite inbound traffic before the filter INPUT chain sees it.
// It returns a reason, or "" when none were found.
func preFilterCaveat(family domain.FamilyRuleSet) string {
	for _, table := range family.Tables {
		if table.Name == domain.TableRaw || table.Name == domain.TableMangle {
			if hasRuleWithTargets(table, preFilterTargets, false) {
				return fmt.Sprintf(reasonPreFilterFormat, table.Name)
			}
		}
		if table.Name == domain.TableNat && hasRuleWithTargets(table, rewritingTargets, true) {
			return fmt.Sprintf(reasonPreFilterFormat, table.Name)
		}
	}
	return ""
}

// hasRuleWithTargets scans the chains that can see inbound traffic. DNAT
// rules inside the chains Docker itself creates are skipped when
// exemptDockerDNAT is set, because their rewrites lead to the FORWARD path and
// are reported through the Docker verdict. Nothing else is exempt.
func hasRuleWithTargets(table domain.Table, targets []domain.TargetName, exemptDockerDNAT bool) bool {
	for _, chain := range table.Chains {
		if slices.Contains(nonInboundChains, chain.Name) {
			continue
		}
		isDockerChain := slices.Contains(dockerNatChains, chain.Name)
		for _, rule := range chain.Rules {
			if exemptDockerDNAT && isDockerChain && rule.Target == domain.TargetDnat {
				continue
			}
			if slices.Contains(targets, rule.Target) {
				return true
			}
		}
	}
	return false
}

// applyCaveat downgrades Allowed and Blocked to Unknown when a pre-filter
// caveat exists. Restricted and Unknown are left alone.
func applyCaveat(verdict domain.FirewallVerdict, reason string, caveat string) (domain.FirewallVerdict, string) {
	if caveat == "" {
		return verdict, reason
	}
	if verdict == domain.FirewallAllowed || verdict == domain.FirewallBlocked {
		return domain.FirewallUnknown, caveat
	}
	return verdict, reason
}
