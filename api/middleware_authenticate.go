package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/gin-gonic/gin"
	"github.com/swuecho/chat_backend/auth"
	"github.com/swuecho/chat_backend/sqlc_queries"
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
	Token     *jwt.Token
	Claims    jwt.MapClaims
	UserID    string
	Role      string
	TokenType string
	Valid     bool
	Error     *APIError
}

type contextKey string

const (
	roleContextKey contextKey = "role"
	userContextKey contextKey = "user"
	guidContextKey contextKey = "guid"
)

// ginExtractBearerToken extracts bearer token from Gin context
func ginExtractBearerToken(c *gin.Context) string {
	bearerToken := c.GetHeader("Authorization")
	tokenParts := strings.Split(bearerToken, " ")
	if len(tokenParts) == 2 {
		return tokenParts[1]
	}
	return ""
}

// ginParseAndValidateJWT parses and validates JWT with provided secret
func ginParseAndValidateJWT(bearerToken string, expectedTokenType string, jwtSecret sqlc_queries.JwtSecret) *AuthTokenResult {
	result := &AuthTokenResult{}

	if bearerToken == "" {
		err := ErrAuthInvalidCredentials
		err.Detail = "Authorization token required"
		result.Error = &err
		return result
	}

	jwtSigningKey := []byte(jwtSecret.Secret)
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

	tokenType, ok := claims["token_type"].(string)
	if !ok {
		// Legacy forever tokens were generated before the token_type claim existed.
		// Treat them as access tokens so they remain usable.
		if expectedTokenType == "" || expectedTokenType == auth.TokenTypeAccess {
			tokenType = auth.TokenTypeAccess
		} else {
			apiErr := ErrAuthInvalidCredentials
			apiErr.Detail = "Token type not found in token"
			result.Error = &apiErr
			return result
		}
	}

	if expectedTokenType != "" && tokenType != expectedTokenType {
		apiErr := ErrAuthInvalidCredentials
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

// GinAdminAuthMiddleware - Gin middleware for admin authentication
func GinAdminAuthMiddleware(jwtSecret sqlc_queries.JwtSecret) gin.HandlerFunc {
	return func(c *gin.Context) {
		bearerToken := ginExtractBearerToken(c)
		result := ginParseAndValidateJWT(bearerToken, auth.TokenTypeAccess, jwtSecret)

		if result.Error != nil {
			result.Error.GinResponse(c)
			c.Abort()
			return
		}

		// Admin-only check
		if result.Role != "admin" {
			apiErr := ErrAuthAdminRequired
			apiErr.Detail = "Admin privileges required"
			apiErr.GinResponse(c)
			c.Abort()
			return
		}

		// Set user context
		SetUserContext(c, result.UserID, result.Role)
		c.Next()
	}
}

// GinUserAuthMiddleware - Gin middleware for user authentication
func GinUserAuthMiddleware(jwtSecret sqlc_queries.JwtSecret) gin.HandlerFunc {
	return func(c *gin.Context) {
		bearerToken := ginExtractBearerToken(c)
		result := ginParseAndValidateJWT(bearerToken, auth.TokenTypeAccess, jwtSecret)

		if result.Error != nil {
			result.Error.GinResponse(c)
			c.Abort()
			return
		}

		// Set user context
		SetUserContext(c, result.UserID, result.Role)
		c.Next()
	}
}

// GinAdminOnlyMiddleware - Gin middleware that checks for admin role (use after auth middleware)
func GinAdminOnlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role := GetUserRole(c)
		if role != "admin" {
			apiErr := ErrAuthAdminRequired
			apiErr.Detail = "Admin privileges required for this endpoint"
			apiErr.GinResponse(c)
			c.Abort()
			return
		}
		c.Next()
	}
}

// IsChatSnapshotUUID checks if the request is for a chat snapshot UUID (used for public access)
func IsChatSnapshotUUID(r *http.Request) bool {
	const snapshotPrefix = "/api/uuid/chat_snapshot/"
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
