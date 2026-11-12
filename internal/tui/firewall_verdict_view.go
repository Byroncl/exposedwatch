package tui

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

// firewallVerdictOrder lists the verdicts from the most worrying to the least,
// which is also the default order of the Firewall ports table.
var firewallVerdictOrder = []domain.FirewallVerdict{
	domain.FirewallAllowed, domain.FirewallRestricted, domain.FirewallUnknown,
	domain.FirewallDockerPublished, domain.FirewallBlocked, domain.FirewallNotExposed,
}

// firewallWord is the short word shown for a verdict.
func firewallWord(verdict domain.FirewallVerdict) string {
	switch verdict {
	case domain.FirewallAllowed:
		return firewallWordAllowed
	case domain.FirewallRestricted:
		return firewallWordRestricted
	case domain.FirewallBlocked:
		return firewallWordBlocked
	case domain.FirewallNotExposed:
		return firewallWordNotExposed
	case domain.FirewallDockerPublished:
		return firewallWordDockerShort
	default:
		return firewallWordUnknown
	}
}

// firewallRank orders verdicts: the lower, the more worrying.
func firewallRank(verdict domain.FirewallVerdict) int {
	for rank, candidate := range firewallVerdictOrder {
		if candidate == verdict {
			return rank
		}
	}
	return len(firewallVerdictOrder)
}

// firewallColor paints a verdict chip. Blocked is the calm one, allowed the loud one.
func firewallColor(verdict domain.FirewallVerdict) lipgloss.AdaptiveColor {
	switch verdict {
	case domain.FirewallAllowed:
		return colorHigh
	case domain.FirewallRestricted:
		return colorMedium
	case domain.FirewallBlocked:
		return colorSafe
	case domain.FirewallNotExposed:
		return colorInfo
	case domain.FirewallDockerPublished:
		return colorLow
	default:
		return colorWarning
	}
}

// chipForVerdict is the colored-background style of a verdict chip.
func (palette styles) chipForVerdict(verdict domain.FirewallVerdict) lipgloss.Style {
	return palette.chipText.Background(firewallColor(verdict))
}

// countFirewallVerdicts counts the findings per verdict.
func countFirewallVerdicts(findings []domain.FirewallFinding) map[domain.FirewallVerdict]int {
	counts := make(map[domain.FirewallVerdict]int, len(firewallVerdictOrder))
	for _, finding := range findings {
		counts[finding.Verdict]++
	}
	return counts
}
