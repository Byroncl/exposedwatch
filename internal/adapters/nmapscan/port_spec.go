package nmapscan

import (
	"fmt"
	"strconv"
	"strings"
)

// PortSpec is a validated nmap port list such as "22,80,443,8000-8100".
// The empty PortSpec means "scan the most common ports".
type PortSpec string

// ParsePortSpec validates a port list: only digits, commas and hyphens, each
// item a port or an ascending range inside 1-65535.
func ParsePortSpec(raw string) (PortSpec, error) {
	if raw == "" {
		return "", nil
	}
	reject := func(reason string) (PortSpec, error) {
		return "", fmt.Errorf(invalidPortSpecFormat, ErrInvalidPortSpec, raw, reason)
	}
	if len(raw) > maxPortSpecLength {
		return reject(reasonTooLong)
	}
	for _, character := range raw {
		isDigit := character >= '0' && character <= '9'
		if !isDigit && character != ',' && character != '-' {
			return reject(reasonPortSpecCharsets)
		}
	}
	for _, item := range strings.Split(raw, portListSeparator) {
		if reason := validatePortItem(item); reason != "" {
			return reject(reason)
		}
	}
	return PortSpec(raw), nil
}

// validatePortItem returns an empty string when item is a port or a range.
func validatePortItem(item string) string {
	bounds := strings.Split(item, portRangeSeparator)
	if len(bounds) > portRangePartsCount {
		return reasonBadPortItem
	}
	numbers := make([]int, 0, len(bounds))
	for _, bound := range bounds {
		number, err := strconv.Atoi(bound)
		if err != nil {
			return reasonBadPortItem
		}
		if number < minPortNumber || number > maxPortNumber {
			return reasonPortOutOfRange
		}
		numbers = append(numbers, number)
	}
	if len(numbers) == portRangePartsCount && numbers[0] > numbers[1] {
		return reasonRangeOrder
	}
	return ""
}
