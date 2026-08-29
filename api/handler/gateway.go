package handler

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"hash"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/middleware"
	"github.com/swuecho/chat_backend/svc"
)

type GatewayHandler struct {
	service       *svc.GatewayService
	proxyURL      string
	retentionDays int
	captureBytes  int
	cleanupMu     sync.Mutex
	lastCleanup   time.Time
}

func NewGatewayHandler(service *svc.GatewayService, proxyURL ...string) *GatewayHandler {
	h := &GatewayHandler{service: service, retentionDays: 7, captureBytes: 64 * 1024}
	if len(proxyURL) > 0 {
		h.proxyURL = proxyURL[0]
	}
	return h
}

func (h *GatewayHandler) WithObservability(retentionDays, captureBytes int) *GatewayHandler {
	if retentionDays > 0 {
		h.retentionDays = retentionDays
	}
	if captureBytes > 0 {
		h.captureBytes = captureBytes
	}
	return h
}

func (h *GatewayHandler) Register(r *mux.Router) {
	r.Handle("/models", h.authenticate(http.HandlerFunc(h.models))).Methods(http.MethodGet)
	r.Handle("/chat/completions", h.authenticate(http.HandlerFunc(h.chatCompletions))).Methods(http.MethodPost)
}

type gatewayKeyContextKey struct{}

func openAIError(w http.ResponseWriter, status int, message, errorType, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(openAIErrorResponse{Error: openAIErrorDetail{Message: message, Type: errorType, Code: code}})
}

func (h *GatewayHandler) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := middleware.ExtractBearerToken(r)
		if !strings.HasPrefix(token, "sk-chat-") {
			openAIError(w, http.StatusUnauthorized, "Invalid API key", "invalid_request_error", "invalid_api_key")
			return
		}
		sum := sha256.Sum256([]byte(token))
		key, err := h.service.CredentialByHash(r.Context(), hex.EncodeToString(sum[:]))
		if err != nil || key.Status != "active" || (key.ExpiresAt != nil && !key.ExpiresAt.After(time.Now())) {
			openAIError(w, http.StatusUnauthorized, "Invalid or expired API key", "invalid_request_error", "invalid_api_key")
			return
		}
		count, err := h.service.CountRecentRequests(r.Context(), key.ID)
		if err != nil {
			openAIError(w, http.StatusInternalServerError, "Unable to check rate limit", "server_error", "internal_error")
			return
		}
		if count >= int64(key.RequestsPerMinute) {
			openAIError(w, http.StatusTooManyRequests, "API key rate limit exceeded", "rate_limit_error", "rate_limit_exceeded")
			return
		}
		_ = h.service.TouchKey(r.Context(), key.ID)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), gatewayKeyContextKey{}, key)))
	})
}

