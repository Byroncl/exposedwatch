package app

import (
	"testing"
	"time"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

func hostPort(number int, bind string) domain.Port {
	return domain.Port{Number: number, Protocol: domain.TCP, BindAddress: bind, Source: domain.SourceHost, Process: "daemon"}
}

func dockerPort(number int, bind string) domain.Port {
	return domain.Port{Number: number, Protocol: domain.TCP, BindAddress: bind, Source: domain.SourceDocker, Container: "web"}
}

func baselineOf(entries ...domain.BaselineEntry) domain.Baseline {
	return domain.Baseline{
		Version: domain.BaselineVersion, CreatedAt: time.Date(2026, 11, 11, 10, 0, 0, 0, time.UTC),
		ToolVersion: "test", Entries: entries,
	}
}

func entryOf(port domain.Port) domain.BaselineEntry { return domain.BaselineEntryOf(port) }

func TestCompareToBaselineUnchangedIsNoDrift(testRun *testing.T) {
	ports := []domain.Port{hostPort(22, "0.0.0.0"), dockerPort(8080, "0.0.0.0")}
	drift := CompareToBaseline(baselineOf(entryOf(ports[0]), entryOf(ports[1])), ports, nil)
	if len(drift.New) != 0 || len(drift.Closed) != 0 || len(drift.Unverified) != 0 {
		testRun.Fatalf("got %+v", drift)
	}
}

func TestCompareToBaselineFindsNewAndClosed(testRun *testing.T) {
	known := hostPort(22, "0.0.0.0")
	gone := hostPort(3306, "0.0.0.0")
	fresh := hostPort(8080, "0.0.0.0")
	drift := CompareToBaseline(baselineOf(entryOf(known), entryOf(gone)), []domain.Port{known, fresh}, nil)
	if len(drift.New) != 1 || drift.New[0].Number != 8080 || drift.New[0].Process != "daemon" {
		testRun.Fatalf("new: %+v", drift.New)
	}
	if len(drift.Closed) != 1 || drift.Closed[0] != entryOf(gone) {
		testRun.Fatalf("closed: %+v", drift.Closed)
	}
}

func TestCompareToBaselineLoopbackToWildcardIsNew(testRun *testing.T) {
	before := hostPort(5432, "127.0.0.1")
	after := hostPort(5432, "0.0.0.0")
	drift := CompareToBaseline(baselineOf(entryOf(before)), []domain.Port{after}, nil)
	if len(drift.New) != 1 || drift.New[0].BindAddress != "0.0.0.0" {
		testRun.Fatalf("new: %+v", drift.New)
	}
	if len(drift.Closed) != 1 || drift.Closed[0].BindAddress != "127.0.0.1" {
		testRun.Fatalf("closed: %+v", drift.Closed)
	}
}

func TestCompareToBaselineIgnoresProcessAndContainerNames(testRun *testing.T) {
	renamed := hostPort(22, "0.0.0.0")
	renamed.Process = "dropbear"
	drift := CompareToBaseline(baselineOf(entryOf(hostPort(22, "0.0.0.0"))), []domain.Port{renamed}, nil)
	if len(drift.New) != 0 || len(drift.Closed) != 0 {
		testRun.Fatalf("got %+v", drift)
	}
}

func TestCompareToBaselineSourceIsPartOfTheIdentity(testRun *testing.T) {
	drift := CompareToBaseline(baselineOf(entryOf(hostPort(8080, "0.0.0.0"))), []domain.Port{dockerPort(8080, "0.0.0.0")}, nil)
	if len(drift.New) != 1 || len(drift.Closed) != 1 {
		testRun.Fatalf("got %+v", drift)
	}
}

func TestCompareToBaselineNeverReportsAFailedSourceAsClosed(testRun *testing.T) {
	baseline := baselineOf(entryOf(hostPort(22, "0.0.0.0")), entryOf(dockerPort(8080, "0.0.0.0")))
	drift := CompareToBaseline(baseline, []domain.Port{hostPort(22, "0.0.0.0")}, []domain.Source{domain.SourceDocker})
	if len(drift.Closed) != 0 {
		testRun.Fatalf("docker ports are unknown, not closed: %+v", drift.Closed)
	}
	if len(drift.Unverified) != 1 || drift.Unverified[0].Source != domain.SourceDocker || drift.Unverified[0].Entries != 1 {
		testRun.Fatalf("unverified: %+v", drift.Unverified)
	}
}

func TestCompareToBaselineListsFailedSourcesWithoutEntries(testRun *testing.T) {
	drift := CompareToBaseline(baselineOf(entryOf(hostPort(22, "0.0.0.0"))), []domain.Port{hostPort(22, "0.0.0.0")}, []domain.Source{domain.SourceDocker, domain.SourceDocker})
	if len(drift.Unverified) != 1 || drift.Unverified[0].Source != domain.SourceDocker || drift.Unverified[0].Entries != 0 {
		testRun.Fatalf("got %+v", drift.Unverified)
	}
}

func TestCompareToBaselineIgnoresExternalPorts(testRun *testing.T) {
	external := domain.Port{Number: 22, Protocol: domain.TCP, BindAddress: "192.0.2.10", Source: domain.SourceExternal}
	drift := CompareToBaseline(baselineOf(), []domain.Port{external}, nil)
	if len(drift.New) != 0 {
		testRun.Fatalf("external ports are not part of a baseline: %+v", drift.New)
	}
}

func TestCompareToBaselineCollapsesDuplicatesAndOrdersDeterministically(testRun *testing.T) {
	ports := []domain.Port{
		hostPort(9000, "0.0.0.0"), hostPort(80, "0.0.0.0"), hostPort(80, "0.0.0.0"), dockerPort(8000, "0.0.0.0"),
		{Number: 53, Protocol: domain.UDP, BindAddress: "0.0.0.0", Source: domain.SourceHost},
	}
	baseline := baselineOf(entryOf(hostPort(7000, "0.0.0.0")), entryOf(dockerPort(1, "0.0.0.0")), entryOf(hostPort(6000, "0.0.0.0")))
	drift := CompareToBaseline(baseline, ports, nil)
	wantNew := []int{8000, 80, 9000, 53}
	if len(drift.New) != len(wantNew) {
		testRun.Fatalf("new: %+v", drift.New)
	}
	for index, want := range wantNew {
		if drift.New[index].Number != want {
			testRun.Fatalf("new[%d]=%d want %d (%+v)", index, drift.New[index].Number, want, drift.New)
		}
	}
	wantClosed := []int{1, 6000, 7000}
	for index, want := range wantClosed {
		if drift.Closed[index].Port != want {
			testRun.Fatalf("closed[%d]=%d want %d", index, drift.Closed[index].Port, want)
		}
	}
}

func TestCompareToBaselineNormalizesAnEmptyBindLikeTheBaseline(testRun *testing.T) {
	port := domain.Port{Number: 80, Protocol: domain.TCP, Source: domain.SourceDocker}
	drift := CompareToBaseline(baselineOf(entryOf(port)), []domain.Port{port}, nil)
	if len(drift.New) != 0 || len(drift.Closed) != 0 {
		testRun.Fatalf("got %+v", drift)
	}
}

func TestCompareToBaselineUsesTheCanonicalBindOnBothSides(testRun *testing.T) {
	stored := domain.BaselineEntry{Protocol: domain.TCP, Port: 80, BindAddress: "FE80::1", Source: domain.SourceHost}
	current := hostPort(80, "fe80::1%eth0")
	drift := CompareToBaseline(baselineOf(stored), []domain.Port{current}, nil)
	if len(drift.New) != 0 || len(drift.Closed) != 0 {
		testRun.Fatalf("got %+v", drift)
	}
}
