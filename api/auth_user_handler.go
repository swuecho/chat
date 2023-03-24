package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/swuecho/chatgpt_backend/auth"
	"github.com/swuecho/chatgpt_backend/sqlc_queries"
)

type AuthUserHandler struct {
	q       *sqlc_queries.Queries
	service *AuthUserService
}

func NewAuthUserHandler(service *AuthUserService) *AuthUserHandler {
	return &AuthUserHandler{service: service}
}

func (h *AuthUserHandler) Register(router *mux.Router) {
	router.HandleFunc("/users", h.CreateUser).Methods(http.MethodPost)
	router.HandleFunc("/users/{id}", h.GetUserByID).Methods(http.MethodGet)
	router.HandleFunc("/users/{id}", h.UpdateUser).Methods(http.MethodPut)
	router.HandleFunc("/signup", h.SignUp).Methods(http.MethodPost)
	router.HandleFunc("/login", h.Login).Methods(http.MethodPost)
	router.HandleFunc("/verify", h.verify).Methods(http.MethodPost)
	router.HandleFunc("/config", h.configHandler).Methods(http.MethodPost)
	// rate limit handler
	router.HandleFunc("/admin/rate_limit", h.UpdateRateLimit).Methods(http.MethodPost)
	// user stats handler
	router.HandleFunc("/admin/user_stats", h.UserStatHandler).Methods(http.MethodPost)

}

func (h *AuthUserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var userParams sqlc_queries.CreateAuthUserParams
	err := json.NewDecoder(r.Body).Decode(&userParams)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	user, err := h.service.CreateAuthUser(r.Context(), userParams)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(user)
}

func (h *AuthUserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid user ID", http.StatusBadRequest)
		return
	}
	user, err := h.service.GetAuthUserByID(r.Context(), int32(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(user)
}

func (h *AuthUserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid user ID", http.StatusBadRequest)
		return
	}
	var userParams sqlc_queries.UpdateAuthUserParams
	err = json.NewDecoder(r.Body).Decode(&userParams)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	userParams.ID = int32(id)
	user, err := h.service.UpdateAuthUser(r.Context(), userParams)
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
		http.Error(w, "Failed to generate password hash: internal server error", http.StatusInternalServerError)
		return
	}
	userParams := sqlc_queries.CreateAuthUserParams{
		Password: hash,
		Email:    params.Email,
		Username: params.Email,
	}

	user, err := h.service.CreateAuthUser(r.Context(), userParams)
	if err != nil {
		http.Error(w, fmt.Errorf("failed to create user %w", err).Error(), http.StatusInternalServerError)
		return
	}

	tokenString, err := auth.GenerateToken(user.ID, user.Role(), appConfig.JWT.SECRET, appConfig.JWT.AUD)
	if err != nil {
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	cookie := &http.Cookie{
		Name:     "jwt",
		Value:    tokenString,
		HttpOnly: false, // TODO: change to true when https
		Path:     "/",
	}
	http.SetCookie(w, cookie)
	w.Header().Set("Content-Type", "application/json")
	expiresIn := time.Now().Add(time.Hour * 8).Unix()
	json.NewEncoder(w).Encode(TokenResult{AccessToken: tokenString, ExpiresIn: int(expiresIn)})
	w.WriteHeader(http.StatusOK)
}

