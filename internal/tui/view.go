package tui

import "strings"

// View renders the whole screen. It is pure: it only reads the model.
func (model Model) View() string {
	frame := model.computeLayout()
	headerPart := append(model.headerLines(frame.width), model.warningViewLines(frame)...)
	var bodyPart []string
	if model.viewState == ViewHelp {
		bodyPart = model.helpLines(frame.width)
	} else {
		bodyPart = model.tableLines(frame)
		if frame.showDetail {
			bodyPart = append(bodyPart, model.detailLines(frame.width)...)
		}
	}
	footerPart := []string{model.footerLine(frame.width)}
	return strings.Join(fitLines(headerPart, bodyPart, footerPart, frame.height), "\n")
}

func (model Model) warningViewLines(frame layout) []string {
	lines := make([]string, 0, len(frame.warnings))
	for _, warning := range frame.warnings {
		lines = append(lines, model.styles.warning.MaxWidth(frame.width).Render(warning))
	}
	return lines
}
