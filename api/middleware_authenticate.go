package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	jwt "github.com/golang-jwt/jwt/v5"
)

func CheckPermission(userID int, ctx context.Context) bool {
	contextUserID, ok := ctx.Value("user_id").(int)
	if !ok {
		return false
	}
	role, ok := ctx.Value("role").(string)
	if !ok {
		return false
	}

	switch role {
	case "admin":
		return true
	case "member":
		return userID == contextUserID
	default:
		return false
	}
}

type AuthTokenResult struct {
	Token  *jwt.Token
	Claims jwt.MapClaims
	UserID string
	Role   string
	Valid  bool
	Error  *APIError
}

func extractBearerToken(r *http.Request) string {
	// Extract from Authorization header for access tokens
	bearerToken := r.Header.Get("Authorization")
	tokenParts := strings.Split(bearerToken, " ")
	if len(tokenParts) == 2 {
		return tokenParts[1]
	}
	return ""
}

func createUserContext(r *http.Request, userID, role string) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, userID)
	ctx = context.WithValue(ctx, roleContextKey, role)
	return r.WithContext(ctx)
}

func parseAndValidateJWT(bearerToken string) *AuthTokenResult {
	result := &AuthTokenResult{}

	if bearerToken == "" {
		err := ErrAuthInvalidCredentials
		err.Detail = "Authorization token required"
		result.Error = &err
		return result
	}

	jwtSigningKey := []byte(jwtSecretAndAud.Secret)
	token, err := jwt.Parse(bearerToken, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("invalid JWT signing method")
		}
		return jwtSigningKey, nil
	})

	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.Detail = "Invalid authorization token"
		result.Error = &apiErr
		return result
	}

	if !token.Valid {
		apiErr := ErrAuthInvalidCredentials
		apiErr.Detail = "Token is not valid"
		result.Error = &apiErr
		return result
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		apiErr := ErrAuthInvalidCredentials
		apiErr.Detail = "Cannot parse token claims"
		result.Error = &apiErr
		return result
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		apiErr := ErrAuthInvalidCredentials
		apiErr.Detail = "User ID not found in token"
		result.Error = &apiErr
		return result
	}

	role, ok := claims["role"].(string)
	if !ok {
		apiErr := ErrAuthInvalidCredentials
		apiErr.Detail = "User role not found in token"
		result.Error = &apiErr
		return result
	}

	result.Token = token
	result.Claims = claims
	result.UserID = userID
	result.Role = role
	result.Valid = true
	return result
}

type contextKey string

const (
	roleContextKey contextKey = "role"
	userContextKey contextKey = "user"
	guidContextKey contextKey = "guid"
)
const snapshotPrefix = "/api/uuid/chat_snapshot/"

func IsChatSnapshotUUID(r *http.Request) bool {
	// Check http method is GET
	if r.Method != http.MethodGet {
		return false
	}
	// Check if request url path has the required prefix and does not have "/all" suffix
	if strings.HasPrefix(r.URL.Path, snapshotPrefix) && !strings.HasSuffix(r.URL.Path, "/all") {
		return true
	}
	return false
}

func AdminOnlyHander(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userRole, ok := ctx.Value(roleContextKey).(string)
		if !ok {
			apiErr := ErrAuthAdminRequired
			apiErr.Detail = "User role information not found"
			RespondWithAPIError(w, apiErr)
			return
		}
		if userRole != "admin" {
			apiErr := ErrAuthAdminRequired
			apiErr.Detail = "Current user does not have admin role"
			RespondWithAPIError(w, apiErr)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func AdminOnlyHandlerFunc(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userRole, ok := ctx.Value(roleContextKey).(string)
		if !ok {
			apiErr := ErrAuthAdminRequired
			apiErr.Detail = "User role information not found"
			RespondWithAPIError(w, apiErr)
			return
		}
		if userRole != "admin" {
			apiErr := ErrAuthAdminRequired
			apiErr.Detail = "Current user does not have admin role"
			RespondWithAPIError(w, apiErr)
			return
		}
		handlerFunc(w, r)
	}
}

// AdminRouteMiddleware applies admin-only protection to all routes in a subrouter
func AdminRouteMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userRole, ok := ctx.Value(roleContextKey).(string)
		if !ok {
			apiErr := ErrAuthAdminRequired
			apiErr.Detail = "User role information not found"
			RespondWithAPIError(w, apiErr)
			return
		}
		if userRole != "admin" {
			apiErr := ErrAuthAdminRequired
			apiErr.Detail = "Admin privileges required for this endpoint"
			RespondWithAPIError(w, apiErr)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// AdminAuthMiddleware - Authentication middleware specifically for admin routes
func AdminAuthMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bearerToken := extractBearerToken(r)
		result := parseAndValidateJWT(bearerToken)

		if result.Error != nil {
			RespondWithAPIError(w, *result.Error)
			return
		}

		// Admin-only check
		if result.Role != "admin" {
			apiErr := ErrAuthAdminRequired
			apiErr.Detail = "Admin privileges required"
			RespondWithAPIError(w, apiErr)
			return
		}

		// Add user context and proceed
		handler.ServeHTTP(w, createUserContext(r, result.UserID, result.Role))
	})
}

// UserAuthMiddleware - Authentication middleware for regular user routes
func UserAuthMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bearerToken := extractBearerToken(r)
		result := parseAndValidateJWT(bearerToken)

		if result.Error != nil {
			RespondWithAPIError(w, *result.Error)
			return
		}

		// Add user context and proceed (no role restrictions for user middleware)
		handler.ServeHTTP(w, createUserContext(r, result.UserID, result.Role))
	})
}

func IsAuthorizedMiddleware(handler http.Handler) http.Handler {
	noAuthPaths := map[string]bool{
		"/":            true,
		"/favicon.ico": true,
		"/api/login":   true,
		"/api/signup":  true,
		"/api/tts":     true,
		"/api/errors":  true,
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := noAuthPaths[r.URL.Path]; ok || strings.HasPrefix(r.URL.Path, "/static") || IsChatSnapshotUUID(r) {
			handler.ServeHTTP(w, r)
			return
		}

		bearerToken := extractBearerToken(r)
		result := parseAndValidateJWT(bearerToken)

		if result.Error != nil {
			RespondWithAPIError(w, *result.Error)
			return
		}

		if result.Valid {
			// superuser
			if strings.HasPrefix(r.URL.Path, "/admin") && result.Role != "admin" {
				apiErr := ErrAuthAdminRequired
				apiErr.Detail = "This endpoint requires admin privileges"
				RespondWithAPIError(w, apiErr)
				return
			}

			// TODO: get trace id and add it to context
			//traceID := r.Header.Get("X-Request-Id")
			//if len(traceID) > 0 {
			//ctx = context.WithValue(ctx, guidContextKey, traceID)
			//}
			// Store user ID and role in the request context
			// pass token to request
			handler.ServeHTTP(w, createUserContext(r, result.UserID, result.Role))
		}
	})
}
