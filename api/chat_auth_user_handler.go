package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/auth"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

type AuthUserHandler struct {
	service *AuthUserService
}

// isHTTPS checks if the request is using HTTPS or if we're in production
func isHTTPS(r *http.Request) bool {
	// Check if request is HTTPS
	if r.TLS != nil {
		return true
	}
	
	// Check common proxy headers
	if r.Header.Get("X-Forwarded-Proto") == "https" {
		return true
	}
	
	if r.Header.Get("X-Forwarded-Ssl") == "on" {
		return true
	}
	
	// Check if environment indicates production
	env := os.Getenv("ENV")
	if env == "" {
		env = os.Getenv("ENVIRONMENT")
	}
	if env == "" {
		env = os.Getenv("NODE_ENV")
	}
	
	return env == "production" || env == "prod"
}

func NewAuthUserHandler(sqlc_q *sqlc_queries.Queries) *AuthUserHandler {
	userService := NewAuthUserService(sqlc_q)
	return &AuthUserHandler{service: userService}
}

func (h *AuthUserHandler) Register(router *mux.Router) {
	// Authenticated user routes
	router.HandleFunc("/users", h.GetUserByID).Methods(http.MethodGet)
	router.HandleFunc("/users/{id}", h.UpdateSelf).Methods(http.MethodPut)
	router.HandleFunc("/token_10years", h.ForeverToken).Methods(http.MethodGet)
}

func (h *AuthUserHandler) RegisterPublicRoutes(router *mux.Router) {
	// Public routes (no authentication required)
	router.HandleFunc("/signup", h.SignUp).Methods(http.MethodPost)
	router.HandleFunc("/login", h.Login).Methods(http.MethodPost)
	router.HandleFunc("/auth/refresh", h.RefreshToken).Methods(http.MethodPost)
	router.HandleFunc("/logout", h.Logout).Methods(http.MethodPost)
}

func (h *AuthUserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var userParams sqlc_queries.CreateAuthUserParams
	err := json.NewDecoder(r.Body).Decode(&userParams)
	if err != nil {
		RespondWithAPIError(w, ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error()))
		return
	}
	user, err := h.service.CreateAuthUser(r.Context(), userParams)
	if err != nil {
		RespondWithAPIError(w, WrapError(err, "Failed to create user"))
		return
	}
	json.NewEncoder(w).Encode(user)
}

func (h *AuthUserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r.Context())
	if err != nil {
		RespondWithAPIError(w, ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}
	user, err := h.service.GetAuthUserByID(r.Context(), userID)
	if err != nil {
		RespondWithAPIError(w, ErrResourceNotFound("user"))
		return
	}
	json.NewEncoder(w).Encode(user)
}

func (h *AuthUserHandler) UpdateSelf(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r.Context())
	if err != nil {
		RespondWithAPIError(w, ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}

	var userParams sqlc_queries.UpdateAuthUserParams
	err = json.NewDecoder(r.Body).Decode(&userParams)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	userParams.ID = userID
	user, err := h.service.q.UpdateAuthUser(r.Context(), userParams)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(user)
}

func (h *AuthUserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	// get user id from var
	// to int32
	var userParams sqlc_queries.UpdateAuthUserByEmailParams
	err := json.NewDecoder(r.Body).Decode(&userParams)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	user, err := h.service.q.UpdateAuthUserByEmail(r.Context(), userParams)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(user)
}

type LoginParams struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthUserHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	var params LoginParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, "Invalid request: unable to decode JSON body", http.StatusBadRequest)
		return
	}
	hash, err := auth.GeneratePasswordHash(params.Password)
	if err != nil {
		RespondWithAPIError(w, ErrInternalUnexpected.WithMessage("Failed to generate password hash").WithDebugInfo(err.Error()))
		return
	}
	userParams := sqlc_queries.CreateAuthUserParams{
		Password: hash,
		Email:    params.Email,
		Username: params.Email,
	}

	user, err := h.service.CreateAuthUser(r.Context(), userParams)
	if err != nil {
		RespondWithAPIError(w, WrapError(err, "Failed to create user"))
		return
	}
	lifetime := time.Duration(jwtSecretAndAud.Lifetime) * time.Hour
	tokenString, err := auth.GenerateToken(user.ID, user.Role(), jwtSecretAndAud.Secret, jwtSecretAndAud.Audience, lifetime)
	if err != nil {
		RespondWithAPIError(w, ErrInternalUnexpected.WithMessage("Failed to generate token").WithDebugInfo(err.Error()))
		return
	}

	// Set refresh token as httpOnly cookie (long-lived)
	refreshTokenLifetime := time.Duration(24*7) * time.Hour // 7 days
	refreshToken, err := auth.GenerateToken(user.ID, user.Role(), jwtSecretAndAud.Secret, jwtSecretAndAud.Audience, refreshTokenLifetime)
	if err != nil {
		RespondWithAPIError(w, ErrInternalUnexpected.WithMessage("Failed to generate refresh token").WithDebugInfo(err.Error()))
		return
	}

	refreshCookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HttpOnly: true,  // Secure from XSS
		Secure:   isHTTPS(r), // Set to true when using HTTPS
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   int(refreshTokenLifetime.Seconds()),
	}
	http.SetCookie(w, refreshCookie)

	w.Header().Set("Content-Type", "application/json")
	// Access token expires in 30 minutes (short-lived)
	accessTokenLifetime := time.Duration(30) * time.Minute
	expiresIn := time.Now().Add(accessTokenLifetime).Unix()
	json.NewEncoder(w).Encode(TokenResult{AccessToken: tokenString, ExpiresIn: int(expiresIn)})
	w.WriteHeader(http.StatusOK)
}

