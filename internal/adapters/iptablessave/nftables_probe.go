package iptablessave

import (
	"context"
	"errors"
	"slices"
	"strings"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

// probeNftables learns whether native nftables rules exist next to the
// iptables ones. A native ruleset can only add drops on top of iptables, so
// the result decides how far Allowed and Restricted verdicts can be trusted.
// Only a cancelled context is an error: a missing or failing nft is recorded
// as state and note.
func (loader *RuleSetLoader) probeNftables(ctx context.Context, ruleSet *domain.RuleSet) error {
	binary, err := loader.locator.Find(nftBinaryName)
	if err != nil {
		ruleSet.Nftables = domain.NftablesNone
		ruleSet.Notes = append(ruleSet.Notes, nftAbsentNote)
		return nil
	}
	commandContext, cancel := context.WithTimeout(ctx, commandTimeout)
	defer cancel()
	result, err := loader.runner.Run(commandContext, binary, nftListTablesArguments)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}
		return recordNftablesFailure(ruleSet)
	}
	if result.ExitCode != 0 || result.Truncated {
		return recordNftablesFailure(ruleSet)
	}
	state, understood := classifyNftTables(string(result.Stdout))
	if !understood {
		return recordNftablesFailure(ruleSet)
	}
	ruleSet.Nftables = state
	if state == domain.NftablesNative {
		ruleSet.Notes = append(ruleSet.Notes, nftNativeNote)
	}
	return nil
}

func recordNftablesFailure(ruleSet *domain.RuleSet) error {
	ruleSet.Nftables = domain.NftablesProbeFailed
	ruleSet.Notes = append(ruleSet.Notes, nftCheckFailedNote)
	return nil
}

// classifyNftTables reads `nft list tables` output. A table is native when
// its family is not ip or ip6, or when its name is not one that iptables-nft
// creates. Known gap: a hand-made native `table ip filter` looks exactly like
// the iptables-nft one and is not detected. The second result is false when
// the output is not understood.
func classifyNftTables(output string) (domain.NftablesState, bool) {
	state := domain.NftablesNone
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != nftTableFields || fields[0] != nftTableKeyword {
			return domain.NftablesProbeFailed, false
		}
		family := fields[nftFamilyPosition]
		name := fields[nftNamePosition]
		isIPFamily := family == nftFamilyIPv4 || family == nftFamilyIPv6
		if !isIPFamily || !slices.Contains(iptablesNftTableNames, name) {
			state = domain.NftablesNative
		}
	}
	return state, true
}
