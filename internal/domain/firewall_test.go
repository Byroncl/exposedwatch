package domain

import (
	"net/netip"
	"testing"
)

func TestFirewallVerdictString(testRun *testing.T) {
	testCases := map[FirewallVerdict]string{
		FirewallAllowed:         "allowed",
		FirewallRestricted:      "restricted",
		FirewallBlocked:         "blocked",
		FirewallUnknown:         "unknown",
		FirewallNotExposed:      "not-exposed",
		FirewallDockerPublished: "docker-published",
		FirewallVerdict(99):     "unknown",
	}
	for verdict, expected := range testCases {
		if verdict.String() != expected {
			testRun.Errorf("verdict %d = %q, want %q", int(verdict), verdict.String(), expected)
		}
	}
}

func TestFirewallVerdictZeroValueIsUnknown(testRun *testing.T) {
	var verdict FirewallVerdict
	if verdict != FirewallUnknown {
		testRun.Fatalf("zero verdict = %v", verdict)
	}
}

func TestPortRangeContains(testRun *testing.T) {
	portRange := PortRange{First: 8000, Last: 8010}
	testCases := []struct {
		port     int
		expected bool
	}{{7999, false}, {8000, true}, {8005, true}, {8010, true}, {8011, false}}
	for _, testCase := range testCases {
		if portRange.Contains(testCase.port) != testCase.expected {
			testRun.Errorf("Contains(%d) != %v", testCase.port, testCase.expected)
		}
	}
}

func TestRuleSetFamilyAndLookups(testRun *testing.T) {
	filterTable := Table{
		Name: TableFilter,
		Chains: []Chain{
			{Name: ChainInput, Policy: TargetDrop, Rules: []Rule{{Target: TargetAccept}}},
			{Name: "custom"},
		},
	}
	ruleSet := RuleSet{
		IPv4: FamilyRuleSet{Family: FamilyIPv4, Available: true, Tables: []Table{filterTable}},
		IPv6: FamilyRuleSet{Family: FamilyIPv6},
	}
	if ruleSet.Family(FamilyIPv4).Family != FamilyIPv4 || ruleSet.Family(FamilyIPv6).Available {
		testRun.Fatal("Family lookup returned the wrong set")
	}
	table, found := ruleSet.IPv4.Table(TableFilter)
	if !found || table.Name != TableFilter {
		testRun.Fatal("filter table not found")
	}
	if _, found := ruleSet.IPv4.Table(TableRaw); found {
		testRun.Fatal("raw table should be missing")
	}
	chain, found := table.Chain(ChainInput)
	if !found || chain.Policy != TargetDrop {
		testRun.Fatal("INPUT chain not found")
	}
	if _, found := table.Chain("missing"); found {
		testRun.Fatal("missing chain reported as found")
	}
	if ruleSet.IPv4.RuleCount() != 1 || ruleSet.IPv6.RuleCount() != 0 {
		testRun.Fatalf("rule counts = %d/%d", ruleSet.IPv4.RuleCount(), ruleSet.IPv6.RuleCount())
	}
	if ruleSet.HasTables() != true || (RuleSet{}).HasTables() {
		testRun.Fatal("HasTables is wrong")
	}
}

func TestRuleCoversPrefixHelpers(testRun *testing.T) {
	if !IsLoopbackPrefix(netip.MustParsePrefix("127.0.0.0/8")) || !IsLoopbackPrefix(netip.MustParsePrefix("::1/128")) {
		testRun.Error("loopback prefixes not recognised")
	}
	if IsLoopbackPrefix(netip.MustParsePrefix("0.0.0.0/0")) || IsLoopbackPrefix(netip.MustParsePrefix("10.0.0.0/8")) {
		testRun.Error("non-loopback prefix flagged as loopback")
	}
}
