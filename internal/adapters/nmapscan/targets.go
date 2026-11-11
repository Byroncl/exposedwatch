package nmapscan

import (
	"fmt"
	"net/netip"
	"strings"
)

// Target is one validated scan target: an IP literal, a CIDR block or a
// conservative hostname.
type Target string

// TargetList is an ordered list of targets.
type TargetList []Target

// ParseTargets validates and normalizes raw user input. It is the only way
// untrusted text becomes a Target.
func ParseTargets(rawTargets []string) (TargetList, error) {
	if len(rawTargets) == 0 {
		return nil, ErrNoTargets
	}
	if len(rawTargets) > maxTargetCount {
		return nil, fmt.Errorf(invalidTargetFormat, ErrInvalidTarget, "", fmt.Sprintf(tooManyTargetsFormat, maxTargetCount))
	}
	targets := make(TargetList, 0, len(rawTargets))
	for _, rawTarget := range rawTargets {
		target, err := parseTarget(rawTarget)
		if err != nil {
			return nil, err
		}
		targets = append(targets, target)
	}
	return targets, nil
}

// Validate re-checks every entry; the scanner calls it again before building
// the command so a hand-built TargetList cannot bypass validation.
func (targets TargetList) Validate() error {
	if len(targets) == 0 {
		return ErrNoTargets
	}
	if len(targets) > maxTargetCount {
		return fmt.Errorf(invalidTargetFormat, ErrInvalidTarget, "", fmt.Sprintf(tooManyTargetsFormat, maxTargetCount))
	}
	for _, target := range targets {
		parsed, err := parseTarget(string(target))
		if err != nil {
			return err
		}
		if parsed != target {
			return fmt.Errorf(invalidTargetFormat, ErrInvalidTarget, string(target), reasonBadCharacters)
		}
	}
	return nil
}

func parseTarget(raw string) (Target, error) {
	reject := func(reason string) (Target, error) {
		return "", fmt.Errorf(invalidTargetFormat, ErrInvalidTarget, raw, reason)
	}
	switch {
	case strings.TrimSpace(raw) == "":
		return reject(reasonEmpty)
	case len(raw) > maxHostnameLength:
		return reject(reasonTooLong)
	case strings.HasPrefix(raw, "-"):
		return reject(reasonLeadingHyphen)
	}
	if strings.Contains(raw, "%") {
		return reject(reasonZoneNotAllowed)
	}
	if address, err := netip.ParseAddr(raw); err == nil {
		if address.Is6() {
			return reject(reasonIPv6NotSupported)
		}
		return Target(address.String()), nil
	}
	if strings.Contains(raw, "/") {
		return parsePrefixTarget(raw, reject)
	}
	return parseHostnameTarget(raw, reject)
}

func parsePrefixTarget(raw string, reject func(string) (Target, error)) (Target, error) {
	prefix, err := netip.ParsePrefix(raw)
	if err != nil {
		return reject(reasonBadCharacters)
	}
	if prefix.Addr().Is6() {
		return reject(reasonIPv6NotSupported)
	}
	if prefix.Bits() < minIPv4PrefixBits {
		return reject(reasonPrefixTooWide)
	}
	return Target(prefix.String()), nil
}

func parseHostnameTarget(raw string, reject func(string) (Target, error)) (Target, error) {
	hostname := strings.ToLower(raw)
	labels := strings.Split(hostname, hostnameSeparator)
	for _, label := range labels {
		if !isValidLabel(label) {
			return reject(reasonBadLabel)
		}
	}
	if !startsWithLetter(labels[len(labels)-1]) {
		return reject(reasonNumericHostname)
	}
	return Target(hostname), nil
}

func isValidLabel(label string) bool {
	if label == "" || len(label) > maxLabelLength {
		return false
	}
	if label[0] == labelHyphen || label[len(label)-1] == labelHyphen {
		return false
	}
	for _, character := range label {
		isLetter := character >= 'a' && character <= 'z'
		isDigit := character >= '0' && character <= '9'
		if !isLetter && !isDigit && character != labelHyphen {
			return false
		}
	}
	return true
}

// startsWithLetter reports whether the label begins with a letter. Real TLDs
// never start with a digit, which also rejects numeric forms such as
// 0x7f000001 or 127.1 that nmap's resolver would read as IPv4 addresses.
func startsWithLetter(label string) bool {
	first := label[0]
	return first >= 'a' && first <= 'z'
}
