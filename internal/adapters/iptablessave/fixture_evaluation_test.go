package iptablessave

import (
	"context"
	"testing"

	"github.com/Byroncl/exposedwatch/internal/app"
	"github.com/Byroncl/exposedwatch/internal/domain"
)

// These tests run the whole path (fixture text -> parser -> evaluator) with a
// fake command runner. Nothing here touches real firewall rules.

type fixtureExpectation struct {
	protocol domain.Protocol
	number   int
	bind     string
	source   domain.Source
	expected domain.FirewallVerdict
}

const (
	anyIPv4Bind  = "0.0.0.0"
	anyDualBind  = "::"
	documentedV6 = "2001:db8::10"
)

func evaluateFixtures(testRun *testing.T, ipv4Fixture string, ipv6Fixture string, expectations []fixtureExpectation) {
	testRun.Helper()
	loader, _ := newLoaderWithFixtures(testRun, ipv6Fixture != "", ipv4Fixture, ipv6Fixture)
	var listening []domain.Port
	for _, expectation := range expectations {
		source := expectation.source
		if source == "" {
			source = domain.SourceHost
		}
		listening = append(listening, domain.Port{Number: expectation.number, Protocol: expectation.protocol, BindAddress: expectation.bind, Source: source})
	}
	report, err := app.NewFirewallService(loader).Evaluate(context.Background(), listening)
	if err != nil {
		testRun.Fatal(err)
	}
	for _, expectation := range expectations {
		found := false
		for _, finding := range report.Findings {
			port := finding.Port
			if port.Number == expectation.number && port.Protocol == expectation.protocol && port.BindAddress == expectation.bind {
				found = true
				testRun.Logf("%s + %s/%d@%s -> %s (%s)", ipv4Fixture, port.Protocol, port.Number, port.BindAddress, finding.Verdict, finding.Reason)
				if finding.Verdict != expectation.expected {
					testRun.Errorf("%s/%d@%s = %s (%s), want %s", port.Protocol, port.Number, port.BindAddress, finding.Verdict, finding.Reason, expectation.expected)
				}
			}
		}
		if !found {
			testRun.Errorf("no finding for %s/%d@%s", expectation.protocol, expectation.number, expectation.bind)
		}
	}
}

func TestFixtureNoRules(testRun *testing.T) {
	evaluateFixtures(testRun, "a_no_rules.txt", "", []fixtureExpectation{
		{domain.TCP, 22, anyIPv4Bind, "", domain.FirewallAllowed},
	})
}

func TestFixtureUfw(testRun *testing.T) {
	evaluateFixtures(testRun, "b_ufw.txt", "", []fixtureExpectation{
		{domain.TCP, 22, anyIPv4Bind, "", domain.FirewallAllowed},
		{domain.TCP, 80, anyIPv4Bind, "", domain.FirewallAllowed},
		{domain.TCP, 443, anyIPv4Bind, "", domain.FirewallAllowed},
		{domain.TCP, 9999, anyIPv4Bind, "", domain.FirewallBlocked},
		{domain.TCP, 445, anyIPv4Bind, "", domain.FirewallBlocked},
		{domain.UDP, 137, anyIPv4Bind, "", domain.FirewallBlocked},
		{domain.UDP, 5353, anyIPv4Bind, "", domain.FirewallUnknown},
		{domain.UDP, 68, anyIPv4Bind, "", domain.FirewallUnknown},
		{domain.TCP, 2222, anyIPv4Bind, "", domain.FirewallUnknown},
		{domain.TCP, 2223, anyIPv4Bind, "", domain.FirewallUnknown},
		{domain.TCP, 631, "127.0.0.1", "", domain.FirewallNotExposed},
	})
}

func TestFixtureDocker(testRun *testing.T) {
	evaluateFixtures(testRun, "c_docker.txt", "", []fixtureExpectation{
		{domain.TCP, 22, anyIPv4Bind, "", domain.FirewallAllowed},
		{domain.TCP, 8080, anyIPv4Bind, domain.SourceDocker, domain.FirewallDockerPublished},
	})
}

