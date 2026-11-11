package tui

import (
	"context"

	"github.com/Byroncl/exposedwatch/internal/app"
)

// Scanner produces exposure reports; *app.ScanService satisfies it.
type Scanner interface {
	Scan(ctx context.Context) (app.Report, error)
}
