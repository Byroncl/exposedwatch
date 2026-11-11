package domain

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func validBaselineEntry() BaselineEntry {
	return BaselineEntry{Protocol: TCP, Port: 22, BindAddress: "0.0.0.0", Source: SourceHost}
}

func validBaseline() Baseline {
	return Baseline{
		Version:     BaselineVersion,
		CreatedAt:   time.Date(2026, 11, 11, 10, 0, 0, 0, time.UTC),
		ToolVersion: "v0.4.0",
		Entries:     []BaselineEntry{validBaselineEntry()},
	}
}

func TestBaselineEntryValidation(testRun *testing.T) {
	testCases := []struct {
		name   string
		mutate func(entry *BaselineEntry)
		valid  bool
	}{
		{"valid tcp host", func(*BaselineEntry) {}, true},
		{"valid udp docker ipv6", func(entry *BaselineEntry) { entry.Protocol, entry.Source, entry.BindAddress = UDP, SourceDocker, "::" }, true},
		{"unknown protocol", func(entry *BaselineEntry) { entry.Protocol = "sctp" }, false},
		{"empty protocol", func(entry *BaselineEntry) { entry.Protocol = "" }, false},
		{"port zero", func(entry *BaselineEntry) { entry.Port = 0 }, false},
		{"negative port", func(entry *BaselineEntry) { entry.Port = -1 }, false},
		{"port too large", func(entry *BaselineEntry) { entry.Port = 65536 }, false},
		{"highest port", func(entry *BaselineEntry) { entry.Port = 65535 }, true},
		{"bind not an address", func(entry *BaselineEntry) { entry.BindAddress = "example.com" }, false},
		{"empty bind", func(entry *BaselineEntry) { entry.BindAddress = "" }, false},
		{"external source", func(entry *BaselineEntry) { entry.Source = SourceExternal }, false},
		{"empty source", func(entry *BaselineEntry) { entry.Source = "" }, false},
	}
	for _, testCase := range testCases {
		testRun.Run(testCase.name, func(testRun *testing.T) {
			entry := validBaselineEntry()
			testCase.mutate(&entry)
			err := entry.Validate()
			if testCase.valid && err != nil {
				testRun.Fatalf("want valid, got %v", err)
			}
			if !testCase.valid && !errors.Is(err, ErrInvalidBaseline) {
				testRun.Fatalf("want ErrInvalidBaseline, got %v", err)
			}
		})
	}
}

func TestBaselineValidation(testRun *testing.T) {
	testCases := []struct {
		name   string
		mutate func(baseline *Baseline)
		valid  bool
	}{
		{"valid", func(*Baseline) {}, true},
		{"empty entries are valid", func(baseline *Baseline) { baseline.Entries = nil }, true},
		{"wrong version", func(baseline *Baseline) { baseline.Version = 2 }, false},
		{"zero version", func(baseline *Baseline) { baseline.Version = 0 }, false},
		{"zero creation time", func(baseline *Baseline) { baseline.CreatedAt = time.Time{} }, false},
		{"tool version too long", func(baseline *Baseline) { baseline.ToolVersion = strings.Repeat("v", MaxToolVersionLength+1) }, false},
		{"tool version with control character", func(baseline *Baseline) { baseline.ToolVersion = "v1\n2" }, false},
		{"invalid entry", func(baseline *Baseline) { baseline.Entries[0].Port = 0 }, false},
		{"too many entries", func(baseline *Baseline) {
			baseline.Entries = make([]BaselineEntry, MaxBaselineEntries+1)
			for index := range baseline.Entries {
				baseline.Entries[index] = validBaselineEntry()
			}
		}, false},
	}
	for _, testCase := range testCases {
		testRun.Run(testCase.name, func(testRun *testing.T) {
			baseline := validBaseline()
			testCase.mutate(&baseline)
			err := baseline.Validate()
			if testCase.valid != (err == nil) || (err != nil && !errors.Is(err, ErrInvalidBaseline)) {
				testRun.Fatalf("valid=%v err=%v", testCase.valid, err)
			}
		})
	}
}

func TestBaselineErrorsNeverEchoTheValues(testRun *testing.T) {
	entry := validBaselineEntry()
	entry.BindAddress = "secret-host.internal"
	if err := entry.Validate(); err == nil || strings.Contains(err.Error(), "secret-host") {
		testRun.Fatalf("error must not echo the value: %v", err)
	}
}

func TestBaselineEntryOfDropsEverythingButTheIdentity(testRun *testing.T) {
	port := Port{Number: 5432, Protocol: TCP, BindAddress: "10.0.0.5", Process: "postgres", Container: "db", Source: SourceDocker}
	want := BaselineEntry{Protocol: TCP, Port: 5432, BindAddress: "10.0.0.5", Source: SourceDocker}
	if got := BaselineEntryOf(port); got != want {
		testRun.Fatalf("got %+v want %+v", got, want)
	}
}

