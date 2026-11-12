package telegram

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Byroncl/exposedwatch/internal/domain"
	"github.com/Byroncl/exposedwatch/internal/ports"
)

const (
	fakeToken  = domain.BotToken("123456789:AA-fake_TOKEN+xyz")
	fakeChatID = domain.ChatID("-1009876543210")
)

func sampleAlert() domain.Alert {
	return domain.Alert{
		Kind:     domain.AlertKindNewFinding,
		Severity: domain.High,
		Title:    "Database exposed",
		Body:     "Bind the datastore to 127.0.0.1.",
		Host:     "web-01",
		Port:     5432,
		Protocol: domain.TCP,
		Process:  "postgres",
		RuleID:   "database-exposed",
	}
}

func newTestClient(baseURL string, httpClient *http.Client) *Client {
	return New(Options{
		BaseURL:    baseURL,
		Token:      fakeToken,
		ChatID:     fakeChatID,
		HTTPClient: httpClient,
		Wait:       func(context.Context, time.Duration) error { return nil },
	})
}

func assertNoToken(testRun *testing.T, err error) {
	testRun.Helper()
	if err == nil {
		testRun.Fatal("expected an error")
	}
	secret := fakeToken.Reveal()
	for _, variant := range []string{secret, "AA-fake_TOKEN", "123456789"} {
		if strings.Contains(err.Error(), variant) {
			testRun.Fatalf("error leaks the token (variant length %d)", len(variant))
		}
	}
}

func TestSendPostsPlainTextJSON(testRun *testing.T) {
	var gotPath string
	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		gotPath = request.URL.Path
		if request.Method != http.MethodPost {
			testRun.Errorf("method %s", request.Method)
		}
		if err := json.NewDecoder(request.Body).Decode(&gotBody); err != nil {
			testRun.Errorf("decode: %v", err)
		}
		_, _ = writer.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	if err := newTestClient(server.URL, server.Client()).Send(context.Background(), sampleAlert()); err != nil {
		testRun.Fatalf("send: %v", err)
	}
	if gotPath != "/bot"+fakeToken.Reveal()+"/sendMessage" {
		testRun.Fatalf("unexpected path shape")
	}
	if gotBody["chat_id"] != string(fakeChatID) || gotBody["disable_web_page_preview"] != true {
		testRun.Fatalf("body %v", gotBody)
	}
	if _, present := gotBody["parse_mode"]; present {
		testRun.Fatal("parse_mode must not be sent")
	}
	if text, _ := gotBody["text"].(string); !strings.HasPrefix(text, "HIGH") {
		testRun.Fatalf("text %q", text)
	}
}

func TestSendFailuresNeverLeakTheToken(testRun *testing.T) {
	echoServer := func(status int, body string) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
			writer.WriteHeader(status)
			_, _ = writer.Write([]byte(body))
		}))
	}
	tokenText := fakeToken.Reveal()
	testCases := []struct {
		name   string
		status int
		body   string
		target error
	}{
		{"server error without json", 500, "boom " + tokenText, ports.ErrNotificationRejected},
		{"ok false echoing the token", 400, `{"ok":false,"description":"bad bot` + tokenText + `/sendMessage"}`, ports.ErrNotificationRejected},
		{"ok false with bot path", 401, `{"ok":false,"description":"Unauthorized: /bot` + tokenText + `/sendMessage"}`, ports.ErrNotificationRejected},
		{"status 200 but ok false", 200, `{"ok":false,"description":"` + tokenText + `"}`, ports.ErrNotificationRejected},
		{"not json at all", 200, "<html>" + tokenText + "</html>", ports.ErrNotificationRejected},
	}
	for _, testCase := range testCases {
		testRun.Run(testCase.name, func(testRun *testing.T) {
			server := echoServer(testCase.status, testCase.body)
			defer server.Close()
			err := newTestClient(server.URL, server.Client()).Send(context.Background(), sampleAlert())
			assertNoToken(testRun, err)
			if !errors.Is(err, testCase.target) {
				testRun.Fatalf("want %v, got %v", testCase.target, err)
			}
		})
	}
}

func TestSendTransportFailuresNeverLeakTheToken(testRun *testing.T) {
	closedServer := httptest.NewServer(http.NotFoundHandler())
	closedURL := closedServer.URL
	closedServer.Close()

	slowServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		time.Sleep(300 * time.Millisecond)
	}))
	defer slowServer.Close()

	canceledContext, cancel := context.WithCancel(context.Background())
	cancel()

	testCases := []struct {
		name       string
		baseURL    string
		httpClient *http.Client
		ctx        context.Context
		target     error
	}{
		{"connection refused", closedURL, nil, context.Background(), ports.ErrNotificationFailed},
		{"client timeout", slowServer.URL, &http.Client{Timeout: 50 * time.Millisecond}, context.Background(), ports.ErrNotificationFailed},
		{"canceled context", slowServer.URL, nil, canceledContext, context.Canceled},
		{"malformed base url", "http://[::1", nil, context.Background(), ports.ErrNotificationFailed},
		{"base url without scheme", "://nope", nil, context.Background(), ports.ErrNotificationFailed},
		{"unsupported scheme", "ftp://example.invalid", nil, context.Background(), ports.ErrNotificationFailed},
	}
	for _, testCase := range testCases {
		testRun.Run(testCase.name, func(testRun *testing.T) {
			err := newTestClient(testCase.baseURL, testCase.httpClient).Send(testCase.ctx, sampleAlert())
			assertNoToken(testRun, err)
			if !errors.Is(err, testCase.target) {
				testRun.Fatalf("want %v, got %v", testCase.target, err)
			}
		})
	}
}

