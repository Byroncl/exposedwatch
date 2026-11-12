package app

import (
	"fmt"
	"net/netip"
	"slices"
	"strings"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

// matchKind is the three-valued outcome of matching one rule.
type matchKind int

const (
	matchNo matchKind = iota
	matchYes
	matchMaybe
)

// ruleMatch says whether a rule applies to the probe. restriction is
// non-empty when the rule only applies to a subset of sources or interfaces;
// uncertainty explains a matchMaybe.
type ruleMatch struct {
	kind        matchKind
	restriction string
	uncertainty string
}

// matchRule evaluates the supported parts of a rule against the probe. A
// definite mismatch always wins over uncertainty, so an unsupported match
// module on a rule for another port does not make this port unknown.
func matchRule(rule domain.Rule, probe connectionProbe) ruleMatch {
	if rule.Protocol != "" && rule.Protocol != probe.protocol {
		return ruleMatch{kind: matchNo}
	}
	if len(rule.DestinationPorts) > 0 && !containsPort(rule.DestinationPorts, probe.port) {
		return ruleMatch{kind: matchNo}
	}
	if isLoopbackInterface(rule.InInterface) || rule.OutInterface != "" {
		return ruleMatch{kind: matchNo}
	}
	if len(rule.ConnectionStates) > 0 && !slices.Contains(rule.ConnectionStates, domain.StateNew) {
		return ruleMatch{kind: matchNo}
	}
	if rule.TCPFlags != nil && !matchesSynPacket(*rule.TCPFlags) {
		return ruleMatch{kind: matchNo}
	}
	if len(rule.SourceCIDRs) > 0 && allLoopbackPrefixes(rule.SourceCIDRs) {
		return ruleMatch{kind: matchNo}
	}
	uncertainty := ""
	addressTypeKind, addressTypeUncertainty := matchAddressTypes(rule.AddressTypes)
	if addressTypeKind == matchNo {
		return ruleMatch{kind: matchNo}
	}
	if addressTypeKind == matchMaybe {
		uncertainty = addressTypeUncertainty
	}
	destinationKind := matchDestination(rule.DestinationCIDRs, probe)
	if destinationKind == matchNo {
		return ruleMatch{kind: matchNo}
	}
	if destinationKind == matchMaybe && uncertainty == "" {
		uncertainty = reasonDestinationDepends
	}
	if rule.Unsupported && uncertainty == "" {
		uncertainty = unsupportedUncertainty(rule)
	}
	if uncertainty != "" {
		return ruleMatch{kind: matchMaybe, uncertainty: uncertainty}
	}
	return ruleMatch{kind: matchYes, restriction: describeRestriction(rule)}
}

func unsupportedUncertainty(rule domain.Rule) string {
	return fmt.Sprintf(reasonUnsupportedFormat, rule.UnsupportedReason)
}

func containsPort(ranges []domain.PortRange, port int) bool {
	return slices.ContainsFunc(ranges, func(portRange domain.PortRange) bool {
		return portRange.Contains(port)
	})
}

func allLoopbackPrefixes(prefixes []netip.Prefix) bool {
	return !slices.ContainsFunc(prefixes, func(prefix netip.Prefix) bool {
		return !domain.IsLoopbackPrefix(prefix)
	})
}

// matchesSynPacket reports whether the first packet of a TCP connection (only
// SYN set) satisfies the flag match.
func matchesSynPacket(flags domain.TCPFlagsMatch) bool {
	for _, flag := range flags.Mask {
		shouldBeSet := flag == domain.TCPFlagSyn
		if slices.Contains(flags.Set, flag) != shouldBeSet {
			return false
		}
	}
	return true
}

// matchAddressTypes evaluates --dst-type for traffic addressed to this host.
func matchAddressTypes(types []domain.AddressType) (matchKind, string) {
	if len(types) == 0 || slices.Contains(types, domain.AddressTypeLocal) {
		return matchYes, ""
	}
	for _, addressType := range types {
		if !slices.Contains(nonUnicastAddressTypes, addressType) {
			return matchMaybe, fmt.Sprintf(reasonAddressTypeFormat, string(addressType))
		}
	}
	return matchNo, ""
}

// matchDestination compares -d with the bind address of the socket.
func matchDestination(destinations []netip.Prefix, probe connectionProbe) matchKind {
	if len(destinations) == 0 {
		return matchYes
	}
	if allLoopbackPrefixes(destinations) {
		return matchNo
	}
	if probe.hasBind {
		for _, prefix := range destinations {
			if prefix.Contains(probe.bind) {
				return matchYes
			}
		}
		return matchNo
	}
	for _, prefix := range destinations {
		if prefix.Bits() == fullPrefixBits {
			return matchYes
		}
	}
	return matchMaybe
}

// describeRestriction names the sources and interface a rule is limited to,
// or returns "" when the rule applies to every source.
func describeRestriction(rule domain.Rule) string {
	var parts []string
	if len(rule.SourceCIDRs) > 0 && !slices.ContainsFunc(rule.SourceCIDRs, func(prefix netip.Prefix) bool {
		return prefix.Bits() == fullPrefixBits
	}) {
		texts := make([]string, 0, len(rule.SourceCIDRs))
		for _, prefix := range rule.SourceCIDRs {
			texts = append(texts, prefix.String())
		}
		parts = append(parts, fmt.Sprintf(sourceRestrictionFormat, strings.Join(texts, valueJoiner)))
	}
	if rule.InInterface != "" {
		parts = append(parts, fmt.Sprintf(interfaceRestrictionFormat, string(rule.InInterface)))
	}
	return strings.Join(parts, restrictionJoiner)
}

// isLoopbackInterface recognises lo and wildcard patterns such as lo+.
func isLoopbackInterface(name domain.InterfaceName) bool {
	text := string(name)
	return text == loopbackInterface ||
		(strings.HasPrefix(text, loopbackPatternPrefix) && strings.HasSuffix(text, interfaceWildcardSuffix))
}