func (h *AuthUserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var loginParams LoginParams
	err := json.NewDecoder(r.Body).Decode(&loginParams)
	if err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	logger.Warn(fmt.Sprintf("loginParams: %v", loginParams))
	user, err := h.service.Authenticate(r.Context(), loginParams.Email, loginParams.Password)
	if err != nil {
		http.Error(w, fmt.Errorf("invalid email or password: %w", err).Error(), http.StatusUnauthorized)
		return
	}
	token, err := auth.GenerateToken(user.ID, user.Role(), appConfig.JWT.SECRET, appConfig.JWT.AUD)

	if err != nil {
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	cookie := http.Cookie{
		Name:     "jwt",
		Value:    token,
		Path:     "/",
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
	}

	http.SetCookie(w, &cookie)

	w.Header().Set("Content-Type", "application/json")
	expiresIn := time.Now().Add(time.Hour * 8).Unix()
	json.NewEncoder(w).Encode(TokenResult{AccessToken: token, ExpiresIn: int(expiresIn)})
	w.WriteHeader(http.StatusOK)

}

func (h *AuthUserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("jwt")
	if err != nil {
		http.Error(w, "missing token in cookie", http.StatusBadRequest)
	}

	tokenString := c.Value
	if len(tokenString) == 0 {
		http.Error(w, "empty token string", http.StatusBadRequest)
	}

	cookie, err := h.service.Logout(tokenString)
	if err != nil {
		http.Error(w, "logout fail", http.StatusBadRequest)
	}
	http.SetCookie(w, cookie)
	w.WriteHeader(http.StatusOK)
}

type TokenRequest struct {
	Token string `json:"token"`
}

func (h *AuthUserHandler) verify(w http.ResponseWriter, r *http.Request) {
	var req TokenRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Invalid request body: %v", err)
		return
	}

	fmt.Printf("Received token: %s\n", req.Token)
	token := req.Token

	if token == "" {
		http.Error(w, "Secret key is empty", http.StatusBadRequest)
		return
	}
	// AuthSecretKey := os.Getenv("AUTH_SECRET_KEY")
	_, err = auth.ValidateToken(token, appConfig.JWT.SECRET)
	if err != nil {
		http.Error(w, fmt.Errorf("密钥无效 | Secret key is invalid %w", err).Error(), http.StatusUnauthorized)
		return
	}

	// Send a JSON response with status 'Success', message 'Verify successfully', and null data payload
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"data": %v}`, token)
}

type ModelConfig struct {
	ApiModel     string `json:"apiModel"`
	ReverseProxy string `json:"reverseProxy"`
	TimeoutMs    int    `json:"timeoutMs"`
	SocksProxy   string `json:"socksProxy"`
}

func (h *AuthUserHandler) configHandler(w http.ResponseWriter, r *http.Request) {
	apiModel := "ChatGPTAPI" // Replace with your actual API model
	timeoutMs := 5000        // Replace with your actual timeout
	socksProxy := "-"
	if os.Getenv("SOCKS_PROXY_HOST") != "" && os.Getenv("SOCKS_PROXY_PORT") != "" {
		socksProxy = os.Getenv("SOCKS_PROXY_HOST") + ":" + os.Getenv("SOCKS_PROXY_PORT")
	}
	reverseProxy := os.Getenv("API_REVERSE_PROXY")

	// Create a ModelConfig object and populate it with values
	config := ModelConfig{
		ApiModel:     apiModel,
		TimeoutMs:    timeoutMs,
		ReverseProxy: reverseProxy,
		SocksProxy:   socksProxy,
	}

	// Encode the ModelConfig object as JSON and send it as the response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "Success", "data": config})

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
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Retrieve user account from the database by email address
	user, err := h.q.GetUserByEmail(context.Background(), req.Email)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Generate temporary password
	tempPassword := auth.GenerateRandomPassword()

	// Hash temporary password
	hashedPassword, err := auth.GeneratePasswordHash(tempPassword)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Update user account with new hashed password
	err = h.q.UpdateUserPassword(
		context.Background(),
		sqlc_queries.UpdateUserPasswordParams{
			Email:    req.Email,
			Password: hashedPassword,
		})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Send email to the user with temporary password and instructions
	err = SendPasswordResetEmail(user.Email, tempPassword)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
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
	err = h.q.UpdateUserPassword(context.Background(), sqlc_queries.UpdateUserPasswordParams{
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
	Email                  string `json:"email"`
	TotalChatMessages      int64  `json:"totalChatMessages"`
	TotalChatMessages3Days int64  `json:"totalChatMessages3Days"`
	RateLimit              int32  `json:"rateLimit"`
}

func (h *AuthUserHandler) UserStatHandler(w http.ResponseWriter, r *http.Request) {
	var pagination Pagination
	err := json.NewDecoder(r.Body).Decode(&pagination)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userStatsRows, total, err := h.service.GetUserStat(r.Context(), pagination)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create a new []interface{} slice with same length as userStatsRows
	data := make([]interface{}, len(userStatsRows))

	// Copy the contents of userStatsRows into data
	for i, v := range userStatsRows {
		data[i] = UserStat{
			Email:                  v.UserEmail,
			TotalChatMessages:      v.TotalChatMessages,
			TotalChatMessages3Days: v.TotalChatMessages3Days,
			RateLimit:              v.RateLimit,
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
		map[string]interface{}{
			"rate": rate,
		})
}