func TestBaselineEntryOfTreatsAnEmptyBindAsTheWildcard(testRun *testing.T) {
	entry := BaselineEntryOf(Port{Number: 80, Protocol: TCP, Source: SourceDocker})
	if entry.BindAddress != WildcardBindAddress {
		testRun.Fatalf("got %q", entry.BindAddress)
	}
}

func TestNormalizedSortsAndCollapsesDuplicates(testRun *testing.T) {
	baseline := validBaseline()
	baseline.Entries = []BaselineEntry{
		{Protocol: TCP, Port: 80, BindAddress: "0.0.0.0", Source: SourceHost},
		{Protocol: UDP, Port: 53, BindAddress: "0.0.0.0", Source: SourceHost},
		{Protocol: TCP, Port: 22, BindAddress: "127.0.0.1", Source: SourceHost},
		{Protocol: TCP, Port: 22, BindAddress: "0.0.0.0", Source: SourceHost},
		{Protocol: TCP, Port: 22, BindAddress: "0.0.0.0", Source: SourceHost},
		{Protocol: TCP, Port: 9000, BindAddress: "0.0.0.0", Source: SourceDocker},
	}
	original := append([]BaselineEntry(nil), baseline.Entries...)
	normalized := baseline.Normalized()
	want := []BaselineEntry{
		{Protocol: TCP, Port: 9000, BindAddress: "0.0.0.0", Source: SourceDocker},
		{Protocol: TCP, Port: 22, BindAddress: "0.0.0.0", Source: SourceHost},
		{Protocol: TCP, Port: 22, BindAddress: "127.0.0.1", Source: SourceHost},
		{Protocol: TCP, Port: 80, BindAddress: "0.0.0.0", Source: SourceHost},
		{Protocol: UDP, Port: 53, BindAddress: "0.0.0.0", Source: SourceHost},
	}
	if len(normalized.Entries) != len(want) {
		testRun.Fatalf("got %+v", normalized.Entries)
	}
	for index := range want {
		if normalized.Entries[index] != want[index] {
			testRun.Fatalf("index %d: got %+v want %+v", index, normalized.Entries[index], want[index])
		}
	}
	for index := range original {
		if baseline.Entries[index] != original[index] {
			testRun.Fatal("Normalized must not modify the receiver")
		}
	}
}

func TestBindAddressesAreCanonicalized(testRun *testing.T) {
	testCases := []struct{ name, bind, want string }{
		{"uppercase ipv6", "2001:DB8::A", "2001:db8::a"},
		{"ipv4 mapped ipv6", "::ffff:10.0.0.1", "10.0.0.1"},
		{"zoned ipv6", "fe80::1%eth0", "fe80::1"},
		{"empty", "", "0.0.0.0"},
		{"plain ipv4", "10.0.0.1", "10.0.0.1"},
		{"ipv6 wildcard stays", "::", "::"},
		{"ipv4 wildcard stays", "0.0.0.0", "0.0.0.0"},
	}
	for _, testCase := range testCases {
		testRun.Run(testCase.name, func(testRun *testing.T) {
			got := BaselineEntryOf(Port{Number: 80, Protocol: TCP, BindAddress: testCase.bind, Source: SourceHost})
			if got.BindAddress != testCase.want {
				testRun.Fatalf("got %q want %q", got.BindAddress, testCase.want)
			}
		})
	}
	if BaselineEntryOf(Port{BindAddress: "0.0.0.0"}) == BaselineEntryOf(Port{BindAddress: "::"}) {
		testRun.Fatal("0.0.0.0 and :: are different identities")
	}
}

func TestNormalizedCanonicalizesAndCollapses(testRun *testing.T) {
	baseline := validBaseline()
	baseline.Entries = []BaselineEntry{
		{Protocol: TCP, Port: 80, BindAddress: "FE80::1%eth0", Source: SourceHost},
		{Protocol: TCP, Port: 80, BindAddress: "fe80::1", Source: SourceHost},
		{Protocol: TCP, Port: 80, BindAddress: "::ffff:1.2.3.4", Source: SourceHost},
		{Protocol: TCP, Port: 80, BindAddress: "1.2.3.4", Source: SourceHost},
	}
	normalized := baseline.Normalized()
	if len(normalized.Entries) != 2 || normalized.Entries[0].BindAddress != "1.2.3.4" || normalized.Entries[1].BindAddress != "fe80::1" {
		testRun.Fatalf("got %+v", normalized.Entries)
	}
}