func TestFixtureSourceRestricted(testRun *testing.T) {
	evaluateFixtures(testRun, "d_source_restricted.txt", "", []fixtureExpectation{
		{domain.TCP, 5432, anyIPv4Bind, "", domain.FirewallRestricted},
		{domain.TCP, 22, anyIPv4Bind, "", domain.FirewallAllowed},
		{domain.TCP, 9999, anyIPv4Bind, "", domain.FirewallBlocked},
	})
}

func TestFixtureUnsupportedMatches(testRun *testing.T) {
	evaluateFixtures(testRun, "e_unsupported_matches.txt", "", []fixtureExpectation{
		{domain.TCP, 3306, anyIPv4Bind, "", domain.FirewallUnknown},
		{domain.TCP, 6379, anyIPv4Bind, "", domain.FirewallUnknown},
		{domain.TCP, 22, anyIPv4Bind, "", domain.FirewallAllowed},
		{domain.TCP, 80, anyIPv4Bind, "", domain.FirewallBlocked},
	})
}

func TestFixtureIPv6AndDualStack(testRun *testing.T) {
	evaluateFixtures(testRun, "d_source_restricted.txt", "f_ipv6.txt", []fixtureExpectation{
		{domain.TCP, 443, documentedV6, "", domain.FirewallRestricted},
		{domain.TCP, 443, anyDualBind, "", domain.FirewallRestricted},
		{domain.TCP, 22, anyDualBind, "", domain.FirewallAllowed},
		{domain.TCP, 9999, anyDualBind, "", domain.FirewallBlocked},
		{domain.TCP, 9999, anyIPv4Bind, "", domain.FirewallBlocked},
	})
}

func TestFixtureIPv6ToolMissingMakesDualStackUnknown(testRun *testing.T) {
	evaluateFixtures(testRun, "d_source_restricted.txt", "", []fixtureExpectation{
		{domain.TCP, 9999, anyDualBind, "", domain.FirewallUnknown},
		{domain.TCP, 22, anyDualBind, "", domain.FirewallAllowed},
	})
}

func TestFixtureChainCycle(testRun *testing.T) {
	evaluateFixtures(testRun, "g_chain_cycle.txt", "", []fixtureExpectation{
		{domain.TCP, 7000, anyIPv4Bind, "", domain.FirewallUnknown},
		{domain.TCP, 7001, anyIPv4Bind, "", domain.FirewallBlocked},
	})
}

func TestFixtureRawAndMangleDrops(testRun *testing.T) {
	evaluateFixtures(testRun, "h_raw_mangle_drop.txt", "", []fixtureExpectation{
		{domain.TCP, 22, anyIPv4Bind, "", domain.FirewallUnknown},
		{domain.TCP, 23, anyIPv4Bind, "", domain.FirewallUnknown},
	})
}

func TestFixtureNonRootReportsUnavailableNotFailure(testRun *testing.T) {
	locator := fixtureLocator(testRun, true)
	ipv4Path, _ := locator.Find(ipv4SaveBinaryName)
	runner := &fakeRunner{results: map[string]CommandResult{
		ipv4Path: {Stderr: readFixture(testRun, "i_not_root.txt"), ExitCode: rootRequiredExit},
	}}
	report, err := app.NewFirewallService(NewRuleSetLoader(runner, locator)).Evaluate(context.Background(), []domain.Port{
		{Number: 22, Protocol: domain.TCP, BindAddress: anyIPv4Bind, Source: domain.SourceHost},
	})
	if err != nil || report.Available || len(report.Findings) != 1 || report.Findings[0].Verdict != domain.FirewallUnknown {
		testRun.Fatalf("report=%+v err=%v", report, err)
	}
	if report.Findings[0].Reason != "needs root to read firewall rules" {
		testRun.Fatalf("reason = %q", report.Findings[0].Reason)
	}
}

