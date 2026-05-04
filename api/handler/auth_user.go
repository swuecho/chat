package handler

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/auth"
	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/middleware"
	"github.com/swuecho/chat_backend/sqlc_queries"
	"github.com/swuecho/chat_backend/svc"
	"log/slog"
)

// Token lifetime constants.
const (
	AccessTokenLifetime         = 30 * time.Minute
	DefaultRefreshTokenLifetime = 7 * 24 * time.Hour
	MobileRefreshTokenLifetime  = 90 * 24 * time.Hour
	RefreshTokenName            = "refresh_token"
	MobileClientHeader          = "X-Chat-Client"
	MobileClientValue           = "mobile"
)

// AuthUserHandler handles authentication HTTP requests.
type AuthUserHandler struct {
	service          *svc.AuthUserService
	jwtSecret        string
	audience         string
	defaultRateLimit int32
}

// NewAuthUserHandler creates a new AuthUserHandler.
func NewAuthUserHandler(sqlc_q *sqlc_queries.Queries, jwtSecret, audience string, defaultRateLimit int32) *AuthUserHandler {
	return &AuthUserHandler{
		service:          svc.NewAuthUserService(sqlc_q, jwtSecret, defaultRateLimit),
		jwtSecret:        jwtSecret,
		audience:         audience,
		defaultRateLimit: defaultRateLimit,
	}
}

// --- Cookie helpers ---

func isHTTPS(r *http.Request) bool {
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" || r.Header.Get("X-Forwarded-Ssl") == "on" {
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

func createSecureRefreshCookie(name, value string, maxAge int, r *http.Request) *http.Cookie {
	sameSite := http.SameSiteLaxMode
	if isHTTPS(r) {
		sameSite = http.SameSiteStrictMode
	}

	var domain string
	host := r.Host
	if host != "" && !strings.HasPrefix(host, "localhost") && !strings.HasPrefix(host, "127.0.0.1") {
		if i := strings.IndexByte(host, ':'); i >= 0 {
			domain = host[:i]
		} else {
			domain = host
		}
	}

	cookie := &http.Cookie{
		Name: name, Value: value, HttpOnly: true,
		Secure: isHTTPS(r), SameSite: sameSite, Path: "/", MaxAge: maxAge,
	}
	if domain != "" && domain != "localhost" && domain != "127.0.0.1" {
		cookie.Domain = domain
	}
	return cookie
}

func refreshTokenLifetimeForRequest(r *http.Request) time.Duration {
	if strings.EqualFold(r.Header.Get(MobileClientHeader), MobileClientValue) {
		return MobileRefreshTokenLifetime
	}
	return DefaultRefreshTokenLifetime
}

// --- Route registration ---

func (h *AuthUserHandler) Register(router *mux.Router) {
	router.HandleFunc("/users", h.GetUserByID).Methods(http.MethodGet)
	router.HandleFunc("/users/{id}", h.UpdateSelf).Methods(http.MethodPut)
	router.HandleFunc("/token_10years", h.ForeverToken).Methods(http.MethodGet)
}

func (h *AuthUserHandler) RegisterPublicRoutes(router *mux.Router) {
	router.HandleFunc("/signup", h.SignUp).Methods(http.MethodPost)
	router.HandleFunc("/login", h.Login).Methods(http.MethodPost)
	router.HandleFunc("/auth/refresh", h.RefreshToken).Methods(http.MethodPost)
	router.HandleFunc("/logout", h.Logout).Methods(http.MethodPost)
}

// --- CRUD handlers ---

func (h *AuthUserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var params sqlc_queries.CreateAuthUserParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error()))
		return
	}
	user, err := h.service.CreateAuthUser(r.Context(), params)
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(err, "Failed to create user"))
		return
	}
	json.NewEncoder(w).Encode(user)
}

func (h *AuthUserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r.Context())
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}
	user, err := h.service.GetAuthUserByID(r.Context(), userID)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrResourceNotFound("user"))
		return
	}
	json.NewEncoder(w).Encode(user)
}

func (h *AuthUserHandler) UpdateSelf(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r.Context())
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}
	var params sqlc_queries.UpdateAuthUserParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error()))
		return
	}
	params.ID = userID
	user, err := h.service.UpdateAuthUser(r.Context(), params)
	if err != nil {
		slog.Error("Failed to update user", "error", err)
		dto.RespondWithAPIError(w, dto.ErrInternalUnexpected.WithMessage("Failed to update user").WithDebugInfo(err.Error()))
		return
	}
	json.NewEncoder(w).Encode(user)
}

func (h *AuthUserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	var params sqlc_queries.UpdateAuthUserByEmailParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error()))
		return
	}
	user, err := h.service.UpdateAuthUserByEmail(r.Context(), params)
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(dto.MapDatabaseError(err), "Failed to update user"))
		return
	}
	json.NewEncoder(w).Encode(user)
}

// --- Auth handlers ---

