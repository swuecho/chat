package provider

import (
	"context"
	"golang.org/x/time/rate"
)

type titleGenerationHandler struct {
	limiter *rate.Limiter
}

func newTitleGenerationHandler() *titleGenerationHandler {
	return &titleGenerationHandler{limiter: rate.NewLimiter(rate.Inf, 1)}
}

func (h *titleGenerationHandler) CheckModelAccess(context.Context, string, string, int32) error {
	return nil
}

func (h *titleGenerationHandler) Config() Config {
	return Config{RateLimiter: h.limiter}
}
