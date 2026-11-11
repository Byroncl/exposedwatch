package tui

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

// styles holds every Lip Gloss style the view uses.
type styles struct {
	critical lipgloss.Style
	high     lipgloss.Style
	medium   lipgloss.Style
	low      lipgloss.Style
	info     lipgloss.Style
	title    lipgloss.Style
	dim      lipgloss.Style
	warning  lipgloss.Style
	heading  lipgloss.Style
	selected lipgloss.Style
	plain    lipgloss.Style
}

func newStyles() styles {
	base := lipgloss.NewStyle().Inline(true)
	return styles{
		critical: base.Foreground(colorCritical).Bold(true),
		high:     base.Foreground(colorHigh),
		medium:   base.Foreground(colorMedium),
		low:      base.Foreground(colorLow),
		info:     base.Foreground(colorInfo),
		title:    base.Foreground(colorAccent).Bold(true),
		dim:      base.Foreground(colorDim),
		warning:  base.Foreground(colorWarning),
		heading:  base.Bold(true).Underline(true),
		selected: base.Reverse(true),
		plain:    base,
	}
}

// forSeverity returns the style that paints the given severity.
func (palette styles) forSeverity(severity domain.Severity) lipgloss.Style {
	switch severity {
	case domain.Critical:
		return palette.critical
	case domain.High:
		return palette.high
	case domain.Medium:
		return palette.medium
	case domain.Low:
		return palette.low
	default:
		return palette.info
	}
}
