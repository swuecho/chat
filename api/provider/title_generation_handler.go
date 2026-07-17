package provider

import (
	"context"

	"github.com/swuecho/chat_backend/sqlc_queries"
	"golang.org/x/time/rate"
)

type titleGenerationHandler struct {
	q       *sqlc_queries.Queries
	limiter *rate.Limiter
}

func newTitleGenerationHandler(q *sqlc_queries.Queries) *titleGenerationHandler {
	return &titleGenerationHandler{q: q, limiter: rate.NewLimiter(rate.Inf, 1)}
}

func (h *titleGenerationHandler) Queries() *sqlc_queries.Queries { return h.q }

func (h *titleGenerationHandler) CheckModelAccess(context.Context, string, string, int32) error {
	return nil
}

func (h *titleGenerationHandler) Config() Config {
	return Config{RateLimiter: h.limiter}
}
