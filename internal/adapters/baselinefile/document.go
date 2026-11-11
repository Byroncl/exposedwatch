package baselinefile

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

// documentEntry is the stable JSON shape of one baseline entry.
type documentEntry struct {
	Protocol    string `json:"protocol"`
	Port        int    `json:"port"`
	BindAddress string `json:"bind_address"`
	Source      string `json:"source"`
}

// document is the stable JSON shape of the baseline file.
type document struct {
	Version     int             `json:"version"`
	CreatedAt   string          `json:"created_at"`
	ToolVersion string          `json:"tool_version"`
	Entries     []documentEntry `json:"entries"`
}

// encodeDocument validates and normalizes the baseline and renders it as
// indented, newline terminated JSON.
func encodeDocument(baseline domain.Baseline) ([]byte, error) {
	if err := baseline.Validate(); err != nil {
		return nil, err
	}
	normalized := baseline.Normalized()
	wire := document{
		Version:     normalized.Version,
		CreatedAt:   normalized.CreatedAt.UTC().Format(time.RFC3339),
		ToolVersion: normalized.ToolVersion,
		Entries:     make([]documentEntry, 0, len(normalized.Entries)),
	}
	for _, entry := range normalized.Entries {
		wire.Entries = append(wire.Entries, documentEntry{
			Protocol: string(entry.Protocol), Port: entry.Port, BindAddress: entry.BindAddress, Source: string(entry.Source),
		})
	}
	content, err := json.MarshalIndent(wire, "", fileIndentation)
	if err != nil {
		return nil, writeFailure(encodeMessage, err)
	}
	return append(content, fileTrailer...), nil
}

// decodeDocument parses strictly: unknown fields and trailing data are
// rejected, then the baseline is validated and normalized. Errors never
// quote the input.
func decodeDocument(content []byte) (domain.Baseline, error) {
	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.DisallowUnknownFields()
	var wire document
	if err := decoder.Decode(&wire); err != nil {
		return domain.Baseline{}, ErrMalformedBaseline
	}
	if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
		return domain.Baseline{}, ErrMalformedBaseline
	}
	createdAt, err := time.Parse(time.RFC3339, wire.CreatedAt)
	if err != nil {
		return domain.Baseline{}, fmt.Errorf(errorDetailFormat, domain.ErrInvalidBaseline, createdAtInvalidText)
	}
	baseline := domain.Baseline{
		Version:     wire.Version,
		CreatedAt:   createdAt.UTC(),
		ToolVersion: wire.ToolVersion,
		Entries:     make([]domain.BaselineEntry, 0, len(wire.Entries)),
	}
	for _, entry := range wire.Entries {
		baseline.Entries = append(baseline.Entries, domain.BaselineEntry{
			Protocol: domain.Protocol(entry.Protocol), Port: entry.Port, BindAddress: entry.BindAddress, Source: domain.Source(entry.Source),
		})
	}
	if err := baseline.Validate(); err != nil {
		return domain.Baseline{}, err
	}
	return baseline.Normalized(), nil
}
