package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Byroncl/exposedwatch/internal/domain"
	"github.com/Byroncl/exposedwatch/internal/ports"
)

// WaitFunc pauses for the duration or until the context ends.
type WaitFunc func(ctx context.Context, duration time.Duration) error

// Options configure a Client. Only Token and ChatID are required.
type Options struct {
	BaseURL    string // default DefaultBaseURL; injectable for tests
	Token      domain.BotToken
	ChatID     domain.ChatID
	HTTPClient *http.Client // default: a client with RequestTimeout
	Wait       WaitFunc     // default: a context-aware timer
}

// Client implements ports.Notifier over the Telegram Bot API.
type Client struct {
	baseURL    string
	baseURLOK  bool
	token      domain.BotToken
	chatID     domain.ChatID
	httpClient *http.Client
	wait       WaitFunc
	redactor   tokenRedactor
}

// New builds a client. TLS verification, proxy settings and every other
// transport default are left untouched; only redirects are refused, because
// the bot token is part of the request path and must never be replayed to
// another host.
func New(options Options) *Client {
	baseURL := strings.TrimRight(options.BaseURL, "/")
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	httpClient := &http.Client{Timeout: RequestTimeout}
	if options.HTTPClient != nil {
		clientCopy := *options.HTTPClient
		httpClient = &clientCopy
	}
	if httpClient.CheckRedirect == nil {
		httpClient.CheckRedirect = refuseRedirects
	}
	wait := options.Wait
	if wait == nil {
		wait = waitWithTimer
	}
	return &Client{
		baseURL:    baseURL,
		baseURLOK:  isAbsoluteHTTPURL(baseURL),
		token:      options.Token,
		chatID:     options.ChatID,
		httpClient: httpClient,
		wait:       wait,
		redactor:   newTokenRedactor(options.Token),
	}
}

func refuseRedirects(*http.Request, []*http.Request) error {
	return http.ErrUseLastResponse
}

func isAbsoluteHTTPURL(rawURL string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Host == "" {
		return false
	}
	return parsed.Scheme == "http" || parsed.Scheme == "https"
}

