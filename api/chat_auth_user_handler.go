package main

import (
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/swuecho/chat_backend/auth"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

// Constants for token management
const (
	AccessTokenLifetime  = 30 * time.Minute
	RefreshTokenLifetime = 7 * 24 * time.Hour // 7 days
	RefreshTokenName     = "refresh_token"
)

// LoginParams holds login/signup request parameters
type LoginParams struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthUserHandler struct {
	service *AuthUserService
}

func NewAuthUserHandler(sqlc_q *sqlc_queries.Queries) *AuthUserHandler {
	userService := NewAuthUserService(sqlc_q)
	return &AuthUserHandler{service: userService}
}

// GinRegister registers authenticated routes with Gin router
func (h *AuthUserHandler) GinRegister(rg *gin.RouterGroup) {
	rg.GET("/users", h.GinGetUserByID)
	rg.PUT("/users/:id", h.GinUpdateSelf)
	rg.GET("/token_10years", h.GinForeverToken)
}

// GinRegisterPublic registers public routes with Gin router
func (h *AuthUserHandler) GinRegisterPublic(rg *gin.RouterGroup) {
	rg.POST("/signup", h.GinSignUp)
	rg.POST("/login", h.GinLogin)
	rg.POST("/auth/refresh", h.GinRefreshToken)
	rg.POST("/logout", h.GinLogout)
}

// =============================================================================
// Gin Handlers
// =============================================================================

// ginIsHTTPS checks if the request is using HTTPS
func ginIsHTTPS(c *gin.Context) bool {
	if c.Request.TLS != nil {
		return true
	}
	if c.GetHeader("X-Forwarded-Proto") == "https" {
		return true
	}
	if c.GetHeader("X-Forwarded-Ssl") == "on" {
		return true
	}
	env := os.Getenv("ENV")
	if env == "" {
		env = os.Getenv("ENVIRONMENT")
	}
	if env == "" {
		env = os.Getenv("NODE_ENV")
	}
	return env == "production" || env == "prod"
}

// ginCreateSecureRefreshCookie creates a secure cookie for Gin
func ginCreateSecureRefreshCookie(name, value string, maxAge int, c *gin.Context) *http.Cookie {
	sameSite := http.SameSiteLaxMode
	if ginIsHTTPS(c) {
		sameSite = http.SameSiteStrictMode
	}

	var domain string
	host := c.Request.Host
	if host != "" && !strings.HasPrefix(host, "localhost") && !strings.HasPrefix(host, "127.0.0.1") {
		if strings.Contains(host, ":") {
			domain = strings.Split(host, ":")[0]
		} else {
			domain = host
		}
	}

	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		HttpOnly: true,
		Secure:   ginIsHTTPS(c),
		SameSite: sameSite,
		Path:     "/",
		MaxAge:   maxAge,
	}

	if domain != "" && domain != "localhost" && domain != "127.0.0.1" {
		cookie.Domain = domain
	}

	return cookie
}

func (h *AuthUserHandler) GinGetUserByID(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		ErrAuthInvalidCredentials.WithDebugInfo(err.Error()).GinResponse(c)
		return
	}
	user, err := h.service.GetAuthUserByID(c.Request.Context(), userID)
	if err != nil {
		ErrResourceNotFound("user").GinResponse(c)
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *AuthUserHandler) GinUpdateSelf(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		ErrAuthInvalidCredentials.WithDebugInfo(err.Error()).GinResponse(c)
		return
	}

	var userParams sqlc_queries.UpdateAuthUserParams
	if err := c.ShouldBindJSON(&userParams); err != nil {
		ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error()).GinResponse(c)
		return
	}
	userParams.ID = userID
	user, err := h.service.q.UpdateAuthUser(c.Request.Context(), userParams)
	if err != nil {
		log.WithError(err).Error("Failed to update user")
		ErrInternalUnexpected.WithMessage("Failed to update user").WithDebugInfo(err.Error()).GinResponse(c)
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *AuthUserHandler) GinForeverToken(c *gin.Context) {
	lifetime := time.Duration(10*365*24) * time.Hour
	userID, err := GetUserID(c)
	if err != nil {
		ErrAuthInvalidCredentials.WithDebugInfo(err.Error()).GinResponse(c)
		return
	}
	userRole := GetUserRole(c)
	token, err := auth.GenerateToken(userID, userRole, jwtSecretAndAud.Secret, jwtSecretAndAud.Audience, lifetime, auth.TokenTypeAccess)
	if err != nil {
		ErrInternalUnexpected.WithMessage("Failed to generate token").WithDebugInfo(err.Error()).GinResponse(c)
		return
	}
	expiresIn := time.Now().Add(lifetime).Unix()
	c.JSON(http.StatusOK, TokenResult{AccessToken: token, ExpiresIn: int(expiresIn)})
}

