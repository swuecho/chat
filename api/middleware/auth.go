// Package middleware provides HTTP middleware for the chat application.
package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/swuecho/chat_backend/auth"
	"github.com/swuecho/chat_backend/dto"
)

// Context keys for storing values in request context.
type contextKey string

const (
	RoleContextKey contextKey = "role"
	UserContextKey contextKey = "user"
	GuidContextKey contextKey = "guid"
)

// JWTSecret holds the JWT signing secret used across middleware.
var JWTSecret string

// SetJWTSecret sets the JWT secret for middleware.
func SetJWTSecret(secret string) {
	JWTSecret = secret
}

// ExtractBearerToken extracts the bearer token from an Authorization header.
func ExtractBearerToken(r *http.Request) string {
	bearerToken := r.Header.Get("Authorization")
	tokenParts := strings.Split(bearerToken, " ")
	if len(tokenParts) == 2 {
		return tokenParts[1]
	}
	return ""
}

// CreateUserContext adds user ID and role to the request context.
func CreateUserContext(r *http.Request, userID, role string) *http.Request {
	ctx := context.WithValue(r.Context(), UserContextKey, userID)
	ctx = context.WithValue(ctx, RoleContextKey, role)
	return r.WithContext(ctx)
}

// ParseAndValidateJWT parses and validates a JWT token.
func ParseAndValidateJWT(bearerToken string, expectedTokenType string) *AuthTokenResult {
	result := &AuthTokenResult{}

	if bearerToken == "" {
		err := dto.ErrAuthInvalidCredentials
		err.Detail = "Authorization token required"
		result.Error = &err
		return result
	}

	jwtSigningKey := []byte(JWTSecret)
	token, err := jwt.Parse(bearerToken, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("invalid JWT signing method")
		}
		return jwtSigningKey, nil
	})

	if err != nil {
		apiErr := dto.ErrAuthInvalidCredentials
		apiErr.Detail = "Invalid authorization token"
		result.Error = &apiErr
		return result
	}

	if !token.Valid {
		apiErr := dto.ErrAuthInvalidCredentials
		apiErr.Detail = "Token is not valid"
		result.Error = &apiErr
		return result
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		apiErr := dto.ErrAuthInvalidCredentials
		apiErr.Detail = "Cannot parse token claims"
		result.Error = &apiErr
		return result
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		apiErr := dto.ErrAuthInvalidCredentials
		apiErr.Detail = "User ID not found in token"
		result.Error = &apiErr
		return result
	}

	role, ok := claims["role"].(string)
	if !ok {
		apiErr := dto.ErrAuthInvalidCredentials
		apiErr.Detail = "User role not found in token"
		result.Error = &apiErr
		return result
	}

	tokenType, ok := claims["token_type"].(string)
	if !ok {
		if expectedTokenType == "" || expectedTokenType == auth.TokenTypeAccess {
			tokenType = auth.TokenTypeAccess
		} else {
			apiErr := dto.ErrAuthInvalidCredentials
			apiErr.Detail = "Token type not found in token"
			result.Error = &apiErr
			return result
		}
	}

	if expectedTokenType != "" && tokenType != expectedTokenType {
		apiErr := dto.ErrAuthInvalidCredentials
		apiErr.Detail = "Token type is not valid for this operation"
		result.Error = &apiErr
		return result
	}

	result.Token = token
	result.Claims = claims
	result.UserID = userID
	result.Role = role
	result.TokenType = tokenType
	result.Valid = true
	return result
}

// AuthTokenResult holds the result of JWT parsing and validation.
type AuthTokenResult struct {
	Token     *jwt.Token
	Claims    jwt.MapClaims
	UserID    string
	Role      string
	TokenType string
	Valid     bool
	Error     *dto.APIError
}

// GetUserID extracts the user ID from a request context.
func GetUserID(ctx context.Context) (int32, error) {
	userIdValue := ctx.Value(UserContextKey)
	if userIdValue == nil {
		return 0, fmt.Errorf("no user ID in context")
	}
	return parseInt32(userIdValue.(string))
}

func parseInt32(s string) (int32, error) {
	var n int32
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}

// AdminAuthMiddleware provides authentication + admin authorization.
func AdminAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bearerToken := ExtractBearerToken(r)
		result := ParseAndValidateJWT(bearerToken, auth.TokenTypeAccess)

		if result.Error != nil {
			dto.RespondWithAPIError(w, *result.Error)
			return
		}

		if result.Role != "admin" {
			apiErr := dto.ErrAuthAdminRequired
			apiErr.Detail = "Admin privileges required"
			dto.RespondWithAPIError(w, apiErr)
			return
		}

		next.ServeHTTP(w, CreateUserContext(r, result.UserID, result.Role))
	})
}

// UserAuthMiddleware provides authentication for regular user routes.
func UserAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bearerToken := ExtractBearerToken(r)
		result := ParseAndValidateJWT(bearerToken, auth.TokenTypeAccess)

		if result.Error != nil {
			dto.RespondWithAPIError(w, *result.Error)
			return
		}

		next.ServeHTTP(w, CreateUserContext(r, result.UserID, result.Role))
	})
}

// AdminOnlyHandler wraps a handler to require admin role.
func AdminOnlyHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userRole, ok := r.Context().Value(RoleContextKey).(string)
		if !ok || userRole != "admin" {
			apiErr := dto.ErrAuthAdminRequired
			apiErr.Detail = "Admin privileges required"
			dto.RespondWithAPIError(w, apiErr)
			return
		}
		h.ServeHTTP(w, r)
	})
}

// IsChatSnapshotUUID checks if the request is for a public chat snapshot.
func IsChatSnapshotUUID(r *http.Request) bool {
	const snapshotPrefix = "/api/uuid/chat_snapshot/"
	if r.Method != http.MethodGet {
		return false
	}
	return strings.HasPrefix(r.URL.Path, snapshotPrefix) && !strings.HasSuffix(r.URL.Path, "/all")
}
