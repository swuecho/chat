package provider

import (
	"context"
	"golang.org/x/time/rate"
)

type titleGenerationHandler struct {
	q       QueryStore
	limiter *rate.Limiter
}

func newTitleGenerationHandler(q QueryStore) *titleGenerationHandler {
	return &titleGenerationHandler{q: q, limiter: rate.NewLimiter(rate.Inf, 1)}
}

func (h *titleGenerationHandler) Queries() QueryStore { return h.q }

func (h *titleGenerationHandler) CheckModelAccess(context.Context, string, string, int32) error {
	return nil
}

func (h *titleGenerationHandler) Config() Config {
	return Config{RateLimiter: h.limiter}
}
