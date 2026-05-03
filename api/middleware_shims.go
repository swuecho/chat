// Package main — Middleware shims that re-export from the middleware package.
// These exist for backward compatibility with handler code that uses the
// un-prefixed function names. New code should import middleware/ directly.
package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/middleware"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

// --- Re-exported middleware constructors ---

var (
	AdminAuthMiddleware = middleware.AdminAuthMiddleware
	UserAuthMiddleware  = middleware.UserAuthMiddleware
	parseAndValidateJWT = middleware.ParseAndValidateJWT
)

// --- Context key constants ---

const (
	roleContextKey = middleware.RoleContextKey
	userContextKey = middleware.UserContextKey
	guidContextKey = middleware.GuidContextKey
)

// --- Auth helpers (kept in main because they use strconv for parsing) ---

func getUserID(ctx context.Context) (int32, error) {
	userIdValue := ctx.Value(userContextKey)
	if userIdValue == nil {
		return 0, fmt.Errorf("no user Id in context")
	}
	userIDStr, _ := userIdValue.(string)
	userIDInt, err := strconv.ParseInt(userIDStr, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid user ID: %s", userIDStr)
	}
	return int32(userIDInt), nil
}

func getContextWithUser(userID int) context.Context {
	return context.WithValue(context.Background(), userContextKey, strconv.Itoa(userID))
}

// --- Admin-only wrappers ---

func AdminOnlyHandler(h http.Handler) http.Handler {
	return middleware.AdminOnlyHandler(h)
}

func AdminOnlyHandlerFunc(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userRole, ok := ctx.Value(roleContextKey).(string)
		if !ok || userRole != "admin" {
			apiErr := dto.ErrAuthAdminRequired
			apiErr.Detail = "Admin privileges required"
			dto.RespondWithAPIError(w, apiErr)
			return
		}
		handlerFunc(w, r)
	}
}

func IsChatSnapshotUUID(r *http.Request) bool {
	return middleware.IsChatSnapshotUUID(r)
}

// --- Rate limiting shim ---

func RateLimitByUserID(q *sqlc_queries.Queries) func(http.Handler) http.Handler {
	return middleware.RateLimitByUserID(q, appConfig.OPENAI.RATELIMIT)
}

// --- Gzip shim ---

func makeGzipHandler(next http.Handler) http.Handler {
	return middleware.MakeGzipHandler(next)
}

// --- Request time tracking shim ---

var requestTracker = middleware.NewLastRequestTracker()

func UpdateLastRequestTime(next http.Handler) http.Handler {
	return middleware.UpdateLastRequestTime(requestTracker)(next)
}

// lastRequest retained for backward compatibility.
var lastRequest time.Time
