package telegram

import (
	"regexp"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

var ipv4Pattern = regexp.MustCompile(`\b\d{1,3}(\.\d{1,3}){3}\b`)

func TestFormatAlertStartsWithSeverityAndListsLocalFacts(testRun *testing.T) {
	text := FormatAlert(sampleAlert())
	lines := strings.Split(text, "\n")
	if !strings.HasPrefix(lines[0], "HIGH") {
		testRun.Fatalf("first line %q", lines[0])
	}
	for _, want := range []string{"web-01", "5432/tcp", "postgres", "database-exposed", "Bind the datastore"} {
		if !strings.Contains(text, want) {
			testRun.Errorf("missing %q in %q", want, text)
		}
	}
}

func TestFormatAlertContainsNoRemoteAddress(testRun *testing.T) {
	alert := sampleAlert()
	alert.Body = "Close the port."
	if ipv4Pattern.MatchString(FormatAlert(alert)) {
		testRun.Fatal("an IPv4 address appeared in a message built from local data")
	}
}

func TestFormatAlertIsBoundedWithEllipsis(testRun *testing.T) {
	alert := sampleAlert()
	alert.Body = strings.Repeat("é", 3*MaxMessageRunes)
	text := FormatAlert(alert)
	if utf8.RuneCountInString(text) != MaxMessageRunes {
		testRun.Fatalf("runes %d", utf8.RuneCountInString(text))
	}
	if !strings.HasSuffix(text, "…") || !utf8.ValidString(text) {
		testRun.Fatal("expected a valid string ending with an ellipsis")
	}
}

func TestFormatAlertStripsControlCharacters(testRun *testing.T) {
	alert := sampleAlert()
	alert.Process = "evil\x1b[31m\rname"
	text := FormatAlert(alert)
	if strings.ContainsAny(text, "\x1b\r") {
		testRun.Fatalf("control characters survived: %q", text)
	}
}

func TestFormatAlertSummaryAndTest(testRun *testing.T) {
	summary := FormatAlert(domain.Alert{
		Kind: domain.AlertKindSummary, Severity: domain.Critical, Title: "2 findings", Host: "web-01", Body: "- a\n- b",
	})
	if !strings.HasPrefix(summary, "CRITICAL") || !strings.Contains(summary, "- a\n- b") || strings.Contains(summary, "Port:") {
		testRun.Fatalf("summary %q", summary)
	}
	testText := FormatAlert(domain.Alert{Kind: domain.AlertKindTest, Severity: domain.Info, Title: "exposedwatch test message from web-01"})
	if !strings.HasPrefix(testText, "INFO") {
		testRun.Fatalf("test %q", testText)
	}
}
