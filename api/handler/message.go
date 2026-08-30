package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/httpx"
	"github.com/swuecho/chat_backend/models"
	"github.com/swuecho/chat_backend/svc"
)

type ChatMessageHandler struct {
	service         *svc.ChatMessageService
	sessionSvc      *svc.ChatSessionService
	conversationSvc *svc.SessionConversationService
	chatService     *svc.ChatService
}

func NewChatMessageHandler(service *svc.ChatMessageService, sessionSvc *svc.ChatSessionService, conversationSvc *svc.SessionConversationService, chatService *svc.ChatService) *ChatMessageHandler {
	return &ChatMessageHandler{
		service:         service,
		sessionSvc:      sessionSvc,
		conversationSvc: conversationSvc,
		chatService:     chatService,
	}
}

func (h *ChatMessageHandler) Register(router *mux.Router) {
	router.HandleFunc("/chat_messages", endpoint(h.CreateChatMessage)).Methods(http.MethodPost)
	router.HandleFunc("/chat_messages/{id}", endpoint(h.GetChatMessageByID)).Methods(http.MethodGet)
	router.HandleFunc("/chat_messages/{id}", endpoint(h.UpdateChatMessage)).Methods(http.MethodPut)
	router.HandleFunc("/chat_messages/{id}", endpoint(h.DeleteChatMessage)).Methods(http.MethodDelete)
	router.HandleFunc("/chat_messages", endpoint(h.GetAllChatMessages)).Methods(http.MethodGet)

	router.HandleFunc("/uuid/chat_messages/{uuid}", endpoint(h.GetChatMessageByUUID)).Methods(http.MethodGet)
	router.HandleFunc("/uuid/chat_messages/{uuid}", endpoint(h.UpdateChatMessageByUUID)).Methods(http.MethodPut)
	router.HandleFunc("/uuid/chat_messages/{uuid}", endpoint(h.DeleteChatMessageByUUID)).Methods(http.MethodDelete)
	router.HandleFunc("/uuid/chat_messages/{uuid}/generate-suggestions", endpoint(h.GenerateMoreSuggestions)).Methods(http.MethodPost)
	router.HandleFunc("/uuid/chat_messages/chat_sessions/{uuid}", endpoint(h.GetChatHistoryBySessionUUID)).Methods(http.MethodGet)
	router.HandleFunc("/uuid/chat_messages/chat_sessions/{uuid}", endpoint(h.DeleteChatMessagesBySessionUUID)).Methods(http.MethodDelete)
}

func (h *ChatMessageHandler) CreateChatMessage(w http.ResponseWriter, r *http.Request) error {
	var request chatMessageRequest
	err := DecodeJSON(r, &request)
	if err != nil {
		return dto.ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error())
	}
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}
	message, err := h.service.CreateChatMessage(r.Context(), svc.CreateChatMessageInput{
		ChatSessionUUID: request.ChatSessionUUID, UUID: request.UUID, Role: request.Role,
		Content: request.Content, ReasoningContent: request.ReasoningContent, Model: request.Model,
		TokenCount: request.TokenCount, Score: request.Score, UserID: userID,
		CreatedBy: userID, UpdatedBy: userID, LLMSummary: request.LLMSummary,
		Raw: request.Raw, Artifacts: request.Artifacts, SuggestedQuestions: request.SuggestedQuestions,
	})
	if err != nil {
		return dto.WrapError(dto.MapDatabaseError(err), "Failed to create chat message")
	}
	return respondJSON(w, http.StatusCreated, messageResponse(message))
}

func (h *ChatMessageHandler) GetChatMessageByID(w http.ResponseWriter, r *http.Request) error {
	id, err := positiveInt32Param(r, "id")
	if err != nil {
		return err
	}
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}
	message, err := h.service.GetChatMessageByID(r.Context(), id, userID)
	if err != nil {
		return dto.WrapError(dto.MapDatabaseError(err), "Failed to get chat message")
	}
	return respondJSON(w, http.StatusOK, messageResponse(message))
}

func (h *ChatMessageHandler) UpdateChatMessage(w http.ResponseWriter, r *http.Request) error {
	id, err := positiveInt32Param(r, "id")
	if err != nil {
		return err
	}
	var request chatMessageRequest
	err = DecodeJSON(r, &request)
	if err != nil {
		return dto.ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error())
	}
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}
	message, err := h.service.UpdateChatMessage(r.Context(), svc.UpdateChatMessageInput{
		ID: id, Role: request.Role, Content: request.Content, Score: request.Score,
		UserID: userID, UpdatedBy: userID,
		Artifacts: request.Artifacts, SuggestedQuestions: request.SuggestedQuestions,
	})
	if err != nil {
		return dto.WrapError(dto.MapDatabaseError(err), "Failed to update chat message")
	}
	return respondJSON(w, http.StatusOK, messageResponse(message))
}