func (h *GatewayHandler) models(w http.ResponseWriter, r *http.Request) {
	models, err := h.service.ListModelNames(r.Context())
	if err != nil {
		openAIError(w, http.StatusInternalServerError, "Unable to list models", "server_error", "internal_error")
		return
	}
	data := make([]gatewayModelHTTPResponse, 0, len(models))
	for _, model := range models {
		data = append(data, gatewayModelHTTPResponse{ID: model, Object: "model", OwnedBy: "chat"})
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(gatewayModelListHTTPResponse{Object: "list", Data: data})
}

type gatewayEnvelope struct {
	Model  string `json:"model"`
	Stream bool   `json:"stream"`
}

type gatewayUsage struct {
	PromptTokens     int32 `json:"prompt_tokens"`
	CompletionTokens int32 `json:"completion_tokens"`
	TotalTokens      int32 `json:"total_tokens"`
}

func (h *GatewayHandler) chatCompletions(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		openAIError(w, http.StatusBadRequest, "Unable to read request", "invalid_request_error", "invalid_request")
		return
	}
	var envelope gatewayEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		openAIError(w, http.StatusBadRequest, "Invalid JSON request body", "invalid_request_error", "invalid_json")
		return
	}
	if envelope.Model == "" {
		openAIError(w, http.StatusBadRequest, "model is required", "invalid_request_error", "missing_required_parameter")
		return
	}
	model, err := h.service.ModelByName(r.Context(), envelope.Model)
	if err != nil || !model.IsEnable || model.ApiType != "openai" {
		openAIError(w, http.StatusNotFound, "The requested model is not available", "invalid_request_error", "model_not_found")
		return
	}
	providerKey := os.Getenv(model.ApiAuthKey)
	if providerKey == "" {
		openAIError(w, http.StatusBadGateway, "The model provider is not configured", "server_error", "provider_not_configured")
		return
	}
	key := r.Context().Value(gatewayKeyContextKey{}).(svc.GatewayCredential)
	requestHash := sha256.Sum256(body)
	requestSample, requestTruncated := boundedSample(body, h.captureBytes)
	recordID, err := h.service.CreateRequest(r.Context(), svc.CreateGatewayRequestInput{
		APIKeyID: key.ID, UserID: key.UserID, ChatModelID: model.ID, RequestedModel: model.Name, Provider: model.ApiType, Stream: envelope.Stream,
		RequestBytes: int64(len(body)), RequestSHA256: hex.EncodeToString(requestHash[:]), RequestSample: requestSample,
		RequestTruncated: requestTruncated, RequestClassification: classifyRequest(body), RetentionUntil: time.Now().AddDate(0, 0, h.retentionDays),
	})
	if err != nil {
		slog.Error("Unable to create gateway audit record", "model", model.Name, "error", err)
	}
	go h.deleteExpiredRequests()
	started := time.Now()
	upstream, err := http.NewRequestWithContext(r.Context(), http.MethodPost, completionURL(model.Url), bytes.NewReader(body))
	if err != nil {
		h.finish(recordID, started, "failed", gatewayUsage{}, "", "invalid_provider_url")
		openAIError(w, http.StatusBadGateway, "Invalid provider URL", "server_error", "provider_error")
		return
	}
	copyEndToEndHeaders(upstream.Header, r.Header)
	upstream.Header.Del("Authorization")
	upstream.Header.Del("Cookie")
	header := model.ApiAuthHeader
	if header == "" {
		header = "Authorization"
	}
	if strings.EqualFold(header, "Authorization") && !strings.HasPrefix(strings.ToLower(providerKey), "bearer ") {
		providerKey = "Bearer " + providerKey
	}
	upstream.Header.Set(header, providerKey)
	timeout := time.Duration(model.HttpTimeOut) * time.Second
	if timeout <= 0 {
		timeout = 120 * time.Second
	}
	client, err := h.httpClient(timeout)
	if err != nil {
		h.finish(recordID, started, "failed", gatewayUsage{}, "", "invalid_proxy_url")
		openAIError(w, http.StatusBadGateway, "The configured provider proxy URL is invalid", "server_error", "invalid_proxy_url")
		return
	}
	resp, err := client.Do(upstream)
	if err != nil {
		slog.Error("Gateway provider request failed", "model", model.Name, "url", completionURL(model.Url), "error", err)
		status := "failed"
		code := "provider_error"
		if errors.Is(err, context.Canceled) {
			status, code = "cancelled", "client_cancelled"
		}
		h.finish(recordID, started, status, gatewayUsage{}, "", code)
		if status != "cancelled" {
			openAIError(w, http.StatusBadGateway, "Provider request failed", "server_error", code)
		}
		return
	}
	defer resp.Body.Close()
	if envelope.Stream {
		h.proxyStream(w, resp, recordID, started)
		return
	}
	h.proxyResponse(w, resp, recordID, started)
}

func (h *GatewayHandler) httpClient(timeout time.Duration) (*http.Client, error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DisableCompression = true
	if h.proxyURL != "" {
		proxy, err := url.Parse(h.proxyURL)
		if err != nil {
			return nil, err
		}
		transport.Proxy = http.ProxyURL(proxy)
	}
	return &http.Client{Transport: transport, Timeout: timeout}, nil
}

