package app

import (
	"context"
	"fmt"
	"slices"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

// chainWalker walks the chains of one table for one probe, in rule order.
type chainWalker struct {
	ctx     context.Context
	visits  int
	table   domain.Table
	probe   connectionProbe
	outcome walkOutcome
	active  map[domain.ChainName]bool
	// noRules is true when the whole rule set holds no rule at all.
	noRules bool
}

// walkInputChain evaluates the INPUT chain of table and applies its policy.
func walkInputChain(ctx context.Context, table domain.Table, probe connectionProbe, noRules bool) walkOutcome {
	walker := &chainWalker{ctx: ctx, table: table, probe: probe, active: map[domain.ChainName]bool{}, noRules: noRules}
	input, found := table.Chain(domain.ChainInput)
	if !found {
		walker.outcome.markUncertain(reasonMissingInputChain)
		return walker.outcome
	}
	terminated := walker.walk(input, "", 0)
	if !terminated && walker.outcome.uncertainty == "" {
		walker.applyPolicy(input.Policy)
	}
	return walker.outcome
}

func (walker *chainWalker) applyPolicy(policy domain.TargetName) {
	switch policy {
	case domain.TargetAccept:
		reason := reasonPolicyAccept
		if walker.noRules {
			reason = reasonNoRulesAccept
		}
		walker.outcome.recordAccept("", reason)
	case domain.TargetDrop:
		reason := fmt.Sprintf(reasonPolicyBlockedFormat, policy, walker.probe.protocol, walker.probe.port)
		walker.outcome.recordDrop("", reason)
	default:
		walker.outcome.markUncertain(fmt.Sprintf(reasonPolicyUnsupportedFormat, string(policy)))
	}
}

// chainFlow tells the walk of a chain what to do after a rule was applied.
type chainFlow int

const (
	// flowContinue goes on with the next rule of the chain.
	flowContinue chainFlow = iota
	// flowReturn leaves the chain without a decision (RETURN, goto).
	flowReturn
	// flowDecided means every packet of the walked subset was decided, or the
	// evaluation became uncertain; the caller must stop too.
	flowDecided
)

// walk evaluates the rules of one chain. restriction describes the subset of
// sources the walk is limited to ("" for all sources). It reports whether
// every packet of that subset was decided inside the chain, which means the
// caller must not continue with its own later rules.
func (walker *chainWalker) walk(chain domain.Chain, restriction string, depth int) bool {
	walker.active[chain.Name] = true
	defer delete(walker.active, chain.Name)
	for _, rule := range chain.Rules {
		if walker.outcome.uncertainty != "" || walker.overBudget() {
			return true
		}
		match := matchRule(rule, walker.probe)
		if match.kind == matchNo {
			continue
		}
		if match.kind == matchMaybe {
			if isPassiveRule(rule) {
				continue
			}
			walker.outcome.markUncertain(match.uncertainty)
			return true
		}
		switch walker.applyRule(chain.Name, rule, match, restriction, depth) {
		case flowReturn:
			return false
		case flowDecided:
			return true
		}
	}
	return false
}

// isPassiveRule reports whether a rule can never change a packet's fate
// whether it matches or not: logging rules and counting rules.
func isPassiveRule(rule domain.Rule) bool {
	if rule.JumpChain != "" || rule.GotoChain != "" {
		return false
	}
	return rule.Target == "" || slices.Contains(passiveTargets, rule.Target)
}

// applyRule handles a rule that matches.
func (walker *chainWalker) applyRule(chain domain.ChainName, rule domain.Rule, match ruleMatch, restriction string, depth int) chainFlow {
	combined := combineRestrictions(restriction, match.restriction)
	ruleIsRestricted := match.restriction != ""
	switch {
	case rule.GotoChain != "":
		return walker.followChain(rule.GotoChain, combined, depth, ruleIsRestricted, true)
	case rule.JumpChain != "":
		return walker.followChain(rule.JumpChain, combined, depth, ruleIsRestricted, false)
	case rule.Target == domain.TargetAccept:
		walker.outcome.recordAccept(combined, fmt.Sprintf(reasonAllowedByRuleFormat, chain, describeRule(rule)))
		return decidedUnlessRestricted(ruleIsRestricted)
	case rule.Target == domain.TargetDrop || rule.Target == domain.TargetReject:
		walker.outcome.recordDrop(combined, fmt.Sprintf(reasonBlockedByRuleFormat, chain, describeRule(rule)))
		return decidedUnlessRestricted(ruleIsRestricted)
	case rule.Target == domain.TargetReturn:
		if ruleIsRestricted {
			walker.outcome.markUncertain(reasonRestrictedReturn)
			return flowDecided
		}
		return flowReturn
	case isPassiveRule(rule):
		return flowContinue
	default:
		walker.outcome.markUncertain(fmt.Sprintf(reasonUnknownTargetFormat, string(rule.Target)))
		return flowDecided
	}
}

func decidedUnlessRestricted(ruleIsRestricted bool) chainFlow {
	if ruleIsRestricted {
		return flowContinue
	}
	return flowDecided
}

// followChain enters a user chain through a jump or a goto.
func (walker *chainWalker) followChain(name domain.ChainName, restriction string, depth int, ruleIsRestricted bool, isGoto bool) chainFlow {
	target, found := walker.table.Chain(name)
	switch {
	case !found:
		walker.outcome.markUncertain(fmt.Sprintf(reasonUnknownChainFormat, string(name)))
		return flowDecided
	case walker.active[name]:
		walker.outcome.markUncertain(fmt.Sprintf(reasonCycleFormat, string(name)))
		return flowDecided
	case depth+1 > maxChainDepth:
		walker.outcome.markUncertain(fmt.Sprintf(reasonTooDeepFormat, string(name)))
		return flowDecided
	}
	terminated := walker.walk(target, restriction, depth+1)
	switch {
	case walker.outcome.uncertainty != "":
		return flowDecided
	case isGoto && ruleIsRestricted && !terminated:
		// Some sources leave this chain for good without a decision, so the
		// rules after this one only apply to the rest.
		walker.outcome.markUncertain(reasonRestrictedGoto)
		return flowDecided
	case isGoto && !ruleIsRestricted:
		// A goto never comes back: this chain is over either way.
		if terminated {
			return flowDecided
		}
		return flowReturn
	case terminated && !ruleIsRestricted:
		return flowDecided
	default:
		return flowContinue
	}
}

func combineRestrictions(context string, rule string) string {
	switch {
	case context == "":
		return rule
	case rule == "":
		return context
	default:
		return context + restrictionJoiner + rule
	}
}

// overBudget counts one rule visit and reports whether the evaluation must
// stop: the step budget is spent or the context was cancelled.
func (walker *chainWalker) overBudget() bool {
	walker.visits++
	if walker.visits > maxRuleVisits {
		walker.outcome.markUncertain(fmt.Sprintf(reasonStepBudgetFormat, maxRuleVisits))
		return true
	}
	if walker.visits%walkCancellationCheckInterval == 0 && walker.ctx.Err() != nil {
		walker.outcome.markUncertain(reasonCancelled)
		return true
	}
	return false
}