type LoginParams struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthUserHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	var params LoginParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		slog.Warn("Failed to decode signup", "error", err, "ip", r.RemoteAddr, "action", "signup_decode_error")
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("Invalid request: unable to decode JSON body").WithDebugInfo(err.Error()))
		return
	}

	hash, err := auth.GeneratePasswordHash(params.Password)
	if err != nil {
		slog.Error("Failed to hash password", "email", params.Email, "error", err)
		dto.RespondWithAPIError(w, dto.ErrInternalUnexpected.WithMessage("Failed to generate password hash").WithDebugInfo(err.Error()))
		return
	}

	user, err := h.service.CreateAuthUser(r.Context(), sqlc_queries.CreateAuthUserParams{
		Password: hash, Email: params.Email, Username: params.Email,
	})
	if err != nil {
		slog.Error("Failed to create user", "email", params.Email, "error", err)
		dto.RespondWithAPIError(w, dto.WrapError(err, "Failed to create user"))
		return
	}

	accessToken, err := auth.GenerateToken(user.ID, user.Role(), h.jwtSecret, h.audience, AccessTokenLifetime, auth.TokenTypeAccess)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrInternalUnexpected.WithMessage("Failed to generate token").WithDebugInfo(err.Error()))
		return
	}

	refreshLifetime := refreshTokenLifetimeForRequest(r)
	refreshToken, err := auth.GenerateToken(user.ID, user.Role(), h.jwtSecret, h.audience, refreshLifetime, auth.TokenTypeRefresh)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrInternalUnexpected.WithMessage("Failed to generate refresh token").WithDebugInfo(err.Error()))
		return
	}

	http.SetCookie(w, createSecureRefreshCookie(RefreshTokenName, refreshToken, int(refreshLifetime.Seconds()), r))

	slog.Info("User signup successful", "user_id", user.ID, "email", user.Email, "action", "signup_success")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.TokenResult{AccessToken: accessToken, ExpiresIn: int(time.Now().Add(AccessTokenLifetime).Unix())})
}

func (h *AuthUserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var params LoginParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		slog.Warn("Failed to decode login", "error", err, "ip", r.RemoteAddr, "action", "login_decode_error")
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error()))
		return
	}

	user, err := h.service.Authenticate(r.Context(), params.Email, params.Password)
	if err != nil {
		slog.Warn("User login failed", "email", params.Email, "ip", r.RemoteAddr, "error", err, "action", "login_failed")
		dto.RespondWithAPIError(w, dto.ErrAuthInvalidEmailOrPassword.WithDebugInfo(err.Error()))
		return
	}

	accessToken, err := auth.GenerateToken(user.ID, user.Role(), h.jwtSecret, h.audience, AccessTokenLifetime, auth.TokenTypeAccess)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrInternalUnexpected.WithMessage("Failed to generate access token").WithDebugInfo(err.Error()))
		return
	}

	refreshLifetime := refreshTokenLifetimeForRequest(r)
	refreshToken, err := auth.GenerateToken(user.ID, user.Role(), h.jwtSecret, h.audience, refreshLifetime, auth.TokenTypeRefresh)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrInternalUnexpected.WithMessage("Failed to generate refresh token").WithDebugInfo(err.Error()))
		return
	}

	http.SetCookie(w, createSecureRefreshCookie(RefreshTokenName, refreshToken, int(refreshLifetime.Seconds()), r))

	slog.Info("User login successful", "user_id", user.ID, "email", user.Email, "action", "login_success")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.TokenResult{AccessToken: accessToken, ExpiresIn: int(time.Now().Add(AccessTokenLifetime).Unix())})
}

func (h *AuthUserHandler) ForeverToken(w http.ResponseWriter, r *http.Request) {
	lifetime := time.Duration(10*365*24) * time.Hour
	userId, _ := getUserID(r.Context())
	userRole, _ := r.Context().Value(middleware.RoleContextKey).(string)

	token, err := auth.GenerateToken(userId, userRole, h.jwtSecret, h.audience, lifetime, auth.TokenTypeAccess)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrInternalUnexpected.WithMessage("Failed to generate token").WithDebugInfo(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.TokenResult{AccessToken: token, ExpiresIn: int(time.Now().Add(lifetime).Unix())})
}

func (h *AuthUserHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	slog.Info("Token refresh attempt", "ip", r.RemoteAddr, "action", "refresh_attempt")

	refreshCookie, err := r.Cookie(RefreshTokenName)
	if err != nil {
		slog.Warn("Missing refresh token cookie", "ip", r.RemoteAddr, "error", err, "action", "refresh_missing_cookie")
		dto.RespondWithAPIError(w, dto.ErrAuthInvalidCredentials.WithMessage("Missing refresh token"))
		return
	}

	result := middleware.ParseAndValidateJWT(refreshCookie.Value, auth.TokenTypeRefresh, h.jwtSecret)
	if result.Error != nil {
		slog.Warn("Invalid refresh token", "ip", r.RemoteAddr, "error", result.Error.Detail, "action", "refresh_invalid_token")
		dto.RespondWithAPIError(w, *result.Error)
		return
	}

	userIDInt, err := strconv.ParseInt(result.UserID, 10, 32)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrAuthInvalidCredentials.WithMessage("Invalid user ID in token"))
		return
	}

	accessToken, err := auth.GenerateToken(int32(userIDInt), result.Role, h.jwtSecret, h.audience, AccessTokenLifetime, auth.TokenTypeAccess)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrInternalUnexpected.WithMessage("Failed to generate access token").WithDebugInfo(err.Error()))
		return
	}

	slog.Info("Token refresh successful", "user_id", userIDInt, "action", "refresh_success")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.TokenResult{AccessToken: accessToken, ExpiresIn: int(time.Now().Add(AccessTokenLifetime).Unix())})
}

func (h *AuthUserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, createSecureRefreshCookie(RefreshTokenName, "", -1, r))
	w.WriteHeader(http.StatusOK)
}
