// Package risk evaluates ports against a table of exposure rules.
package risk

import (
	"sort"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

// Evaluate returns findings for the given ports, sorted by severity
// descending and then port number ascending. Loopback-only ports yield none.
func Evaluate(ports []domain.Port) []domain.Finding {
	findings := make([]domain.Finding, 0, len(ports))
	for _, p := range ports {
		if f, ok := evaluatePort(p); ok {
			findings = append(findings, f)
		}
	}
	sort.SliceStable(findings, func(i, j int) bool {
		a, b := findings[i], findings[j]
		if a.Severity != b.Severity {
			return a.Severity > b.Severity
		}
		return a.Port.Number < b.Port.Number
	})
	return findings
}

func evaluatePort(p domain.Port) (domain.Finding, bool) {
	exp := p.Exposure()
	if exp == domain.Loopback {
		return domain.Finding{}, false
	}
	for _, r := range builtinRules {
		if r.matches(p.Number) {
			if exp >= r.MinExposure {
				return r.finding(p), true
			}
			return domain.Finding{}, false
		}
	}
	if exp >= genericRule.MinExposure {
		return genericRule.finding(p), true
	}
	return domain.Finding{}, false
}

func (r Rule) matches(port int) bool {
	for _, n := range r.Ports {
		if n == port {
			return true
		}
	}
	return false
}

func (r Rule) finding(p domain.Port) domain.Finding {
	return domain.Finding{
		Port:           p,
		Severity:       r.Severity,
		RuleID:         r.ID,
		Title:          r.Title,
		Recommendation: r.Recommendation,
	}
}
