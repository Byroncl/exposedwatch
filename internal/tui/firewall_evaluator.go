package tui

import (
	"context"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

// FirewallEvaluator evaluates listening ports against the host firewall;
// *app.FirewallService satisfies it. A nil evaluator disables the Firewall
// module and the firewall is never read.
type FirewallEvaluator interface {
	Evaluate(ctx context.Context, listening []domain.Port) (domain.FirewallReport, error)
}
