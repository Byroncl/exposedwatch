package ports

import (
	"context"
	"errors"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

// ErrFirewallUnavailable means no firewall tool could be found on the host.
var ErrFirewallUnavailable = errors.New("ports: no firewall tool found")

// ErrFirewallNeedsRoot means the firewall rules cannot be read without root.
var ErrFirewallNeedsRoot = errors.New("ports: reading firewall rules needs root")

// FirewallRuleSource reads the host firewall rules.
type FirewallRuleSource interface {
	LoadRuleSet(ctx context.Context) (domain.RuleSet, error)
}

// ErrFirewallOutputTruncated means the firewall tool printed more than the
// reader keeps, so the rules are incomplete and must not be trusted.
var ErrFirewallOutputTruncated = errors.New("ports: firewall rules output was truncated")

// ErrFirewallOutputInvalid means the firewall rules could not be understood.
var ErrFirewallOutputInvalid = errors.New("ports: firewall rules output is invalid")
