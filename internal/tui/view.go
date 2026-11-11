package tui

import "strings"

// View renders the whole screen. It is pure: it only reads the model.
func (model Model) View() string {
	frame := model.computeLayout()
	headerPart := append(model.headerLines(frame.width), model.warningViewLines(frame)...)
	var bodyPart []string
	if model.viewState == ViewHelp {
		bodyPart = model.helpPanel(frame)
	} else {
		bodyPart = model.workArea(frame)
	}
	footerPart := []string{model.footerLine(frame.width)}
	fitted := fitLines(headerPart, bodyPart, footerPart, frame.height)
	for index, line := range fitted {
		fitted[index] = clampWidth(line, frame.width)
	}
	return strings.Join(fitted, "\n")
}

func (model Model) warningViewLines(frame layout) []string {
	lines := make([]string, 0, len(frame.warnings))
	for _, warning := range frame.warnings {
		lines = append(lines, model.styles.warning.Render(clampWidth(warning, frame.width)))
	}
	return lines
}
