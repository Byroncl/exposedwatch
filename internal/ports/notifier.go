package ports

import (
	"context"
	"errors"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

// Notifier delivers one alert to the outside world. Implementations must
// never put secrets (tokens, URLs containing them) into returned errors.
type Notifier interface {
	Send(ctx context.Context, alert domain.Alert) error
}

// Notification failure classes. Adapters wrap them with redacted details.
var (
	// ErrNotifierNotConfigured means a credential or destination is missing.
	ErrNotifierNotConfigured = errors.New("notifier is not configured")
	// ErrNotificationRejected means the service answered but refused the alert.
	ErrNotificationRejected = errors.New("notification rejected by the service")
	// ErrNotificationRateLimited means the service asked to slow down.
	ErrNotificationRateLimited = errors.New("notification rate limited by the service")
	// ErrNotificationFailed means the alert could not reach the service.
	ErrNotificationFailed = errors.New("notification could not be delivered")
)
