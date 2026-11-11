package tui

import (
	"context"

	"github.com/Byroncl/exposedwatch/internal/app"
)

// ConnectionSnapshotter produces connection snapshots; *app.ConnectionService
// satisfies it. A nil snapshotter disables the Connections tab.
type ConnectionSnapshotter interface {
	Snapshot(ctx context.Context) (app.ConnectionSnapshot, error)
}
