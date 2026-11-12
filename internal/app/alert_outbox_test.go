package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

const (
	testInitialBackoff = time.Minute
	testMaxBackoff     = 8 * time.Minute
	testTimeToLive     = time.Hour
)

type scriptedNotifier struct {
	failures map[string]int // title -> remaining failures; negative means always
	sent     []string
	attempts []string
}

func (notifier *scriptedNotifier) Send(_ context.Context, alert domain.Alert) error {
	notifier.attempts = append(notifier.attempts, alert.Title)
	if notifier.failures[alert.Title] != 0 {
		if notifier.failures[alert.Title] > 0 {
			notifier.failures[alert.Title]--
		}
		return errors.New("simulated failure")
	}
	notifier.sent = append(notifier.sent, alert.Title)
	return nil
}

func queued(titles ...string) []QueuedAlert {
	entries := make([]QueuedAlert, 0, len(titles))
	for _, title := range titles {
		entries = append(entries, QueuedAlert{Alert: domain.Alert{Title: title}})
	}
	return entries
}

func newTestOutbox(clock *fakeClock) *AlertOutbox {
	return NewAlertOutbox(OutboxPolicy{
		InitialBackoff: testInitialBackoff, MaxBackoff: testMaxBackoff, TimeToLive: testTimeToLive,
	}, clock)
}

func TestOutboxDeliversEverythingOnce(testRun *testing.T) {
	outbox := newTestOutbox(newTestClock())
	outbox.Add(queued("a", "b"))
	result := outbox.Deliver(context.Background(), &scriptedNotifier{})
	if result.Sent != 2 || len(result.Failures) != 0 || outbox.Pending() != 0 {
		testRun.Fatalf("%+v pending=%d", result, outbox.Pending())
	}
}

func TestOutboxBacksOffExponentiallyUpToTheCap(testRun *testing.T) {
	clock := newTestClock()
	outbox := newTestOutbox(clock)
	outbox.Add(queued("a"))
	notifier := &scriptedNotifier{failures: map[string]int{"a": -1}}
	wantBackoffs := []time.Duration{time.Minute, 2 * time.Minute, 4 * time.Minute, 8 * time.Minute, 8 * time.Minute}
	for round, want := range wantBackoffs {
		result := outbox.Deliver(context.Background(), notifier)
		if len(result.Failures) != 1 || result.Failures[0].RetryIn != want || result.Failures[0].Attempt != round+1 || result.Failures[0].Dropped {
			testRun.Fatalf("round %d: %+v", round, result.Failures)
		}
		// Too early: nothing is attempted.
		clock.advance(want - time.Second)
		before := len(notifier.attempts)
		outbox.Deliver(context.Background(), notifier)
		if len(notifier.attempts) != before {
			testRun.Fatalf("round %d: retried before the backoff elapsed", round)
		}
		clock.advance(time.Second)
	}
}

func TestOutboxRecoversAfterAnOutage(testRun *testing.T) {
	clock := newTestClock()
	outbox := newTestOutbox(clock)
	outbox.Add(queued("a"))
	notifier := &scriptedNotifier{failures: map[string]int{"a": 5}}
	for round := 0; round < 5; round++ {
		outbox.Deliver(context.Background(), notifier)
		clock.advance(testMaxBackoff)
	}
	result := outbox.Deliver(context.Background(), notifier)
	if result.Sent != 1 || outbox.Pending() != 0 {
		testRun.Fatalf("%+v pending=%d", result, outbox.Pending())
	}
}

func TestAFailingAlertDoesNotBlockNewerOnes(testRun *testing.T) {
	clock := newTestClock()
	outbox := newTestOutbox(clock)
	notifier := &scriptedNotifier{failures: map[string]int{"old": -1}}
	outbox.Add(queued("old", "second"))
	first := outbox.Deliver(context.Background(), notifier)
	if first.Sent != 1 || notifier.sent[0] != "second" {
		testRun.Fatalf("%+v", first)
	}
	clock.advance(10 * time.Second) // "old" is still backing off
	outbox.Add(queued("newer"))
	second := outbox.Deliver(context.Background(), notifier)
	if second.Sent != 1 || notifier.sent[1] != "newer" || outbox.Pending() != 1 {
		testRun.Fatalf("%+v sent=%v pending=%d", second, notifier.sent, outbox.Pending())
	}
}

func TestOutboxDropsAfterTheLifetimeAndTellsTheOwner(testRun *testing.T) {
	clock := newTestClock()
	outbox := newTestOutbox(clock)
	dropped := 0
	outbox.Add([]QueuedAlert{{Alert: domain.Alert{Title: "a"}, OnDrop: func() { dropped++ }}})
	notifier := &scriptedNotifier{failures: map[string]int{"a": -1}}
	outbox.Deliver(context.Background(), notifier)
	clock.advance(testTimeToLive)
	attemptsBefore := len(notifier.attempts)
	result := outbox.Deliver(context.Background(), notifier)
	if len(result.Failures) != 1 || !result.Failures[0].Dropped || !errors.Is(result.Failures[0].Err, ErrAlertExpired) {
		testRun.Fatalf("%+v", result)
	}
	if dropped != 1 || outbox.Pending() != 0 || len(notifier.attempts) != attemptsBefore {
		testRun.Fatalf("dropped=%d pending=%d attempts=%d", dropped, outbox.Pending(), len(notifier.attempts))
	}
}

func TestOutboxOverflowDropsTheOldestAndTellsTheOwner(testRun *testing.T) {
	outbox := newTestOutbox(newTestClock())
	dropped := 0
	var entries []QueuedAlert
	for index := 0; index < MaxPendingAlerts+10; index++ {
		entries = append(entries, QueuedAlert{Alert: domain.Alert{Title: "x"}, OnDrop: func() { dropped++ }})
	}
	if count := outbox.Add(entries); count != 10 || dropped != 10 || outbox.Pending() != MaxPendingAlerts {
		testRun.Fatalf("count=%d dropped=%d pending=%d", count, dropped, outbox.Pending())
	}
}

func TestOutboxStopsOnCancellationWithoutConsumingAttempts(testRun *testing.T) {
	outbox := newTestOutbox(newTestClock())
	outbox.Add(queued("a", "b"))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	result := outbox.Deliver(ctx, &scriptedNotifier{})
	if result.Sent != 0 || len(result.Failures) != 0 || outbox.Pending() != 2 {
		testRun.Fatalf("%+v pending=%d", result, outbox.Pending())
	}
}
