package handler

import (
	"encoding/base64"
	"net/http"
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
	r.HandleFunc("/api-keys", endpoint(h.list)).Methods(http.MethodGet)
	r.HandleFunc("/api-keys", endpoint(h.create)).Methods(http.MethodPost)
	r.HandleFunc("/api-keys/{id}", endpoint(h.revoke)).Methods(http.MethodDelete)
	r.HandleFunc("/api-keys/{id}/usage", endpoint(h.usage)).Methods(http.MethodGet)
	r.HandleFunc("/api-keys/{id}/requests", endpoint(h.requests)).Methods(http.MethodGet)
	r.HandleFunc("/api-keys/{id}/requests/{requestId}", endpoint(h.requestDetail)).Methods(http.MethodGet)
}

func (h *APIKeyHandler) list(w http.ResponseWriter, r *http.Request) error {
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}
	keys, err := h.service.List(r.Context(), userID)
	if err != nil {
		return dto.ErrInternalUnexpected.WithDebugInfo(err.Error())
	}
	return respondJSON(w, http.StatusOK, keys)
}

func (h *APIKeyHandler) create(w http.ResponseWriter, r *http.Request) error {
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}
	var input createAPIKeyRequest
	if err := DecodeJSON(r, &input); err != nil {
		return dto.ErrValidationInvalidInput("Invalid request body").WithDebugInfo(err.Error())
	}
	key, err := h.service.Create(r.Context(), userID, input.Name, input.expiration(), input.RequestsPerMinute)
	if err != nil {
		return dto.ErrInternalUnexpected.WithDebugInfo(err.Error())
	}
	return respondJSON(w, http.StatusCreated, key)
}

func (h *APIKeyHandler) revoke(w http.ResponseWriter, r *http.Request) error {
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}
	id, err := positiveInt64Param(r, "id")
	if err != nil {
		return err
	}
	count, err := h.service.Revoke(r.Context(), id, userID)
	if err != nil {
		return dto.ErrInternalUnexpected.WithDebugInfo(err.Error())
	}
	if count == 0 {
		return dto.ErrResourceNotFound("API key")
	}
	return noContent(w)
}

func (h *APIKeyHandler) usage(w http.ResponseWriter, r *http.Request) error {
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}
	id, err := positiveInt64Param(r, "id")
	if err != nil {
		return err
	}
	usage, err := h.service.Usage(r.Context(), id, userID)
	if err != nil {
		return dto.ErrInternalUnexpected.WithDebugInfo(err.Error())
	}
	return respondJSON(w, http.StatusOK, usage)
}

func (h *APIKeyHandler) requests(w http.ResponseWriter, r *http.Request) error {
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}
	keyID, err := positiveInt64Param(r, "id")
	if err != nil {
		return err
	}
	requests, err := h.service.Requests(r.Context(), svc.APIKeyRequestsQuery{KeyID: keyID, UserID: userID, Limit: 100})
	if err != nil {
		return dto.ErrInternalUnexpected.WithDebugInfo(err.Error())
	}
	return respondJSON(w, http.StatusOK, requests)
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

func (h *APIKeyHandler) requestDetail(w http.ResponseWriter, r *http.Request) error {
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}
	keyID, err := positiveInt64Param(r, "id")
	if err != nil {
		return err
	}
	requestID, err := positiveInt64Param(r, "requestId")
	if err != nil {
		return err
	}
	record, err := h.service.RequestDetail(r.Context(), requestID, keyID, userID)
	if err != nil {
		return dto.ErrResourceNotFound("Gateway request")
	}
	var completedAt *time.Time
	if record.CompletedAt.Valid {
		completedAt = &record.CompletedAt.Time
	}
	return respondJSON(w, http.StatusOK, gatewayRequestDetailHTTPResponse{ID: record.ID, RequestUUID: record.RequestUuid,
		RequestedModel: record.RequestedModel, Provider: record.Provider, Status: record.Status, Stream: record.Stream,
		PromptTokens: record.PromptTokens, CompletionTokens: record.CompletionTokens, TotalTokens: record.TotalTokens,
		LatencyMs: record.LatencyMs, ProviderRequestID: record.ProviderRequestID, ErrorCode: record.ErrorCode,
		RequestBytes: record.RequestBytes, ResponseBytes: record.ResponseBytes, RequestSHA256: record.RequestSha256,
		ResponseSHA256: record.ResponseSha256, RequestTruncated: record.RequestTruncated, ResponseTruncated: record.ResponseTruncated,
		RequestClassification: record.RequestClassification, ResponseClassification: record.ResponseClassification,
		CreatedAt: record.CreatedAt, CompletedAt: completedAt, RetentionUntil: record.RetentionUntil,
		RequestCapture: sampleForAPI(record.RequestSample), ResponseCapture: sampleForAPI(record.ResponseSample)})
}
