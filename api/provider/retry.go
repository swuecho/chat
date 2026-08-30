package provider

import (
	"context"
	"math/rand"
	"time"

	"github.com/swuecho/chat_backend/domain"
)

// RetryPolicy is a reusable policy value. MaxAttempts includes the initial
// attempt. It retries only normalized provider failures marked retryable.
type RetryPolicy struct {
	MaxAttempts  int
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Jitter       func(time.Duration) time.Duration
	Sleep        func(context.Context, time.Duration) error
}

// RandomJitter returns a delay transform in the range
// [delay*(1-fraction), delay*(1+fraction)].
func RandomJitter(fraction float64) func(time.Duration) time.Duration {
	if fraction < 0 {
		fraction = -fraction
	}
	if fraction > 1 {
		fraction = 1
	}
	return func(delay time.Duration) time.Duration {
		factor := 1 - fraction + rand.Float64()*(2*fraction)
		return time.Duration(float64(delay) * factor)
	}
}

func (p RetryPolicy) delay(attempt int) time.Duration {
	delay := p.InitialDelay
	for i := 1; i < attempt; i++ {
		if p.MaxDelay > 0 && delay >= p.MaxDelay/2 {
			delay = p.MaxDelay
			break
		}
		delay *= 2
	}
	if p.MaxDelay > 0 && delay > p.MaxDelay {
		delay = p.MaxDelay
	}
	if p.Jitter != nil {
		delay = p.Jitter(delay)
	}
	return delay
}

// Retry executes operation according to policy and stops immediately on
// cancellation, non-retryable failures, or exhaustion.
func Retry[T any](ctx context.Context, policy RetryPolicy, operation func(context.Context) (T, error)) (T, error) {
	var zero T
	attempts := policy.MaxAttempts
	if attempts < 1 {
		attempts = 1
	}
	for attempt := 1; attempt <= attempts; attempt++ {
		value, err := operation(ctx)
		if err == nil {
			return value, nil
		}
		if !domain.IsProviderRetryable(err) || attempt == attempts {
			return zero, err
		}
		sleep := policy.Sleep
		if sleep == nil {
			sleep = sleepWithContext
		}
		if err := sleep(ctx, policy.delay(attempt)); err != nil {
			return zero, domain.NewProviderFailure("retry", "wait", err)
		}
	}
	return zero, nil
}

func sleepWithContext(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
