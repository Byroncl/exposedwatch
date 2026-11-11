package tui

import "github.com/charmbracelet/lipgloss"

// Palette (ANSI 256 color codes). Every color is adaptive so the interface is
// readable on light and dark terminal backgrounds. lipgloss.AdaptiveColor is a
// struct and cannot be a constant, hence the variables.
var (
	colorCritical = lipgloss.AdaptiveColor{Light: "160", Dark: "203"}
	colorHigh     = lipgloss.AdaptiveColor{Light: "166", Dark: "209"}
	colorMedium   = lipgloss.AdaptiveColor{Light: "136", Dark: "220"}
	colorLow      = lipgloss.AdaptiveColor{Light: "31", Dark: "75"}
	colorInfo     = lipgloss.AdaptiveColor{Light: "244", Dark: "245"}
	colorSafe     = lipgloss.AdaptiveColor{Light: "28", Dark: "42"}

	colorAccent     = lipgloss.AdaptiveColor{Light: "25", Dark: "81"}
	colorWarning    = lipgloss.AdaptiveColor{Light: "130", Dark: "220"}
	colorDim        = lipgloss.AdaptiveColor{Light: "243", Dark: "245"}
	colorBorder     = lipgloss.AdaptiveColor{Light: "247", Dark: "240"}
	colorSelected   = lipgloss.AdaptiveColor{Light: "254", Dark: "237"}
	colorNewBadge   = lipgloss.AdaptiveColor{Light: "127", Dark: "213"}
	colorChipText   = lipgloss.AdaptiveColor{Light: "231", Dark: "232"}
	colorOpened     = lipgloss.AdaptiveColor{Light: "166", Dark: "209"}
	colorClosed     = lipgloss.AdaptiveColor{Light: "28", Dark: "42"}
	colorTabActive  = lipgloss.AdaptiveColor{Light: "25", Dark: "81"}
	colorHeadingRow = lipgloss.AdaptiveColor{Light: "238", Dark: "250"}
)
