// Package requestctx owns typed values attached to an HTTP request context.
// It is dependency-free so middleware and handlers can share it without cycles.
package requestctx

import (
	"context"
	"fmt"
)

type contextKey uint8

const (
	principalKey contextKey = iota
	requestIDKey
)

// Principal is the authenticated caller established by auth middleware.
type Principal struct {
	UserID int32
	Role   string
}

func WithPrincipal(ctx context.Context, principal Principal) context.Context {
	return context.WithValue(ctx, principalKey, principal)
}

func PrincipalFrom(ctx context.Context) (Principal, error) {
	principal, ok := ctx.Value(principalKey).(Principal)
	if !ok || principal.UserID <= 0 {
		return Principal{}, fmt.Errorf("authenticated principal is missing")
	}
	return principal, nil
}

func UserID(ctx context.Context) (int32, error) {
	principal, err := PrincipalFrom(ctx)
	if err != nil {
		return 0, err
	}
	return principal.UserID, nil
}

func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

func RequestID(ctx context.Context) string {
	requestID, _ := ctx.Value(requestIDKey).(string)
	return requestID
}
