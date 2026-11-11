// Package ports declares the interfaces the application core uses to reach
// the outside world.
package ports

import (
	"context"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

// PortSource discovers ports from one origin (host sockets, Docker, ...).
type PortSource interface {
	ListPorts(ctx context.Context) ([]domain.Port, error)
}
