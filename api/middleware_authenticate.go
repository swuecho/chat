package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/middleware"
)

// Re-export middleware functions for backward compatibility.
var (
	AdminAuthMiddleware  = middleware.AdminAuthMiddleware
	UserAuthMiddleware   = middleware.UserAuthMiddleware
	parseAndValidateJWT  = middleware.ParseAndValidateJWT
)

// Context key constants matching middleware package.
const (
	roleContextKey = middleware.RoleContextKey
	userContextKey = middleware.UserContextKey
	guidContextKey = middleware.GuidContextKey
)

// getUserID extracts the user ID from context.
func getUserID(ctx context.Context) (int32, error) {
	userIdValue := ctx.Value(userContextKey)
	if userIdValue == nil {
		return 0, fmt.Errorf("no user Id in context")
	}
	userIDStr := userIdValue.(string)
	userIDInt, err := strconv.ParseInt(userIDStr, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid user ID: %s", userIDStr)
	}
	return int32(userIDInt), nil
}

func getContextWithUser(userID int) context.Context {
	return context.WithValue(context.Background(), userContextKey, strconv.Itoa(userID))
}

// AdminOnlyHandler wraps a handler to require admin role.
func AdminOnlyHandler(h http.Handler) http.Handler {
	return middleware.AdminOnlyHandler(h)
}

// AdminOnlyHandlerFunc wraps a handler func to require admin role.
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
