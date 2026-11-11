package app

import (
	"reflect"
	"testing"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

func makeTestPort(number int, protocol domain.Protocol, bindAddress string, source domain.Source) domain.Port {
	return domain.Port{Number: number, Protocol: protocol, BindAddress: bindAddress, Source: source}
}

func TestDetectChangesWithoutBaselineYieldsNothing(testRun *testing.T) {
	current := []domain.Port{makeTestPort(22, domain.TCP, "0.0.0.0", domain.SourceHost)}
	if changes := DetectChanges(nil, current); len(changes) != 0 {
		testRun.Fatalf("changes = %v, want none", changes)
	}
}

func TestDetectChangesEmptyBaselineReportsEverythingAsOpened(testRun *testing.T) {
	current := []domain.Port{makeTestPort(22, domain.TCP, "0.0.0.0", domain.SourceHost)}
	changes := DetectChanges([]domain.Port{}, current)
	want := []Change{{Kind: ChangeOpened, Port: current[0]}}
	if !reflect.DeepEqual(changes, want) {
		testRun.Fatalf("changes = %v, want %v", changes, want)
	}
}

func TestDetectChangesOpenedAndClosed(testRun *testing.T) {
	kept := makeTestPort(22, domain.TCP, "0.0.0.0", domain.SourceHost)
	gone := makeTestPort(3306, domain.TCP, "0.0.0.0", domain.SourceHost)
	added := makeTestPort(8080, domain.TCP, "0.0.0.0", domain.SourceDocker)
	changes := DetectChanges([]domain.Port{kept, gone}, []domain.Port{kept, added})
	want := []Change{{Kind: ChangeOpened, Port: added}, {Kind: ChangeClosed, Port: gone}}
	if !reflect.DeepEqual(changes, want) {
		testRun.Fatalf("changes = %v, want %v", changes, want)
	}
}

func TestDetectChangesUnchangedYieldsNothing(testRun *testing.T) {
	ports := []domain.Port{makeTestPort(22, domain.TCP, "0.0.0.0", domain.SourceHost)}
	if changes := DetectChanges(ports, ports); len(changes) != 0 {
		testRun.Fatalf("changes = %v, want none", changes)
	}
}

func TestDetectChangesIgnoresDuplicatesAndReordering(testRun *testing.T) {
	first := makeTestPort(22, domain.TCP, "0.0.0.0", domain.SourceHost)
	second := makeTestPort(80, domain.TCP, "0.0.0.0", domain.SourceHost)
	previous := []domain.Port{first, second}
	current := []domain.Port{second, first, first, second}
	if changes := DetectChanges(previous, current); len(changes) != 0 {
		testRun.Fatalf("changes = %v, want none", changes)
	}
	added := makeTestPort(443, domain.TCP, "0.0.0.0", domain.SourceHost)
	changes := DetectChanges(previous, []domain.Port{added, added, first, second})
	if len(changes) != 1 || changes[0].Kind != ChangeOpened {
		testRun.Fatalf("duplicate opened port produced %v", changes)
	}
}

func TestDetectChangesSourceBindAndProtocolAreIdentity(testRun *testing.T) {
	base := makeTestPort(80, domain.TCP, "0.0.0.0", domain.SourceHost)
	variants := []domain.Port{
		makeTestPort(80, domain.TCP, "0.0.0.0", domain.SourceDocker),
		makeTestPort(80, domain.TCP, "127.0.0.1", domain.SourceHost),
		makeTestPort(80, domain.UDP, "0.0.0.0", domain.SourceHost),
	}
	for _, variant := range variants {
		changes := DetectChanges([]domain.Port{base}, []domain.Port{variant})
		want := []Change{{Kind: ChangeOpened, Port: variant}, {Kind: ChangeClosed, Port: base}}
		if !reflect.DeepEqual(changes, want) {
			testRun.Errorf("variant %+v: changes = %v, want %v", variant, changes, want)
		}
	}
}

func TestDetectChangesProcessChangeIsNotAChange(testRun *testing.T) {
	before := makeTestPort(80, domain.TCP, "0.0.0.0", domain.SourceHost)
	after := before
	after.Process = "nginx"
	if changes := DetectChanges([]domain.Port{before}, []domain.Port{after}); len(changes) != 0 {
		testRun.Fatalf("changes = %v, want none", changes)
	}
}

func TestDetectChangesOrderingIsStable(testRun *testing.T) {
	previous := []domain.Port{
		makeTestPort(900, domain.TCP, "0.0.0.0", domain.SourceHost),
		makeTestPort(100, domain.TCP, "0.0.0.0", domain.SourceHost),
	}
	current := []domain.Port{
		makeTestPort(700, domain.TCP, "0.0.0.0", domain.SourceHost),
		makeTestPort(300, domain.TCP, "0.0.0.0", domain.SourceHost),
	}
	wantNumbers := []int{300, 700, 100, 900}
	wantKinds := []ChangeKind{ChangeOpened, ChangeOpened, ChangeClosed, ChangeClosed}
	for run := 0; run < 5; run++ {
		changes := DetectChanges(previous, current)
		for index, change := range changes {
			if change.Port.Number != wantNumbers[index] || change.Kind != wantKinds[index] {
				testRun.Fatalf("run %d index %d: %+v", run, index, change)
			}
		}
		if len(changes) != len(wantNumbers) {
			testRun.Fatalf("len = %d", len(changes))
		}
	}
}

func TestChangeKindString(testRun *testing.T) {
	if ChangeOpened.String() != "opened" || ChangeClosed.String() != "closed" {
		testRun.Fatalf("names: %q %q", ChangeOpened, ChangeClosed)
	}
	if ChangeKind(99).String() != "unknown" {
		testRun.Fatal("unknown kind name")
	}
}

func TestIdentityOfIgnoresProcessAndContainer(testRun *testing.T) {
	left := makeTestPort(80, domain.TCP, "0.0.0.0", domain.SourceHost)
	right := left
	right.Process, right.Container = "nginx", "web"
	if IdentityOf(left) != IdentityOf(right) {
		testRun.Fatal("identity must ignore process and container")
	}
}
