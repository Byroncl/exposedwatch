package iptablessave

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

// parsePortRanges reads a comma separated list of ports and first:last ranges.
func parsePortRanges(text string) ([]domain.PortRange, error) {
	var ranges []domain.PortRange
	for _, part := range strings.Split(text, valueSeparator) {
		first, last, isRange := strings.Cut(part, portRangeSeparator)
		firstPort, err := parsePort(first)
		if err != nil {
			return nil, err
		}
		lastPort := firstPort
		if isRange {
			if lastPort, err = parsePort(last); err != nil {
				return nil, err
			}
		}
		if lastPort < firstPort {
			return nil, fmt.Errorf(invalidPortDetail, part)
		}
		ranges = append(ranges, domain.PortRange{First: firstPort, Last: lastPort})
	}
	return ranges, nil
}

func parsePort(text string) (int, error) {
	port, err := strconv.Atoi(text)
	if err != nil || port < minimumPort || port > maximumPort {
		return 0, fmt.Errorf(invalidPortDetail, text)
	}
	return port, nil
}
