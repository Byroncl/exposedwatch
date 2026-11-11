package tui

import "fmt"

// helpPanel lists every key and the meaning of the header counters.
func (model Model) helpPanel(frame layout) []string {
	lines := make([]string, 0, len(keyBindings)+2)
	for _, binding := range keyBindings {
		lines = append(lines, fmt.Sprintf(helpKeyDescriptionFmt, binding.Keys, binding.Description))
	}
	lines = append(lines, model.styles.dim.Render(helpCountersNote), model.styles.dim.Render(helpCloseHint))
	return model.styles.renderPanel(panelHelpTitle, lines, frame.width)
}
