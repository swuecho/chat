package handler

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

type APIKeyHandler struct{ db *sqlc_queries.Queries }

func NewAPIKeyHandler(db *sqlc_queries.Queries) *APIKeyHandler { return &APIKeyHandler{db: db} }

func (h *APIKeyHandler) Register(r *mux.Router) {
	r.HandleFunc("/api-keys", h.list).Methods(http.MethodGet)
	r.HandleFunc("/api-keys", h.create).Methods(http.MethodPost)
	r.HandleFunc("/api-keys/{id}", h.revoke).Methods(http.MethodDelete)
	r.HandleFunc("/api-keys/{id}/usage", h.usage).Methods(http.MethodGet)
}

type apiKeyView struct {
	ID                int64      `json:"id"`
	Name              string     `json:"name"`
	KeyPrefix         string     `json:"keyPrefix"`
	Status            string     `json:"status"`
	RequestsPerMinute int32      `json:"requestsPerMinute"`
	ExpiresAt         *time.Time `json:"expiresAt"`
	LastUsedAt        *time.Time `json:"lastUsedAt"`
	CreatedAt         time.Time  `json:"createdAt"`
}

func keyView(k sqlc_queries.VirtualApiKey) apiKeyView {
	var expiresAt, lastUsedAt *time.Time
	if k.ExpiresAt.Valid {
		expiresAt = &k.ExpiresAt.Time
	}
	if k.LastUsedAt.Valid {
		lastUsedAt = &k.LastUsedAt.Time
	}
	return apiKeyView{k.ID, k.Name, k.KeyPrefix, k.Status, k.RequestsPerMinute, expiresAt, lastUsedAt, k.CreatedAt}
}

func (h *APIKeyHandler) list(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r.Context())
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrAuthInvalidCredentials)
		return
	}
	keys, err := h.db.ListVirtualAPIKeysByUser(r.Context(), userID)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrInternalUnexpected.WithDebugInfo(err.Error()))
		return
	}
	views := make([]apiKeyView, 0, len(keys))
	for _, key := range keys {
		views = append(views, keyView(key))
	}
	dto.RespondWithJSON(w, http.StatusOK, views)
}

func (h *APIKeyHandler) create(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r.Context())
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrAuthInvalidCredentials)
		return
	}
	var input struct {
		Name              string `json:"name"`
		ExpiresAt         string `json:"expiresAt"`
		RequestsPerMinute int32  `json:"requestsPerMinute"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("Invalid request body"))
		return
	}
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" || len(input.Name) > 100 {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("Key name must be between 1 and 100 characters"))
		return
	}
	if input.RequestsPerMinute == 0 {
		input.RequestsPerMinute = 60
	}
	if input.RequestsPerMinute < 1 || input.RequestsPerMinute > 10000 {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("requestsPerMinute must be between 1 and 10000"))
		return
	}
	expires := sql.NullTime{}
	if input.ExpiresAt != "" {
		t, err := time.Parse(time.RFC3339, input.ExpiresAt)
		if err != nil || !t.After(time.Now()) {
			dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("expiresAt must be a future RFC3339 timestamp"))
			return
		}
		expires = sql.NullTime{Time: t, Valid: true}
	}
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		dto.RespondWithAPIError(w, dto.ErrInternalUnexpected)
		return
	}
	plaintext := "sk-chat-" + base64.RawURLEncoding.EncodeToString(secret)
	sum := sha256.Sum256([]byte(plaintext))
	prefix := plaintext
	if len(prefix) > 18 {
		prefix = prefix[:18]
	}
	key, err := h.db.CreateVirtualAPIKey(r.Context(), sqlc_queries.CreateVirtualAPIKeyParams{
		UserID: userID, Name: input.Name, KeyPrefix: prefix, KeyHash: hex.EncodeToString(sum[:]),
		ExpiresAt: expires, RequestsPerMinute: input.RequestsPerMinute,
	})
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrInternalUnexpected.WithDebugInfo(err.Error()))
		return
	}
	dto.RespondWithJSON(w, http.StatusCreated, struct {
		apiKeyView
		Key string `json:"key"`
	}{keyView(key), plaintext})
}

func (h *APIKeyHandler) revoke(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r.Context())
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrAuthInvalidCredentials)
		return
	}
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("Invalid API key ID"))
		return
	}
	count, err := h.db.RevokeVirtualAPIKey(r.Context(), sqlc_queries.RevokeVirtualAPIKeyParams{ID: id, UserID: userID})
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrInternalUnexpected.WithDebugInfo(err.Error()))
		return
	}
	if count == 0 {
		dto.RespondWithAPIError(w, dto.ErrResourceNotFound("API key"))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *APIKeyHandler) usage(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r.Context())
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrAuthInvalidCredentials)
		return
	}
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("Invalid API key ID"))
		return
	}
	if _, err := h.db.VirtualAPIKeyByIDAndUser(r.Context(), sqlc_queries.VirtualAPIKeyByIDAndUserParams{ID: id, UserID: userID}); err != nil {
		dto.RespondWithAPIError(w, dto.ErrResourceNotFound("API key"))
		return
	}
	usage, err := h.db.GatewayUsageByKey(r.Context(), sqlc_queries.GatewayUsageByKeyParams{ApiKeyID: id, UserID: userID})
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrInternalUnexpected.WithDebugInfo(err.Error()))
		return
	}
	dto.RespondWithJSON(w, http.StatusOK, usage)
}
