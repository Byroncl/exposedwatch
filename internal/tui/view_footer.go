package tui

import (
	"fmt"
	"strings"
)

func (model Model) footerLine(width int) string {
	if model.viewState == ViewFilterInput {
		prompt := fmt.Sprintf(filterPromptFormat, model.filterInput, filterCursor)
		return model.styles.warning.MaxWidth(width).Render(prompt)
	}
	hints := make([]string, 0, len(keyBindings))
	for _, binding := range keyBindings {
		if binding.FooterHint != "" {
			hints = append(hints, binding.FooterHint)
		}
	}
	line := strings.Join(hints, countSeparator) + model.statusSuffix()
	return model.styles.dim.MaxWidth(width).Render(line)
}

func (model Model) statusSuffix() string {
	if model.filterText == "" {
		return fmt.Sprintf(statusFormatNoFilter, model.sortMode, model.minSeverity)
	}
	return fmt.Sprintf(statusFormat, model.sortMode, model.minSeverity, model.filterText)
}
