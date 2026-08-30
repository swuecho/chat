package handler

import (
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/auth"
	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/httpx"
	"github.com/swuecho/chat_backend/middleware"
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
func NewAuthUserHandler(service *svc.AuthUserService, jwtSecret, audience string, defaultRateLimit int32) *AuthUserHandler {
	return &AuthUserHandler{
		service:          service,
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
	router.HandleFunc("/users", endpoint(h.GetUserByID)).Methods(http.MethodGet)
	router.HandleFunc("/users/{id}", endpoint(h.UpdateSelf)).Methods(http.MethodPut)
	router.HandleFunc("/token_10years", endpoint(h.ForeverToken)).Methods(http.MethodGet)
}

func (h *AuthUserHandler) RegisterPublicRoutes(router *mux.Router) {
	router.HandleFunc("/signup", endpoint(h.SignUp)).Methods(http.MethodPost)
	router.HandleFunc("/login", endpoint(h.Login)).Methods(http.MethodPost)
	router.HandleFunc("/auth/refresh", endpoint(h.RefreshToken)).Methods(http.MethodPost)
	router.HandleFunc("/logout", endpoint(h.Logout)).Methods(http.MethodPost)
}

// --- CRUD handlers ---

func (h *AuthUserHandler) CreateUser(w http.ResponseWriter, r *http.Request) error {
	var request createAuthUserRequest
	if err := DecodeJSON(r, &request); err != nil {
		return dto.ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error())
	}
	user, err := h.service.CreateAuthUser(r.Context(), request.input())
	if err != nil {
		return dto.WrapError(err, "Failed to create user")
	}
	return respondJSON(w, http.StatusCreated, authUserResponse(user))
}

func (h *AuthUserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) error {
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}
	user, err := h.service.GetAuthUserByID(r.Context(), userID)
	if err != nil {
		return dto.ErrResourceNotFound("user")
	}
	return respondJSON(w, http.StatusOK, authUserResponse(user))
}

func (h *AuthUserHandler) UpdateSelf(w http.ResponseWriter, r *http.Request) error {
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}
	var request updateAuthUserRequest
	if err := DecodeJSON(r, &request); err != nil {
		return dto.ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error())
	}
	user, err := h.service.UpdateAuthUser(r.Context(), request.selfInput(userID))
	if err != nil {
		return dto.ErrInternalUnexpected.WithMessage("Failed to update user").WithDebugInfo(err.Error())
	}
	return respondJSON(w, http.StatusOK, updatedUserResponse(user.FirstName, user.LastName, user.Email))
}

func (h *AuthUserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) error {
	var request updateAuthUserRequest
	if err := DecodeJSON(r, &request); err != nil {
		return dto.ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error())
	}
	user, err := h.service.UpdateAuthUserByEmail(r.Context(), request.emailInput())
	if err != nil {
		return dto.WrapError(dto.MapDatabaseError(err), "Failed to update user")
	}
	return respondJSON(w, http.StatusOK, updatedUserResponse(user.FirstName, user.LastName, user.Email))
}

// --- Auth handlers ---

type LoginParams struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (p *LoginParams) Validate() error {
	p.Email = strings.TrimSpace(p.Email)
	if p.Email == "" || p.Password == "" {
		return httpx.Invalid("email and password are required")
	}
	return nil
}

func (h *AuthUserHandler) SignUp(w http.ResponseWriter, r *http.Request) error {
	var params LoginParams
	if err := DecodeJSON(r, &params); err != nil {
		slog.Warn("Failed to decode signup", "error", err, "ip", r.RemoteAddr, "action", "signup_decode_error")
		return dto.ErrValidationInvalidInput("Invalid request: unable to decode JSON body").WithDebugInfo(err.Error())
	}

	hash, err := auth.GeneratePasswordHash(params.Password)
	if err != nil {
		slog.Error("Failed to hash password", "email", params.Email, "error", err)
		return dto.ErrInternalUnexpected.WithMessage("Failed to generate password hash").WithDebugInfo(err.Error())
	}

	user, err := h.service.CreateAuthUser(r.Context(), svc.CreateAuthUserInput{
		Password: hash, Email: params.Email, Username: params.Email,
	})
	if err != nil {
		slog.Error("Failed to create user", "email", params.Email, "error", err)
		return dto.WrapError(err, "Failed to create user")
	}

	accessToken, err := auth.GenerateToken(user.ID, user.Role(), h.jwtSecret, h.audience, AccessTokenLifetime, auth.TokenTypeAccess)
	if err != nil {
		return dto.ErrInternalUnexpected.WithMessage("Failed to generate token").WithDebugInfo(err.Error())
	}

	refreshLifetime := refreshTokenLifetimeForRequest(r)
	refreshToken, err := auth.GenerateToken(user.ID, user.Role(), h.jwtSecret, h.audience, refreshLifetime, auth.TokenTypeRefresh)
	if err != nil {
		return dto.ErrInternalUnexpected.WithMessage("Failed to generate refresh token").WithDebugInfo(err.Error())
	}

	http.SetCookie(w, createSecureRefreshCookie(RefreshTokenName, refreshToken, int(refreshLifetime.Seconds()), r))

	slog.Info("User signup successful", "user_id", user.ID, "email", user.Email, "action", "signup_success")

	return respondJSON(w, http.StatusCreated, dto.TokenResult{AccessToken: accessToken, ExpiresIn: int(time.Now().Add(AccessTokenLifetime).Unix())})
}

