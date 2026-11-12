// Package telegram delivers alerts through the Telegram Bot API.
package telegram

import "time"

// Endpoint and transport limits.
const (
	// DefaultBaseURL is the public Telegram Bot API.
	DefaultBaseURL = "https://api.telegram.org"
	// RequestTimeout bounds one HTTP request, including reading the answer.
	RequestTimeout = 10 * time.Second
	// MaxResponseBytes bounds how much of an answer is ever read.
	MaxResponseBytes = 64 * 1024
	// MaxRetryAfter is the longest wait honored after an HTTP 429.
	MaxRetryAfter = 30 * time.Second
	// MaxMessageRunes is Telegram's limit for the text of one message.
	MaxMessageRunes = 4096
)

// Request building.
const (
	botPathPrefix       = "/bot"
	sendMessagePath     = "/sendMessage"
	contentTypeHeader   = "Content-Type"
	contentTypeJSON     = "application/json"
	retryAfterHeader    = "Retry-After"
	defaultRetryAfter   = time.Second
	maxRateLimitedTries = 2
)

// Message layout.
const (
	truncationSuffix    = "…"
	maxDescriptionRunes = 100
	headerFormat        = "%s %s"
	fieldHostFormat     = "Host: %s"
	fieldPortFormat     = "Port: %s"
	fieldExposureFormat = "Exposure: %s"
	fieldProcessFormat  = "Process: %s"
	fieldRuleFormat     = "Rule: %s"
	fieldFixFormat      = "Fix: %s"
	lineSeparator       = "\n"
	blankLineSeparator  = "\n\n"
)

// Redaction.
const (
	redactionText        = "***"
	botSegmentLabel      = "bot"
	caseInsensitiveFlag  = "(?i)"
	alternationSeparator = "|"
	// tokenFragmentLength is how many consecutive token characters in remote
	// text make it unsafe to echo.
	tokenFragmentLength = 6
	// encodedAlphabet holds the characters of base64 (standard and URL
	// alphabets, padding) and hex, used to split text into decodable fields.
	encodedAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/=-_"
	// descriptionCharset is the conservative alphabet of an echoed Telegram
	// description such as "Bad Request: chat not found".
	descriptionCharset = `^[A-Za-z0-9 .,:;'()_!?-]+$`
	// Placeholders printed instead of values that hold the token.
	redactorPlaceholder = "telegram.tokenRedactor{***}"
	clientPlaceholder   = "telegram.Client{***}"
)

// retryAfterCeilingSeconds clamps remote retry_after values before they are
// converted to a time.Duration. It is one second above MaxRetryAfter, so a
// clamped value is still refused as too long and cannot overflow.
const retryAfterCeilingSeconds = int64(MaxRetryAfter/time.Second) + 1

// Error details. None of them may ever contain the request URL.
const (
	invalidBaseURLDetail   = "base URL must be an absolute http or https URL"
	buildRequestDetail     = "request could not be built"
	canceledDetail         = "request canceled"
	timeoutDetail          = "request timed out"
	transportDetailFormat  = "request failed: %s"
	httpStatusDetailFormat = "Telegram answered HTTP %d: %s"
	genericStatusFormat    = "Telegram answered HTTP %d (error_code %d)"
	rateLimitDetailFormat  = "Telegram asked to retry after %s"
	marshalDetail          = "message could not be encoded"
	errorFormat            = "%s: %s"
)
