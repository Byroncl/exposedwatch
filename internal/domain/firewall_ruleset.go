package domain

import "net/netip"

// Chain is an ordered list of rules with an optional policy (built-in chains).
type Chain struct {
	Name   ChainName
	Policy TargetName
	Rules  []Rule
}

// Table is a set of chains.
type Table struct {
	Name   TableName
	Chains []Chain
}

// Chain finds a chain by name.
func (table Table) Chain(name ChainName) (Chain, bool) {
	for _, chain := range table.Chains {
		if chain.Name == name {
			return chain, true
		}
	}
	return Chain{}, false
}

// FamilyRuleSet holds the tables of one address family. Available is false
// when the tool for the family could not be run.
type FamilyRuleSet struct {
	Family    AddressFamily
	Available bool
	Tables    []Table
}

// Table finds a table by name.
func (family FamilyRuleSet) Table(name TableName) (Table, bool) {
	for _, table := range family.Tables {
		if table.Name == name {
			return table, true
		}
	}
	return Table{}, false
}

// RuleCount returns the number of rules across all tables.
func (family FamilyRuleSet) RuleCount() int {
	total := 0
	for _, table := range family.Tables {
		for _, chain := range table.Chains {
			total += len(chain.Rules)
		}
	}
	return total
}

// RuleSet is everything read from the iptables family of tools.
type RuleSet struct {
	// Backend names the tool and version, or BackendUnknown.
	Backend string
	IPv4    FamilyRuleSet
	IPv6    FamilyRuleSet
	// Notes are caveats found while reading the rules.
	Notes []string
	// Nftables is the result of the native nftables probe.
	Nftables NftablesState
}

// Family returns the rule set of one address family.
func (ruleSet RuleSet) Family(family AddressFamily) FamilyRuleSet {
	if family == FamilyIPv6 {
		return ruleSet.IPv6
	}
	return ruleSet.IPv4
}

// HasTables reports whether any family contains at least one table.
func (ruleSet RuleSet) HasTables() bool {
	return len(ruleSet.IPv4.Tables) > 0 || len(ruleSet.IPv6.Tables) > 0
}

// IsLoopbackPrefix reports whether prefix lies wholly inside loopback space.
func IsLoopbackPrefix(prefix netip.Prefix) bool {
	if !prefix.Addr().IsLoopback() {
		return false
	}
	if prefix.Addr().Is4() {
		return prefix.Bits() >= loopbackMinimumBitsIPv4
	}
	return prefix.Bits() >= loopbackMinimumBitsIPv6
}
