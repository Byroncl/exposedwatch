package app

import (
	"context"
	"errors"
	"time"

	"github.com/Byroncl/exposedwatch/internal/domain"
	"github.com/Byroncl/exposedwatch/internal/ports"
)

// ErrAlertExpired is reported when an alert could not be delivered within
// its lifetime and was given up on.
var ErrAlertExpired = errors.New(alertExpiredMessage)

// OutboxPolicy bounds the retry behaviour of an AlertOutbox.
type OutboxPolicy struct {
	// InitialBackoff is the wait after the first failed attempt.
	InitialBackoff time.Duration
	// MaxBackoff caps the doubling wait between attempts.
	MaxBackoff time.Duration
	// TimeToLive is how long an alert is retried before it is dropped.
	TimeToLive time.Duration
}

// QueuedAlert is an alert waiting for delivery. OnDrop, when set, is called
// if the alert is finally given up on (lifetime over or queue overflow), so
// its owner can forget the finding and let it alert again later.
type QueuedAlert struct {
	Alert  domain.Alert
	OnDrop func()
}

type pendingAlert struct {
	queued      QueuedAlert
	enqueuedAt  time.Time
	nextAttempt time.Time
	backoff     time.Duration
	attempts    int
}

// DeliveryFailure describes one failed send or one dropped alert.
type DeliveryFailure struct {
	Alert   domain.Alert
	Err     error
	Attempt int
	// RetryIn is the wait before the next attempt; zero when dropped.
	RetryIn time.Duration
	// Dropped is true when the alert was given up on.
	Dropped bool
}

// DeliveryResult summarises one Deliver call.
type DeliveryResult struct {
	Sent     int
	Failures []DeliveryFailure
}

// AlertOutbox holds alerts until a notifier accepts them. Every alert backs
// off on its own (doubling up to a cap), so a failing alert never blocks the
// newer ones, and it is retried until its lifetime ends. The queue is
// bounded: the oldest alerts are dropped first.
type AlertOutbox struct {
	policy  OutboxPolicy
	clock   Clock
	pending []pendingAlert
}

// NewAlertOutbox builds an outbox with the given retry policy.
func NewAlertOutbox(policy OutboxPolicy, clock Clock) *AlertOutbox {
	return &AlertOutbox{policy: policy, clock: clock}
}

// Add queues alerts and returns how many old alerts were dropped to keep
// the queue within MaxPendingAlerts.
func (outbox *AlertOutbox) Add(entries []QueuedAlert) int {
	now := outbox.clock.Now()
	for _, entry := range entries {
		outbox.pending = append(outbox.pending, pendingAlert{queued: entry, enqueuedAt: now})
	}
	overflow := len(outbox.pending) - MaxPendingAlerts
	if overflow <= 0 {
		return 0
	}
	for _, entry := range outbox.pending[:overflow] {
		entry.drop()
	}
	outbox.pending = append([]pendingAlert(nil), outbox.pending[overflow:]...)
	return overflow
}

// Pending returns how many alerts still wait for delivery.
func (outbox *AlertOutbox) Pending() int {
	return len(outbox.pending)
}

// Deliver tries every alert that is due, in order. Failed alerts stay with
// a longer backoff; expired alerts are dropped. It stops without consuming
// attempts when the context ends.
func (outbox *AlertOutbox) Deliver(ctx context.Context, notifier ports.Notifier) DeliveryResult {
	var result DeliveryResult
	var remaining []pendingAlert
	now := outbox.clock.Now()
	for index, entry := range outbox.pending {
		if ctx.Err() != nil {
			remaining = append(remaining, outbox.pending[index:]...)
			break
		}
		if now.Sub(entry.enqueuedAt) >= outbox.policy.TimeToLive {
			entry.drop()
			result.Failures = append(result.Failures, DeliveryFailure{
				Alert: entry.queued.Alert, Err: ErrAlertExpired, Attempt: entry.attempts, Dropped: true,
			})
			continue
		}
		if now.Before(entry.nextAttempt) {
			remaining = append(remaining, entry)
			continue
		}
		err := notifier.Send(ctx, entry.queued.Alert)
		if err == nil {
			result.Sent++
			continue
		}
		if ctx.Err() != nil {
			remaining = append(remaining, outbox.pending[index:]...)
			break
		}
		entry.attempts++
		entry.backoff = outbox.nextBackoff(entry.backoff)
		entry.nextAttempt = now.Add(entry.backoff)
		result.Failures = append(result.Failures, DeliveryFailure{
			Alert: entry.queued.Alert, Err: err, Attempt: entry.attempts, RetryIn: entry.backoff,
		})
		remaining = append(remaining, entry)
	}
	outbox.pending = remaining
	return result
}

// nextBackoff doubles the previous wait, starting at InitialBackoff and
// never exceeding MaxBackoff.
func (outbox *AlertOutbox) nextBackoff(previous time.Duration) time.Duration {
	next := outbox.policy.InitialBackoff
	if previous > 0 {
		next = previous * backoffMultiplier
	}
	if next > outbox.policy.MaxBackoff {
		next = outbox.policy.MaxBackoff
	}
	return next
}

func (entry pendingAlert) drop() {
	if entry.queued.OnDrop != nil {
		entry.queued.OnDrop()
	}
}
