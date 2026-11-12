package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

// FirewallEvaluator decides, for each listening port, whether a new inbound
// connection would get through the host firewall. It is pure and
// conservative: whenever a rule cannot be evaluated the verdict is Unknown,
// never an assurance. The result is information only.
type FirewallEvaluator struct{}

// NewFirewallEvaluator returns an evaluator.
func NewFirewallEvaluator() FirewallEvaluator {
	return FirewallEvaluator{}
}

// familyResult is the verdict for one address family.
type familyResult struct {
	family  domain.AddressFamily
	verdict domain.FirewallVerdict
	reason  string
}

// EvaluatePort returns the firewall finding for one host or Docker port.
func (evaluator FirewallEvaluator) EvaluatePort(ctx context.Context, ruleSet domain.RuleSet, port domain.Port) domain.FirewallFinding {
	if finding, decided := classifyWithoutRules(port); decided {
		return finding
	}
	address, specific, families := parseBind(port.BindAddress)
	probe := connectionProbe{port: port.Number, protocol: port.Protocol, bind: address, hasBind: specific}
	results := make([]familyResult, 0, len(families))
	for _, family := range families {
		results = append(results, evaluateFamily(ctx, ruleSet, family, probe))
	}
	return buildFinding(port, results)
}

// EvaluateUnavailable returns the finding for a port when the rules could not
// be read: ports that need no rules keep their verdict and every other port
// is Unknown with the given reason.
func (evaluator FirewallEvaluator) EvaluateUnavailable(port domain.Port, reason string) domain.FirewallFinding {
	if finding, decided := classifyWithoutRules(port); decided {
		return finding
	}
	return domain.FirewallFinding{Port: port, Verdict: domain.FirewallUnknown, Reason: reason}
}

// classifyWithoutRules handles the ports whose verdict does not depend on the
// rules: loopback-only binds and Docker-published ports.
func classifyWithoutRules(port domain.Port) (domain.FirewallFinding, bool) {
	switch {
	case port.Exposure() == domain.Loopback:
		return domain.FirewallFinding{Port: port, Verdict: domain.FirewallNotExposed, Reason: reasonLoopback}, true
	case port.Source == domain.SourceDocker:
		return domain.FirewallFinding{Port: port, Verdict: domain.FirewallDockerPublished, Reason: reasonDockerPublished}, true
	default:
		return domain.FirewallFinding{}, false
	}
}

func evaluateFamily(ctx context.Context, ruleSet domain.RuleSet, familyName domain.AddressFamily, probe connectionProbe) familyResult {
	family := ruleSet.Family(familyName)
	result := familyResult{family: familyName, verdict: domain.FirewallUnknown}
	if !family.Available {
		result.reason = fmt.Sprintf(reasonFamilyUnavailableFormat, toolName(familyName))
		return result
	}
	if len(family.Tables) == 0 {
		result.reason = reasonNoTables
		return result
	}
	filter, found := family.Table(domain.TableFilter)
	if !found {
		result.reason = reasonNoFilterTable
		return result
	}
	outcome := walkInputChain(ctx, filter, probe, family.RuleCount() == 0)
	verdict, reason := outcome.verdict()
	verdict, reason = applyCaveat(verdict, reason, preFilterCaveat(family))
	result.verdict, result.reason = applyNftablesState(verdict, reason, ruleSet.Nftables)
	return result
}

func toolName(family domain.AddressFamily) string {
	if family == domain.FamilyIPv6 {
		return ipv6ToolName
	}
	return ipv4ToolName
}

// buildFinding merges the per-family results of one port.
func buildFinding(port domain.Port, results []familyResult) domain.FirewallFinding {
	if len(results) == 1 {
		return domain.FirewallFinding{Port: port, Verdict: results[0].verdict, Reason: results[0].reason}
	}
	verdict := results[0].verdict
	reasons := make([]string, 0, len(results))
	for index, result := range results {
		if index > 0 {
			verdict = combineVerdicts(verdict, result.verdict)
		}
		reasons = append(reasons, fmt.Sprintf(reasonFamilyFormat, result.family, result.reason))
	}
	return domain.FirewallFinding{Port: port, Verdict: verdict, Reason: strings.Join(reasons, reasonJoiner)}
}

// applyNftablesState keeps what iptables can prove and drops what it cannot.
// A native nftables ruleset can only add drops, so an iptables DROP stays
// valid, while Allowed and Restricted are only safe when no native rules may
// exist.
func applyNftablesState(verdict domain.FirewallVerdict, reason string, state domain.NftablesState) (domain.FirewallVerdict, string) {
	if verdict != domain.FirewallAllowed && verdict != domain.FirewallRestricted {
		return verdict, reason
	}
	switch state {
	case domain.NftablesNative:
		return domain.FirewallUnknown, reasonNativeNftables
	case domain.NftablesProbeFailed:
		return domain.FirewallUnknown, reasonNftablesCheckFailed
	default:
		return verdict, reason
	}
}
