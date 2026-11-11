package main

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

func newTestLogger(output *bytes.Buffer) watchLogger {
	return watchLogger{
		writer:  output,
		clock:   &steppingClock{current: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)},
		secrets: []domain.BotToken{domain.BotToken(testSecretToken)},
	}
}

func TestLoggerWritesOneCleanLine(testRun *testing.T) {
	var output bytes.Buffer
	logger := newTestLogger(&output)
	hostile := "a\x1b[31m\x00b\u0085c" + string(rune(0x202E)) + "d" + string(rune(0x200B)) + "e\nCRITICAL: forged\r\tend " + testSecretToken
	logger.Logf("event: %s", hostile)
	line := output.String()
	if strings.Count(line, "\n") != 1 || !strings.HasSuffix(line, "\n") {
		testRun.Fatalf("not a single line: %q", line)
	}
	for _, character := range line[:len(line)-1] {
		if character < ' ' || character == 0x7f || character == 0x85 || character == 0x202E || character == 0x200B {
			testRun.Fatalf("control or invisible character %U in %q", character, line)
		}
	}
	if strings.Contains(line, testSecretToken) || !strings.Contains(line, "***") || !strings.HasPrefix(line, "2026-01-02T03:04:05Z ") {
		testRun.Fatalf("line %q", line)
	}
}

func TestLoggerNeverPrintsItsSecrets(testRun *testing.T) {
	var output bytes.Buffer
	logger := newTestLogger(&output)
	values := map[string]any{"value": logger, "pointer": &logger}
	for name, value := range values {
		for _, verb := range []string{"%v", "%+v", "%#v", "%s", "%q", "%x"} {
			if rendered := fmt.Sprintf(verb, value); strings.Contains(rendered, "AA-fake") || strings.Contains(rendered, "123456789") ||
				strings.Contains(strings.ToLower(rendered), "6e7a6f") {
				testRun.Errorf("%s with %s leaks: %.60q", name, verb, rendered)
			}
		}
	}
}
