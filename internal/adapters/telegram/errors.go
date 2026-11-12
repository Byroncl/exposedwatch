package telegram

import "fmt"

// deliveryError is the only error type the client returns. Its detail is
// redacted before construction, and it only ever wraps a port-level sentinel
// and, optionally, a bare context error. It never wraps a *url.Error, which
// would carry the request URL (and with it the bot token).
type deliveryError struct {
	sentinel error
	cause    error
	detail   string
}

// Error implements the error interface.
func (failure *deliveryError) Error() string {
	return fmt.Sprintf(errorFormat, failure.sentinel, failure.detail)
}

// Unwrap exposes the sentinel and the optional context cause to errors.Is.
func (failure *deliveryError) Unwrap() []error {
	if failure.cause == nil {
		return []error{failure.sentinel}
	}
	return []error{failure.sentinel, failure.cause}
}
