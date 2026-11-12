package iptablessave

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Byroncl/exposedwatch/internal/domain"
	"github.com/Byroncl/exposedwatch/internal/ports"
)

const (
	fakeIPv4Path     = "/fake/iptables-save"
	fakeIPv6Path     = "/fake/ip6tables-save"
	rootRequiredExit = 1
)

// fakeRunner answers by binary path and records every call.
type fakeRunner struct {
	results map[string]CommandResult
	errors  map[string]error
	calls   []fakeCall
}

type fakeCall struct {
	binary    string
	arguments []string
}

func (runner *fakeRunner) Run(ctx context.Context, binary string, arguments []string) (CommandResult, error) {
	runner.calls = append(runner.calls, fakeCall{binary: binary, arguments: arguments})
	if err := runner.errors[binary]; err != nil {
		return CommandResult{}, err
	}
	return runner.results[binary], nil
}

// fixtureLocator returns a locator that knows both fake binaries.
func fixtureLocator(testRun *testing.T, includeIPv6 bool) BinaryLocator {
	testRun.Helper()
	directory := testRun.TempDir()
	writeBinary(testRun, directory, ipv4SaveBinaryName, fileModeExecutable)
	if includeIPv6 {
		writeBinary(testRun, directory, ipv6SaveBinaryName, fileModeExecutable)
	}
	return BinaryLocator{Directories: []string{directory}, PathLookup: failingPathLookup}
}

func newLoaderWithFixtures(testRun *testing.T, includeIPv6 bool, ipv4Fixture string, ipv6Fixture string) (*RuleSetLoader, *fakeRunner) {
	testRun.Helper()
	locator := fixtureLocator(testRun, includeIPv6)
	ipv4Path, _ := locator.Find(ipv4SaveBinaryName)
	runner := &fakeRunner{results: map[string]CommandResult{
		ipv4Path: {Stdout: readFixture(testRun, ipv4Fixture)},
	}, errors: map[string]error{}}
	if includeIPv6 {
		ipv6Path, _ := locator.Find(ipv6SaveBinaryName)
		runner.results[ipv6Path] = CommandResult{Stdout: readFixture(testRun, ipv6Fixture)}
	}
	return NewRuleSetLoader(runner, locator), runner
}

func TestLoaderReadsBothFamiliesWithFixedArguments(testRun *testing.T) {
	loader, runner := newLoaderWithFixtures(testRun, true, "b_ufw.txt", "f_ipv6.txt")
	ruleSet, err := loader.LoadRuleSet(context.Background())
	if err != nil {
		testRun.Fatal(err)
	}
	if !ruleSet.IPv4.Available || !ruleSet.IPv6.Available || ruleSet.IPv4.Family != domain.FamilyIPv4 || ruleSet.IPv6.Family != domain.FamilyIPv6 {
		testRun.Fatalf("ruleSet = %+v", ruleSet)
	}
	if ruleSet.Backend != "iptables v1.8.11" {
		testRun.Fatalf("backend = %q", ruleSet.Backend)
	}
	if len(runner.calls) != 2 {
		testRun.Fatalf("calls = %d", len(runner.calls))
	}
	for _, call := range runner.calls {
		if len(call.arguments) != 0 {
			testRun.Fatalf("save commands must run without arguments, got %v", call.arguments)
		}
	}
}

func TestLoaderIPv6ToolMissingIsANoteNotAnError(testRun *testing.T) {
	loader, _ := newLoaderWithFixtures(testRun, false, "a_no_rules.txt", "")
	ruleSet, err := loader.LoadRuleSet(context.Background())
	if err != nil {
		testRun.Fatal(err)
	}
	if ruleSet.IPv6.Available || !ruleSet.IPv4.Available {
		testRun.Fatalf("availability v4=%v v6=%v", ruleSet.IPv4.Available, ruleSet.IPv6.Available)
	}
	if !hasNote(ruleSet.Notes, "ip6tables-save") {
		testRun.Fatalf("notes = %v", ruleSet.Notes)
	}
}

func TestLoaderMissingIPv4ToolIsUnavailable(testRun *testing.T) {
	loader := NewRuleSetLoader(&fakeRunner{}, BinaryLocator{Directories: []string{testRun.TempDir()}})
	_, err := loader.LoadRuleSet(context.Background())
	if !errors.Is(err, ports.ErrFirewallUnavailable) {
		testRun.Fatalf("err=%v", err)
	}
}

func TestLoaderMapsRealNonRootErrorToNeedsRoot(testRun *testing.T) {
	locator := fixtureLocator(testRun, true)
	ipv4Path, _ := locator.Find(ipv4SaveBinaryName)
	runner := &fakeRunner{results: map[string]CommandResult{
		ipv4Path: {Stderr: readFixture(testRun, "i_not_root.txt"), ExitCode: rootRequiredExit},
	}}
	_, err := NewRuleSetLoader(runner, locator).LoadRuleSet(context.Background())
	if !errors.Is(err, ports.ErrFirewallNeedsRoot) {
		testRun.Fatalf("err=%v", err)
	}
}

