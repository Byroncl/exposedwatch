package main

import (
	"context"
	"time"

	"github.com/Byroncl/exposedwatch/internal/app"
	"github.com/Byroncl/exposedwatch/internal/ports"
)

// watchLoop is the headless scan, evaluate and notify cycle.
type watchLoop struct {
	scan      scanFunc
	evaluator *app.AlertEvaluator
	outbox    *app.AlertOutbox
	notifier  ports.Notifier // nil when notifications are disabled
	logger    watchLogger
	interval  time.Duration
	once      bool
	wait      func(ctx context.Context, duration time.Duration) bool
}

// run loops until the context ends (exit 0) or, with once, after a single
// iteration, whose outcome becomes the exit code.
func (loop *watchLoop) run(ctx context.Context) ExitCode {
	for {
		if ctx.Err() != nil {
			loop.logger.Logf(logShutdownText)
			return ExitOK
		}
		code := loop.iterate(ctx)
		if loop.once {
			return code
		}
		if !loop.wait(ctx, loop.interval) {
			loop.logger.Logf(logShutdownText)
			return ExitOK
		}
	}
}

// iterate performs one scan and one delivery round. A scan failure never
// stops pending alerts from being retried.
func (loop *watchLoop) iterate(ctx context.Context) ExitCode {
	code := ExitOK
	report, err := loop.scan(ctx)
	switch {
	case err != nil && ctx.Err() != nil:
		return ExitOK
	case err != nil:
		loop.logger.Logf(logScanFailedFormat, err)
		code = ExitRuntimeError
	default:
		for _, sourceError := range report.SourceErrors {
			loop.logger.Logf(logSourceUnavailableFormat, sourceError.Source, sourceError.Err)
		}
		loop.queue(loop.evaluator.EvaluateDetailed(report))
	}
	if loop.notifier != nil && loop.outbox.Pending() > 0 && loop.deliver(ctx) {
		code = ExitRuntimeError
	}
	return code
}

// queue hands evaluated alerts to the outbox. When an alert is finally given
// up on, its findings are forgotten so they can alert again later.
func (loop *watchLoop) queue(evaluated []app.EvaluatedAlert) {
	if loop.notifier == nil {
		for _, entry := range evaluated {
			loop.logger.Logf(logAlertFormat, alertLabel(entry.Alert), entry.Alert.Host)
		}
		return
	}
	queued := make([]app.QueuedAlert, 0, len(evaluated))
	for _, entry := range evaluated {
		identities := entry.Identities
		queued = append(queued, app.QueuedAlert{
			Alert:  entry.Alert,
			OnDrop: func() { loop.evaluator.Forget(identities) },
		})
	}
	if dropped := loop.outbox.Add(queued); dropped > 0 {
		loop.logger.Logf(logQueueOverflowFormat, dropped)
	}
}

// deliver sends the alerts that are due and reports whether any send failed.
// Failed alerts stay in the outbox with an exponential backoff until their
// lifetime ends. A send in flight when shutdown starts gets a grace period.
func (loop *watchLoop) deliver(ctx context.Context) bool {
	deliveryContext, release := graceContext(ctx, shutdownGracePeriod)
	defer release()
	result := loop.outbox.Deliver(deliveryContext, loop.notifier)
	if result.Sent > 0 {
		loop.logger.Logf(logAlertSentCountFormat, result.Sent)
	}
	for _, failure := range result.Failures {
		if failure.Dropped {
			loop.logger.Logf(logDeliveryDroppedFormat, alertLabel(failure.Alert), failure.Err)
			continue
		}
		loop.logger.Logf(logDeliveryFailedFormat, failure.Attempt, failure.RetryIn, failure.Err)
	}
	return len(result.Failures) > 0
}

// graceContext stays alive for the grace period after the parent ends, so
// an alert being sent during shutdown is not cut off mid-request.
func graceContext(parent context.Context, grace time.Duration) (context.Context, context.CancelFunc) {
	detached, cancel := context.WithCancel(context.WithoutCancel(parent))
	stopWatching := context.AfterFunc(parent, func() {
		time.AfterFunc(grace, cancel)
	})
	return detached, func() {
		stopWatching()
		cancel()
	}
}
