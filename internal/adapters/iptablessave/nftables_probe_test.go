package iptablessave

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

// loaderWithNft builds a loader whose runner answers iptables-save with the
// no-rules fixture and nft with the given result. includeNft false means the
// nft binary is not installed.
func loaderWithNft(testRun *testing.T, includeNft bool, nftResult CommandResult) (*RuleSetLoader, *fakeRunner) {
	testRun.Helper()
	directory := testRun.TempDir()
	writeBinary(testRun, directory, ipv4SaveBinaryName, fileModeExecutable)
	writeBinary(testRun, directory, ipv6SaveBinaryName, fileModeExecutable)
	locator := BinaryLocator{Directories: []string{directory}, PathLookup: failingPathLookup}
	ipv4Path, _ := locator.Find(ipv4SaveBinaryName)
	ipv6Path, _ := locator.Find(ipv6SaveBinaryName)
	runner := &fakeRunner{results: map[string]CommandResult{
		ipv4Path: {Stdout: readFixture(testRun, "a_no_rules.txt")},
		ipv6Path: {Stdout: readFixture(testRun, "a_no_rules.txt")},
	}}
	if includeNft {
		nftPath := writeBinary(testRun, directory, nftBinaryName, fileModeExecutable)
		runner.results[nftPath] = nftResult
	}
	return NewRuleSetLoader(runner, locator), runner
}

func TestNftProbeClassifiesOutput(testRun *testing.T) {
	testCases := []struct {
		name     string
		result   CommandResult
		include  bool
		expected domain.NftablesState
		note     string
	}{
		{"binary absent assumes none", CommandResult{}, false, domain.NftablesNone, "nft not found"},
		{"only iptables-nft tables", CommandResult{Stdout: readFixture(testRun, "nft_iptables_tables_only.txt")}, true, domain.NftablesNone, ""},
		{"empty output", CommandResult{}, true, domain.NftablesNone, ""},
		{"inet filter is native", CommandResult{Stdout: readFixture(testRun, "nft_inet_filter.txt")}, true, domain.NftablesNative, "native nftables"},
		{"several tables with native ones", CommandResult{Stdout: readFixture(testRun, "nft_several_tables.txt")}, true, domain.NftablesNative, "native nftables"},
		{"root error is unknown", CommandResult{Stderr: readFixture(testRun, "nft_not_root.txt"), ExitCode: 1}, true, domain.NftablesProbeFailed, "could not check nftables"},
		{"garbage output is unknown", CommandResult{Stdout: readFixture(testRun, "nft_garbage.txt")}, true, domain.NftablesProbeFailed, "could not check nftables"},
		{"truncated output is unknown", CommandResult{Stdout: readFixture(testRun, "nft_iptables_tables_only.txt"), Truncated: true}, true, domain.NftablesProbeFailed, "could not check nftables"},
	}
	for _, testCase := range testCases {
		testRun.Run(testCase.name, func(testRun *testing.T) {
			loader, _ := loaderWithNft(testRun, testCase.include, testCase.result)
			ruleSet, err := loader.LoadRuleSet(context.Background())
			if err != nil {
				testRun.Fatal(err)
			}
			if ruleSet.Nftables != testCase.expected {
				testRun.Fatalf("state = %v, want %v (notes %v)", ruleSet.Nftables, testCase.expected, ruleSet.Notes)
			}
			if testCase.note != "" && !hasNote(ruleSet.Notes, testCase.note) {
				testRun.Fatalf("notes = %v, want one containing %q", ruleSet.Notes, testCase.note)
			}
		})
	}
}

func TestNftProbeUsesFixedArgumentsAndNoShell(testRun *testing.T) {
	loader, runner := loaderWithNft(testRun, true, CommandResult{})
	if _, err := loader.LoadRuleSet(context.Background()); err != nil {
		testRun.Fatal(err)
	}
	found := false
	for _, call := range runner.calls {
		if strings.HasSuffix(call.binary, nftBinaryName) {
			found = true
			if strings.Join(call.arguments, " ") != "list tables" {
				testRun.Fatalf("arguments = %v", call.arguments)
			}
		}
	}
	if !found {
		testRun.Fatal("nft was never run")
	}
}

func TestNftProbeFailureDoesNotFailTheLoad(testRun *testing.T) {
	loader, _ := loaderWithNft(testRun, true, CommandResult{Stderr: []byte("boom"), ExitCode: 3})
	ruleSet, err := loader.LoadRuleSet(context.Background())
	if err != nil || ruleSet.Nftables != domain.NftablesProbeFailed {
		testRun.Fatalf("ruleSet=%+v err=%v", ruleSet, err)
	}
}

func TestNftProbePropagatesCancellation(testRun *testing.T) {
	loader, runner := loaderWithNft(testRun, true, CommandResult{})
	for binary := range runner.results {
		if strings.HasSuffix(binary, nftBinaryName) {
			runner.errors = map[string]error{binary: context.Canceled}
		}
	}
	if _, err := loader.LoadRuleSet(context.Background()); !errors.Is(err, context.Canceled) {
		testRun.Fatalf("err=%v", err)
	}
}

func TestClassifyNftTablesDocumentsTheNativeIPFilterGap(testRun *testing.T) {
	// A hand-made native "table ip filter" is indistinguishable from the one
	// iptables-nft creates, so it is not detected. This test pins that gap.
	state, ok := classifyNftTables("table ip filter\n")
	if !ok || state != domain.NftablesNone {
		testRun.Fatalf("state=%v ok=%v", state, ok)
	}
}
