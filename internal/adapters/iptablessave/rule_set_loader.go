// Package iptablessave reads the host firewall rules by running iptables-save
// and ip6tables-save and parsing their text output.
package iptablessave

import (
	"context"
	"errors"
	"fmt"

	"github.com/Byroncl/exposedwatch/internal/domain"
	"github.com/Byroncl/exposedwatch/internal/ports"
)

// RuleSetLoader implements ports.FirewallRuleSource over the iptables-save
// family of tools.
type RuleSetLoader struct {
	runner  CommandRunner
	locator BinaryLocator
}

// NewRuleSetLoader builds a loader from an injectable runner and locator.
func NewRuleSetLoader(runner CommandRunner, locator BinaryLocator) *RuleSetLoader {
	return &RuleSetLoader{runner: runner, locator: locator}
}

// NewDefaultRuleSetLoader returns the production loader.
func NewDefaultRuleSetLoader() *RuleSetLoader {
	return NewRuleSetLoader(NewExecRunner(), NewBinaryLocator())
}

// LoadRuleSet runs both save tools. A missing iptables-save is
// ports.ErrFirewallUnavailable and a permission failure is
// ports.ErrFirewallNeedsRoot. A problem with ip6tables-save alone only adds a
// note and leaves the IPv6 rules unavailable.
func (loader *RuleSetLoader) LoadRuleSet(ctx context.Context) (domain.RuleSet, error) {
	ruleSet := domain.RuleSet{
		Backend: domain.BackendUnknown,
		IPv4:    domain.FamilyRuleSet{Family: domain.FamilyIPv4},
		IPv6:    domain.FamilyRuleSet{Family: domain.FamilyIPv6},
	}
	ipv4Binary, err := loader.locator.Find(ipv4SaveBinaryName)
	if err != nil {
		return domain.RuleSet{}, fmt.Errorf("%w: %v", ports.ErrFirewallUnavailable, err)
	}
	ipv4Output, err := loader.runSave(ctx, ipv4Binary)
	if err != nil {
		return domain.RuleSet{}, err
	}
	ruleSet.IPv4.Available = true
	ruleSet.IPv4.Tables = ipv4Output.tables
	ruleSet.Notes = append(ruleSet.Notes, ipv4Output.notes...)
	if ipv4Output.version != "" {
		ruleSet.Backend = fmt.Sprintf(backendFormat, ipv4Output.version)
	}

	if err := loader.loadIPv6(ctx, &ruleSet); err != nil {
		return domain.RuleSet{}, err
	}
	if err := loader.probeNftables(ctx, &ruleSet); err != nil {
		return domain.RuleSet{}, err
	}
	return ruleSet, nil
}

// loadIPv6 fills the IPv6 half. Anything but a cancelled context only adds a
// note and leaves the IPv6 rules unavailable.
func (loader *RuleSetLoader) loadIPv6(ctx context.Context, ruleSet *domain.RuleSet) error {
	ipv6Binary, err := loader.locator.Find(ipv6SaveBinaryName)
	if err != nil {
		ruleSet.Notes = append(ruleSet.Notes, ipv6UnavailableNote)
		return nil
	}
	ipv6Output, err := loader.runSave(ctx, ipv6Binary)
	switch {
	case err == nil:
		ruleSet.IPv6.Available = true
		ruleSet.IPv6.Tables = ipv6Output.tables
	case errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded):
		return err
	default:
		ruleSet.Notes = append(ruleSet.Notes, ipv6FailedNote)
	}
	return nil
}

// runSave runs one save binary and parses its output.
func (loader *RuleSetLoader) runSave(ctx context.Context, binary string) (parsedOutput, error) {
	commandContext, cancel := context.WithTimeout(ctx, commandTimeout)
	defer cancel()
	result, err := loader.runner.Run(commandContext, binary, nil)
	if err != nil {
		return parsedOutput{}, err
	}
	if result.ExitCode != 0 {
		return parsedOutput{}, failureError(binary, result)
	}
	if result.Truncated {
		return parsedOutput{}, fmt.Errorf(truncatedFormat, ports.ErrFirewallOutputTruncated, binary)
	}
	parsed, err := parseSaveOutput(commandContext, result.Stdout)
	if err != nil {
		if errors.Is(err, ErrInvalidOutput) {
			return parsedOutput{}, fmt.Errorf(invalidOutputFormat, ports.ErrFirewallOutputInvalid, err)
		}
		return parsedOutput{}, err
	}
	return parsed, nil
}

func failureError(binary string, result CommandResult) error {
	detail := boundedMessage(string(result.Stderr))
	if detail == "" {
		detail = boundedMessage(string(result.Stdout))
	}
	if needsRoot(result) {
		return fmt.Errorf(needsRootFormat, ports.ErrFirewallNeedsRoot, detail)
	}
	return fmt.Errorf(commandFailedFormat, ErrCommandFailed, binary, result.ExitCode, detail)
}
