package provider

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/swuecho/chat_backend/domain"
)

func TestRetryRetriesOnlyRetryableProviderFailures(t *testing.T) {
	attempts := 0
	result, err := Retry(context.Background(), RetryPolicy{MaxAttempts: 3, Sleep: func(context.Context, time.Duration) error { return nil }}, func(context.Context) (string, error) {
		attempts++
		if attempts < 3 {
			return "", domain.NewProviderHTTPFailure("test", "request", http.StatusServiceUnavailable, errors.New("down"))
		}
		return "ok", nil
	})
	if err != nil || result != "ok" || attempts != 3 {
		t.Fatalf("result=%q err=%v attempts=%d", result, err, attempts)
	}
}

func TestRetryStopsOnNonRetryableFailure(t *testing.T) {
	attempts := 0
	_, err := Retry(context.Background(), RetryPolicy{MaxAttempts: 3}, func(context.Context) (string, error) {
		attempts++
		return "", domain.NewProviderHTTPFailure("test", "request", http.StatusBadRequest, errors.New("bad"))
	})
	if err == nil || attempts != 1 {
		t.Fatalf("err=%v attempts=%d", err, attempts)
	}
}

func TestRetryWaitHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := Retry(ctx, RetryPolicy{MaxAttempts: 2, InitialDelay: time.Hour}, func(context.Context) (string, error) {
		return "", domain.NewProviderHTTPFailure("test", "request", http.StatusTooManyRequests, errors.New("busy"))
	})
	if failure, ok := domain.AsProviderFailure(err); !ok || failure.Kind != domain.ProviderFailureCanceled {
		t.Fatalf("unexpected error: %v", err)
	}
}