func TestSendDoesNotFollowRedirects(testRun *testing.T) {
	var targetHits atomic.Int32
	target := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		targetHits.Add(1)
	}))
	defer target.Close()
	redirector := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		http.Redirect(writer, request, target.URL+request.URL.Path, http.StatusTemporaryRedirect)
	}))
	defer redirector.Close()

	err := newTestClient(redirector.URL, redirector.Client()).Send(context.Background(), sampleAlert())
	assertNoToken(testRun, err)
	if targetHits.Load() != 0 {
		testRun.Fatal("redirect target was contacted: token path would have been sent elsewhere")
	}
}

func TestSendRetriesOnceAfterRateLimit(testRun *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		if requests.Add(1) == 1 {
			writer.WriteHeader(http.StatusTooManyRequests)
			_, _ = io.WriteString(writer, `{"ok":false,"error_code":429,"description":"Too Many Requests","parameters":{"retry_after":2}}`)
			return
		}
		_, _ = io.WriteString(writer, `{"ok":true}`)
	}))
	defer server.Close()

	var waited []time.Duration
	client := New(Options{
		BaseURL: server.URL, Token: fakeToken, ChatID: fakeChatID, HTTPClient: server.Client(),
		Wait: func(_ context.Context, duration time.Duration) error {
			waited = append(waited, duration)
			return nil
		},
	})
	if err := client.Send(context.Background(), sampleAlert()); err != nil {
		testRun.Fatalf("send: %v", err)
	}
	if requests.Load() != 2 || len(waited) != 1 || waited[0] != 2*time.Second {
		testRun.Fatalf("requests=%d waited=%v", requests.Load(), waited)
	}
}

func TestSendRateLimitIsBounded(testRun *testing.T) {
	testCases := []struct {
		name         string
		retryAfter   int
		wantRequests int32
		wantWaits    int
	}{
		{"second 429 gives up", 1, 2, 1},
		{"retry_after above the cap is not waited", int(MaxRetryAfter/time.Second) + 1, 1, 0},
	}
	for _, testCase := range testCases {
		testRun.Run(testCase.name, func(testRun *testing.T) {
			var requests atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
				requests.Add(1)
				writer.WriteHeader(http.StatusTooManyRequests)
				_, _ = fmt.Fprintf(writer, `{"ok":false,"parameters":{"retry_after":%d}}`, testCase.retryAfter)
			}))
			defer server.Close()
			waits := 0
			client := New(Options{
				BaseURL: server.URL, Token: fakeToken, ChatID: fakeChatID, HTTPClient: server.Client(),
				Wait: func(context.Context, time.Duration) error { waits++; return nil },
			})
			err := client.Send(context.Background(), sampleAlert())
			if !errors.Is(err, ports.ErrNotificationRateLimited) {
				testRun.Fatalf("got %v", err)
			}
			assertNoToken(testRun, err)
			if requests.Load() != testCase.wantRequests || waits != testCase.wantWaits {
				testRun.Fatalf("requests=%d waits=%d", requests.Load(), waits)
			}
		})
	}
}

func TestSendRateLimitWaitHonorsCancellation(testRun *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusTooManyRequests)
		_, _ = io.WriteString(writer, `{"ok":false,"parameters":{"retry_after":1}}`)
	}))
	defer server.Close()
	client := New(Options{
		BaseURL: server.URL, Token: fakeToken, ChatID: fakeChatID, HTTPClient: server.Client(),
		Wait: func(context.Context, time.Duration) error { return context.Canceled },
	})
	err := client.Send(context.Background(), sampleAlert())
	if !errors.Is(err, context.Canceled) {
		testRun.Fatalf("got %v", err)
	}
	assertNoToken(testRun, err)
}

func TestSendRequiresCredentials(testRun *testing.T) {
	missingToken := New(Options{BaseURL: "http://localhost", ChatID: fakeChatID})
	missingChat := New(Options{BaseURL: "http://localhost", Token: fakeToken})
	for name, client := range map[string]*Client{"token": missingToken, "chat": missingChat} {
		if err := client.Send(context.Background(), sampleAlert()); !errors.Is(err, ports.ErrNotifierNotConfigured) {
			testRun.Errorf("%s: got %v", name, err)
		}
	}
}

func TestResponseBodyIsBounded(testRun *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusBadGateway)
		_, _ = writer.Write([]byte(strings.Repeat("x", 5*MaxResponseBytes)))
	}))
	defer server.Close()
	err := newTestClient(server.URL, server.Client()).Send(context.Background(), sampleAlert())
	if err == nil || len(err.Error()) > 1000 {
		testRun.Fatalf("error should be short, got %v", err)
	}
}

func TestBotTokenNeverPrintsItself(testRun *testing.T) {
	for _, verb := range []string{"%s", "%v", "%+v", "%#v", "%q", "%x"} {
		if strings.Contains(fmt.Sprintf(verb, fakeToken), "AA-fake") {
			testRun.Errorf("verb %s leaks the token", verb)
		}
	}
}
