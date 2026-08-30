package handler

import (
	"encoding/base64"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/svc"
)

type APIKeyHandler struct{ service *svc.APIKeyService }

func NewAPIKeyHandler(service *svc.APIKeyService) *APIKeyHandler {
	return &APIKeyHandler{service: service}
}

func (h *APIKeyHandler) Register(r *mux.Router) {
	r.HandleFunc("/api-keys", h.list).Methods(http.MethodGet)
	r.HandleFunc("/api-keys", h.create).Methods(http.MethodPost)
	r.HandleFunc("/api-keys/{id}", h.revoke).Methods(http.MethodDelete)
	r.HandleFunc("/api-keys/{id}/usage", h.usage).Methods(http.MethodGet)
	r.HandleFunc("/api-keys/{id}/requests", h.requests).Methods(http.MethodGet)
	r.HandleFunc("/api-keys/{id}/requests/{requestId}", h.requestDetail).Methods(http.MethodGet)
}

func (h *APIKeyHandler) list(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r.Context())
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrAuthInvalidCredentials)
		return
	}
	keys, err := h.service.List(r.Context(), userID)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrInternalUnexpected.WithDebugInfo(err.Error()))
		return
	}
	dto.RespondWithJSON(w, http.StatusOK, keys)
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
	if err := DecodeJSON(r, &input); err != nil {
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
	var expires *time.Time
	if input.ExpiresAt != "" {
		t, err := time.Parse(time.RFC3339, input.ExpiresAt)
		if err != nil || !t.After(time.Now()) {
			dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("expiresAt must be a future RFC3339 timestamp"))
			return
		}
		expires = &t
	}
	key, err := h.service.Create(r.Context(), userID, input.Name, expires, input.RequestsPerMinute)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrInternalUnexpected.WithDebugInfo(err.Error()))
		return
	}
	dto.RespondWithJSON(w, http.StatusCreated, key)
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
	count, err := h.service.Revoke(r.Context(), id, userID)
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
	usage, err := h.service.Usage(r.Context(), id, userID)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrInternalUnexpected.WithDebugInfo(err.Error()))
		return
	}
	dto.RespondWithJSON(w, http.StatusOK, usage)
}

func (h *APIKeyHandler) requests(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r.Context())
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrAuthInvalidCredentials)
		return
	}
	keyID, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("Invalid API key ID"))
		return
	}
	requests, err := h.service.Requests(r.Context(), svc.APIKeyRequestsQuery{KeyID: keyID, UserID: userID, Limit: 100})
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrInternalUnexpected.WithDebugInfo(err.Error()))
		return
	}
	dto.RespondWithJSON(w, http.StatusOK, requests)
}

type capturedSample struct {
	Encoding string `json:"encoding"`
	Text     string `json:"text,omitempty"`
	Base64   string `json:"base64,omitempty"`
}

func sampleForAPI(sample []byte) capturedSample {
	if utf8.Valid(sample) && !strings.ContainsRune(string(sample), '\x00') {
		return capturedSample{Encoding: "utf-8", Text: string(sample)}
	}
	return capturedSample{Encoding: "base64", Base64: base64.StdEncoding.EncodeToString(sample)}
}

func (h *APIKeyHandler) requestDetail(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r.Context())
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrAuthInvalidCredentials)
		return
	}
	keyID, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("Invalid API key ID"))
		return
	}
	requestID, err := strconv.ParseInt(mux.Vars(r)["requestId"], 10, 64)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("Invalid gateway request ID"))
		return
	}
	record, err := h.service.RequestDetail(r.Context(), requestID, keyID, userID)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrResourceNotFound("Gateway request"))
		return
	}
	var completedAt *time.Time
	if record.CompletedAt.Valid {
		completedAt = &record.CompletedAt.Time
	}
	dto.RespondWithJSON(w, http.StatusOK, gatewayRequestDetailHTTPResponse{ID: record.ID, RequestUUID: record.RequestUuid,
		RequestedModel: record.RequestedModel, Provider: record.Provider, Status: record.Status, Stream: record.Stream,
		PromptTokens: record.PromptTokens, CompletionTokens: record.CompletionTokens, TotalTokens: record.TotalTokens,
		LatencyMs: record.LatencyMs, ProviderRequestID: record.ProviderRequestID, ErrorCode: record.ErrorCode,
		RequestBytes: record.RequestBytes, ResponseBytes: record.ResponseBytes, RequestSHA256: record.RequestSha256,
		ResponseSHA256: record.ResponseSha256, RequestTruncated: record.RequestTruncated, ResponseTruncated: record.ResponseTruncated,
		RequestClassification: record.RequestClassification, ResponseClassification: record.ResponseClassification,
		CreatedAt: record.CreatedAt, CompletedAt: completedAt, RetentionUntil: record.RetentionUntil,
		RequestCapture: sampleForAPI(record.RequestSample), ResponseCapture: sampleForAPI(record.ResponseSample)})
}
