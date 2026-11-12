package iptablessave

import (
	"context"
	"fmt"
	"strings"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

// parsedOutput is the structured content of one save command's output.
type parsedOutput struct {
	tables  []domain.Table
	version string
	notes   []string
}

// tableBuilder collects the chains of the table being read. Chains are
// indexed by name so large inputs stay linear.
type tableBuilder struct {
	name   domain.TableName
	chains []domain.Chain
	index  map[domain.ChainName]int
}

func newTableBuilder(name domain.TableName) *tableBuilder {
	return &tableBuilder{name: name, index: map[domain.ChainName]int{}}
}

func (builder *tableBuilder) chainIndex(name domain.ChainName) int {
	if position, found := builder.index[name]; found {
		return position
	}
	builder.chains = append(builder.chains, domain.Chain{Name: name})
	builder.index[name] = len(builder.chains) - 1
	return len(builder.chains) - 1
}

// finish turns jumps to user chains into JumpChain references.
func (builder *tableBuilder) finish() domain.Table {
	table := domain.Table{Name: builder.name, Chains: builder.chains}
	for chainIndex := range table.Chains {
		rules := table.Chains[chainIndex].Rules
		for ruleIndex := range rules {
			target := domain.ChainName(rules[ruleIndex].Target)
			if target == "" {
				continue
			}
			if _, isChain := builder.index[target]; isChain {
				rules[ruleIndex].JumpChain = target
				rules[ruleIndex].Target = ""
			}
		}
	}
	return table
}

// parseSaveOutput parses iptables-save / ip6tables-save text. Unknown lines
// and tables that are not closed by COMMIT are errors rather than being
// accepted partially, so truncated output is never trusted. A rule whose
// arguments cannot be read is kept as an unsupported rule.
func parseSaveOutput(ctx context.Context, output []byte) (parsedOutput, error) {
	var parsed parsedOutput
	var current *tableBuilder
	for lineIndex, rawLine := range strings.Split(string(output), "\n") {
		if lineIndex%parseCancellationCheckInterval == 0 {
			if err := ctx.Err(); err != nil {
				return parsedOutput{}, err
			}
		}
		line := strings.TrimSpace(rawLine)
		lineNumber := lineIndex + 1
		var err error
		switch {
		case line == "":
		case strings.HasPrefix(line, commentPrefix):
			readComment(line, &parsed)
		case strings.HasPrefix(line, tablePrefix):
			if current != nil {
				err = fmt.Errorf("%s", unterminatedTableDetail)
			}
			current = newTableBuilder(domain.TableName(strings.TrimPrefix(line, tablePrefix)))
		case line == commitLine:
			if current == nil {
				err = fmt.Errorf("%s", commitOutsideTableDetail)
				break
			}
			parsed.tables = append(parsed.tables, current.finish())
			current = nil
		case strings.HasPrefix(line, chainPrefix):
			err = readChainDeclaration(line, current)
		default:
			err = readRuleLine(line, current)
		}
		if err != nil {
			return parsedOutput{}, fmt.Errorf(invalidLineFormat, ErrInvalidOutput, lineNumber, err)
		}
	}
	if current != nil {
		return parsedOutput{}, fmt.Errorf(invalidLineFormat, ErrInvalidOutput, len(output), unterminatedTableDetail)
	}
	return parsed, nil
}

func readComment(line string, parsed *parsedOutput) {
	if strings.HasPrefix(line, generatedPrefix) {
		fields := strings.Fields(line)
		if len(fields) > generatedVersionPos && strings.HasPrefix(fields[generatedVersionPos], generatedVersionFlag) && parsed.version == "" {
			parsed.version = fields[generatedVersionPos]
		}
		return
	}
	if strings.Contains(line, legacyWarningMarker) {
		parsed.notes = append(parsed.notes, legacyTablesNote)
	}
}

func readChainDeclaration(line string, current *tableBuilder) error {
	fields := strings.Fields(strings.TrimPrefix(line, chainPrefix))
	if current == nil || len(fields) < minimumChainFields {
		return fmt.Errorf("%s", invalidChainDetail)
	}
	index := current.chainIndex(domain.ChainName(fields[0]))
	if fields[1] != userChainPolicy {
		current.chains[index].Policy = domain.TargetName(fields[1])
	}
	return nil
}

func readRuleLine(line string, current *tableBuilder) error {
	text := stripCounters(line)
	tokens, err := tokenizeRule(text)
	if err != nil {
		return err
	}
	if len(tokens) < appendRuleMinTokens || tokens[0] != appendRulePrefix {
		return fmt.Errorf("%s", unknownLineDetail)
	}
	if current == nil {
		return fmt.Errorf("%s", ruleOutsideTableDetail)
	}
	chain, rule, err := parseRule(text, tokens)
	if err != nil {
		rule.Unsupported = true
		rule.UnsupportedReason = fmt.Sprintf(unsupportedUnparseableFormat, err)
		if rule.Target == "" {
			rule.Target = domain.TargetUnparseable
		}
	}
	index := current.chainIndex(chain)
	current.chains[index].Rules = append(current.chains[index].Rules, rule)
	return nil
}

// stripCounters removes the optional "[packets:bytes] " prefix that
// iptables-save -c adds.
func stripCounters(line string) string {
	if strings.HasPrefix(line, counterPrefix) {
		if _, rest, found := strings.Cut(line, counterSuffix); found {
			return rest
		}
	}
	return line
}
