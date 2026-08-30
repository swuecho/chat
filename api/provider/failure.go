package provider

import (
	"errors"

	"github.com/cenkalti/backoff/v4"
	openai "github.com/sashabaranov/go-openai"
	"github.com/swuecho/chat_backend/domain"
)

func normalizeFailure(provider, operation string, err error) error {
	var statusErr interface{ HTTPStatusCode() int }
	if errors.As(err, &statusErr) && statusErr.HTTPStatusCode() != 0 {
		return domain.NewProviderHTTPFailure(provider, operation, statusErr.HTTPStatusCode(), err)
	}
	var apiErr *openai.APIError
	if errors.As(err, &apiErr) && apiErr.HTTPStatusCode != 0 {
		return domain.NewProviderHTTPFailure(provider, operation, apiErr.HTTPStatusCode, err)
	}
	var requestErr *openai.RequestError
	if errors.As(err, &requestErr) && requestErr.HTTPStatusCode != 0 {
		return domain.NewProviderHTTPFailure(provider, operation, requestErr.HTTPStatusCode, err)
	}
	return domain.NewProviderFailure(provider, operation, err)
}

func stopRetryingUnlessProviderFailureIsRetryable(err error) error {
	if err == nil || domain.IsProviderRetryable(err) {
		return err
	}
	return backoff.Permanent(err)
}

func classifiedFailure(provider, operation string, kind domain.ProviderFailureKind, retryable bool, err error) error {
	return &domain.ProviderFailure{Provider: provider, Operation: operation, Kind: kind, Retryable: retryable, Err: err}
}