func (h *AuthUserHandler) GinSignUp(c *gin.Context) {
	var params LoginParams
	if err := c.ShouldBindJSON(&params); err != nil {
		log.WithFields(log.Fields{
			"error":  err.Error(),
			"ip":     c.ClientIP(),
			"action": "signup_decode_error",
		}).Warn("Failed to decode signup request")
		ErrValidationInvalidInput("Invalid request: unable to decode JSON body").WithDebugInfo(err.Error()).GinResponse(c)
		return
	}

	log.WithFields(log.Fields{
		"email":  params.Email,
		"ip":     c.ClientIP(),
		"action": "signup_attempt",
	}).Info("User signup attempt")

	hash, err := auth.GeneratePasswordHash(params.Password)
	if err != nil {
		log.WithFields(log.Fields{
			"email": params.Email,
			"error": err.Error(),
		}).Error("Failed to generate password hash")
		ErrInternalUnexpected.WithMessage("Failed to generate password hash").WithDebugInfo(err.Error()).GinResponse(c)
		return
	}

	userParams := sqlc_queries.CreateAuthUserParams{
		Password: hash,
		Email:    params.Email,
		Username: params.Email,
	}

	user, err := h.service.CreateAuthUser(c.Request.Context(), userParams)
	if err != nil {
		log.WithFields(log.Fields{
			"email": params.Email,
			"error": err.Error(),
		}).Error("Failed to create user")
		WrapError(err, "Failed to create user").GinResponse(c)
		return
	}

	tokenString, err := auth.GenerateToken(user.ID, user.Role(), jwtSecretAndAud.Secret, jwtSecretAndAud.Audience, AccessTokenLifetime, auth.TokenTypeAccess)
	if err != nil {
		log.WithFields(log.Fields{
			"user_id": user.ID,
			"error":   err.Error(),
		}).Error("Failed to generate access token")
		ErrInternalUnexpected.WithMessage("Failed to generate token").WithDebugInfo(err.Error()).GinResponse(c)
		return
	}

	refreshToken, err := auth.GenerateToken(user.ID, user.Role(), jwtSecretAndAud.Secret, jwtSecretAndAud.Audience, RefreshTokenLifetime, auth.TokenTypeRefresh)
	if err != nil {
		log.WithFields(log.Fields{
			"user_id": user.ID,
			"error":   err.Error(),
		}).Error("Failed to generate refresh token")
		ErrInternalUnexpected.WithMessage("Failed to generate refresh token").WithDebugInfo(err.Error()).GinResponse(c)
		return
	}

	refreshCookie := ginCreateSecureRefreshCookie(RefreshTokenName, refreshToken, int(RefreshTokenLifetime.Seconds()), c)
	http.SetCookie(c.Writer, refreshCookie)

	log.WithFields(log.Fields{
		"user_id": user.ID,
		"email":   user.Email,
		"action":  "signup_success",
	}).Info("User signup successful")

	expiresIn := time.Now().Add(AccessTokenLifetime).Unix()
	c.JSON(http.StatusOK, TokenResult{AccessToken: tokenString, ExpiresIn: int(expiresIn)})
}