func completionURL(raw string) string {
	trimmed := strings.TrimRight(raw, "/")
	if strings.HasSuffix(trimmed, "/chat/completions") {
		return trimmed
	}
	if u, err := url.Parse(trimmed); err == nil && u.Path != "" && u.Path != "/" && !strings.HasSuffix(u.Path, "/v1") {
		return trimmed
	}
	return trimmed + "/chat/completions"
}

func (h *GatewayHandler) proxyResponse(w http.ResponseWriter, resp *http.Response, id int64, started time.Time) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		h.finish(id, started, "failed", gatewayUsage{}, "", "provider_read_error")
		if !errors.Is(err, context.Canceled) {
			openAIError(w, http.StatusBadGateway, "Unable to read provider response", "server_error", "provider_error")
		}
		return
	}
	var parsed struct {
		ID    string       `json:"id"`
		Usage gatewayUsage `json:"usage"`
	}
	_ = json.Unmarshal(body, &parsed)
	observation := observeBytes(body, h.captureBytes)
	status, code := "succeeded", ""
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		status, code = "failed", "provider_http_error"
	}
	h.finishObserved(id, started, status, parsed.Usage, parsed.ID, code, observation, classifyResponse(resp, false))
	replaceEndToEndHeaders(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	_, _ = w.Write(body)
}

func (h *GatewayHandler) proxyStream(w http.ResponseWriter, resp *http.Response, id int64, started time.Time) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		h.finish(id, started, "failed", gatewayUsage{}, "", "stream_unsupported")
		openAIError(w, http.StatusInternalServerError, "Streaming is not supported", "server_error", "stream_error")
		return
	}
	replaceEndToEndHeaders(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	observation := newBodyObservation(h.captureBytes)
	_, err := io.Copy(&flushWriter{writer: w, flusher: flusher, observation: observation}, resp.Body)
	if err != nil {
		status, code := "failed", "provider_stream_error"
		if errors.Is(resp.Request.Context().Err(), context.Canceled) {
			status, code = "cancelled", "client_cancelled"
		}
		h.finishObserved(id, started, status, gatewayUsage{}, resp.Header.Get("x-request-id"), code, observation, classifyResponse(resp, true))
		return
	}
	status, code := "succeeded", ""
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		status, code = "failed", "provider_http_error"
	}
	h.finishObserved(id, started, status, gatewayUsage{}, resp.Header.Get("x-request-id"), code, observation, classifyResponse(resp, true))
}

type flushWriter struct {
	writer      io.Writer
	flusher     http.Flusher
	observation *bodyObservation
}

func (w *flushWriter) Write(p []byte) (int, error) {
	n, err := w.writer.Write(p)
	if n > 0 {
		_, _ = w.observation.Write(p[:n])
	}
	w.flusher.Flush()
	return n, err
}

var hopByHopHeaders = []string{
	"Connection", "Proxy-Connection", "Keep-Alive", "Proxy-Authenticate",
	"Proxy-Authorization", "Te", "Trailer", "Transfer-Encoding", "Upgrade",
}

func copyEndToEndHeaders(dst, src http.Header) {
	for key, values := range src {
		dst[http.CanonicalHeaderKey(key)] = append([]string(nil), values...)
	}
	for _, token := range strings.Split(src.Get("Connection"), ",") {
		if token = strings.TrimSpace(token); token != "" {
			dst.Del(token)
		}
	}
	for _, key := range hopByHopHeaders {
		dst.Del(key)
	}
}

func replaceEndToEndHeaders(dst, src http.Header) {
	for key := range dst {
		dst.Del(key)
	}
	copyEndToEndHeaders(dst, src)
	// Provider cookies must not be scoped to the gateway's own origin.
	dst.Del("Set-Cookie")
}

type bodyObservation struct {
	hash      hash.Hash
	sample    []byte
	limit     int
	byteCount int64
}

func newBodyObservation(limit int) *bodyObservation {
	return &bodyObservation{hash: sha256.New(), sample: make([]byte, 0, limit), limit: limit}
}

