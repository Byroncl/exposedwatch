package ports

import (
	"context"
	"errors"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

// ErrConnectionsTruncated is returned together with a usable, partial list
// when a source hit its connection limit. Callers should show the list and a
// notice instead of treating it as a failure.
var ErrConnectionsTruncated = errors.New("ports: connection list truncated")

// ConnectionSource lists the connections currently open on this host.
type ConnectionSource interface {
	ListConnections(ctx context.Context) ([]domain.Connection, error)
}
