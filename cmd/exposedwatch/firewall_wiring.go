package main

import (
	"context"
	"fmt"

	"github.com/Byroncl/exposedwatch/internal/adapters/iptablessave"
	"github.com/Byroncl/exposedwatch/internal/app"
	"github.com/Byroncl/exposedwatch/internal/domain"
	"github.com/Byroncl/exposedwatch/internal/tui"
)

// firewallEvaluator evaluates listening ports against the host firewall;
// *app.FirewallService satisfies it.
type firewallEvaluator interface {
	Evaluate(ctx context.Context, listening []domain.Port) (domain.FirewallReport, error)
}

// buildFirewallService reads the rules through iptables-save and
// ip6tables-save. Nothing is read until Evaluate runs.
func buildFirewallService() *app.FirewallService {
	return app.NewFirewallService(iptablessave.NewDefaultRuleSetLoader())
}

// buildFirewallEvaluator returns the firewall evaluator of the tui, or nil
// when --no-firewall disabled it.
func buildFirewallEvaluator(options TuiOptions) tui.FirewallEvaluator {
	if options.NoFirewall {
		return nil
	}
	return buildFirewallService()
}

// evaluateFirewall evaluates the ports. An evaluation error is not fatal: the
// report then says the firewall could not be read and every host or Docker
// port is unknown. Verdicts are information only.
func evaluateFirewall(ctx context.Context, evaluator firewallEvaluator, listening []domain.Port) domain.FirewallReport {
	firewallReport, err := evaluator.Evaluate(ctx, listening)
	if err != nil {
		return failedFirewallReport(listening, err)
	}
	return firewallReport
}

func failedFirewallReport(listening []domain.Port, cause error) domain.FirewallReport {
	reason := fmt.Sprintf(firewallReadFailedFormat, cause)
	failed := domain.FirewallReport{Backend: domain.BackendUnknown, Notes: []string{reason}}
	for _, port := range listening {
		if port.Source == domain.SourceExternal {
			continue
		}
		failed.Findings = append(failed.Findings, domain.FirewallFinding{Port: port, Verdict: domain.FirewallUnknown, Reason: reason})
	}
	return failed
}