func (h *ChatMessageHandler) DeleteChatMessage(w http.ResponseWriter, r *http.Request) error {
	id, err := positiveInt32Param(r, "id")
	if err != nil {
		return err
	}
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}
	err = h.service.DeleteChatMessage(r.Context(), svc.DeleteChatMessageCommand{ID: id, UserID: userID})
	if err != nil {
		return dto.WrapError(dto.MapDatabaseError(err), "Failed to delete chat message")
	}
	return respondStatus(w, http.StatusOK)
}

func (h *ChatMessageHandler) GetAllChatMessages(w http.ResponseWriter, r *http.Request) error {
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}
	messages, err := h.service.GetAllChatMessages(r.Context(), userID)
	if err != nil {
		return dto.WrapError(dto.MapDatabaseError(err), "Failed to get chat messages")
	}
	return respondJSON(w, http.StatusOK, messageResponses(messages))
}

func (h *ChatMessageHandler) GetChatMessageByUUID(w http.ResponseWriter, r *http.Request) error {
	uuidStr := mux.Vars(r)["uuid"]
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}
	message, err := h.service.GetChatMessageByUUID(r.Context(), uuidStr, userID)
	if err != nil {
		return dto.WrapError(dto.MapDatabaseError(err), "Failed to get chat message")
	}
	return respondJSON(w, http.StatusOK, messageResponse(message))
}

func (h *ChatMessageHandler) UpdateChatMessageByUUID(w http.ResponseWriter, r *http.Request) error {
	var simpleMsg dto.SimpleChatMessage
	err := DecodeJSON(r, &simpleMsg)
	if err != nil {
		return dto.ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error())
	}
	tokenCount, _ := getTokenCount(simpleMsg.Text)
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}
	message, err := h.service.UpdateChatMessageByUUID(r.Context(), svc.UpdateChatMessageByUUIDInput{
		UUID: simpleMsg.Uuid, Content: simpleMsg.Text, TokenCount: int32(tokenCount), IsPin: simpleMsg.IsPin, UserID: userID,
	})
	if err != nil {
		return dto.WrapError(dto.MapDatabaseError(err), "Failed to update chat message")
	}
	return respondJSON(w, http.StatusOK, messageResponse(message))
}

func (h *ChatMessageHandler) DeleteChatMessageByUUID(w http.ResponseWriter, r *http.Request) error {
	uuidStr := mux.Vars(r)["uuid"]
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}
	err = h.service.DeleteChatMessageByUUID(r.Context(), svc.DeleteChatMessageByUUIDCommand{UUID: uuidStr, UserID: userID})
	if err != nil {
		return dto.WrapError(dto.MapDatabaseError(err), "Failed to delete chat message")
	}
	return respondStatus(w, http.StatusOK)
}

func (h *ChatMessageHandler) GetChatMessagesBySessionUUID(w http.ResponseWriter, r *http.Request) error {
	uuidStr := mux.Vars(r)["uuid"]
	page, err := httpx.ParsePage(r)
	if err != nil {
		return err
	}

	messages, err := h.service.GetChatMessagesBySessionUUID(r.Context(), svc.SessionMessagesPageQuery{SessionUUID: uuidStr, Page: svc.PageWindow{Limit: page.Limit, Offset: page.Offset}})
	if err != nil {
		return dto.WrapError(dto.MapDatabaseError(err), "Failed to get chat messages")
	}

	simpleMsgs := make([]dto.SimpleChatMessage, len(messages))
	for i, message := range messages {
		var artifacts []dto.Artifact
		if message.Artifacts != nil {
			err := json.Unmarshal(message.Artifacts, &artifacts)
			if err != nil {
				artifacts = []dto.Artifact{}
			}
		}

		simpleMsgs[i] = dto.SimpleChatMessage{
			DateTime:  message.UpdatedAt.Format(time.RFC3339),
			Text:      message.Content,
			Inversion: message.Role != "user",
			Error:     false,
			Loading:   false,
			Artifacts: artifacts,
		}
	}
	return respondJSON(w, http.StatusOK, simpleMsgs)
}

