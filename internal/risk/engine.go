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
	for _, port := range ports {
		if finding, found := evaluatePort(port); found {
			findings = append(findings, finding)
		}
	}
	sort.SliceStable(findings, func(left, right int) bool {
		leftFinding, rightFinding := findings[left], findings[right]
		if leftFinding.Severity != rightFinding.Severity {
			return leftFinding.Severity > rightFinding.Severity
		}
		return leftFinding.Port.Number < rightFinding.Port.Number
	})
	return findings
}

func evaluatePort(port domain.Port) (domain.Finding, bool) {
	exposure := port.Exposure()
	if exposure == domain.Loopback {
		return domain.Finding{}, false
	}
	for _, rule := range builtinRules {
		if rule.matches(port.Number) {
			if exposure >= rule.MinExposure {
				return rule.finding(port), true
			}
			return domain.Finding{}, false
		}
	}
	if exposure >= genericRule.MinExposure {
		return genericRule.finding(port), true
	}
	return domain.Finding{}, false
}

func (rule Rule) matches(portNumber int) bool {
	for _, candidatePort := range rule.Ports {
		if candidatePort == portNumber {
			return true
		}
	}
	return false
}

func (rule Rule) finding(port domain.Port) domain.Finding {
	return domain.Finding{
		Port:           port,
		Severity:       rule.Severity,
		RuleID:         rule.ID,
		Title:          rule.Title,
		Recommendation: rule.Recommendation,
	}
}