func (h *AuthUserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var loginParams LoginParams
	err := json.NewDecoder(r.Body).Decode(&loginParams)
	if err != nil {
		RespondWithAPIError(w, ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error()))
		return
	}
	user, err := h.service.Authenticate(r.Context(), loginParams.Email, loginParams.Password)
	if err != nil {
		RespondWithAPIError(w, ErrAuthInvalidEmailOrPassword.WithDebugInfo(err.Error()))
		return
	}
	// Generate short-lived access token (30 minutes)
	accessTokenLifetime := time.Duration(30) * time.Minute
	accessToken, err := auth.GenerateToken(user.ID, user.Role(), jwtSecretAndAud.Secret, jwtSecretAndAud.Audience, accessTokenLifetime)
	if err != nil {
		RespondWithAPIError(w, ErrInternalUnexpected.WithMessage("Failed to generate access token").WithDebugInfo(err.Error()))
		return
	}

	// Generate long-lived refresh token (7 days)
	refreshTokenLifetime := time.Duration(24*7) * time.Hour
	refreshToken, err := auth.GenerateToken(user.ID, user.Role(), jwtSecretAndAud.Secret, jwtSecretAndAud.Audience, refreshTokenLifetime)
	if err != nil {
		RespondWithAPIError(w, ErrInternalUnexpected.WithMessage("Failed to generate refresh token").WithDebugInfo(err.Error()))
		return
	}

	// Set refresh token as httpOnly cookie
	refreshCookie := http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,  // Secure from XSS
		Secure:   isHTTPS(r), // Set to true when using HTTPS
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(refreshTokenLifetime.Seconds()),
	}
	http.SetCookie(w, &refreshCookie)

	w.Header().Set("Content-Type", "application/json")
	expiresIn := time.Now().Add(accessTokenLifetime).Unix()
	json.NewEncoder(w).Encode(TokenResult{AccessToken: accessToken, ExpiresIn: int(expiresIn)})
	w.WriteHeader(http.StatusOK)

}

