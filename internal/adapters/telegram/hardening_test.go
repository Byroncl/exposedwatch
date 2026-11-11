package telegram

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Byroncl/exposedwatch/internal/ports"
)

const fragmentLength = 6

// leaksToken reports whether text contains the token, or any run of
// fragmentLength consecutive characters of it, compared case-insensitively,
// or an encoding of it.
func leaksToken(text string) bool {
	secret := fakeToken.Reveal()
	lowerText := strings.ToLower(text)
	lowerSecret := strings.ToLower(secret)
	for start := 0; start+fragmentLength <= len(lowerSecret); start++ {
		if strings.Contains(lowerText, lowerSecret[start:start+fragmentLength]) {
			return true
		}
	}
	encodings := []string{
		base64.StdEncoding.EncodeToString([]byte(secret)),
		base64.RawStdEncoding.EncodeToString([]byte(secret)),
		base64.URLEncoding.EncodeToString([]byte(secret)),
		base64.RawURLEncoding.EncodeToString([]byte(secret)),
		hex.EncodeToString([]byte(secret)),
	}
	for _, encoded := range encodings {
		if strings.Contains(lowerText, strings.ToLower(encoded)) {
			return true
		}
	}
	return false
}

func percentEncodeEverything(text string) string {
	var builder strings.Builder
	for index := 0; index < len(text); index++ {
		fmt.Fprintf(&builder, "%%%02x", text[index])
	}
	return builder.String()
}

func TestHostileDescriptionsNeverEchoTheToken(testRun *testing.T) {
	secret := fakeToken.Reveal()
	hostile := map[string]string{
		"lowercase percent escapes": "Bad Request " + percentEncodeEverything(secret),
		"uppercase token":           "Bad Request " + strings.ToUpper(secret),
		"std base64":                "Bad Request " + base64.StdEncoding.EncodeToString([]byte(secret)),
		"raw base64":                "Bad Request " + base64.RawStdEncoding.EncodeToString([]byte(secret)),
		"url base64":                "Bad Request " + base64.RawURLEncoding.EncodeToString([]byte(secret)),
		"hex":                       "Bad Request " + hex.EncodeToString([]byte(secret)),
		"six character fragment":    "Bad Request AA-fake_",
		"markup":                    "<b>Bad Request</b>",
		"too long":                  strings.Repeat("a", 101),
		"newline":                   "Bad Request\nCRITICAL: forged",
	}
	for name, description := range hostile {
		testRun.Run(name, func(testRun *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
				writer.WriteHeader(http.StatusBadRequest)
				_, _ = fmt.Fprintf(writer, `{"ok":false,"error_code":400,"description":%q}`, description)
			}))
			defer server.Close()
			err := newTestClient(server.URL, server.Client()).Send(context.Background(), sampleAlert())
			if err == nil || leaksToken(err.Error()) {
				testRun.Fatalf("error leaks or is missing: %v", err)
			}
			if !strings.Contains(err.Error(), "400") || strings.Contains(err.Error(), "Bad Request") || strings.Contains(err.Error(), "forged") {
				testRun.Fatalf("expected the generic message with the status only: %v", err)
			}
		})
	}
}

func TestNormalDescriptionsAreKept(testRun *testing.T) {
	for _, description := range []string{"Bad Request: chat not found", "Forbidden: bot was blocked by the user", "Unauthorized"} {
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
			writer.WriteHeader(http.StatusBadRequest)
			_, _ = fmt.Fprintf(writer, `{"ok":false,"error_code":400,"description":%q}`, description)
		}))
		err := newTestClient(server.URL, server.Client()).Send(context.Background(), sampleAlert())
		server.Close()
		if err == nil || !strings.Contains(err.Error(), description) {
			testRun.Errorf("description %q lost: %v", description, err)
		}
	}
}

func TestSecretHoldersNeverPrintTheToken(testRun *testing.T) {
	client := newTestClient("http://localhost", nil)
	redactor := newTokenRedactor(fakeToken)
	values := map[string]any{
		"client value":     *client,
		"client pointer":   client,
		"redactor value":   redactor,
		"redactor pointer": &redactor,
		"options value":    Options{Token: fakeToken, ChatID: fakeChatID},
		"options pointer":  &Options{Token: fakeToken, ChatID: fakeChatID},
	}
	for name, value := range values {
		for _, verb := range []string{"%v", "%+v", "%#v", "%s", "%q", "%x"} {
			if rendered := fmt.Sprintf(verb, value); leaksToken(rendered) {
				testRun.Errorf("%s with %s leaks the token: %.80q", name, verb, rendered)
			}
		}
	}
}

func TestFormatAlertSingleLineFieldsCannotInjectLinesOrSpoofText(testRun *testing.T) {
	alert := sampleAlert()
	alert.Process = "a\nCRITICAL: Fake"
	rightToLeftOverride := string(rune(0x202E))
	zeroWidthSpace := string(rune(0x200B))
	byteOrderMark := string(rune(0xFEFF))
	isolateStart := string(rune(0x2066))
	isolateEnd := string(rune(0x2069))
	alert.Host = "web" + rightToLeftOverride + "-01"
	alert.Title = "Data" + zeroWidthSpace + "base" + byteOrderMark + " exposed\r\n"
	alert.RuleID = "rule" + isolateStart + "x" + isolateEnd
	text := FormatAlert(alert)
	for _, forbidden := range []string{rightToLeftOverride, zeroWidthSpace, byteOrderMark, isolateStart, isolateEnd, "\r"} {
		if strings.Contains(text, forbidden) {
			testRun.Fatalf("%q survived in %q", forbidden, text)
		}
	}
	for _, line := range strings.Split(text, "\n") {
		if strings.HasPrefix(line, "CRITICAL") {
			testRun.Fatalf("injected line in %q", text)
		}
	}
	if got := strings.Count(text, "\n"); got != 6 {
		testRun.Fatalf("expected 7 lines, got %d newlines in %q", got, text)
	}
}

func TestRetryAfterExtremesNeverWaitOrOverflow(testRun *testing.T) {
	for _, retryAfter := range []string{"9223372037", "9223372036854775807", "-5", "0"} {
		testRun.Run(retryAfter, func(testRun *testing.T) {
			var requests atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
				requests.Add(1)
				writer.WriteHeader(http.StatusTooManyRequests)
				_, _ = fmt.Fprintf(writer, `{"ok":false,"parameters":{"retry_after":%s}}`, retryAfter)
			}))
			defer server.Close()
			waits := 0
			client := New(Options{
				BaseURL: server.URL, Token: fakeToken, ChatID: fakeChatID, HTTPClient: server.Client(),
				Wait: func(context.Context, time.Duration) error { waits++; return nil },
			})
			err := client.Send(context.Background(), sampleAlert())
			if !errors.Is(err, ports.ErrNotificationRateLimited) || waits != 0 || requests.Load() != 1 {
				testRun.Fatalf("err=%v waits=%d requests=%d", err, waits, requests.Load())
			}
		})
	}
}
