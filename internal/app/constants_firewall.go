package app

import "github.com/Byroncl/exposedwatch/internal/domain"

// maxChainDepth bounds how deep user chains may be nested before the
// evaluation gives up with an unknown verdict.
const maxChainDepth = 16

// Walk limits. The step budget bounds how many rule visits one evaluation may
// make, so a hostile fan-out of chain jumps cannot run for minutes.
const (
	maxRuleVisits                 = 50000
	walkCancellationCheckInterval = 100
)

// dockerNatChains are the only nat chains whose DNAT rules are expected: the
// ones Docker itself creates.
var dockerNatChains = []domain.ChainName{
	"DOCKER", "DOCKER-INGRESS", "DOCKER_OUTPUT", "DOCKER_POSTROUTING",
}

// Joiners used to build explanations.
const (
	restrictionJoiner = " and "
	reasonJoiner      = "; "
	valueJoiner       = ","
)

// Explanation texts. Reasons are short and stable so a UI can show them as is.
const (
	reasonNoRulesAccept           = "no firewall rules (policy ACCEPT)"
	reasonPolicyAccept            = "allowed by INPUT policy ACCEPT"
	reasonPolicyBlockedFormat     = "INPUT policy %s and no rule allows %s/%d"
	reasonAllowedByRuleFormat     = "allowed by rule in %s: %s"
	reasonBlockedByRuleFormat     = "blocked by rule in %s: %s"
	reasonRestrictedFormat        = "restricted to %s"
	reasonBlockedForFormat        = "blocked for %s, allowed otherwise"
	reasonUnsupportedFormat       = "contains rules the engine cannot evaluate (%s)"
	reasonDestinationDepends      = "a rule depends on the destination address of a wildcard listener"
	reasonAddressTypeFormat       = "a rule matches address type %s, which cannot be evaluated"
	reasonUnknownTargetFormat     = "a rule uses target %s, which cannot be evaluated"
	reasonRestrictedReturn        = "a rule returns only for some sources, so later rules cannot be evaluated"
	reasonRestrictedGoto          = "a rule jumps for good for only some sources, so later rules cannot be evaluated"
	reasonCycleFormat             = "firewall chains form a cycle at %s"
	reasonTooDeepFormat           = "firewall chains are nested too deep at %s"
	reasonUnknownChainFormat      = "a rule jumps to unknown chain %s"
	reasonMissingInputChain       = "no INPUT chain found in the filter table"
	reasonNoFilterTable           = "no filter table found in the iptables rules"
	reasonNoTables                = "no iptables rules found; native nftables rules are not evaluated yet"
	reasonPolicyUnsupportedFormat = "INPUT chain policy %q cannot be evaluated"
	reasonPreFilterFormat         = "rules in the %s table can drop or redirect traffic before the INPUT chain"
	reasonNeedsRoot               = "needs root to read firewall rules"
	reasonNoTool                  = "no iptables tool found"
	reasonFamilyUnavailableFormat = "%s rules are not available"
	reasonLoopback                = "bound to loopback only: not reachable from other hosts"
	reasonDockerPublished         = "published by Docker: traffic usually bypasses INPUT/ufw rules; DOCKER-USER or FORWARD rules may still apply"
	reasonFamilyFormat            = "%s: %s"
	sourceRestrictionFormat       = "source %s"
	interfaceRestrictionFormat    = "interface %s"
	ipv4ToolName                  = "iptables"
	ipv6ToolName                  = "ip6tables"
	failureFormat                 = "app: read firewall rules: %w"
)

// loopbackInterface is the only interface name the engine knows is loopback.
const (
	loopbackInterface       = "lo"
	loopbackPatternPrefix   = "lo"
	interfaceWildcardSuffix = "+"
)

// Targets that never decide the fate of a packet and only let evaluation
// continue.
var passiveTargets = []domain.TargetName{
	domain.TargetLog, domain.TargetNflog, domain.TargetUlog,
}

// Targets that can drop, divert or reroute a packet in tables evaluated
// before the filter table.
var preFilterTargets = []domain.TargetName{
	domain.TargetDrop, domain.TargetReject, domain.TargetNfqueue, domain.TargetQueue, domain.TargetTproxy,
	domain.TargetCt, domain.TargetNotrack, domain.TargetSynproxy, domain.TargetSecmark, domain.TargetConnsecmark,
}

// Targets that can change the destination port of a packet in the nat table.
var rewritingTargets = []domain.TargetName{
	domain.TargetDnat, domain.TargetRedirect, domain.TargetTproxy,
}

// Built-in chains that never see inbound traffic before the INPUT chain.
var nonInboundChains = []domain.ChainName{
	domain.ChainForward, domain.ChainOutput, domain.ChainPostrouting,
}

// Address types that can never be the destination of a unicast inbound
// connection to this host.
var nonUnicastAddressTypes = []domain.AddressType{
	domain.AddressTypeMulticast, domain.AddressTypeBroadcast, domain.AddressTypeAnycast,
}

// fullPrefixBits is the prefix length of a rule that covers every address.
const fullPrefixBits = 0

// reasonNoDecision is only reachable if a walk ends without any decision,
// which the evaluator never lets happen; it keeps the verdict honest anyway.
const reasonNoDecision = "no firewall decision could be determined"

// More explanation texts.
const (
	reasonStepBudgetFormat    = "too many rules to evaluate (limit %d)"
	reasonCancelled           = "evaluation cancelled"
	reasonOutputTruncated     = "the firewall rules output was truncated, so the rules cannot be trusted"
	reasonOutputInvalid       = "the firewall rules could not be understood"
	reasonNativeNftables      = "native nftables rules present and not evaluated yet"
	reasonNftablesCheckFailed = "could not check nftables"
	ruleDescriptionAny        = "any"
	ruleDestinationPortFormat = "dport %s"
	ruleSourceFormat          = "from %s"
	ruleDestinationFormat     = "to %s"
	ruleInterfaceFormat       = "in %s"
	ruleTargetFormat          = "-> %s"
	ruleSeparator             = " "
	portRangeFormat           = "%d:%d"
)
