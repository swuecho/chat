package provider

import (
	"errors"
	"net/http"
	"testing"

	"github.com/cenkalti/backoff/v4"
	"github.com/swuecho/chat_backend/domain"
)

func TestProviderRetryAdapterMarksOnlyNonRetryableErrorsPermanent(t *testing.T) {
	retryable := domain.NewProviderHTTPFailure("test", "request", http.StatusServiceUnavailable, errors.New("down"))
	if got := stopRetryingUnlessProviderFailureIsRetryable(retryable); got != retryable {
		t.Fatalf("retryable error was wrapped: %v", got)
	}

	nonRetryable := domain.NewProviderHTTPFailure("test", "request", http.StatusBadRequest, errors.New("bad request"))
	got := stopRetryingUnlessProviderFailureIsRetryable(nonRetryable)
	var permanent *backoff.PermanentError
	if !errors.As(got, &permanent) || !errors.Is(got, nonRetryable) {
		t.Fatalf("non-retryable error was not permanent: %v", got)
	}
}
