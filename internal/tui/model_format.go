package tui

import "fmt"

// Format keeps the model out of fmt output, whatever the verb: it holds the
// bot token of the settings form, and fmt cannot call methods of the
// unexported fields that hold it.
func (model Model) Format(state fmt.State, _ rune) {
	_, _ = fmt.Fprint(state, modelPlaceholder)
}