func (h *AuthUserHandler) GinLogin(c *gin.Context) {
	var loginParams LoginParams
	if err := c.ShouldBindJSON(&loginParams); err != nil {
		log.WithFields(log.Fields{
			"error":  err.Error(),
			"ip":     c.ClientIP(),
			"action": "login_decode_error",
		}).Warn("Failed to decode login request")
		ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error()).GinResponse(c)
		return
	}

	log.WithFields(log.Fields{
		"email":  loginParams.Email,
		"ip":     c.ClientIP(),
		"action": "login_attempt",
	}).Info("User login attempt")

	user, err := h.service.Authenticate(c.Request.Context(), loginParams.Email, loginParams.Password)
	if err != nil {
		log.WithFields(log.Fields{
			"email":  loginParams.Email,
			"ip":     c.ClientIP(),
			"error":  err.Error(),
			"action": "login_failed",
		}).Warn("User login failed")
		ErrAuthInvalidEmailOrPassword.WithDebugInfo(err.Error()).GinResponse(c)
		return
	}

	accessToken, err := auth.GenerateToken(user.ID, user.Role(), jwtSecretAndAud.Secret, jwtSecretAndAud.Audience, AccessTokenLifetime, auth.TokenTypeAccess)
	if err != nil {
		log.WithFields(log.Fields{
			"user_id": user.ID,
			"error":   err.Error(),
		}).Error("Failed to generate access token")
		ErrInternalUnexpected.WithMessage("Failed to generate access token").WithDebugInfo(err.Error()).GinResponse(c)
		return
	}

	refreshToken, err := auth.GenerateToken(user.ID, user.Role(), jwtSecretAndAud.Secret, jwtSecretAndAud.Audience, RefreshTokenLifetime, auth.TokenTypeRefresh)
	if err != nil {
		log.WithFields(log.Fields{
			"user_id": user.ID,
			"error":   err.Error(),
		}).Error("Failed to generate refresh token")
		ErrInternalUnexpected.WithMessage("Failed to generate refresh token").WithDebugInfo(err.Error()).GinResponse(c)
		return
	}

	refreshCookie := ginCreateSecureRefreshCookie(RefreshTokenName, refreshToken, int(RefreshTokenLifetime.Seconds()), c)
	http.SetCookie(c.Writer, refreshCookie)

	log.WithFields(log.Fields{
		"user_id": user.ID,
		"email":   user.Email,
		"action":  "login_success",
	}).Info("User login successful")

	expiresIn := time.Now().Add(AccessTokenLifetime).Unix()
	c.JSON(http.StatusOK, TokenResult{AccessToken: accessToken, ExpiresIn: int(expiresIn)})
}

func (h *AuthUserHandler) GinRefreshToken(c *gin.Context) {
	log.WithFields(log.Fields{
		"ip":     c.ClientIP(),
		"action": "refresh_attempt",
	}).Info("Token refresh attempt")

	refreshCookie, err := c.Request.Cookie(RefreshTokenName)
	if err != nil {
		log.WithFields(log.Fields{
			"ip":     c.ClientIP(),
			"error":  err.Error(),
			"action": "refresh_missing_cookie",
		}).Warn("Missing refresh token cookie")
		ErrAuthInvalidCredentials.WithMessage("Missing refresh token").GinResponse(c)
		return
	}

	result := ginParseAndValidateJWT(refreshCookie.Value, auth.TokenTypeRefresh, jwtSecretAndAud)
	if result.Error != nil {
		log.WithFields(log.Fields{
			"ip":     c.ClientIP(),
			"error":  result.Error.Detail,
			"action": "refresh_invalid_token",
		}).Warn("Invalid refresh token")
		result.Error.GinResponse(c)
		return
	}

	userIDInt, err := strconv.ParseInt(result.UserID, 10, 32)
	if err != nil {
		log.WithFields(log.Fields{
			"user_id": result.UserID,
			"error":   err.Error(),
			"action":  "refresh_invalid_user_id",
		}).Error("Invalid user ID in refresh token")
		ErrAuthInvalidCredentials.WithMessage("Invalid user ID in token").GinResponse(c)
		return
	}

	accessToken, err := auth.GenerateToken(int32(userIDInt), result.Role, jwtSecretAndAud.Secret, jwtSecretAndAud.Audience, AccessTokenLifetime, auth.TokenTypeAccess)
	if err != nil {
		log.WithFields(log.Fields{
			"user_id": userIDInt,
			"error":   err.Error(),
		}).Error("Failed to generate access token during refresh")
		ErrInternalUnexpected.WithMessage("Failed to generate access token").WithDebugInfo(err.Error()).GinResponse(c)
		return
	}

	log.WithFields(log.Fields{
		"user_id": userIDInt,
		"action":  "refresh_success",
	}).Info("Token refresh successful")

	expiresIn := time.Now().Add(AccessTokenLifetime).Unix()
	c.JSON(http.StatusOK, TokenResult{AccessToken: accessToken, ExpiresIn: int(expiresIn)})
}

func (h *AuthUserHandler) GinLogout(c *gin.Context) {
	log.WithFields(log.Fields{
		"ip":     c.ClientIP(),
		"action": "logout_attempt",
	}).Info("User logout attempt")

	refreshCookie := ginCreateSecureRefreshCookie(RefreshTokenName, "", -1, c)
	http.SetCookie(c.Writer, refreshCookie)

	log.WithFields(log.Fields{
		"ip":     c.ClientIP(),
		"action": "logout_success",
	}).Info("User logout successful")

	c.Status(http.StatusOK)
}
