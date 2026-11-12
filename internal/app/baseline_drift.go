package app

import (
	"slices"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

// UnverifiedSource is a source that could not be read during the scan, with
// the number of baseline entries that therefore could not be checked.
type UnverifiedSource struct {
	Source  domain.Source
	Entries int
}

// Drift is the difference between a baseline and the ports seen now.
type Drift struct {
	// New are the current ports whose identity is not in the baseline.
	New []domain.Port
	// Closed are baseline entries that are no longer exposed.
	Closed []domain.BaselineEntry
	// Unverified are the sources that failed during the scan, with how many
	// baseline entries could not be checked: their ports are unknown, not closed.
	Unverified []UnverifiedSource
}

// CompareToBaseline compares the host and Docker ports against a baseline. It
// is pure and deterministic: results are sorted by source, protocol, port and
// bind address, and duplicates collapse. Ports of other sources (external)
// are ignored. A source listed in unknownSources failed during the scan, so
// its baseline entries are reported as unverified instead of closed.
func CompareToBaseline(baseline domain.Baseline, ports []domain.Port, unknownSources []domain.Source) Drift {
	baselineIdentities := make(map[domain.BaselineEntry]bool, len(baseline.Entries))
	for _, entry := range baseline.Entries {
		baselineIdentities[entry.Canonical()] = true
	}
	currentIdentities := make(map[domain.BaselineEntry]bool, len(ports))
	var drift Drift
	for _, port := range ports {
		if !isBaselineSource(port.Source) {
			continue
		}
		identity := domain.BaselineEntryOf(port)
		if currentIdentities[identity] {
			continue
		}
		currentIdentities[identity] = true
		if !baselineIdentities[identity] {
			drift.New = append(drift.New, port)
		}
	}
	slices.SortFunc(drift.New, func(left, right domain.Port) int {
		return domain.CompareBaselineEntries(domain.BaselineEntryOf(left), domain.BaselineEntryOf(right))
	})
	unverifiedCounts := map[domain.Source]int{}
	for _, source := range unknownSources {
		unverifiedCounts[source] = 0
	}
	for identity := range baselineIdentities {
		switch {
		case slices.Contains(unknownSources, identity.Source):
			unverifiedCounts[identity.Source]++
		case !currentIdentities[identity]:
			drift.Closed = append(drift.Closed, identity)
		}
	}
	slices.SortFunc(drift.Closed, domain.CompareBaselineEntries)
	for source, entries := range unverifiedCounts {
		drift.Unverified = append(drift.Unverified, UnverifiedSource{Source: source, Entries: entries})
	}
	slices.SortFunc(drift.Unverified, func(left, right UnverifiedSource) int {
		return compareSources(left.Source, right.Source)
	})
	return drift
}

func isBaselineSource(source domain.Source) bool {
	return source == domain.SourceHost || source == domain.SourceDocker
}

func compareSources(left, right domain.Source) int {
	return domain.CompareBaselineEntries(domain.BaselineEntry{Source: left}, domain.BaselineEntry{Source: right})
}
