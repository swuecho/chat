package handler

import (
	"encoding/base64"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/apicontract"
	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/sqlc_queries"
	"github.com/swuecho/chat_backend/svc"
)

type APIKeyHandler struct{ service *svc.APIKeyService }

func NewAPIKeyHandler(service *svc.APIKeyService) *APIKeyHandler {
	return &APIKeyHandler{service: service}
}

func (h *APIKeyHandler) Register(r *mux.Router, registry *apicontract.Registry) {
	r.HandleFunc("/api-keys", endpoint(h.list)).Methods(http.MethodGet)
	r.HandleFunc("/api-keys", endpoint(h.create)).Methods(http.MethodPost)
	r.HandleFunc("/api-keys/{id}", endpoint(h.revoke)).Methods(http.MethodDelete)
	r.HandleFunc("/api-keys/{id}/usage", endpoint(h.usage)).Methods(http.MethodGet)
	r.HandleFunc("/api-keys/{id}/requests", endpoint(h.requests)).Methods(http.MethodGet)
	r.HandleFunc("/api-keys/{id}/requests/{requestId}", endpoint(h.requestDetail)).Methods(http.MethodGet)
	security := apicontract.BearerAuth()
	tags := []string{"Admin API keys"}
	apicontract.DocumentJSON[apicontract.NoBody, []svc.APIKeyView](registry, apicontract.Operation{Method: http.MethodGet, Path: "/admin/api-keys", OperationID: "listAPIKeys", Summary: "List virtual API keys", Tags: tags, SuccessStatus: http.StatusOK, Security: security})
	apicontract.DocumentJSON[createAPIKeyRequest, svc.CreatedAPIKey](registry, apicontract.Operation{Method: http.MethodPost, Path: "/admin/api-keys", OperationID: "createAPIKey", Summary: "Create a virtual API key", Tags: tags, SuccessStatus: http.StatusCreated, Security: security})
	idParameter := []apicontract.Parameter{apicontract.PositiveIntegerPathParameter("id")}
	apicontract.DocumentJSON[apicontract.NoBody, apicontract.NoBody](registry, apicontract.Operation{Method: http.MethodDelete, Path: "/admin/api-keys/{id}", OperationID: "revokeAPIKey", Summary: "Revoke a virtual API key", Tags: tags, SuccessStatus: http.StatusNoContent, Security: security, Parameters: idParameter})
	apicontract.DocumentJSON[apicontract.NoBody, []sqlc_queries.GatewayUsageByKeyRow](registry, apicontract.Operation{Method: http.MethodGet, Path: "/admin/api-keys/{id}/usage", OperationID: "getAPIKeyUsage", Summary: "Get virtual API key usage", Tags: tags, SuccessStatus: http.StatusOK, Security: security, Parameters: idParameter})
	apicontract.DocumentJSON[apicontract.NoBody, []gatewayRequestSummaryHTTPResponse](registry, apicontract.Operation{Method: http.MethodGet, Path: "/admin/api-keys/{id}/requests", OperationID: "listAPIKeyRequests", Summary: "List gateway requests for an API key", Tags: tags, SuccessStatus: http.StatusOK, Security: security, Parameters: idParameter})
	apicontract.DocumentJSON[apicontract.NoBody, gatewayRequestDetailHTTPResponse](registry, apicontract.Operation{Method: http.MethodGet, Path: "/admin/api-keys/{id}/requests/{requestId}", OperationID: "getAPIKeyRequest", Summary: "Get a gateway request", Tags: tags, SuccessStatus: http.StatusOK, Security: security, Parameters: []apicontract.Parameter{apicontract.PositiveIntegerPathParameter("id"), apicontract.PositiveIntegerPathParameter("requestId")}})
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
	response := make([]gatewayRequestSummaryHTTPResponse, len(requests))
	for i, record := range requests {
		var completedAt *time.Time
		if record.CompletedAt.Valid {
			completedAt = &record.CompletedAt.Time
		}
		response[i] = gatewayRequestSummaryHTTPResponse{ID: record.ID, RequestUUID: record.RequestUuid,
			RequestedModel: record.RequestedModel, Provider: record.Provider, Status: record.Status, Stream: record.Stream,
			PromptTokens: record.PromptTokens, CompletionTokens: record.CompletionTokens, TotalTokens: record.TotalTokens,
			LatencyMs: record.LatencyMs, RequestBytes: record.RequestBytes, ResponseBytes: record.ResponseBytes,
			RequestTruncated: record.RequestTruncated, ResponseTruncated: record.ResponseTruncated,
			CreatedAt: record.CreatedAt, CompletedAt: completedAt, RetentionUntil: record.RetentionUntil, ErrorCode: record.ErrorCode}
	}
	return respondJSON(w, http.StatusOK, response)
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
