package provider

import (
	"context"

	"golang.org/x/time/rate"
)

// internalCompletionHandler serves app-initiated requests (titles,
// suggestions): no per-user access checks and no rate limiting.
type internalCompletionHandler struct {
	limiter     *rate.Limiter
	openAIProxy string
}

func newInternalCompletionHandler(openAIProxy string) *internalCompletionHandler {
	return &internalCompletionHandler{limiter: rate.NewLimiter(rate.Inf, 1), openAIProxy: openAIProxy}
}

func (h *internalCompletionHandler) CheckModelAccess(context.Context, string, string, int32) error {
	return nil
}

func (h *internalCompletionHandler) Config() Config {
	return Config{RateLimiter: h.limiter, OpenAIProxy: h.openAIProxy}
}
