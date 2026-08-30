package domain

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
)

// ProviderFailureKind classifies failures returned by an upstream LLM provider.
// It is transport independent and is safe for application code to inspect.
type ProviderFailureKind string

const (
	ProviderFailureInvalidRequest  ProviderFailureKind = "invalid_request"
	ProviderFailureAuthentication  ProviderFailureKind = "authentication"
	ProviderFailurePermission      ProviderFailureKind = "permission"
	ProviderFailureRateLimited     ProviderFailureKind = "rate_limited"
	ProviderFailureUnavailable     ProviderFailureKind = "unavailable"
	ProviderFailureTimeout         ProviderFailureKind = "timeout"
	ProviderFailureCanceled        ProviderFailureKind = "canceled"
	ProviderFailureInvalidResponse ProviderFailureKind = "invalid_response"
	ProviderFailureConfiguration   ProviderFailureKind = "configuration"
	ProviderFailureInternal        ProviderFailureKind = "internal"
)

// ProviderFailure is the normalized error contract for all LLM providers.
// Retryable describes whether retrying the operation is generally safe; callers
// must still avoid retrying a stream after content has already been delivered.
type ProviderFailure struct {
	Provider   string
	Operation  string
	Kind       ProviderFailureKind
	Retryable  bool
	StatusCode int
	Err        error
}

func (e *ProviderFailure) Error() string {
	message := fmt.Sprintf("%s provider %s failed", e.Provider, e.Operation)
	if e.Err != nil {
		return message + ": " + e.Err.Error()
	}
	return message
}

func (e *ProviderFailure) Unwrap() error { return e.Err }

// NewProviderFailure normalizes non-HTTP failures such as cancellation,
// deadlines, and network errors. Unknown failures are deliberately not marked
// retryable.
func NewProviderFailure(provider, operation string, err error) error {
	if err == nil {
		return nil
	}
	var existing *ProviderFailure
	if errors.As(err, &existing) {
		return err
	}

	kind, retryable := ProviderFailureInternal, false
	switch {
	case errors.Is(err, context.Canceled):
		kind = ProviderFailureCanceled
	case errors.Is(err, context.DeadlineExceeded):
		kind, retryable = ProviderFailureTimeout, true
	default:
		var netErr net.Error
		if errors.As(err, &netErr) {
			if netErr.Timeout() {
				kind = ProviderFailureTimeout
			} else {
				kind = ProviderFailureUnavailable
			}
			retryable = true
		}
	}
	return &ProviderFailure{Provider: provider, Operation: operation, Kind: kind, Retryable: retryable, Err: err}
}

// NewProviderHTTPFailure classifies an unsuccessful upstream HTTP response.
func NewProviderHTTPFailure(provider, operation string, statusCode int, err error) error {
	kind, retryable := ProviderFailureInternal, false
	switch statusCode {
	case http.StatusBadRequest, http.StatusUnprocessableEntity:
		kind = ProviderFailureInvalidRequest
	case http.StatusUnauthorized:
		kind = ProviderFailureAuthentication
	case http.StatusForbidden:
		kind = ProviderFailurePermission
	case http.StatusRequestTimeout, http.StatusGatewayTimeout:
		kind, retryable = ProviderFailureTimeout, true
	case http.StatusTooManyRequests:
		kind, retryable = ProviderFailureRateLimited, true
	default:
		if statusCode >= 500 {
			kind, retryable = ProviderFailureUnavailable, true
		}
	}
	return &ProviderFailure{Provider: provider, Operation: operation, Kind: kind, Retryable: retryable, StatusCode: statusCode, Err: err}
}

func AsProviderFailure(err error) (*ProviderFailure, bool) {
	var failure *ProviderFailure
	ok := errors.As(err, &failure)
	return failure, ok
}

func IsProviderRetryable(err error) bool {
	failure, ok := AsProviderFailure(err)
	return ok && failure.Retryable
}