func TestFixtureNatLookalikeChainHidesRedirect(testRun *testing.T) {
	evaluateFixtures(testRun, "k_nat_dockerish.txt", "", []fixtureExpectation{
		{domain.TCP, 22, anyIPv4Bind, "", domain.FirewallUnknown},
	})
}

func TestFixtureRawNotrackMakesAllowedUnknown(testRun *testing.T) {
	evaluateFixtures(testRun, "l_raw_notrack.txt", "", []fixtureExpectation{
		{domain.TCP, 22, anyIPv4Bind, "", domain.FirewallUnknown},
	})
}

func TestFixturePortZeroRange(testRun *testing.T) {
	evaluateFixtures(testRun, "n_port_zero_range.txt", "", []fixtureExpectation{
		{domain.TCP, 22, anyIPv4Bind, "", domain.FirewallBlocked},
		{domain.TCP, 500, anyIPv4Bind, "", domain.FirewallBlocked},
		{domain.TCP, 8005, anyIPv4Bind, "", domain.FirewallAllowed},
	})
}

func TestFixtureOddSyntax(testRun *testing.T) {
	evaluateFixtures(testRun, "o_odd_syntax.txt", "", []fixtureExpectation{
		{domain.TCP, 22, anyIPv4Bind, "", domain.FirewallUnknown},
		{domain.TCP, 80, anyIPv4Bind, "", domain.FirewallAllowed},
		{domain.TCP, 443, anyIPv4Bind, "", domain.FirewallUnknown},
	})
}

func TestFixtureTruncatedOutputMakesEverythingUnknown(testRun *testing.T) {
	locator := fixtureLocator(testRun, false)
	ipv4Path, _ := locator.Find(ipv4SaveBinaryName)
	runner := &fakeRunner{results: map[string]CommandResult{ipv4Path: {Stdout: readFixture(testRun, "m_truncated_mid_rule.txt")}}}
	report, err := app.NewFirewallService(NewRuleSetLoader(runner, locator)).Evaluate(context.Background(), []domain.Port{
		{Number: 22, Protocol: domain.TCP, BindAddress: anyIPv4Bind, Source: domain.SourceHost},
	})
	if err != nil || report.Available || report.Findings[0].Verdict != domain.FirewallUnknown {
		testRun.Fatalf("report=%+v err=%v", report, err)
	}
	testRun.Logf("m_truncated_mid_rule.txt + tcp/22 -> %s (%s)", report.Findings[0].Verdict, report.Findings[0].Reason)
}

func TestFixtureNativeNftablesDowngradesAllowed(testRun *testing.T) {
	testCases := []struct {
		name     string
		include  bool
		result   CommandResult
		expected domain.FirewallVerdict
	}{
		{"nft absent", false, CommandResult{}, domain.FirewallAllowed},
		{"iptables-nft tables only", true, CommandResult{Stdout: readFixture(testRun, "nft_iptables_tables_only.txt")}, domain.FirewallAllowed},
		{"native inet filter", true, CommandResult{Stdout: readFixture(testRun, "nft_inet_filter.txt")}, domain.FirewallUnknown},
		{"several tables", true, CommandResult{Stdout: readFixture(testRun, "nft_several_tables.txt")}, domain.FirewallUnknown},
		{"nft needs root", true, CommandResult{Stderr: readFixture(testRun, "nft_not_root.txt"), ExitCode: 1}, domain.FirewallUnknown},
	}
	for _, testCase := range testCases {
		loader, _ := loaderWithNft(testRun, testCase.include, testCase.result)
		report, err := app.NewFirewallService(loader).Evaluate(context.Background(), []domain.Port{
			{Number: 22, Protocol: domain.TCP, BindAddress: anyIPv4Bind, Source: domain.SourceHost},
		})
		if err != nil || report.Findings[0].Verdict != testCase.expected {
			testRun.Errorf("%s: report=%+v err=%v", testCase.name, report, err)
			continue
		}
		testRun.Logf("a_no_rules.txt + nft(%s) + tcp/22 -> %s (%s)", testCase.name, report.Findings[0].Verdict, report.Findings[0].Reason)
	}
}
