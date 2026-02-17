package main

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// Context keys for Gin
const (
	ContextKeyUserID   = "user_id"
	ContextKeyRole     = "role"
	ContextKeyGUID     = "guid"
	ContextKeyToken    = "token"
)

// GetUserID extracts user ID from Gin context as int32
func GetUserID(c *gin.Context) (int32, error) {
	userIDVal, exists := c.Get(ContextKeyUserID)
	if !exists {
		return 0, ErrAuthInvalidCredentials
	}

	switch v := userIDVal.(type) {
	case int32:
		return v, nil
	case int:
		return int32(v), nil
	case int64:
		return int32(v), nil
	case string:
		userID, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			return 0, ErrAuthInvalidCredentials
		}
		return int32(userID), nil
	default:
		return 0, ErrAuthInvalidCredentials
	}
}

// GetUserIDString extracts user ID from Gin context as string
func GetUserIDString(c *gin.Context) (string, bool) {
	userID, exists := c.Get(ContextKeyUserID)
	if !exists {
		return "", false
	}
	switch v := userID.(type) {
	case string:
		return v, true
	case int32:
		return strconv.FormatInt(int64(v), 10), true
	case int:
		return strconv.Itoa(v), true
	default:
		return "", false
	}
}

// GetUserRole extracts user role from Gin context
func GetUserRole(c *gin.Context) string {
	role, exists := c.Get(ContextKeyRole)
	if !exists {
		return ""
	}
	if roleStr, ok := role.(string); ok {
		return roleStr
	}
	return ""
}

// IsAdmin checks if the current user has admin role
func IsAdmin(c *gin.Context) bool {
	return GetUserRole(c) == "admin"
}

// SetUserContext sets user information in Gin context
func SetUserContext(c *gin.Context, userID string, role string) {
	c.Set(ContextKeyUserID, userID)
	c.Set(ContextKeyRole, role)
}

// SetUserContextInt sets user information in Gin context with int32 user ID
func SetUserContextInt(c *gin.Context, userID int32, role string) {
	c.Set(ContextKeyUserID, userID)
	c.Set(ContextKeyRole, role)
}

// GetToken extracts JWT token from Gin context
func GetToken(c *gin.Context) (string, bool) {
	token, exists := c.Get(ContextKeyToken)
	if !exists {
		return "", false
	}
	if tokenStr, ok := token.(string); ok {
		return tokenStr, true
	}
	return "", false
}

// SetToken sets JWT token in Gin context
func SetToken(c *gin.Context, token string) {
	c.Set(ContextKeyToken, token)
}

// GetGUID extracts request GUID from Gin context
func GetGUID(c *gin.Context) string {
	guid, exists := c.Get(ContextKeyGUID)
	if !exists {
		return ""
	}
	if guidStr, ok := guid.(string); ok {
		return guidStr
	}
	return ""
}

// SetGUID sets request GUID in Gin context
func SetGUID(c *gin.Context, guid string) {
	c.Set(ContextKeyGUID, guid)
}

// CheckPermission checks if the current user has permission to access resource for given userID
func GinCheckPermission(resourceUserID int32, c *gin.Context) bool {
	currentUserID, err := GetUserID(c)
	if err != nil {
		return false
	}
	role := GetUserRole(c)

	switch role {
	case "admin":
		return true
	case "member":
		return resourceUserID == currentUserID
	default:
		return false
	}
}
