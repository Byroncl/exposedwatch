package tui

// Reference values used to turn ANSI color indexes into RGB triples in tests.
const (
	ansiStandardColorCount = 16
	ansiCubeStart          = 16
	ansiGrayStart          = 232
	ansiCubeSide           = 6
	ansiCubeSideSquared    = ansiCubeSide * ansiCubeSide
	ansiGrayBase           = 8
	ansiGrayStep           = 10
)

// ansiCubeLevels are the six intensities of each channel of the 6x6x6 cube.
var ansiCubeLevels = [ansiCubeSide]int{0, 95, 135, 175, 215, 255}

// ansiStandardColors are the xterm values of the first 16 indexes.
var ansiStandardColors = [ansiStandardColorCount]rgbColor{
	{0, 0, 0}, {128, 0, 0}, {0, 128, 0}, {128, 128, 0},
	{0, 0, 128}, {128, 0, 128}, {0, 128, 128}, {192, 192, 192},
	{128, 128, 128}, {255, 0, 0}, {0, 255, 0}, {255, 255, 0},
	{0, 0, 255}, {255, 0, 255}, {0, 255, 255}, {255, 255, 255},
}

// rgbColor is a resolved terminal color.
type rgbColor struct {
	red, green, blue int
}

// colorFromAnsiIndex resolves one of the 256 indexed colors.
func colorFromAnsiIndex(index int) rgbColor {
	switch {
	case index < ansiCubeStart:
		return ansiStandardColors[index]
	case index < ansiGrayStart:
		offset := index - ansiCubeStart
		return rgbColor{
			red:   ansiCubeLevels[offset/ansiCubeSideSquared],
			green: ansiCubeLevels[(offset/ansiCubeSide)%ansiCubeSide],
			blue:  ansiCubeLevels[offset%ansiCubeSide],
		}
	default:
		level := ansiGrayBase + ansiGrayStep*(index-ansiGrayStart)
		return rgbColor{level, level, level}
	}
}
