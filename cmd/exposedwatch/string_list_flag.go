package main

import "strings"

// stringListFlag collects values from a flag that may be repeated and whose
// values may be comma separated. It implements flag.Value.
type stringListFlag []string

// String implements flag.Value.
func (list *stringListFlag) String() string {
	return strings.Join(*list, listSeparator)
}

// Set implements flag.Value, splitting on commas. Empty items are kept so the
// validator can reject them instead of silently dropping them.
func (list *stringListFlag) Set(value string) error {
	*list = append(*list, strings.Split(value, listSeparator)...)
	return nil
}