func (h *AuthUserHandler) ForeverToken(w http.ResponseWriter, r *http.Request) {

	lifetime := time.Duration(10*365*24) * time.Hour
	userId, _ := getUserID(r.Context())
	userRole := r.Context().Value(userContextKey).(string)
	token, err := auth.GenerateToken(userId, userRole, jwtSecretAndAud.Secret, jwtSecretAndAud.Audience, lifetime)

	if err != nil {
		RespondWithAPIError(w, ErrInternalUnexpected.WithMessage("Failed to generate token").WithDebugInfo(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	expiresIn := time.Now().Add(lifetime).Unix()
	json.NewEncoder(w).Encode(TokenResult{AccessToken: token, ExpiresIn: int(expiresIn)})
	w.WriteHeader(http.StatusOK)

}

func (h *AuthUserHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	// Get refresh token from httpOnly cookie
	refreshCookie, err := r.Cookie("refresh_token")
	if err != nil {
		RespondWithAPIError(w, ErrAuthInvalidCredentials.WithMessage("Missing refresh token"))
		return
	}

	// Validate refresh token
	result := parseAndValidateJWT(refreshCookie.Value)
	if result.Error != nil {
		RespondWithAPIError(w, *result.Error)
		return
	}

	// Convert UserID string back to int32
	userIDInt, err := strconv.ParseInt(result.UserID, 10, 32)
	if err != nil {
		RespondWithAPIError(w, ErrAuthInvalidCredentials.WithMessage("Invalid user ID in token"))
		return
	}

	// Generate new short-lived access token
	accessTokenLifetime := time.Duration(30) * time.Minute
	accessToken, err := auth.GenerateToken(int32(userIDInt), result.Role, jwtSecretAndAud.Secret, jwtSecretAndAud.Audience, accessTokenLifetime)
	if err != nil {
		RespondWithAPIError(w, ErrInternalUnexpected.WithMessage("Failed to generate access token").WithDebugInfo(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	expiresIn := time.Now().Add(accessTokenLifetime).Unix()
	json.NewEncoder(w).Encode(TokenResult{AccessToken: accessToken, ExpiresIn: int(expiresIn)})
}

func (h *AuthUserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Clear refresh token cookie
	refreshCookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1, // Delete cookie
	}
	http.SetCookie(w, refreshCookie)
	
	w.WriteHeader(http.StatusOK)
}

type TokenRequest struct {
	Token string `json:"token"`
}

type ResetPasswordRequest struct {
	Email string `json:"email"`
}

func (h *AuthUserHandler) ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req ResetPasswordRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		RespondWithAPIError(w, ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error()))
		return
	}

	// Retrieve user account from the database by email address
	user, err := h.service.q.GetUserByEmail(context.Background(), req.Email)
	if err != nil {
		RespondWithAPIError(w, ErrResourceNotFound("user"))
		return
	}

	// Generate temporary password
	tempPassword := auth.GenerateRandomPassword()

	// Hash temporary password
	hashedPassword, err := auth.GeneratePasswordHash(tempPassword)
	if err != nil {
		RespondWithAPIError(w, ErrInternalUnexpected.WithMessage("Failed to hash password").WithDebugInfo(err.Error()))
		return
	}

	// Update user account with new hashed password
	err = h.service.q.UpdateUserPassword(
		context.Background(),
		sqlc_queries.UpdateUserPasswordParams{
			Email:    req.Email,
			Password: hashedPassword,
		})
	if err != nil {
		RespondWithAPIError(w, ErrInternalUnexpected.WithMessage("Failed to update password").WithDebugInfo(err.Error()))
		return
	}

	// Send email to the user with temporary password and instructions
	err = SendPasswordResetEmail(user.Email, tempPassword)
	if err != nil {
		RespondWithAPIError(w, ErrInternalUnexpected.WithMessage("Failed to send password reset email").WithDebugInfo(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
}

func SendPasswordResetEmail(email, tempPassword string) error {
	return nil
}

type ChangePasswordRequest struct {
	Email       string `json:"email"`
	NewPassword string `json:"new_password"`
}

func (h *AuthUserHandler) ChangePasswordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req ChangePasswordRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Hash new password
	hashedPassword, err := auth.GeneratePasswordHash(req.NewPassword)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Update password in the database
	err = h.service.q.UpdateUserPassword(context.Background(), sqlc_queries.UpdateUserPasswordParams{
		Email:    req.Email,
		Password: string(hashedPassword),
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

type UserStat struct {
	Email                            string `json:"email"`
	FirstName                        string `json:"firstName"`
	LastName                         string `json:"lastName"`
	TotalChatMessages                int64  `json:"totalChatMessages"`
	TotalChatMessagesTokenCount      int64  `json:"totalChatMessagesTokenCount"`
	TotalChatMessages3Days           int64  `json:"totalChatMessages3Days"`
	TotalChatMessages3DaysTokenCount int64  `json:"totalChatMessages3DaysTokenCount"`
	AvgChatMessages3DaysTokenCount   int64  `json:"avgChatMessages3DaysTokenCount"`
	RateLimit                        int32  `json:"rateLimit"`
}

func (h *AuthUserHandler) UserStatHandler(w http.ResponseWriter, r *http.Request) {
	var pagination Pagination
	err := json.NewDecoder(r.Body).Decode(&pagination)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userStatsRows, total, err := h.service.GetUserStats(r.Context(), pagination, int32(appConfig.OPENAI.RATELIMIT))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create a new []interface{} slice with same length as userStatsRows
	data := make([]interface{}, len(userStatsRows))

	// Copy the contents of userStatsRows into data
	for i, v := range userStatsRows {
		divider := v.TotalChatMessages3Days
		var avg int64
		if divider > 0 {
			avg = v.TotalTokenCount3Days / v.TotalChatMessages3Days
		} else {
			avg = 0
		}
		data[i] = UserStat{
			Email:                            v.UserEmail,
			FirstName:                        v.FirstName,
			LastName:                         v.LastName,
			TotalChatMessages:                v.TotalChatMessages,
			TotalChatMessages3Days:           v.TotalChatMessages3Days,
			RateLimit:                        v.RateLimit,
			TotalChatMessagesTokenCount:      v.TotalTokenCount,
			TotalChatMessages3DaysTokenCount: v.TotalTokenCount3Days,
			AvgChatMessages3DaysTokenCount:   avg,
		}
	}

	json.NewEncoder(w).Encode(Pagination{
		Page:  pagination.Page,
		Size:  pagination.Size,
		Total: total,
		Data:  data,
	})
}

type RateLimitRequest struct {
	Email     string `json:"email"`
	RateLimit int32  `json:"rateLimit"`
}

func (h *AuthUserHandler) UpdateRateLimit(w http.ResponseWriter, r *http.Request) {
	var rateLimitRequest RateLimitRequest
	err := json.NewDecoder(r.Body).Decode(&rateLimitRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	rate, err := h.service.q.UpdateAuthUserRateLimitByEmail(r.Context(),
		sqlc_queries.UpdateAuthUserRateLimitByEmailParams{
			Email:     rateLimitRequest.Email,
			RateLimit: rateLimitRequest.RateLimit,
		})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(
		map[string]int32{
			"rate": rate,
		})
}
func (h *AuthUserHandler) GetRateLimit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		RespondWithAPIError(w, ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}

	rate, err := h.service.q.GetRateLimit(ctx, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]int32{
		"rate": rate,
	})
}

