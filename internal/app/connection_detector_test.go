package app

import (
	"reflect"
	"testing"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

func makeTestConnection(localPort int, remoteAddress string, remotePort int) domain.Connection {
	return domain.Connection{
		Protocol:      domain.TCP,
		LocalAddress:  "10.0.0.5",
		LocalPort:     localPort,
		RemoteAddress: remoteAddress,
		RemotePort:    remotePort,
		State:         domain.ConnectionEstablished,
		Direction:     domain.ConnectionInbound,
	}
}

func TestDetectConnectionChangesWithoutBaselineYieldsNothing(testRun *testing.T) {
	current := []domain.Connection{makeTestConnection(22, "203.0.113.10", 50000)}
	if changes := DetectConnectionChanges(nil, current); len(changes) != 0 {
		testRun.Fatalf("changes = %v, want none", changes)
	}
}

func TestDetectConnectionChangesEmptyBaselineReportsOpened(testRun *testing.T) {
	current := []domain.Connection{makeTestConnection(22, "203.0.113.10", 50000)}
	changes := DetectConnectionChanges([]domain.Connection{}, current)
	want := []ConnectionChange{{Kind: ChangeOpened, Connection: current[0]}}
	if !reflect.DeepEqual(changes, want) {
		testRun.Fatalf("changes = %v, want %v", changes, want)
	}
}

func TestDetectConnectionChangesOpenedAndClosed(testRun *testing.T) {
	kept := makeTestConnection(22, "203.0.113.10", 50000)
	gone := makeTestConnection(443, "198.51.100.7", 50001)
	added := makeTestConnection(443, "192.0.2.55", 50002)
	changes := DetectConnectionChanges([]domain.Connection{kept, gone}, []domain.Connection{kept, added})
	want := []ConnectionChange{
		{Kind: ChangeOpened, Connection: added},
		{Kind: ChangeClosed, Connection: gone},
	}
	if !reflect.DeepEqual(changes, want) {
		testRun.Fatalf("changes = %v, want %v", changes, want)
	}
}

func TestDetectConnectionChangesUnchangedYieldsNothing(testRun *testing.T) {
	connections := []domain.Connection{makeTestConnection(22, "203.0.113.10", 50000)}
	if changes := DetectConnectionChanges(connections, connections); len(changes) != 0 {
		testRun.Fatalf("changes = %v, want none", changes)
	}
}

func TestDetectConnectionChangesSamePeerNewSourcePortIsNew(testRun *testing.T) {
	first := makeTestConnection(22, "203.0.113.10", 50000)
	second := makeTestConnection(22, "203.0.113.10", 50001)
	changes := DetectConnectionChanges([]domain.Connection{first}, []domain.Connection{second})
	want := []ConnectionChange{
		{Kind: ChangeOpened, Connection: second},
		{Kind: ChangeClosed, Connection: first},
	}
	if !reflect.DeepEqual(changes, want) {
		testRun.Fatalf("changes = %v, want %v", changes, want)
	}
}

func TestDetectConnectionChangesIgnoresDescriptiveFieldsDuplicatesAndOrder(testRun *testing.T) {
	first := makeTestConnection(22, "203.0.113.10", 50000)
	second := makeTestConnection(80, "198.51.100.7", 50001)
	renamed := first
	renamed.ProcessName = "sshd"
	changes := DetectConnectionChanges(
		[]domain.Connection{first, second},
		[]domain.Connection{second, renamed, first},
	)
	if len(changes) != 0 {
		testRun.Fatalf("changes = %v, want none", changes)
	}
}

func TestDetectConnectionChangesOrderIsStable(testRun *testing.T) {
	lowPort := makeTestConnection(22, "203.0.113.10", 50000)
	highPort := makeTestConnection(443, "198.51.100.7", 50001)
	highPortOtherPeer := makeTestConnection(443, "192.0.2.55", 50002)
	changes := DetectConnectionChanges(
		[]domain.Connection{},
		[]domain.Connection{highPort, highPortOtherPeer, lowPort},
	)
	want := []ConnectionChange{
		{Kind: ChangeOpened, Connection: lowPort},
		{Kind: ChangeOpened, Connection: highPortOtherPeer},
		{Kind: ChangeOpened, Connection: highPort},
	}
	if !reflect.DeepEqual(changes, want) {
		testRun.Fatalf("changes = %v, want %v", changes, want)
	}
}
