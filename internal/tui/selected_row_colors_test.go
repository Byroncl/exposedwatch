package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// selectedRowCase is one table whose first row is checked while selected.
var (
	selectionGlyph = []rune(selectionMark)[0]
	borderGlyph    = []rune(lipgloss.RoundedBorder().Left)[0]
)

type selectedRowCase struct {
	name string
	view moduleView
}

type colorProfileCase struct {
	name    string
	profile termenv.Profile
	dark    bool
}

func selectedRowCases() []selectedRowCase {
	return []selectedRowCase{
		{"findings", moduleView{ModulePorts, 0}},
		{"all ports", moduleView{ModulePorts, 1}},
		{"changes", moduleView{ModulePorts, 2}},
		{"connections live", moduleView{ModuleConnections, 0}},
		{"firewall ports", moduleView{ModuleFirewall, 1}},
	}
}

func colorProfileCases() []colorProfileCase {
	return []colorProfileCase{
		{"ansi dark", termenv.ANSI, true},
		{"ansi light", termenv.ANSI, false},
		{"ansi256 dark", termenv.ANSI256, true},
		{"ansi256 light", termenv.ANSI256, false},
		{"truecolor dark", termenv.TrueColor, true},
		{"truecolor light", termenv.TrueColor, false},
	}
}

// rowSpan is the rune range of a table row inside one rendered line.
type rowSpan struct {
	line  int
	start int
	end   int
}

// selectedRowSpan locates the table's selected row: the selection mark that
// sits furthest to the right, up to the panel padding before the border.
func selectedRowSpan(testRun *testing.T, view string) rowSpan {
	testRun.Helper()
	found := rowSpan{line: -1, start: -1}
	for lineIndex, line := range strings.Split(view, "\n") {
		for column, printed := range styledRunes(line) {
			if printed.glyph == selectionGlyph && column > found.start {
				found = rowSpan{line: lineIndex, start: column}
			}
		}
	}
	if found.line < 0 {
		testRun.Fatal("no selected row in view")
	}
	for column, printed := range styledRunes(strings.Split(view, "\n")[found.line]) {
		if printed.glyph == borderGlyph {
			found.end = column - 1
		}
	}
	return found
}

// rowRunes returns the printed runes of the span on its line of the view.
func rowRunes(view string, span rowSpan) []styledRune {
	return styledRunes(strings.Split(view, "\n")[span.line])[span.start:span.end]
}

func selectionBackground(testRun *testing.T) optionalColor {
	testRun.Helper()
	runes := styledRunes(lipgloss.NewStyle().Background(colorSelected).Render("x"))
	if len(runes) == 0 || !runes[0].look.background.present {
		testRun.Fatal("selection color did not render a background")
	}
	return runes[0].look.background
}

func TestSelectedRowKeepsChipColors(testRun *testing.T) {
	for _, profileCase := range colorProfileCases() {
		for _, tableCase := range selectedRowCases() {
			testRun.Run(profileCase.name+"/"+tableCase.name, func(testRun *testing.T) {
				forceColors(testRun, profileCase.profile, profileCase.dark)
				selectedModel := showView(testRun, previewModel(testRun), tableCase.view)
				if selectedModel.cursor != 0 {
					testRun.Fatalf("cursor = %d, want 0", selectedModel.cursor)
				}
				unselectedModel := typeKeys(testRun, selectedModel, "down")
				if unselectedModel.cursor == 0 {
					testRun.Fatal("cursor did not move off the first row")
				}
				selectedView := selectedModel.View()
				span := selectedRowSpan(testRun, selectedView)
				selectedRow := rowRunes(selectedView, span)
				unselectedRow := rowRunes(unselectedModel.View(), span)
				if len(selectedRow) != len(unselectedRow) {
					testRun.Fatalf("row widths differ: selected %d, unselected %d", len(selectedRow), len(unselectedRow))
				}
				assertChipsKeepTheirColors(testRun, selectedRow, unselectedRow)
				assertPlainCellsAreHighlighted(testRun, selectedRow, unselectedRow, selectionBackground(testRun))
			})
		}
	}
}

// assertChipsKeepTheirColors checks every cell that has its own background in
// the unselected row looks identical in the selected row and is legible.
func assertChipsKeepTheirColors(testRun *testing.T, selectedRow, unselectedRow []styledRune) {
	testRun.Helper()
	chipRunes := 0
	for index := 1; index < len(selectedRow); index++ {
		unselected, selected := unselectedRow[index], selectedRow[index]
		if !unselected.look.background.present {
			continue
		}
		chipRunes++
		if selected.look != unselected.look {
			testRun.Errorf("cell %d %q: selected look %+v differs from unselected %+v", index, string(selected.glyph), selected.look, unselected.look)
		}
		look := selected.look
		if !look.foreground.present || look.foreground.color == look.background.color {
			testRun.Errorf("cell %d %q: foreground %+v is not distinct from background %+v", index, string(selected.glyph), look.foreground, look.background)
			continue
		}
		if ratio := contrastRatio(look.foreground.color, look.background.color); ratio < minimumContrastRatio {
			testRun.Errorf("cell %d %q: contrast %.2f below %.1f", index, string(selected.glyph), ratio, minimumContrastRatio)
		}
	}
	if chipRunes == 0 {
		testRun.Fatal("the row has no chip to check")
	}
}

// assertPlainCellsAreHighlighted checks the cells and gaps that are not chips
// carry the selection background and stay readable when they have a color.
func assertPlainCellsAreHighlighted(testRun *testing.T, selectedRow, unselectedRow []styledRune, highlight optionalColor) {
	testRun.Helper()
	for index := 1; index < len(selectedRow); index++ {
		if unselectedRow[index].look.background.present {
			continue
		}
		look := selectedRow[index].look
		if look.background != highlight {
			testRun.Errorf("cell %d %q: background %+v, want the selection background %+v", index, string(selectedRow[index].glyph), look.background, highlight)
			continue
		}
		if look.foreground.present && look.foreground.color == highlight.color {
			testRun.Errorf("cell %d %q: text color equals the selection background", index, string(selectedRow[index].glyph))
		}
	}
}

// TestSelectionBackgroundIsReadable checks the dim and accent text colors keep
// enough contrast against the selection background in both themes.
func TestSelectionBackgroundIsReadable(testRun *testing.T) {
	for _, dark := range []bool{true, false} {
		forceColors(testRun, termenv.TrueColor, dark)
		highlight := selectionBackground(testRun)
		for name, text := range map[string]styleRenderer{"dim": newStyles().dim, "title": newStyles().title} {
			runes := styledRunes(text.Render("x"))
			if !runes[0].look.foreground.present {
				continue
			}
			if ratio := contrastRatio(runes[0].look.foreground.color, highlight.color); ratio < minimumContrastRatio {
				testRun.Errorf("dark=%v %s text contrast %.2f below %.1f", dark, name, ratio, minimumContrastRatio)
			}
		}
	}
}
