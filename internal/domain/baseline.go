package domain

import (
	"cmp"
	"errors"
	"fmt"
	"net/netip"
	"slices"
	"time"
	"unicode"
)

// ErrInvalidBaseline means a baseline or one of its entries breaks the format
// rules. Its messages never include the offending value.
var ErrInvalidBaseline = errors.New(invalidBaselineMessage)

// BaselineEntry is one accepted port. Its identity is protocol, port, bind
// address and source: process and container names, remote addresses and host
// names are deliberately not stored, so a baseline is safe to copy around.
type BaselineEntry struct {
	Protocol    Protocol
	Port        int
	BindAddress string
	Source      Source
}

// Baseline is a versioned snapshot of the ports a host is expected to expose.
type Baseline struct {
	Version     int
	CreatedAt   time.Time
	ToolVersion string
	Entries     []BaselineEntry
}

// BaselineEntryOf keeps only the identity of a port. An empty bind address
// means all interfaces, so it is stored as the wildcard.
func BaselineEntryOf(port Port) BaselineEntry {
	return BaselineEntry{
		Protocol: port.Protocol, Port: port.Number, BindAddress: CanonicalBindAddress(port.BindAddress), Source: port.Source,
	}
}

// Canonical returns the entry with its bind address in canonical form.
func (entry BaselineEntry) Canonical() BaselineEntry {
	entry.BindAddress = CanonicalBindAddress(entry.BindAddress)
	return entry
}

// CanonicalBindAddress gives one spelling to equivalent bind addresses: empty
// means the IPv4 wildcard, IPv4-mapped IPv6 becomes IPv4, an IPv6 zone is
// dropped and letters are lowercase. "0.0.0.0" and "::" stay different. An
// unparseable value is returned unchanged, so validation can reject it.
func CanonicalBindAddress(bindAddress string) string {
	if bindAddress == "" {
		return WildcardBindAddress
	}
	address, err := netip.ParseAddr(bindAddress)
	if err != nil {
		return bindAddress
	}
	return address.WithZone("").Unmap().String()
}

// Validate checks every field of the entry.
func (entry BaselineEntry) Validate() error {
	if entry.Protocol != TCP && entry.Protocol != UDP {
		return baselineInvalid(entryProtocolMessage)
	}
	if entry.Port < MinPortNumber || entry.Port > MaxPortNumber {
		return baselineInvalid(fmt.Sprintf(entryPortFormat, MinPortNumber, MaxPortNumber))
	}
	if _, err := netip.ParseAddr(entry.BindAddress); err != nil {
		return baselineInvalid(entryBindMessage)
	}
	if entry.Source != SourceHost && entry.Source != SourceDocker {
		return baselineInvalid(entrySourceMessage)
	}
	return nil
}

// Validate checks the header and every entry of the baseline.
func (baseline Baseline) Validate() error {
	if baseline.Version != BaselineVersion {
		return baselineInvalid(fmt.Sprintf(baselineVersionFormat, baseline.Version, BaselineVersion))
	}
	if baseline.CreatedAt.IsZero() {
		return baselineInvalid(baselineCreatedAtMessage)
	}
	if len(baseline.ToolVersion) > MaxToolVersionLength || containsControlCharacter(baseline.ToolVersion) {
		return baselineInvalid(baselineToolVersionMessage)
	}
	if len(baseline.Entries) > MaxBaselineEntries {
		return baselineInvalid(fmt.Sprintf(baselineTooManyEntriesFormat, MaxBaselineEntries))
	}
	for index, entry := range baseline.Entries {
		if err := entry.Validate(); err != nil {
			return fmt.Errorf(baselineEntryFormat, index, err)
		}
	}
	return nil
}

// Normalized returns a copy whose entries are canonical, sorted by source, protocol,
// port and bind address, with duplicates collapsed. The receiver is untouched.
func (baseline Baseline) Normalized() Baseline {
	normalized := baseline
	normalized.Entries = slices.Clone(baseline.Entries)
	for index := range normalized.Entries {
		normalized.Entries[index] = normalized.Entries[index].Canonical()
	}
	slices.SortFunc(normalized.Entries, CompareBaselineEntries)
	normalized.Entries = slices.Compact(normalized.Entries)
	if normalized.Entries == nil {
		normalized.Entries = []BaselineEntry{}
	}
	return normalized
}

// CompareBaselineEntries orders entries by source, protocol, port and bind address.
func CompareBaselineEntries(left, right BaselineEntry) int {
	return cmp.Or(
		cmp.Compare(left.Source, right.Source),
		cmp.Compare(left.Protocol, right.Protocol),
		cmp.Compare(left.Port, right.Port),
		cmp.Compare(left.BindAddress, right.BindAddress),
	)
}

func baselineInvalid(detail string) error {
	return fmt.Errorf("%w: %s", ErrInvalidBaseline, detail)
}

func containsControlCharacter(text string) bool {
	return slices.ContainsFunc([]rune(text), unicode.IsControl)
}
