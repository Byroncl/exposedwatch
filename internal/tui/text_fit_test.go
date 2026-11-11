package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestFitWidthTruncatesWithEllipsis(testRun *testing.T) {
	cases := []struct {
		name  string
		text  string
		width int
		want  string
	}{
		{"exact fit", "abcde", 5, "abcde"},
		{"fits with padding", "abc", 5, "abc  "},
		{"one over", "abcdef", 5, "abcd…"},
		{"zero width", "abc", 0, ""},
		{"one cell", "abc", 1, "…"},
		{"one cell exact", "a", 1, "a"},
		{"wide unicode cut", "日本語日本", 6, "日本…"},
	}
	for _, testCase := range cases {
		got := fitWidth(testCase.text, testCase.width)
		if lipgloss.Width(got) > testCase.width {
			testRun.Errorf("%s: width %d exceeds %d", testCase.name, lipgloss.Width(got), testCase.width)
		}
		if got != testCase.want && testCase.name != "wide unicode cut" {
			testRun.Errorf("%s: got %q, want %q", testCase.name, got, testCase.want)
		}
	}
	got := strings.TrimRight(fitWidth("日本語日本", 6), " ")
	if lipgloss.Width(got) > 6 || !strings.HasSuffix(got, "…") {
		testRun.Errorf("wide unicode: got %q", got)
	}
}

func TestFitWidthKeepsEllipsisOutOfTextThatFits(testRun *testing.T) {
	if got := fitWidth("日本", 4); got != "日本" {
		testRun.Fatalf("got %q", got)
	}
}
