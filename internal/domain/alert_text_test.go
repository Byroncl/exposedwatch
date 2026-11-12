package domain

import "testing"

func TestSanitizeLineRemovesInvisibleAndControlCharacters(testRun *testing.T) {
	override := string(rune(0x202E))
	zeroWidth := string(rune(0x200B))
	testCases := []struct {
		name string
		text string
		want string
	}{
		{"bidi override", "a" + override + "b", "ab"},
		{"zero width", "a" + zeroWidth + "b" + string(rune(0xFEFF)), "ab"},
		{"line break becomes a space", "a\nCRITICAL: Fake", "a CRITICAL: Fake"},
		{"other controls are dropped", "a\x00b\x1b[31m", "ab[31m"},
		{"ends are trimmed", "  a\r\n", "a"},
	}
	for _, testCase := range testCases {
		if got := SanitizeLine(testCase.text); got != testCase.want {
			testRun.Errorf("%s: got %q want %q", testCase.name, got, testCase.want)
		}
	}
}

func TestSanitizeBlockKeepsNewlinesOnly(testRun *testing.T) {
	if got := SanitizeBlock("a\r\nb\x00" + string(rune(0x202E)) + "c\td"); got != "a\nbc d" {
		testRun.Fatalf("got %q", got)
	}
}
