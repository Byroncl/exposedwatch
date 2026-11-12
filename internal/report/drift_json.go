package report

import (
	"time"

	"github.com/Byroncl/exposedwatch/internal/app"
	"github.com/Byroncl/exposedwatch/internal/domain"
)

// jsonDriftNew is the stable JSON shape of a port that is not in the baseline.
type jsonDriftNew struct {
	jsonPort
	Severity string `json:"severity"`
	Title    string `json:"title"`
}

// jsonDriftClosed is the stable JSON shape of a baseline entry that is gone.
type jsonDriftClosed struct {
	Port        int    `json:"port"`
	Protocol    string `json:"protocol"`
	BindAddress string `json:"bind_address"`
	Source      string `json:"source"`
}

// jsonDrift is the stable JSON shape of the comparison with a baseline.
type jsonDrift struct {
	BaselineFile string            `json:"baseline_file"`
	CreatedAt    string            `json:"created_at"`
	New          []jsonDriftNew    `json:"new"`
	Closed       []jsonDriftClosed `json:"closed"`
	Notes        []string          `json:"notes"`
}

func toJSONDrift(driftReport app.DriftReport) *jsonDrift {
	drift := &jsonDrift{
		BaselineFile: driftReport.BaselineFile,
		CreatedAt:    driftReport.CreatedAt.UTC().Format(time.RFC3339),
		New:          make([]jsonDriftNew, 0, len(driftReport.New)),
		Closed:       make([]jsonDriftClosed, 0, len(driftReport.Closed)),
		Notes:        append(make([]string, 0, len(driftReport.Notes)), driftReport.Notes...),
	}
	for _, assessment := range driftReport.New {
		drift.New = append(drift.New, jsonDriftNew{
			jsonPort: toJSONPort(assessment.Port), Severity: assessment.Severity.String(), Title: assessment.Title,
		})
	}
	for _, entry := range driftReport.Closed {
		drift.Closed = append(drift.Closed, toJSONDriftClosed(entry))
	}
	return drift
}

func toJSONDriftClosed(entry domain.BaselineEntry) jsonDriftClosed {
	return jsonDriftClosed{Port: entry.Port, Protocol: string(entry.Protocol), BindAddress: entry.BindAddress, Source: string(entry.Source)}
}