func waitWithTimer(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// Send delivers the alert as one plain-text message. After an HTTP 429 it
// waits the advertised retry_after (bounded by MaxRetryAfter) and retries
// once. Every returned error is already free of the token.
func (client *Client) Send(ctx context.Context, alert domain.Alert) error {
	if client.token.IsEmpty() || client.chatID == "" {
		return client.failure(ports.ErrNotifierNotConfigured, nil, "")
	}
	if !client.baseURLOK {
		return client.failure(ports.ErrNotificationFailed, nil, invalidBaseURLDetail)
	}
	payload, err := json.Marshal(sendMessageRequest{
		ChatID:                string(client.chatID),
		Text:                  FormatAlert(alert),
		DisableWebPagePreview: true,
	})
	if err != nil {
		return client.failure(ports.ErrNotificationFailed, nil, marshalDetail)
	}
	for attempt := 1; ; attempt++ {
		retryAfter, sendErr := client.post(ctx, payload)
		if sendErr == nil || retryAfter <= 0 || attempt >= maxRateLimitedTries || retryAfter > MaxRetryAfter {
			return sendErr
		}
		if waitErr := client.wait(ctx, retryAfter); waitErr != nil {
			return client.contextFailure(ctx, waitErr)
		}
	}
}

// post performs one request. A positive retryAfter means "rate limited".
func (client *Client) post(ctx context.Context, payload []byte) (time.Duration, error) {
	requestContext, cancel := context.WithTimeout(ctx, RequestTimeout)
	defer cancel()
	// The token lives in the URL path: this string must never reach an error
	// or a log line.
	endpoint := client.baseURL + botPathPrefix + url.PathEscape(client.token.Reveal()) + sendMessagePath
	request, err := http.NewRequestWithContext(requestContext, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return 0, client.failure(ports.ErrNotificationFailed, nil, buildRequestDetail)
	}
	request.Header.Set(contentTypeHeader, contentTypeJSON)
	response, err := client.httpClient.Do(request)
	if err != nil {
		return 0, client.transportFailure(ctx, err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, MaxResponseBytes))
	if err != nil {
		return 0, client.transportFailure(ctx, err)
	}
	return client.interpret(response, body)
}

// interpret turns an HTTP answer into success, a rate limit or a rejection.
func (client *Client) interpret(response *http.Response, body []byte) (time.Duration, error) {
	var answer sendMessageResponse
	decodeErr := json.Unmarshal(body, &answer)
	succeeded := response.StatusCode >= http.StatusOK && response.StatusCode < http.StatusMultipleChoices
	if succeeded && decodeErr == nil && answer.OK {
		return 0, nil
	}
	if response.StatusCode == http.StatusTooManyRequests {
		retryAfter := retryAfterOf(response, answer)
		return retryAfter, client.failure(ports.ErrNotificationRateLimited, nil,
			fmt.Sprintf(rateLimitDetailFormat, retryAfter))
	}
	// Remote text is echoed only when it is short, plain and free of any
	// part of the token; otherwise only numbers are shown.
	if decodeErr == nil {
		if description, safe := client.redactor.safeDescription(answer.Description); safe {
			return 0, client.failure(ports.ErrNotificationRejected, nil,
				fmt.Sprintf(httpStatusDetailFormat, response.StatusCode, description))
		}
	}
	return 0, client.failure(ports.ErrNotificationRejected, nil,
		fmt.Sprintf(genericStatusFormat, response.StatusCode, answer.ErrorCode))
}

// retryAfterOf reads the advertised wait. Zero means "do not retry".
func retryAfterOf(response *http.Response, answer sendMessageResponse) time.Duration {
	if answer.Parameters.RetryAfter != nil {
		return clampedSeconds(*answer.Parameters.RetryAfter)
	}
	seconds, err := strconv.ParseInt(strings.TrimSpace(response.Header.Get(retryAfterHeader)), 10, 64)
	if err == nil {
		return clampedSeconds(seconds)
	}
	return defaultRetryAfter
}

// clampedSeconds bounds remote seconds BEFORE converting them to a
// time.Duration, which would overflow for huge values.
func clampedSeconds(seconds int64) time.Duration {
	if seconds <= 0 {
		return 0
	}
	if seconds > retryAfterCeilingSeconds {
		seconds = retryAfterCeilingSeconds
	}
	return time.Duration(seconds) * time.Second
}

// transportFailure converts a client error, which is a *url.Error carrying
// the full request URL, into a redacted deliveryError. The *url.Error itself
// is dropped; only a bare context error may be kept for errors.Is.
func (client *Client) transportFailure(ctx context.Context, err error) error {
	if ctx.Err() != nil {
		return client.contextFailure(ctx, ctx.Err())
	}
	var timeoutError net.Error
	if errors.As(err, &timeoutError) && timeoutError.Timeout() {
		return client.failure(ports.ErrNotificationFailed, nil, timeoutDetail)
	}
	reason := err
	var urlError *url.Error
	if errors.As(err, &urlError) {
		reason = urlError.Err
	}
	return client.failure(ports.ErrNotificationFailed, nil,
		fmt.Sprintf(transportDetailFormat, client.redactor.cleanReason(reason.Error())))
}

// String keeps the token out of fmt output: the client holds it.
func (client Client) String() string { return clientPlaceholder }

// GoString keeps the token out of %#v output.
func (client Client) GoString() string { return clientPlaceholder }

// Format keeps the token out of fmt output, whatever the verb.
func (client Client) Format(state fmt.State, _ rune) {
	_, _ = state.Write([]byte(clientPlaceholder))
}

func (client *Client) contextFailure(ctx context.Context, cause error) error {
	detail := canceledDetail
	if errors.Is(cause, context.DeadlineExceeded) {
		detail = timeoutDetail
	}
	if ctx.Err() != nil {
		cause = ctx.Err()
	}
	return client.failure(ports.ErrNotificationFailed, bareContextError(cause), detail)
}

// bareContextError returns context.Canceled or context.DeadlineExceeded
// themselves, so no wrapper (and no URL) can ride along.
func bareContextError(err error) error {
	switch {
	case errors.Is(err, context.Canceled):
		return context.Canceled
	case errors.Is(err, context.DeadlineExceeded):
		return context.DeadlineExceeded
	default:
		return nil
	}
}

func (client *Client) failure(sentinel, cause error, detail string) error {
	return &deliveryError{sentinel: sentinel, cause: cause, detail: client.redactor.Redact(detail)}
}
