package handler

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"golang.org/x/time/rate"
	"log/slog"

	"github.com/swuecho/chat_backend/apicontract"
	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/provider"
	"github.com/swuecho/chat_backend/svc"
)

// ChatHandler handles chat completion and streaming requests.
type ChatHandler struct {
	service         *svc.ChatService
	sessionSvc      *svc.ChatSessionService
	conversationSvc *svc.SessionConversationService
	rateLimitSvc    *svc.SessionRateLimitService
	snapshotSvc     *svc.SessionSnapshotQueryService
	modelSvc        *svc.SessionModelService
	botHistorySvc   *svc.SessionBotHistoryService
	rateLimiter     *rate.Limiter
	openAIKey       string
	openAIProxy     string
}

const sessionTitleGenerationTimeout = 30 * time.Second

// NewChatHandler creates a new ChatHandler.
func NewChatHandler(service *svc.ChatService, sessionSvc *svc.ChatSessionService, conversationSvc *svc.SessionConversationService, rateLimitSvc *svc.SessionRateLimitService, snapshotSvc *svc.SessionSnapshotQueryService, modelSvc *svc.SessionModelService, botHistorySvc *svc.SessionBotHistoryService, rateLimiter *rate.Limiter, openAIKey, openAIProxy string) *ChatHandler {
	return &ChatHandler{
		service:         service,
		sessionSvc:      sessionSvc,
		conversationSvc: conversationSvc,
		rateLimitSvc:    rateLimitSvc,
		snapshotSvc:     snapshotSvc,
		modelSvc:        modelSvc,
		botHistorySvc:   botHistorySvc,
		rateLimiter:     rateLimiter,
		openAIKey:       openAIKey,
		openAIProxy:     openAIProxy,
	}
}

// Register registers chat routes on the given router.
func (h *ChatHandler) Register(router *mux.Router, registry *apicontract.Registry) {
	router.HandleFunc("/chat_stream", streamEndpoint(h.ChatCompletionHandler)).Methods(http.MethodPost)
	apicontract.DocumentJSON[ChatRequest, provider.AnswerEvent](registry, apicontract.Operation{
		Method: http.MethodPost, Path: "/chat_stream", OperationID: "streamChat", Summary: "Stream a chat completion",
		Tags: []string{"Chat"}, SuccessStatus: http.StatusOK, Security: apicontract.BearerAuth(), ResponseContentType: "text/event-stream",
	})
	router.HandleFunc("/chatbot", streamEndpoint(h.ChatBotCompletionHandler)).Methods(http.MethodPost)
	apicontract.RegisterJSON(router, registry, apicontract.Operation{
		Method: http.MethodGet, Path: "/chat_instructions", OperationID: "getChatInstructions",
		Summary: "Get chat artifact instructions", Tags: []string{"Chat"}, SuccessStatus: http.StatusOK, Security: apicontract.BearerAuth(),
	}, h.getChatInstructions)
}

// --- provider.Handler implementation ---

func (h *ChatHandler) Config() provider.Config {
	return provider.Config{
		OpenAIKey:   h.openAIKey,
		OpenAIProxy: h.openAIProxy,
		RateLimiter: h.rateLimiter,
	}
}

// GetChatInstructions returns artifact instruction text.
func (h *ChatHandler) getChatInstructions(_ *http.Request, _ apicontract.NoBody) (dto.ChatInstructionResponse, error) {
	artifactInstruction, err := svc.LoadArtifactInstruction()
	if err != nil {
		slog.Warn("Failed to load artifact instruction", "error", err)
		artifactInstruction = ""
	}
	return dto.ChatInstructionResponse{
		ArtifactInstruction: artifactInstruction,
	}, nil
}
