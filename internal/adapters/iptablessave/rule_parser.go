package iptablessave

import (
	"fmt"
	"net/netip"
	"slices"
	"strings"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

// ruleParser reads the arguments of one -A line into a domain.Rule.
type ruleParser struct {
	rule     domain.Rule
	tokens   []string
	position int
	negated  bool
	stopped  bool
	// sawSinglePort and sawPortList record --dport and --dports, which
	// iptables ANDs when both are present.
	sawSinglePort bool
	sawPortList   bool
}

// parseRule parses the tokens of an "-A CHAIN ..." line. Anything the engine
// cannot evaluate marks the rule Unsupported; only structurally broken input
// is an error.
//
// When the arguments are broken, the rule read so far is returned together
// with the error so the caller can keep it as an unsupported rule.
func parseRule(text string, tokens []string) (domain.ChainName, domain.Rule, error) {
	parser := &ruleParser{rule: domain.Rule{Text: text}, tokens: tokens[appendRuleMinTokens:]}
	chain := domain.ChainName(tokens[1])
	for !parser.stopped && parser.position < len(parser.tokens) {
		if err := parser.readOption(); err != nil {
			return chain, parser.rule, err
		}
	}
	return chain, parser.rule, nil
}

func (parser *ruleParser) markUnsupported(format string, subject string) {
	if parser.rule.Unsupported {
		return
	}
	parser.rule.Unsupported = true
	parser.rule.UnsupportedReason = fmt.Sprintf(format, subject)
}

// argument returns the value of option. A "!" in the value position is the
// legacy placement of a negation and applies to the option.
func (parser *ruleParser) argument(option string) (string, error) {
	value, err := parser.rawArgument(option)
	if err == nil && value == negationToken {
		parser.negated = true
		return parser.rawArgument(option)
	}
	return value, err
}

// rawArgument returns the next token as the value of option without any
// negation handling (module names, comments, targets).
func (parser *ruleParser) rawArgument(option string) (string, error) {
	if parser.position >= len(parser.tokens) {
		return "", fmt.Errorf(missingArgumentDetail, option)
	}
	value := parser.tokens[parser.position]
	parser.position++
	return value, nil
}

// consumeNegation reports whether the option being read was negated with "!"
// and marks the rule unsupported when it was.
func (parser *ruleParser) consumeNegation(option string) bool {
	negated := parser.negated
	parser.negated = false
	if negated {
		parser.markUnsupported(unsupportedNegatedFormat, option)
	}
	return negated
}

func (parser *ruleParser) readOption() error {
	option := parser.tokens[parser.position]
	parser.position++
	switch {
	case option == negationToken:
		parser.negated = true
		return nil
	case option == moduleOption:
		return parser.readModule()
	case option == protocolOption:
		return parser.readProtocol()
	case option == sourceOption || option == sourceLongOption:
		return parser.readAddresses(option, &parser.rule.SourceCIDRs)
	case option == destinationOption || option == destinationLong:
		return parser.readAddresses(option, &parser.rule.DestinationCIDRs)
	case option == inInterfaceOption || option == inInterfaceLong:
		return parser.readInterface(option, &parser.rule.InInterface)
	case option == outInterfaceOption || option == outInterfaceLong:
		return parser.readInterface(option, &parser.rule.OutInterface)
	case option == destinationPortOpt || option == destinationPortLong ||
		option == destinationPortsOpt || option == destinationPortsLong:
		return parser.readDestinationPorts(option)
	case option == connectionStateOpt || option == legacyStateOpt:
		return parser.readConnectionStates(option)
	case option == destinationTypeOpt:
		return parser.readAddressTypes(option)
	case option == commentOption:
		_, err := parser.rawArgument(option)
		return err
	case option == synOption:
		return parser.readSyn(option)
	case option == tcpFlagsOption:
		return parser.readTCPFlags(option)
	case option == jumpOption || option == jumpLongOption:
		return parser.readTarget(option)
	case option == gotoOption || option == gotoLongOption:
		return parser.readGoto(option)
	default:
		return parser.readUnknownOption(option)
	}
}

func (parser *ruleParser) readModule() error {
	module, err := parser.rawArgument(moduleOption)
	if err != nil {
		return err
	}
	if !slices.Contains(supportedModules, module) {
		parser.markUnsupported(unsupportedModuleFormat, module)
	}
	return nil
}

func (parser *ruleParser) readProtocol() error {
	value, err := parser.argument(protocolOption)
	if err != nil {
		return err
	}
	if parser.consumeNegation(protocolOption) {
		return nil
	}
	value = strings.ToLower(value)
	switch value {
	case protocolAll, protocolNumberAll:
	case protocolNumberTCP:
		parser.rule.Protocol = domain.TCP
	case protocolNumberUDP:
		parser.rule.Protocol = domain.UDP
	default:
		if isNumeric(value) {
			parser.markUnsupported(unsupportedProtocolFormat, value)
			return nil
		}
		parser.rule.Protocol = domain.Protocol(value)
	}
	return nil
}

func (parser *ruleParser) readAddresses(option string, destination *[]netip.Prefix) error {
	value, err := parser.argument(option)
	if err != nil {
		return err
	}
	if parser.consumeNegation(option) {
		return nil
	}
	for _, part := range strings.Split(value, valueSeparator) {
		prefix, ok := parseAddressPrefix(part)
		if !ok {
			parser.markUnsupported(unsupportedAddressFormat, part)
			return nil
		}
		*destination = append(*destination, prefix)
	}
	return nil
}

// parseAddressPrefix accepts a CIDR or a bare address (treated as a host).
func parseAddressPrefix(text string) (netip.Prefix, bool) {
	if prefix, err := netip.ParsePrefix(text); err == nil {
		return prefix.Masked(), true
	}
	address, err := netip.ParseAddr(text)
	if err != nil {
		return netip.Prefix{}, false
	}
	return netip.PrefixFrom(address, address.BitLen()), true
}

func (parser *ruleParser) readInterface(option string, destination *domain.InterfaceName) error {
	value, err := parser.argument(option)
	if err != nil {
		return err
	}
	if parser.consumeNegation(option) {
		return nil
	}
	*destination = domain.InterfaceName(value)
	return nil
}

func (parser *ruleParser) readDestinationPorts(option string) error {
	value, err := parser.argument(option)
	if err != nil {
		return err
	}
	ranges, err := parsePortRanges(value)
	if err != nil {
		return err
	}
	if parser.consumeNegation(option) {
		return nil
	}
	if option == destinationPortOpt || option == destinationPortLong {
		parser.sawSinglePort = true
	} else {
		parser.sawPortList = true
	}
	if parser.sawSinglePort && parser.sawPortList {
		parser.markUnsupported("%s", unsupportedPortOptionsDetail)
	}
	parser.rule.DestinationPorts = append(parser.rule.DestinationPorts, ranges...)
	return nil
}

func (parser *ruleParser) readConnectionStates(option string) error {
	value, err := parser.argument(option)
	if err != nil {
		return err
	}
	if parser.consumeNegation(option) {
		return nil
	}
	for _, part := range strings.Split(value, valueSeparator) {
		state := domain.ConntrackState(strings.ToUpper(part))
		if !slices.Contains(knownConnectionStates, state) {
			parser.markUnsupported(unsupportedStateFormat, part)
			return nil
		}
		parser.rule.ConnectionStates = append(parser.rule.ConnectionStates, state)
	}
	return nil
}

func (parser *ruleParser) readAddressTypes(option string) error {
	value, err := parser.argument(option)
	if err != nil {
		return err
	}
	if parser.consumeNegation(option) {
		return nil
	}
	for _, part := range strings.Split(value, valueSeparator) {
		parser.rule.AddressTypes = append(parser.rule.AddressTypes, domain.AddressType(strings.ToUpper(part)))
	}
	return nil
}

func (parser *ruleParser) readSyn(option string) error {
	if parser.consumeNegation(option) {
		return nil
	}
	parser.rule.TCPFlags = &domain.TCPFlagsMatch{
		Mask: []domain.TCPFlag{domain.TCPFlagFin, domain.TCPFlagSyn, domain.TCPFlagRst, domain.TCPFlagAck},
		Set:  []domain.TCPFlag{domain.TCPFlagSyn},
	}
	return nil
}

func (parser *ruleParser) readTCPFlags(option string) error {
	maskText, err := parser.argument(option)
	if err != nil {
		return err
	}
	setText, err := parser.argument(option)
	if err != nil {
		return err
	}
	if parser.consumeNegation(option) {
		return nil
	}
	mask, maskOK := parseTCPFlags(maskText)
	set, setOK := parseTCPFlags(setText)
	if !maskOK || !setOK {
		parser.markUnsupported(unsupportedFlagFormat, maskText+" "+setText)
		return nil
	}
	parser.rule.TCPFlags = &domain.TCPFlagsMatch{Mask: mask, Set: set}
	return nil
}

func parseTCPFlags(text string) ([]domain.TCPFlag, bool) {
	switch text {
	case flagsAll:
		return slices.Clone(allTCPFlags), true
	case flagsNone:
		return nil, true
	}
	var flags []domain.TCPFlag
	for _, part := range strings.Split(text, valueSeparator) {
		flag := domain.TCPFlag(part)
		if !slices.Contains(allTCPFlags, flag) {
			return nil, false
		}
		flags = append(flags, flag)
	}
	return flags, true
}

func (parser *ruleParser) readTarget(option string) error {
	target, err := parser.rawArgument(option)
	if err != nil {
		return err
	}
	parser.rule.Target = domain.TargetName(target)
	// Everything after -j belongs to the target and never changes whether
	// the rule matches.
	parser.stopped = true
	return nil
}

func (parser *ruleParser) readGoto(option string) error {
	chain, err := parser.rawArgument(option)
	if err != nil {
		return err
	}
	parser.rule.GotoChain = domain.ChainName(chain)
	parser.stopped = true
	return nil
}

// readUnknownOption marks the rule unsupported and skips the option's values.
func (parser *ruleParser) readUnknownOption(option string) error {
	parser.markUnsupported(unsupportedOptionFormat, option)
	parser.negated = false
	for parser.position < len(parser.tokens) {
		next := parser.tokens[parser.position]
		if strings.HasPrefix(next, optionPrefix) || next == negationToken {
			break
		}
		parser.position++
	}
	return nil
}

func isNumeric(text string) bool {
	if text == "" {
		return false
	}
	for _, character := range text {
		if character < '0' || character > '9' {
			return false
		}
	}
	return true
}
