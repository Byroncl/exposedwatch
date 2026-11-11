package tui

import "fmt"

func (model Model) helpLines(width int) []string {
	lines := []string{model.styles.heading.Render(helpTitle)}
	for _, binding := range keyBindings {
		lines = append(lines, model.styles.plain.MaxWidth(width).Render(
			fmt.Sprintf(helpKeyDescriptionFmt, binding.Keys, binding.Description)))
	}
	return append(lines, model.styles.dim.MaxWidth(width).Render(helpCloseHint))
}
