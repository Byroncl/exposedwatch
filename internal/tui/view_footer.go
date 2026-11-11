package tui

import (
	"fmt"
	"strings"
)

func (model Model) footerLine(width int) string {
	if model.viewState == ViewFieldEdit {
		prompt := fmt.Sprintf(editHintFormat, model.settings.cursor.label())
		return model.styles.warning.Render(clampWidth(prompt, width))
	}
	if model.viewState == ViewFilterInput {
		prompt := fmt.Sprintf(filterPromptFormat, model.filterInput, filterCursor)
		return model.styles.warning.Render(clampWidth(prompt, width))
	}
	line := strings.Join(model.footerHintTexts(width), footerHintJoin)
	return model.styles.dim.Render(clampWidth(line, width))
}

// footerHintTexts lists the hints of the active view, dropping the least
// important ones first until the line fits the width.
func (model Model) footerHintTexts(width int) []string {
	active := make([]footerHint, 0, len(footerHints))
	for _, hint := range footerHints {
		if hint.shownOn(model) {
			active = append(active, hint)
		}
	}
	for {
		texts := make([]string, 0, len(active))
		for _, hint := range active {
			texts = append(texts, hint.textFor(model))
		}
		dropIndex := mostDroppable(active)
		if len(strings.Join(texts, footerHintJoin)) <= width || dropIndex < 0 {
			return texts
		}
		active = append(active[:dropIndex], active[dropIndex+1:]...)
	}
}

// mostDroppable returns the index of the hint with the highest drop rank, or -1.
func mostDroppable(hints []footerHint) int {
	best, bestRank := -1, dropRankNever
	for index, hint := range hints {
		if hint.DropRank > bestRank {
			best, bestRank = index, hint.DropRank
		}
	}
	return best
}

// textFor returns the hint text, built from the model when the hint is dynamic.
func (hint footerHint) textFor(model Model) string {
	if hint.Label != nil {
		return hint.Label(model)
	}
	return hint.Text
}
