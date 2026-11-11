package app

import (
	"time"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

// SnapshotBaseline turns the host and Docker ports of a scan into a
// normalized baseline. Other sources (external) are not part of a baseline,
// and process or container names are dropped. The creation time is stored in
// UTC with second precision.
func SnapshotBaseline(ports []domain.Port, createdAt time.Time, toolVersion string) domain.Baseline {
	baseline := domain.Baseline{
		Version:     domain.BaselineVersion,
		CreatedAt:   createdAt.UTC().Truncate(time.Second),
		ToolVersion: toolVersion,
	}
	for _, port := range ports {
		if isBaselineSource(port.Source) {
			baseline.Entries = append(baseline.Entries, domain.BaselineEntryOf(port))
		}
	}
	return baseline.Normalized()
}