func TestRootMarkersAreDetectedInStdoutOrStderr(testRun *testing.T) {
	texts := []string{
		"Permission denied (you must be root)",
		"you must be root",
		"iptables-save: Operation not permitted",
	}
	for _, text := range texts {
		if !needsRoot(CommandResult{Stderr: []byte(text)}) || !needsRoot(CommandResult{Stdout: []byte(text)}) {
			testRun.Errorf("%q not detected", text)
		}
	}
	if needsRoot(CommandResult{Stderr: []byte("some other failure")}) {
		testRun.Error("unrelated failure mapped to root")
	}
}

func TestLoaderOtherFailureIsACommandError(testRun *testing.T) {
	locator := fixtureLocator(testRun, false)
	ipv4Path, _ := locator.Find(ipv4SaveBinaryName)
	runner := &fakeRunner{results: map[string]CommandResult{
		ipv4Path: {Stderr: []byte("kernel exploded"), ExitCode: 2},
	}}
	_, err := NewRuleSetLoader(runner, locator).LoadRuleSet(context.Background())
	if !errors.Is(err, ErrCommandFailed) || errors.Is(err, ports.ErrFirewallNeedsRoot) {
		testRun.Fatalf("err=%v", err)
	}
	if !strings.Contains(err.Error(), "kernel exploded") {
		testRun.Fatalf("stderr missing from error: %v", err)
	}
}

func TestLoaderBoundsErrorMessage(testRun *testing.T) {
	locator := fixtureLocator(testRun, false)
	ipv4Path, _ := locator.Find(ipv4SaveBinaryName)
	runner := &fakeRunner{results: map[string]CommandResult{
		ipv4Path: {Stderr: []byte(strings.Repeat("x", 5*maxMessageLength)), ExitCode: 2},
	}}
	_, err := NewRuleSetLoader(runner, locator).LoadRuleSet(context.Background())
	if err == nil || len(err.Error()) > 2*maxMessageLength {
		testRun.Fatalf("error not bounded: %v", err)
	}
}

func TestLoaderIPv6FailureAfterIPv4SuccessIsANote(testRun *testing.T) {
	loader, runner := newLoaderWithFixtures(testRun, true, "a_no_rules.txt", "f_ipv6.txt")
	for binary := range runner.results {
		if strings.HasSuffix(binary, ipv6SaveBinaryName) {
			runner.results[binary] = CommandResult{Stderr: []byte("Permission denied"), ExitCode: 1}
		}
	}
	ruleSet, err := loader.LoadRuleSet(context.Background())
	if err != nil || ruleSet.IPv6.Available || !hasNote(ruleSet.Notes, "ip6tables-save could not be read") {
		testRun.Fatalf("ruleSet=%+v err=%v", ruleSet, err)
	}
}

func TestLoaderPropagatesContextCancellation(testRun *testing.T) {
	locator := fixtureLocator(testRun, true)
	ipv4Path, _ := locator.Find(ipv4SaveBinaryName)
	runner := &fakeRunner{errors: map[string]error{ipv4Path: context.Canceled}}
	_, err := NewRuleSetLoader(runner, locator).LoadRuleSet(context.Background())
	if !errors.Is(err, context.Canceled) {
		testRun.Fatalf("err=%v", err)
	}
}

func TestLoaderUnparseableOutputIsInvalid(testRun *testing.T) {
	locator := fixtureLocator(testRun, false)
	ipv4Path, _ := locator.Find(ipv4SaveBinaryName)
	runner := &fakeRunner{results: map[string]CommandResult{ipv4Path: {Stdout: []byte("garbage\n")}}}
	_, err := NewRuleSetLoader(runner, locator).LoadRuleSet(context.Background())
	if !errors.Is(err, ErrInvalidOutput) {
		testRun.Fatalf("err=%v", err)
	}
}

func TestLoaderImplementsPort(testRun *testing.T) {
	var _ ports.FirewallRuleSource = (*RuleSetLoader)(nil)
}

func TestLoaderEmptyOutputHasNoTables(testRun *testing.T) {
	locator := fixtureLocator(testRun, false)
	ipv4Path, _ := locator.Find(ipv4SaveBinaryName)
	runner := &fakeRunner{results: map[string]CommandResult{ipv4Path: {}}}
	ruleSet, err := NewRuleSetLoader(runner, locator).LoadRuleSet(context.Background())
	if err != nil || ruleSet.HasTables() || ruleSet.Backend != domain.BackendUnknown {
		testRun.Fatalf("ruleSet=%+v err=%v", ruleSet, err)
	}
}

func hasNote(notes []string, fragment string) bool {
	for _, note := range notes {
		if strings.Contains(note, fragment) {
			return true
		}
	}
	return false
}
