package domain

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

func TestProviderFailureClassification(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		kind      ProviderFailureKind
		retryable bool
	}{
		{"canceled", NewProviderFailure("openai", "stream", context.Canceled), ProviderFailureCanceled, false},
		{"deadline", NewProviderFailure("openai", "stream", context.DeadlineExceeded), ProviderFailureTimeout, true},
		{"rate limited", NewProviderHTTPFailure("openai", "request", http.StatusTooManyRequests, errors.New("quota")), ProviderFailureRateLimited, true},
		{"bad request", NewProviderHTTPFailure("claude", "request", http.StatusBadRequest, errors.New("bad model")), ProviderFailureInvalidRequest, false},
		{"unavailable", NewProviderHTTPFailure("gemini", "request", http.StatusServiceUnavailable, errors.New("down")), ProviderFailureUnavailable, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			failure, ok := AsProviderFailure(tt.err)
			if !ok || failure.Kind != tt.kind || failure.Retryable != tt.retryable {
				t.Fatalf("got %#v, ok=%v", failure, ok)
			}
			if IsProviderRetryable(tt.err) != tt.retryable {
				t.Fatalf("unexpected retryability")
			}
		})
	}
}
