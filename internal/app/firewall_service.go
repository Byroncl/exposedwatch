package app

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/Byroncl/exposedwatch/internal/domain"
	"github.com/Byroncl/exposedwatch/internal/ports"
)

// FirewallService reads the host firewall rules and evaluates a list of ports
// against them. It is information only and never changes severities.
type FirewallService struct {
	source    ports.FirewallRuleSource
	evaluator FirewallEvaluator
}

// NewFirewallService builds a service over the given rule source.
func NewFirewallService(source ports.FirewallRuleSource) *FirewallService {
	return &FirewallService{source: source, evaluator: NewFirewallEvaluator()}
}

// Evaluate returns one finding per host or Docker port, ordered by port,
// protocol and bind address. External ports are an outside view and are
// skipped. A missing firewall tool or missing root privileges are not
// failures: the report has Available=false and every exposed host port is
// Unknown with the reason. Any other read failure is returned.
func (service *FirewallService) Evaluate(ctx context.Context, listening []domain.Port) (domain.FirewallReport, error) {
	if service.source == nil {
		return domain.FirewallReport{}, ErrNoFirewallSource
	}
	candidates := hostAndDockerPorts(listening)
	ruleSet, err := service.source.LoadRuleSet(ctx)
	switch {
	case errors.Is(err, ports.ErrFirewallUnavailable):
		return service.unavailableReport(candidates, reasonNoTool), nil
	case errors.Is(err, ports.ErrFirewallNeedsRoot):
		return service.unavailableReport(candidates, reasonNeedsRoot), nil
	case errors.Is(err, ports.ErrFirewallOutputTruncated):
		return service.unavailableReport(candidates, reasonOutputTruncated), nil
	case errors.Is(err, ports.ErrFirewallOutputInvalid):
		return service.unavailableReport(candidates, reasonOutputInvalid), nil
	case err != nil:
		return domain.FirewallReport{}, fmt.Errorf(failureFormat, err)
	}
	report := domain.FirewallReport{Backend: ruleSet.Backend, Available: true, Notes: slices.Clone(ruleSet.Notes)}
	if report.Backend == "" {
		report.Backend = domain.BackendUnknown
	}
	if !ruleSet.HasTables() {
		report.Notes = append(report.Notes, reasonNoTables)
	}
	for _, port := range candidates {
		report.Findings = append(report.Findings, service.evaluator.EvaluatePort(ctx, ruleSet, port))
		if contextErr := ctx.Err(); contextErr != nil {
			return domain.FirewallReport{}, contextErr
		}
	}
	sortFindings(report.Findings)
	return report, nil
}

func (service *FirewallService) unavailableReport(candidates []domain.Port, reason string) domain.FirewallReport {
	report := domain.FirewallReport{Backend: domain.BackendUnknown, Notes: []string{reason}}
	for _, port := range candidates {
		report.Findings = append(report.Findings, service.evaluator.EvaluateUnavailable(port, reason))
	}
	sortFindings(report.Findings)
	return report
}

// hostAndDockerPorts drops the ports that were observed from outside.
func hostAndDockerPorts(listening []domain.Port) []domain.Port {
	var selected []domain.Port
	for _, port := range listening {
		if port.Source != domain.SourceExternal {
			selected = append(selected, port)
		}
	}
	return selected
}

// sortFindings orders findings deterministically.
func sortFindings(findings []domain.FirewallFinding) {
	slices.SortStableFunc(findings, func(first domain.FirewallFinding, second domain.FirewallFinding) int {
		return cmp.Or(
			cmp.Compare(first.Port.Number, second.Port.Number),
			cmp.Compare(first.Port.Protocol, second.Port.Protocol),
			cmp.Compare(first.Port.BindAddress, second.Port.BindAddress),
			cmp.Compare(first.Port.Source, second.Port.Source),
			cmp.Compare(first.Port.Container, second.Port.Container),
		)
	})
}
