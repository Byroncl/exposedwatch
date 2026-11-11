package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/Byroncl/exposedwatch/internal/domain"
	"github.com/Byroncl/exposedwatch/internal/ports"
)

// ConnectionSnapshot is the connections seen at one moment plus their summary.
// Truncated is true when the source hit its limit and the list is partial.
type ConnectionSnapshot struct {
	Connections []domain.Connection
	Summary     ConnectionSummary
	Truncated   bool
}

// ConnectionService reads connections from a source and summarizes them.
type ConnectionService struct {
	source ports.ConnectionSource
}

// NewConnectionService builds a service over the given source.
func NewConnectionService(source ports.ConnectionSource) *ConnectionService {
	return &ConnectionService{source: source}
}

// Snapshot reads and summarizes the current connections. A truncated list is
// returned as a successful, flagged snapshot; any other source failure is
// returned to the caller, which decides how to react.
func (service *ConnectionService) Snapshot(ctx context.Context) (ConnectionSnapshot, error) {
	if service.source == nil {
		return ConnectionSnapshot{}, ErrNoConnectionSource
	}
	connections, err := service.source.ListConnections(ctx)
	truncated := errors.Is(err, ports.ErrConnectionsTruncated)
	if err != nil && !truncated {
		return ConnectionSnapshot{}, fmt.Errorf(connectionSourceFormat, err)
	}
	return ConnectionSnapshot{
		Connections: connections,
		Summary:     SummarizeConnections(connections),
		Truncated:   truncated,
	}, nil
}
