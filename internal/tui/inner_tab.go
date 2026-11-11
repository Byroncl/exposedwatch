package tui

// InnerTab is the position of an inner tab inside its module.
type InnerTab int

const (
	firstInnerTabKey = '1'
	lastInnerTabKey  = '9'
)

// parseInnerTabKey maps the number keys 1 to 9 to the inner tab positions 0 to 8.
func parseInnerTabKey(key string) (InnerTab, bool) {
	runes := []rune(key)
	if len(runes) != 1 || runes[0] < firstInnerTabKey || runes[0] > lastInnerTabKey {
		return 0, false
	}
	return InnerTab(runes[0] - firstInnerTabKey), true
}
