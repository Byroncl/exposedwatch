package tui

import (
	"math"
	"regexp"
	"strconv"
	"strings"
)

// SGR parameters the parser understands.
const (
	sgrReset                = 0
	sgrBold                 = 1
	sgrForegroundFirst      = 30
	sgrForegroundLast       = 37
	sgrExtendedForeground   = 38
	sgrDefaultForeground    = 39
	sgrBackgroundFirst      = 40
	sgrBackgroundLast       = 47
	sgrExtendedBackground   = 48
	sgrDefaultBackground    = 49
	sgrBrightForeground     = 90
	sgrBrightForegroundLast = 97
	sgrBrightBackground     = 100
	sgrBrightBackgroundLast = 107
	sgrBrightOffset         = 8
	sgrModeIndexed          = 5
	sgrModeTrueColor        = 2
	sgrTrueColorWidth       = 3
	sgrParameterSplit       = ";"
)

// Relative luminance constants from WCAG 2.x.
const (
	channelMaximum       = 255.0
	linearThreshold      = 0.03928
	linearDivisor        = 12.92
	gammaOffset          = 0.055
	gammaScale           = 1.055
	gammaExponent        = 2.4
	redWeight            = 0.2126
	greenWeight          = 0.7152
	blueWeight           = 0.0722
	contrastFlare        = 0.05
	minimumContrastRatio = 3.0
)

var sgrSequencePattern = regexp.MustCompile("\x1b\\[([0-9;]*)m")

// optionalColor is a terminal color that may be unset.
type optionalColor struct {
	present bool
	color   rgbColor
}

// sgrLook is the resolved look of one printed cell.
type sgrLook struct {
	foreground optionalColor
	background optionalColor
	bold       bool
}

// styledRune is one printed rune and the look it was printed with.
type styledRune struct {
	glyph rune
	look  sgrLook
}

// styledRunes decodes a rendered line into printed runes with their looks.
func styledRunes(line string) []styledRune {
	var runes []styledRune
	var current sgrLook
	cursor := 0
	for _, bounds := range sgrSequencePattern.FindAllStringSubmatchIndex(line, -1) {
		for _, glyph := range line[cursor:bounds[0]] {
			runes = append(runes, styledRune{glyph: glyph, look: current})
		}
		current = applySGR(current, line[bounds[2]:bounds[3]])
		cursor = bounds[1]
	}
	for _, glyph := range line[cursor:] {
		runes = append(runes, styledRune{glyph: glyph, look: current})
	}
	return runes
}

// applySGR folds one SGR parameter list into a look.
func applySGR(look sgrLook, parameterList string) sgrLook {
	var codes []int
	for _, field := range strings.Split(parameterList, sgrParameterSplit) {
		code, err := strconv.Atoi(field)
		if err != nil {
			code = sgrReset
		}
		codes = append(codes, code)
	}
	for position := 0; position < len(codes); position++ {
		code := codes[position]
		switch {
		case code == sgrReset:
			look = sgrLook{}
		case code == sgrBold:
			look.bold = true
		case code >= sgrForegroundFirst && code <= sgrForegroundLast:
			look.foreground = indexedColor(code - sgrForegroundFirst)
		case code >= sgrBrightForeground && code <= sgrBrightForegroundLast:
			look.foreground = indexedColor(code - sgrBrightForeground + sgrBrightOffset)
		case code >= sgrBackgroundFirst && code <= sgrBackgroundLast:
			look.background = indexedColor(code - sgrBackgroundFirst)
		case code >= sgrBrightBackground && code <= sgrBrightBackgroundLast:
			look.background = indexedColor(code - sgrBrightBackground + sgrBrightOffset)
		case code == sgrDefaultForeground:
			look.foreground = optionalColor{}
		case code == sgrDefaultBackground:
			look.background = optionalColor{}
		case code == sgrExtendedForeground || code == sgrExtendedBackground:
			var resolved optionalColor
			resolved, position = extendedColor(codes, position)
			if code == sgrExtendedForeground {
				look.foreground = resolved
			} else {
				look.background = resolved
			}
		}
	}
	return look
}

// extendedColor reads a 38/48 color starting at position and returns the
// position of its last parameter.
func extendedColor(codes []int, position int) (optionalColor, int) {
	if position+1 >= len(codes) {
		return optionalColor{}, position
	}
	switch codes[position+1] {
	case sgrModeIndexed:
		if position+2 < len(codes) {
			return indexedColor(codes[position+2]), position + 2
		}
	case sgrModeTrueColor:
		if position+1+sgrTrueColorWidth < len(codes) {
			return optionalColor{present: true, color: rgbColor{codes[position+2], codes[position+3], codes[position+4]}}, position + 1 + sgrTrueColorWidth
		}
	}
	return optionalColor{}, position + 1
}

func indexedColor(index int) optionalColor {
	return optionalColor{present: true, color: colorFromAnsiIndex(index)}
}

func relativeLuminance(color rgbColor) float64 {
	linear := func(channel int) float64 {
		scaled := float64(channel) / channelMaximum
		if scaled <= linearThreshold {
			return scaled / linearDivisor
		}
		return math.Pow((scaled+gammaOffset)/gammaScale, gammaExponent)
	}
	return redWeight*linear(color.red) + greenWeight*linear(color.green) + blueWeight*linear(color.blue)
}

// contrastRatio is the WCAG contrast ratio between two colors.
func contrastRatio(first, second rgbColor) float64 {
	firstLuminance, secondLuminance := relativeLuminance(first), relativeLuminance(second)
	lighter, darker := math.Max(firstLuminance, secondLuminance), math.Min(firstLuminance, secondLuminance)
	return (lighter + contrastFlare) / (darker + contrastFlare)
}
