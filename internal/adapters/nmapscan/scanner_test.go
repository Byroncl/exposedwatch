package nmapscan

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Byroncl/exposedwatch/internal/domain"
	"github.com/Byroncl/exposedwatch/internal/ports"
)

var _ ports.PortSource = (*Scanner)(nil)

const (
	fakeBinary      = "/usr/bin/nmap"
	fakeScanTimeout = 90 * time.Second
)

type fakeRunner struct {
	result    CommandResult
	err       error
	calls     int
	binary    string
	arguments []string
	block     bool
}

func (runner *fakeRunner) Run(ctx context.Context, binary string, arguments []string) (CommandResult, error) {
	runner.calls++
	runner.binary = binary
	runner.arguments = arguments
	if runner.block {
		<-ctx.Done()
		return CommandResult{}, ctx.Err()
	}
	return runner.result, runner.err
}

func newTestScanner(runner CommandRunner, targets TargetList, portSpec PortSpec) *Scanner {
	return New(Options{
		NmapPath: fakeBinary, Targets: targets, Ports: portSpec, Timeout: fakeScanTimeout,
		Runner: runner, AddressLister: fakeAddresses("192.168.1.10/24", "127.0.0.1/8"),
	})
}

func TestListPortsBuildsExactArgumentsWithTopPorts(testRun *testing.T) {
	runner := &fakeRunner{result: CommandResult{Stdout: readFixture(testRun, "no_hosts.xml")}}
	_, err := newTestScanner(runner, TargetList{"10.0.0.1", "10.0.0.0/30", "example.com"}, "").ListPorts(context.Background())
	if err != nil {
		testRun.Fatal(err)
	}
	want := []string{
		"-oX", "-", "-Pn", "-n", "--open", "-sT", "-T3", "--host-timeout", "90s",
		"--top-ports", "1000", "--", "10.0.0.1", "10.0.0.0/30", "example.com",
	}
	if runner.binary != fakeBinary || !reflect.DeepEqual(runner.arguments, want) {
		testRun.Fatalf("binary=%q arguments=%q", runner.binary, runner.arguments)
	}
}

func TestListPortsBuildsExactArgumentsWithPortSpec(testRun *testing.T) {
	runner := &fakeRunner{result: CommandResult{Stdout: readFixture(testRun, "no_hosts.xml")}}
	_, err := newTestScanner(runner, TargetList{"10.0.0.1"}, "22,8000-8100").ListPorts(context.Background())
	if err != nil {
		testRun.Fatal(err)
	}
	want := []string{
		"-oX", "-", "-Pn", "-n", "--open", "-sT", "-T3", "--host-timeout", "90s",
		"-p", "22,8000-8100", "--", "10.0.0.1",
	}
	if !reflect.DeepEqual(runner.arguments, want) {
		testRun.Fatalf("arguments=%q", runner.arguments)
	}
}

func TestListPortsUsesLocalTargetsByDefault(testRun *testing.T) {
	runner := &fakeRunner{result: CommandResult{Stdout: readFixture(testRun, "no_hosts.xml")}}
	if _, err := newTestScanner(runner, nil, "").ListPorts(context.Background()); err != nil {
		testRun.Fatal(err)
	}
	tail := runner.arguments[len(runner.arguments)-2:]
	if !reflect.DeepEqual(tail, []string{"--", "192.168.1.10"}) {
		testRun.Fatalf("arguments=%q", runner.arguments)
	}
}

func TestListPortsMapsOpenPortsToExternalDomainPorts(testRun *testing.T) {
	runner := &fakeRunner{result: CommandResult{Stdout: readFixture(testRun, "multiple_hosts.xml")}}
	found, err := newTestScanner(runner, TargetList{"10.0.0.5"}, "").ListPorts(context.Background())
	if err != nil {
		testRun.Fatal(err)
	}
	want := []domain.Port{
		{Number: 80, Protocol: domain.TCP, BindAddress: "10.0.0.5", Source: domain.SourceExternal},
		{Number: 443, Protocol: domain.TCP, BindAddress: "2001:db8::1", Source: domain.SourceExternal},
	}
	if !reflect.DeepEqual(found, want) {
		testRun.Fatalf("got %+v", found)
	}
}

func TestListPortsRevalidatesTargetsBeforeRunning(testRun *testing.T) {
	runner := &fakeRunner{}
	_, err := newTestScanner(runner, TargetList{"--script=x"}, "").ListPorts(context.Background())
	if !errors.Is(err, ErrInvalidTarget) || runner.calls != 0 {
		testRun.Fatalf("err=%v calls=%d", err, runner.calls)
	}
}

func TestListPortsRevalidatesPortSpecAndTimeout(testRun *testing.T) {
	runner := &fakeRunner{}
	if _, err := newTestScanner(runner, TargetList{"10.0.0.1"}, "22;ls").ListPorts(context.Background()); !errors.Is(err, ErrInvalidPortSpec) {
		testRun.Fatalf("port spec err=%v", err)
	}
	scanner := New(Options{Targets: TargetList{"10.0.0.1"}, Timeout: time.Hour * 24, Runner: runner})
	if _, err := scanner.ListPorts(context.Background()); !errors.Is(err, ErrInvalidTimeout) || runner.calls != 0 {
		testRun.Fatalf("timeout err=%v calls=%d", err, runner.calls)
	}
}

func TestListPortsNmapMissing(testRun *testing.T) {
	runner := &fakeRunner{err: ErrNmapUnavailable}
	_, err := newTestScanner(runner, TargetList{"10.0.0.1"}, "").ListPorts(context.Background())
	if !errors.Is(err, ErrNmapUnavailable) {
		testRun.Fatalf("err=%v", err)
	}
}

func TestListPortsNonZeroExitBoundsStderr(testRun *testing.T) {
	runner := &fakeRunner{result: CommandResult{ExitCode: 1, Stderr: []byte(strings.Repeat("x", 5000))}}
	_, err := newTestScanner(runner, TargetList{"10.0.0.1"}, "").ListPorts(context.Background())
	if !errors.Is(err, ErrNmapFailed) {
		testRun.Fatalf("err=%v", err)
	}
	if len(err.Error()) > maxStderrCharacters+200 {
		testRun.Fatalf("error not bounded: %d chars", len(err.Error()))
	}
}

func TestListPortsUnparseableOutput(testRun *testing.T) {
	runner := &fakeRunner{result: CommandResult{Stdout: readFixture(testRun, "malformed.xml")}}
	_, err := newTestScanner(runner, TargetList{"10.0.0.1"}, "").ListPorts(context.Background())
	if !errors.Is(err, ErrInvalidOutput) {
		testRun.Fatalf("err=%v", err)
	}
}

func TestListPortsTimeoutKillsRunAndReportsTimeout(testRun *testing.T) {
	runner := &fakeRunner{block: true}
	scanner := New(Options{
		Targets: TargetList{"10.0.0.1"}, Timeout: MinTimeout, Runner: runner,
		AddressLister: fakeAddresses("192.168.1.10/24"),
	})
	scanner.processLimit = 20 * time.Millisecond
	_, err := scanner.ListPorts(context.Background())
	if !errors.Is(err, ErrScanTimedOut) {
		testRun.Fatalf("err=%v", err)
	}
}

func TestListPortsParentCancellationIsPropagated(testRun *testing.T) {
	runner := &fakeRunner{block: true}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := newTestScanner(runner, TargetList{"10.0.0.1"}, "").ListPorts(ctx)
	if !errors.Is(err, context.Canceled) {
		testRun.Fatalf("err=%v", err)
	}
}