func (h *AuthUserHandler) Login(w http.ResponseWriter, r *http.Request) error {
	var params LoginParams
	if err := DecodeJSON(r, &params); err != nil {
		slog.Warn("Failed to decode login", "error", err, "ip", r.RemoteAddr, "action", "login_decode_error")
		return dto.ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error())
	}

	user, err := h.service.Authenticate(r.Context(), params.Email, params.Password)
	if err != nil {
		slog.Warn("User login failed", "email", params.Email, "ip", r.RemoteAddr, "error", err, "action", "login_failed")
		return dto.ErrAuthInvalidEmailOrPassword.WithDebugInfo(err.Error())
	}

	accessToken, err := auth.GenerateToken(user.ID, user.Role(), h.jwtSecret, h.audience, AccessTokenLifetime, auth.TokenTypeAccess)
	if err != nil {
		return dto.ErrInternalUnexpected.WithMessage("Failed to generate access token").WithDebugInfo(err.Error())
	}

	refreshLifetime := refreshTokenLifetimeForRequest(r)
	refreshToken, err := auth.GenerateToken(user.ID, user.Role(), h.jwtSecret, h.audience, refreshLifetime, auth.TokenTypeRefresh)
	if err != nil {
		return dto.ErrInternalUnexpected.WithMessage("Failed to generate refresh token").WithDebugInfo(err.Error())
	}

	http.SetCookie(w, createSecureRefreshCookie(RefreshTokenName, refreshToken, int(refreshLifetime.Seconds()), r))

	slog.Info("User login successful", "user_id", user.ID, "email", user.Email, "action", "login_success")

	return respondJSON(w, http.StatusOK, dto.TokenResult{AccessToken: accessToken, ExpiresIn: int(time.Now().Add(AccessTokenLifetime).Unix())})
}

func (h *AuthUserHandler) ForeverToken(w http.ResponseWriter, r *http.Request) error {
	lifetime := time.Duration(10*365*24) * time.Hour
	principal, err := httpx.Principal(r)
	if err != nil {
		return err
	}

	token, err := auth.GenerateToken(principal.UserID, principal.Role, h.jwtSecret, h.audience, lifetime, auth.TokenTypeAccess)
	if err != nil {
		return dto.ErrInternalUnexpected.WithMessage("Failed to generate token").WithDebugInfo(err.Error())
	}
	return respondJSON(w, http.StatusOK, dto.TokenResult{AccessToken: token, ExpiresIn: int(time.Now().Add(lifetime).Unix())})
}

func (h *AuthUserHandler) RefreshToken(w http.ResponseWriter, r *http.Request) error {
	slog.Info("Token refresh attempt", "ip", r.RemoteAddr, "action", "refresh_attempt")

	refreshCookie, err := r.Cookie(RefreshTokenName)
	if err != nil {
		slog.Warn("Missing refresh token cookie", "ip", r.RemoteAddr, "error", err, "action", "refresh_missing_cookie")
		return dto.ErrAuthInvalidCredentials.WithMessage("Missing refresh token")
	}

	result := middleware.ParseAndValidateJWT(refreshCookie.Value, auth.TokenTypeRefresh, h.jwtSecret)
	if result.Error != nil {
		slog.Warn("Invalid refresh token", "ip", r.RemoteAddr, "error", result.Error.Detail, "action", "refresh_invalid_token")
		return *result.Error
	}

	userIDInt, err := strconv.ParseInt(result.UserID, 10, 32)
	if err != nil {
		return dto.ErrAuthInvalidCredentials.WithMessage("Invalid user ID in token")
	}

	accessToken, err := auth.GenerateToken(int32(userIDInt), result.Role, h.jwtSecret, h.audience, AccessTokenLifetime, auth.TokenTypeAccess)
	if err != nil {
		return dto.ErrInternalUnexpected.WithMessage("Failed to generate access token").WithDebugInfo(err.Error())
	}

	slog.Info("Token refresh successful", "user_id", userIDInt, "action", "refresh_success")

	return respondJSON(w, http.StatusOK, dto.TokenResult{AccessToken: accessToken, ExpiresIn: int(time.Now().Add(AccessTokenLifetime).Unix())})
}

func (h *AuthUserHandler) Logout(w http.ResponseWriter, r *http.Request) error {
	http.SetCookie(w, createSecureRefreshCookie(RefreshTokenName, "", -1, r))
	return respondStatus(w, http.StatusOK)
}