func (h *ChatMessageHandler) GetChatHistoryBySessionUUID(w http.ResponseWriter, r *http.Request) error {
	uuidStr := mux.Vars(r)["uuid"]
	page, err := httpx.ParsePage(r)
	if err != nil {
		return err
	}
	simpleMsgs, err := h.conversationSvc.History(r.Context(), svc.SessionHistoryQuery{SessionUUID: uuidStr, Page: svc.PageWindow{Limit: page.Limit, Offset: page.Offset}})
	if err != nil {
		return dto.WrapError(dto.MapDatabaseError(err), "Failed to get chat history")
	}
	response := make([]historyMessageHTTPResponse, 0, len(simpleMsgs))
	for _, message := range simpleMsgs {
		response = append(response, historyMessageResponse(message))
	}
	return respondJSON(w, http.StatusOK, response)
}

func (h *ChatMessageHandler) DeleteChatMessagesBySessionUUID(w http.ResponseWriter, r *http.Request) error {
	uuidStr := mux.Vars(r)["uuid"]
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}
	err = h.service.DeleteChatMessagesBySessionUUID(r.Context(), svc.DeleteSessionMessagesCommand{SessionUUID: uuidStr, UserID: userID})
	if err != nil {
		return dto.WrapError(dto.MapDatabaseError(err), "Failed to delete chat messages")
	}
	return respondStatus(w, http.StatusOK)
}

func (h *ChatMessageHandler) GenerateMoreSuggestions(w http.ResponseWriter, r *http.Request) error {
	messageUUID := mux.Vars(r)["uuid"]
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}

	message, err := h.service.GetChatMessageByUUID(r.Context(), messageUUID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return dto.ErrChatMessageNotFound.WithMessage("Message not found").WithDebugInfo(err.Error())
		}
		return dto.WrapError(dto.MapDatabaseError(err), "Failed to get message")
	}

	if message.Role != "assistant" {
		return dto.ErrValidationInvalidInput("Suggestions can only be generated for assistant messages")
	}

	session, err := h.sessionSvc.GetOwnedChatSession(r.Context(), svc.GetOwnedChatSessionQuery{UUID: message.ChatSessionUUID, UserID: userID})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return dto.ErrChatSessionNotFound.WithMessage("Session not found").WithDebugInfo(err.Error())
		}
		return dto.WrapError(dto.MapDatabaseError(err), "Failed to get session")
	}

	if !session.ExploreMode {
		return dto.ErrValidationInvalidInput("Suggestions are only available in explore mode")
	}

	contextMessages, err := h.service.GetLatestMessagesBySessionID(r.Context(), svc.LatestSessionMessagesQuery{SessionUUID: session.UUID, Limit: 6})
	if err != nil {
		return dto.WrapError(dto.MapDatabaseError(err), "Failed to get conversation context")
	}

	var msgs []models.Message
	for _, msg := range contextMessages {
		msgs = append(msgs, models.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	newSuggestions := h.chatService.GenerateSuggestedQuestions(r.Context(), message.Content, msgs)
	if len(newSuggestions) == 0 {
		return dto.CreateAPIError(dto.ErrInternalUnexpected, "Failed to generate suggestions", "no suggestions returned")
	}

	var existingSuggestions []string
	if len(message.SuggestedQuestions) > 0 {
		if err := json.Unmarshal(message.SuggestedQuestions, &existingSuggestions); err != nil {
			existingSuggestions = []string{}
		}
	}

	allSuggestions := append(existingSuggestions, newSuggestions...)

	seenSuggestions := make(map[string]bool)
	var uniqueSuggestions []string
	for _, suggestion := range allSuggestions {
		if !seenSuggestions[suggestion] {
			seenSuggestions[suggestion] = true
			uniqueSuggestions = append(uniqueSuggestions, suggestion)
		}
	}

	suggestionsJSON, err := json.Marshal(uniqueSuggestions)
	if err != nil {
		return dto.CreateAPIError(dto.ErrInternalUnexpected, "Failed to serialize suggestions", err.Error())
	}

	_, err = h.service.UpdateSuggestedQuestions(r.Context(), messageUUID, userID, suggestionsJSON)
	if err != nil {
		return dto.WrapError(dto.MapDatabaseError(err), "Failed to update message with suggestions")
	}

	response := suggestionsHTTPResponse{NewSuggestions: newSuggestions, AllSuggestions: uniqueSuggestions}

	return respondJSON(w, http.StatusOK, response)
}
