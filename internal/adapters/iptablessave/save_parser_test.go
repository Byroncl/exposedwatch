package iptablessave

import (
	"context"
	"errors"
	"net/netip"
	"strings"
	"testing"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

func parseFixture(testRun *testing.T, fileName string) parsedOutput {
	testRun.Helper()
	parsed, err := parseSaveOutput(context.Background(), readFixture(testRun, fileName))
	if err != nil {
		testRun.Fatalf("%s: %v", fileName, err)
	}
	return parsed
}

func findChain(testRun *testing.T, parsed parsedOutput, tableName domain.TableName, chainName domain.ChainName) domain.Chain {
	testRun.Helper()
	for _, table := range parsed.tables {
		if table.Name != tableName {
			continue
		}
		if chain, found := table.Chain(chainName); found {
			return chain
		}
	}
	testRun.Fatalf("chain %s/%s not found", tableName, chainName)
	return domain.Chain{}
}

func TestParseNoRulesFixture(testRun *testing.T) {
	parsed := parseFixture(testRun, "a_no_rules.txt")
	if len(parsed.tables) != 1 || parsed.tables[0].Name != domain.TableFilter {
		testRun.Fatalf("tables = %+v", parsed.tables)
	}
	input := findChain(testRun, parsed, domain.TableFilter, domain.ChainInput)
	if input.Policy != domain.TargetAccept || len(input.Rules) != 0 {
		testRun.Fatalf("input = %+v", input)
	}
	if parsed.version != "v1.8.11" {
		testRun.Fatalf("version = %q", parsed.version)
	}
}

func TestParseUfwFixture(testRun *testing.T) {
	parsed := parseFixture(testRun, "b_ufw.txt")
	input := findChain(testRun, parsed, domain.TableFilter, domain.ChainInput)
	if input.Policy != domain.TargetDrop || len(input.Rules) != 6 {
		testRun.Fatalf("input policy=%q rules=%d", input.Policy, len(input.Rules))
	}
	if input.Rules[0].JumpChain != "ufw-before-logging-input" || input.Rules[0].Target != "" {
		testRun.Fatalf("first INPUT rule = %+v", input.Rules[0])
	}
	userChain := findChain(testRun, parsed, domain.TableFilter, "ufw-user-input")
	if userChain.Policy != "" {
		testRun.Fatalf("user chain policy = %q", userChain.Policy)
	}
	ssh := userChain.Rules[0]
	if ssh.Protocol != domain.TCP || len(ssh.DestinationPorts) != 1 || ssh.DestinationPorts[0] != (domain.PortRange{First: 22, Last: 22}) || ssh.Target != domain.TargetAccept || ssh.Unsupported {
		testRun.Fatalf("ssh rule = %+v", ssh)
	}
	web := userChain.Rules[1]
	if len(web.DestinationPorts) != 2 || web.DestinationPorts[1].First != 443 || web.Unsupported {
		testRun.Fatalf("multiport rule = %+v", web)
	}
	before := findChain(testRun, parsed, domain.TableFilter, "ufw-before-input")
	if before.Rules[0].InInterface != "lo" || before.Rules[0].Target != domain.TargetAccept {
		testRun.Fatalf("loopback rule = %+v", before.Rules[0])
	}
	established := before.Rules[1]
	if len(established.ConnectionStates) != 2 || established.ConnectionStates[1] != domain.StateEstablished {
		testRun.Fatalf("conntrack rule = %+v", established)
	}
	notLocal := findChain(testRun, parsed, domain.TableFilter, "ufw-not-local")
	if len(notLocal.Rules[1].AddressTypes) != 1 || notLocal.Rules[1].AddressTypes[0] != domain.AddressTypeMulticast || notLocal.Rules[1].Target != domain.TargetReturn {
		testRun.Fatalf("addrtype rule = %+v", notLocal.Rules[1])
	}
	limitAccept := userChain.Rules[2]
	if !limitAccept.Unsupported || limitAccept.UnsupportedReason != "module: limit" {
		testRun.Fatalf("limit rule = %+v", limitAccept)
	}
	if !before.Rules[6].Unsupported || !strings.Contains(before.Rules[6].UnsupportedReason, "sport") {
		testRun.Fatalf("sport rule = %+v", before.Rules[6])
	}
	multicast := before.Rules[8]
	if len(multicast.DestinationCIDRs) != 1 || multicast.DestinationCIDRs[0] != netip.MustParsePrefix("224.0.0.251/32") {
		testRun.Fatalf("destination rule = %+v", multicast)
	}
}

func TestParseDockerFixture(testRun *testing.T) {
	parsed := parseFixture(testRun, "c_docker.txt")
	if len(parsed.tables) != 2 {
		testRun.Fatalf("tables = %d", len(parsed.tables))
	}
	input := findChain(testRun, parsed, domain.TableFilter, domain.ChainInput)
	forward := findChain(testRun, parsed, domain.TableFilter, domain.ChainForward)
	if input.Policy != domain.TargetAccept || forward.Policy != domain.TargetDrop || len(forward.Rules) != 6 {
		testRun.Fatalf("input=%+v forward policy=%q rules=%d", input, forward.Policy, len(forward.Rules))
	}
	docker := findChain(testRun, parsed, domain.TableNat, "DOCKER")
	dnat := docker.Rules[1]
	if dnat.Target != domain.TargetDnat || len(dnat.DestinationPorts) != 1 || dnat.DestinationPorts[0].First != 8080 {
		testRun.Fatalf("dnat rule = %+v", dnat)
	}
	if !dnat.Unsupported || !strings.Contains(dnat.UnsupportedReason, "negated") {
		testRun.Fatalf("negated interface should be unsupported: %+v", dnat)
	}
	prerouting := findChain(testRun, parsed, domain.TableNat, domain.ChainPrerouting)
	if prerouting.Rules[0].JumpChain != "DOCKER" {
		testRun.Fatalf("prerouting = %+v", prerouting.Rules[0])
	}
}

func TestParseSourceRestrictedFixture(testRun *testing.T) {
	parsed := parseFixture(testRun, "d_source_restricted.txt")
	input := findChain(testRun, parsed, domain.TableFilter, domain.ChainInput)
	rule := input.Rules[2]
	if len(rule.SourceCIDRs) != 1 || rule.SourceCIDRs[0] != netip.MustParsePrefix("10.0.0.0/8") || rule.DestinationPorts[0].First != 5432 {
		testRun.Fatalf("restricted rule = %+v", rule)
	}
}

func TestParseUnsupportedMatchesFixture(testRun *testing.T) {
	parsed := parseFixture(testRun, "e_unsupported_matches.txt")
	input := findChain(testRun, parsed, domain.TableFilter, domain.ChainInput)
	if !input.Rules[0].Unsupported || input.Rules[0].UnsupportedReason != "module: set" {
		testRun.Fatalf("set rule = %+v", input.Rules[0])
	}
	if !input.Rules[1].Unsupported || input.Rules[1].UnsupportedReason != "module: recent" {
		testRun.Fatalf("recent rule = %+v", input.Rules[1])
	}
	if input.Rules[2].Unsupported {
		testRun.Fatalf("plain rule flagged unsupported: %+v", input.Rules[2])
	}
	if input.Rules[0].Target != domain.TargetAccept || input.Rules[0].DestinationPorts[0].First != 3306 {
		testRun.Fatalf("supported parts must still be parsed: %+v", input.Rules[0])
	}
}

func TestParseIPv6Fixture(testRun *testing.T) {
	parsed := parseFixture(testRun, "f_ipv6.txt")
	input := findChain(testRun, parsed, domain.TableFilter, domain.ChainInput)
	if len(input.Rules) != 5 || input.Rules[2].SourceCIDRs[0] != netip.MustParsePrefix("2001:db8::/32") || input.Rules[3].Protocol != "ipv6-icmp" {
		testRun.Fatalf("input = %+v", input)
	}
	if parsed.version != "v1.8.11" {
		testRun.Fatalf("version = %q", parsed.version)
	}
}

func TestParseChainCycleFixture(testRun *testing.T) {
	parsed := parseFixture(testRun, "g_chain_cycle.txt")
	first := findChain(testRun, parsed, domain.TableFilter, "first-chain")
	if first.Rules[0].JumpChain != "second-chain" {
		testRun.Fatalf("first-chain = %+v", first.Rules)
	}
}

func TestParseRawMangleFixture(testRun *testing.T) {
	parsed := parseFixture(testRun, "h_raw_mangle_drop.txt")
	if len(parsed.tables) != 3 {
		testRun.Fatalf("tables = %d", len(parsed.tables))
	}
	raw := findChain(testRun, parsed, domain.TableRaw, domain.ChainPrerouting)
	mangle := findChain(testRun, parsed, domain.TableMangle, domain.ChainInput)
	if raw.Rules[0].Target != domain.TargetDrop || mangle.Rules[0].Target != domain.TargetDrop {
		testRun.Fatalf("raw=%+v mangle=%+v", raw.Rules, mangle.Rules)
	}
}

func TestParseQuotingNegationAndFlags(testRun *testing.T) {
	parsed := parseFixture(testRun, "j_quoting.txt")
	input := findChain(testRun, parsed, domain.TableFilter, domain.ChainInput)
	if input.Policy != domain.TargetDrop {
		testRun.Fatalf("policy = %q", input.Policy)
	}
	rules := input.Rules
	if len(rules) != 10 {
		testRun.Fatalf("rules = %d", len(rules))
	}
	if rules[0].DestinationPorts[0] != (domain.PortRange{First: 8000, Last: 8010}) || rules[0].Unsupported || rules[0].Target != domain.TargetAccept {
		testRun.Fatalf("range/comment rule = %+v", rules[0])
	}
	if !strings.Contains(rules[0].Text, `"dev servers \"alpha\" beta"`) {
		testRun.Fatalf("rule text lost: %q", rules[0].Text)
	}
	if !rules[1].Unsupported || !strings.Contains(rules[1].UnsupportedReason, "negated") {
		testRun.Fatalf("negated dport = %+v", rules[1])
	}
	if !rules[2].Unsupported || !strings.Contains(rules[2].UnsupportedReason, "negated") {
		testRun.Fatalf("negated source = %+v", rules[2])
	}
	if rules[3].TCPFlags == nil || rules[3].Unsupported {
		testRun.Fatalf("--syn rule = %+v", rules[3])
	}
	if rules[4].TCPFlags == nil || len(rules[4].TCPFlags.Mask) != 4 || rules[4].TCPFlags.Set[0] != domain.TCPFlagSyn {
		testRun.Fatalf("tcp-flags rule = %+v", rules[4])
	}
	if !rules[5].Unsupported || !strings.Contains(rules[5].UnsupportedReason, "sport") {
		testRun.Fatalf("sport rule = %+v", rules[5])
	}
	if rules[6].JumpChain != "audit-chain" || len(rules[6].DestinationPorts) != 2 || rules[6].DestinationPorts[1] != (domain.PortRange{First: 200, Last: 210}) {
		testRun.Fatalf("jump rule = %+v", rules[6])
	}
	if rules[7].GotoChain != "audit-chain" {
		testRun.Fatalf("goto rule = %+v", rules[7])
	}
	if !rules[8].Unsupported || rules[8].UnsupportedReason != "module: frobnicate" {
		testRun.Fatalf("unknown module = %+v", rules[8])
	}
	if rules[9].Unsupported || rules[9].ConnectionStates[0] != domain.StateNew {
		testRun.Fatalf("legacy state rule = %+v", rules[9])
	}
	audit := findChain(testRun, parsed, domain.TableFilter, "audit-chain")
	if audit.Rules[0].Target != domain.TargetLog || audit.Rules[1].Target != domain.TargetReturn {
		testRun.Fatalf("audit chain = %+v", audit.Rules)
	}
	if len(parsed.notes) != 1 || !strings.Contains(parsed.notes[0], "legacy") {
		testRun.Fatalf("notes = %v", parsed.notes)
	}
}

func TestParseRejectsMalformedInput(testRun *testing.T) {
	malformed := map[string]string{
		"unterminated quote": "*filter\n:INPUT ACCEPT [0:0]\n-A INPUT -m comment --comment \"open\nCOMMIT\n",
		"rule before table":  "-A INPUT -j ACCEPT\nCOMMIT\n",
		"unknown line":       "*filter\nthis is not iptables\nCOMMIT\n",
		"bad policy line":    "*filter\n:INPUT\nCOMMIT\n",
		"no commit":          "*filter\n:INPUT ACCEPT [0:0]\n-A INPUT -p tcp -m tcp --dport 80 -j ACCEPT\n",
	}
	for name, content := range malformed {
		_, err := parseSaveOutput(context.Background(), []byte(content))
		if !errors.Is(err, ErrInvalidOutput) {
			testRun.Errorf("%s: err = %v", name, err)
		}
	}
}

func TestParseEmptyOutputHasNoTables(testRun *testing.T) {
	parsed, err := parseSaveOutput(context.Background(), []byte("# Generated by iptables-save v1.8.11 on Tue Nov 11 09:00:00 2026\n"))
	if err != nil || len(parsed.tables) != 0 {
		testRun.Fatalf("parsed=%+v err=%v", parsed, err)
	}
}
