package iptablessave

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Byroncl/exposedwatch/internal/domain"
	"github.com/Byroncl/exposedwatch/internal/ports"
)

const (
	generatedChainCount = 20000
	parseTimeBudget     = time.Second
)

func TestParseRejectsTruncatedOutput(testRun *testing.T) {
	_, err := parseSaveOutput(context.Background(), readFixture(testRun, "m_truncated_mid_rule.txt"))
	if !errors.Is(err, ErrInvalidOutput) {
		testRun.Fatalf("err=%v", err)
	}
	unterminated := "*filter\n:INPUT ACCEPT [0:0]\n-A INPUT -p tcp -m tcp --dport 80 -j ACCEPT\n"
	if _, err := parseSaveOutput(context.Background(), []byte(unterminated)); !errors.Is(err, ErrInvalidOutput) {
		testRun.Fatalf("unterminated table accepted: err=%v", err)
	}
}

func TestLoaderTruncatedOutputIsReported(testRun *testing.T) {
	locator := fixtureLocator(testRun, false)
	ipv4Path, _ := locator.Find(ipv4SaveBinaryName)
	runner := &fakeRunner{results: map[string]CommandResult{
		ipv4Path: {Stdout: readFixture(testRun, "a_no_rules.txt"), Truncated: true},
	}}
	_, err := NewRuleSetLoader(runner, locator).LoadRuleSet(context.Background())
	if !errors.Is(err, ports.ErrFirewallOutputTruncated) {
		testRun.Fatalf("err=%v", err)
	}
}

func TestLoaderInvalidOutputMapsToPortSentinel(testRun *testing.T) {
	locator := fixtureLocator(testRun, false)
	ipv4Path, _ := locator.Find(ipv4SaveBinaryName)
	runner := &fakeRunner{results: map[string]CommandResult{ipv4Path: {Stdout: readFixture(testRun, "m_truncated_mid_rule.txt")}}}
	_, err := NewRuleSetLoader(runner, locator).LoadRuleSet(context.Background())
	if !errors.Is(err, ports.ErrFirewallOutputInvalid) || !errors.Is(err, ErrInvalidOutput) {
		testRun.Fatalf("err=%v", err)
	}
}

func TestParsePortRangesStartingAtZero(testRun *testing.T) {
	parsed := parseFixture(testRun, "n_port_zero_range.txt")
	input := findChain(testRun, parsed, domain.TableFilter, domain.ChainInput)
	if input.Rules[0].Unsupported || input.Rules[0].DestinationPorts[0] != (domain.PortRange{First: 0, Last: 1023}) {
		testRun.Fatalf("rule = %+v", input.Rules[0])
	}
	if input.Rules[1].DestinationPorts[0] != (domain.PortRange{First: 8000, Last: 65535}) {
		testRun.Fatalf("rule = %+v", input.Rules[1])
	}
}

func TestParseOddSyntaxDegradesToUnsupportedRules(testRun *testing.T) {
	parsed := parseFixture(testRun, "o_odd_syntax.txt")
	rules := findChain(testRun, parsed, domain.TableFilter, domain.ChainInput).Rules
	if len(rules) != 5 {
		testRun.Fatalf("rules = %d", len(rules))
	}
	if !rules[0].Unsupported || rules[0].Protocol != "" || !strings.Contains(rules[0].UnsupportedReason, "negated") {
		testRun.Fatalf("legacy negated protocol = %+v", rules[0])
	}
	if rules[1].Unsupported || rules[1].Protocol != domain.TCP {
		testRun.Fatalf("numeric tcp = %+v", rules[1])
	}
	if !rules[2].Unsupported || !strings.Contains(rules[2].UnsupportedReason, "protocol") {
		testRun.Fatalf("numeric 47 = %+v", rules[2])
	}
	if !rules[3].Unsupported || !strings.Contains(rules[3].UnsupportedReason, "dport") {
		testRun.Fatalf("dport + dports = %+v", rules[3])
	}
	if !rules[4].Unsupported || rules[4].Target != domain.TargetUnparseable || rules[4].Text == "" {
		testRun.Fatalf("unparseable rule = %+v", rules[4])
	}
}

func TestParseNumericProtocols(testRun *testing.T) {
	testCases := map[string]domain.Protocol{"6": domain.TCP, "17": domain.UDP}
	for number, expected := range testCases {
		content := "*filter\n:INPUT DROP [0:0]\n-A INPUT -p " + number + " -j ACCEPT\nCOMMIT\n"
		parsed, err := parseSaveOutput(context.Background(), []byte(content))
		if err != nil {
			testRun.Fatal(err)
		}
		if rule := parsed.tables[0].Chains[0].Rules[0]; rule.Protocol != expected || rule.Unsupported {
			testRun.Errorf("-p %s = %+v", number, rule)
		}
	}
}

func generatedChainInput() []byte {
	var builder strings.Builder
	builder.WriteString("*filter\n:INPUT ACCEPT [0:0]\n")
	for index := 0; index < generatedChainCount; index++ {
		fmt.Fprintf(&builder, ":chain-%d - [0:0]\n", index)
	}
	for index := 0; index < generatedChainCount; index++ {
		fmt.Fprintf(&builder, "-A INPUT -j chain-%d\n-A chain-%d -p tcp -m tcp --dport 80 -j ACCEPT\n", index, index)
	}
	builder.WriteString("COMMIT\n")
	return []byte(builder.String())
}

func TestParseLargeInputIsFast(testRun *testing.T) {
	input := generatedChainInput()
	started := time.Now()
	parsed, err := parseSaveOutput(context.Background(), input)
	if err != nil {
		testRun.Fatal(err)
	}
	if elapsed := time.Since(started); elapsed > parseTimeBudget {
		testRun.Fatalf("parsing %d chains took %v", generatedChainCount, elapsed)
	}
	if len(parsed.tables[0].Chains) != generatedChainCount+1 {
		testRun.Fatalf("chains = %d", len(parsed.tables[0].Chains))
	}
	if parsed.tables[0].Chains[1].Rules[0].JumpChain != "" || parsed.tables[0].Chains[0].Rules[0].JumpChain != "chain-0" {
		testRun.Fatal("jumps were not resolved")
	}
}

func TestParseHonoursCancellation(testRun *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := parseSaveOutput(ctx, generatedChainInput()); !errors.Is(err, context.Canceled) {
		testRun.Fatalf("err=%v", err)
	}
}

func TestParseKeepsExclamationMarkInsideComments(testRun *testing.T) {
	content := "*filter\n:INPUT DROP [0:0]\n-A INPUT -p tcp -m tcp --dport 22 -m comment --comment ! -j ACCEPT\nCOMMIT\n"
	parsed, err := parseSaveOutput(context.Background(), []byte(content))
	if err != nil {
		testRun.Fatal(err)
	}
	if rule := parsed.tables[0].Chains[0].Rules[0]; rule.Unsupported || rule.Target != domain.TargetAccept {
		testRun.Fatalf("rule = %+v", rule)
	}
}
