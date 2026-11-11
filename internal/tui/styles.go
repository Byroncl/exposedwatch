package tui

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

// styles holds every Lip Gloss style the view uses. All of them are inline:
// the view does its own padding and borders so widths stay exact.
type styles struct {
	critical    lipgloss.Style
	high        lipgloss.Style
	medium      lipgloss.Style
	low         lipgloss.Style
	info        lipgloss.Style
	safe        lipgloss.Style
	title       lipgloss.Style
	dim         lipgloss.Style
	warning     lipgloss.Style
	heading     lipgloss.Style
	selected    lipgloss.Style
	plain       lipgloss.Style
	border      lipgloss.Style
	panelTitle  lipgloss.Style
	newBadge    lipgloss.Style
	tabActive   lipgloss.Style
	tabInactive lipgloss.Style
	opened      lipgloss.Style
	closed      lipgloss.Style
	chipText    lipgloss.Style
}

func newStyles() styles {
	base := lipgloss.NewStyle().Inline(true)
	chip := base.Foreground(colorChipText).Bold(true)
	return styles{
		critical:    base.Foreground(colorCritical).Bold(true),
		high:        base.Foreground(colorHigh),
		medium:      base.Foreground(colorMedium),
		low:         base.Foreground(colorLow),
		info:        base.Foreground(colorInfo),
		safe:        base.Foreground(colorSafe),
		title:       base.Foreground(colorAccent).Bold(true),
		dim:         base.Foreground(colorDim),
		warning:     base.Foreground(colorWarning),
		heading:     base.Foreground(colorHeadingRow).Bold(true),
		selected:    base.Background(colorSelected).Bold(true),
		plain:       base,
		border:      base.Foreground(colorBorder),
		panelTitle:  base.Foreground(colorAccent).Bold(true),
		newBadge:    chip.Background(colorNewBadge),
		tabActive:   chip.Background(colorTabActive),
		tabInactive: base.Foreground(colorDim),
		opened:      base.Foreground(colorOpened).Bold(true),
		closed:      base.Foreground(colorClosed).Bold(true),
		chipText:    chip,
	}
}

// forSeverity returns the style that paints the given severity as text.
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

// chipFor returns the colored-background style of a severity chip.
func (palette styles) chipFor(severity domain.Severity) lipgloss.Style {
	return palette.chipText.Background(severityColor(severity))
}

func severityColor(severity domain.Severity) lipgloss.AdaptiveColor {
	switch severity {
	case domain.Critical:
		return colorCritical
	case domain.High:
		return colorHigh
	case domain.Medium:
		return colorMedium
	case domain.Low:
		return colorLow
	default:
		return colorInfo
	}
}
