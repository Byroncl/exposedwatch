package app

import (
	"fmt"
	"net/netip"
	"strings"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

// describeRule builds a short explanation from the parsed fields of a rule.
// It never includes the raw rule line, so comments cannot leak into reasons.
func describeRule(rule domain.Rule) string {
	parts := []string{string(rule.Protocol)}
	if rule.Protocol == "" {
		parts[0] = ruleDescriptionAny
	}
	if len(rule.DestinationPorts) > 0 {
		parts = append(parts, fmt.Sprintf(ruleDestinationPortFormat, describePortRanges(rule.DestinationPorts)))
	}
	if len(rule.SourceCIDRs) > 0 {
		parts = append(parts, fmt.Sprintf(ruleSourceFormat, joinPrefixes(rule.SourceCIDRs)))
	}
	if len(rule.DestinationCIDRs) > 0 {
		parts = append(parts, fmt.Sprintf(ruleDestinationFormat, joinPrefixes(rule.DestinationCIDRs)))
	}
	if rule.InInterface != "" {
		parts = append(parts, fmt.Sprintf(ruleInterfaceFormat, string(rule.InInterface)))
	}
	parts = append(parts, fmt.Sprintf(ruleTargetFormat, describeTarget(rule)))
	return strings.Join(parts, ruleSeparator)
}

func describeTarget(rule domain.Rule) string {
	switch {
	case rule.JumpChain != "":
		return string(rule.JumpChain)
	case rule.GotoChain != "":
		return string(rule.GotoChain)
	default:
		return string(rule.Target)
	}
}

func describePortRanges(ranges []domain.PortRange) string {
	texts := make([]string, 0, len(ranges))
	for _, portRange := range ranges {
		if portRange.First == portRange.Last {
			texts = append(texts, fmt.Sprint(portRange.First))
			continue
		}
		texts = append(texts, fmt.Sprintf(portRangeFormat, portRange.First, portRange.Last))
	}
	return strings.Join(texts, valueJoiner)
}

func joinPrefixes(prefixes []netip.Prefix) string {
	texts := make([]string, 0, len(prefixes))
	for _, prefix := range prefixes {
		texts = append(texts, prefix.String())
	}
	return strings.Join(texts, valueJoiner)
}