func (o *bodyObservation) Write(p []byte) (int, error) {
	o.byteCount += int64(len(p))
	_, _ = o.hash.Write(p)
	remaining := o.limit - len(o.sample)
	if remaining > 0 {
		if remaining > len(p) {
			remaining = len(p)
		}
		o.sample = append(o.sample, p[:remaining]...)
	}
	return len(p), nil
}

func (o *bodyObservation) digest() string  { return hex.EncodeToString(o.hash.Sum(nil)) }
func (o *bodyObservation) truncated() bool { return o.byteCount > int64(len(o.sample)) }

func observeBytes(body []byte, limit int) *bodyObservation {
	o := newBodyObservation(limit)
	_, _ = o.Write(body)
	return o
}

func boundedSample(body []byte, limit int) ([]byte, bool) {
	if len(body) <= limit {
		return append([]byte(nil), body...), false
	}
	return append([]byte(nil), body[:limit]...), true
}

func classifyRequest(body []byte) json.RawMessage {
	classification := gatewayRequestClassification{Format: "openai_chat_completions"}
	var payload map[string]json.RawMessage
	if json.Unmarshal(body, &payload) != nil {
		encoded, _ := json.Marshal(classification)
		return encoded
	}
	var messages []struct {
		Role    string          `json:"role"`
		Content json.RawMessage `json:"content"`
	}
	if json.Unmarshal(payload["messages"], &messages) == nil {
		roles := make(map[string]int)
		multimodal := false
		for _, message := range messages {
			roles[message.Role]++
			if len(message.Content) > 0 && message.Content[0] == '[' {
				multimodal = true
			}
		}
		classification.MessageCount = len(messages)
		classification.Roles = roles
		classification.Multimodal = multimodal
	}
	_, classification.HasTools = payload["tools"]
	_, classification.HasResponseFormat = payload["response_format"]
	encoded, _ := json.Marshal(classification)
	return encoded
}

func classifyResponse(resp *http.Response, stream bool) json.RawMessage {
	encoded, _ := json.Marshal(gatewayResponseClassification{Stream: stream, StatusCode: resp.StatusCode,
		ContentType: resp.Header.Get("Content-Type"), ContentEncoding: resp.Header.Get("Content-Encoding"),
		Successful: resp.StatusCode >= 200 && resp.StatusCode < 300})
	return encoded
}

func (h *GatewayHandler) finish(id int64, started time.Time, status string, usage gatewayUsage, providerID, code string) {
	h.finishObserved(id, started, status, usage, providerID, code, nil, json.RawMessage(`{}`))
}

func (h *GatewayHandler) finishObserved(id int64, started time.Time, status string, usage gatewayUsage, providerID, code string, observation *bodyObservation, classification json.RawMessage) {
	if id == 0 {
		return
	}
	responseBytes := int64(0)
	responseHash := ""
	responseSample := []byte{}
	responseTruncated := false
	if observation != nil {
		responseBytes = observation.byteCount
		responseHash = observation.digest()
		responseSample = append([]byte(nil), observation.sample...)
		responseTruncated = observation.truncated()
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := h.service.CompleteRequest(ctx, svc.CompleteGatewayRequestInput{
			ID: id, Status: status, PromptTokens: usage.PromptTokens, CompletionTokens: usage.CompletionTokens, TotalTokens: usage.TotalTokens,
			LatencyMs: time.Since(started).Milliseconds(), ProviderRequestID: providerID, ErrorCode: code,
			ResponseBytes: responseBytes, ResponseSHA256: responseHash, ResponseSample: responseSample,
			ResponseTruncated: responseTruncated, ResponseClassification: classification,
		}); err != nil {
			slog.Error("Unable to complete gateway audit record", "requestID", id, "error", err)
		}
	}()
}

func (h *GatewayHandler) deleteExpiredRequests() {
	h.cleanupMu.Lock()
	if time.Since(h.lastCleanup) < time.Hour {
		h.cleanupMu.Unlock()
		return
	}
	h.lastCleanup = time.Now()
	h.cleanupMu.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := h.service.PurgeExpiredSamples(ctx); err != nil {
		slog.Error("Unable to purge expired gateway request samples", "error", err)
	}
}
