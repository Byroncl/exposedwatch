package iptablessave

import "strings"

// needsRoot reports whether a failed save command complained about missing
// privileges. The message may arrive on either stream.
func needsRoot(result CommandResult) bool {
	text := string(result.Stderr) + "\n" + string(result.Stdout)
	for _, marker := range needsRootMarkers {
		if strings.Contains(text, marker) {
			return true
		}
	}
	return false
}

// boundedMessage trims text and cuts it to maxMessageLength characters.
func boundedMessage(text string) string {
	trimmed := strings.TrimSpace(text)
	if len(trimmed) > maxMessageLength {
		return trimmed[:maxMessageLength]
	}
	return trimmed
}
