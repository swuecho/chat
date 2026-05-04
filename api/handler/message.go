package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/samber/lo"
	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/models"
	"github.com/swuecho/chat_backend/sqlc_queries"
	"github.com/swuecho/chat_backend/svc"
)

type ChatMessageHandler struct {
	service     *svc.ChatMessageService
	sessionSvc  *svc.ChatSessionService
	openAIKey   string
	openAIProxy string
}

func NewChatMessageHandler(sqlc_q *sqlc_queries.Queries, openAIKey, openAIProxy string) *ChatMessageHandler {
	return &ChatMessageHandler{
		service:     svc.NewChatMessageService(sqlc_q),
		sessionSvc:  svc.NewChatSessionService(sqlc_q),
		openAIKey:   openAIKey,
		openAIProxy: openAIProxy,
	}
}

func (h *ChatMessageHandler) Register(router *mux.Router) {
	router.HandleFunc("/chat_messages", h.CreateChatMessage).Methods(http.MethodPost)
	router.HandleFunc("/chat_messages/{id}", h.GetChatMessageByID).Methods(http.MethodGet)
	router.HandleFunc("/chat_messages/{id}", h.UpdateChatMessage).Methods(http.MethodPut)
	router.HandleFunc("/chat_messages/{id}", h.DeleteChatMessage).Methods(http.MethodDelete)
	router.HandleFunc("/chat_messages", h.GetAllChatMessages).Methods(http.MethodGet)

	router.HandleFunc("/uuid/chat_messages/{uuid}", h.GetChatMessageByUUID).Methods(http.MethodGet)
	router.HandleFunc("/uuid/chat_messages/{uuid}", h.UpdateChatMessageByUUID).Methods(http.MethodPut)
	router.HandleFunc("/uuid/chat_messages/{uuid}", h.DeleteChatMessageByUUID).Methods(http.MethodDelete)
	router.HandleFunc("/uuid/chat_messages/{uuid}/generate-suggestions", h.GenerateMoreSuggestions).Methods(http.MethodPost)
	router.HandleFunc("/uuid/chat_messages/chat_sessions/{uuid}", h.GetChatHistoryBySessionUUID).Methods(http.MethodGet)
	router.HandleFunc("/uuid/chat_messages/chat_sessions/{uuid}", h.DeleteChatMessagesBySesionUUID).Methods(http.MethodDelete)
}

func (h *ChatMessageHandler) CreateChatMessage(w http.ResponseWriter, r *http.Request) {
	var messageParams sqlc_queries.CreateChatMessageParams
	err := json.NewDecoder(r.Body).Decode(&messageParams)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error()))
		return
	}
	message, err := h.service.CreateChatMessage(r.Context(), messageParams)
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(dto.MapDatabaseError(err), "Failed to create chat message"))
		return
	}
	json.NewEncoder(w).Encode(message)
}

func (h *ChatMessageHandler) GetChatMessageByID(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("invalid chat message ID"))
		return
	}
	message, err := h.service.GetChatMessageByID(r.Context(), int32(id))
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(dto.MapDatabaseError(err), "Failed to get chat message"))
		return
	}
	json.NewEncoder(w).Encode(message)
}

func (h *ChatMessageHandler) UpdateChatMessage(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("invalid chat message ID"))
		return
	}
	var messageParams sqlc_queries.UpdateChatMessageParams
	err = json.NewDecoder(r.Body).Decode(&messageParams)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error()))
		return
	}
	messageParams.ID = int32(id)
	message, err := h.service.UpdateChatMessage(r.Context(), messageParams)
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(dto.MapDatabaseError(err), "Failed to update chat message"))
		return
	}
	json.NewEncoder(w).Encode(message)
}

func (h *ChatMessageHandler) DeleteChatMessage(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("invalid chat message ID"))
		return
	}
	err = h.service.DeleteChatMessage(r.Context(), int32(id))
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(dto.MapDatabaseError(err), "Failed to delete chat message"))
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *ChatMessageHandler) GetAllChatMessages(w http.ResponseWriter, r *http.Request) {
	messages, err := h.service.GetAllChatMessages(r.Context())
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(dto.MapDatabaseError(err), "Failed to get chat messages"))
		return
	}
	json.NewEncoder(w).Encode(messages)
}

func (h *ChatMessageHandler) GetChatMessageByUUID(w http.ResponseWriter, r *http.Request) {
	uuidStr := mux.Vars(r)["uuid"]
	message, err := h.service.GetChatMessageByUUID(r.Context(), uuidStr)
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(dto.MapDatabaseError(err), "Failed to get chat message"))
		return
	}
	json.NewEncoder(w).Encode(message)
}

func (h *ChatMessageHandler) UpdateChatMessageByUUID(w http.ResponseWriter, r *http.Request) {
	var simpleMsg dto.SimpleChatMessage
	err := json.NewDecoder(r.Body).Decode(&simpleMsg)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error()))
		return
	}
	var messageParams sqlc_queries.UpdateChatMessageByUUIDParams
	messageParams.Uuid = simpleMsg.Uuid
	messageParams.Content = simpleMsg.Text
	tokenCount, _ := getTokenCount(simpleMsg.Text)
	messageParams.TokenCount = int32(tokenCount)
	messageParams.IsPin = simpleMsg.IsPin
	message, err := h.service.UpdateChatMessageByUUID(r.Context(), messageParams)
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(dto.MapDatabaseError(err), "Failed to update chat message"))
		return
	}
	json.NewEncoder(w).Encode(message)
}

func (h *ChatMessageHandler) DeleteChatMessageByUUID(w http.ResponseWriter, r *http.Request) {
	uuidStr := mux.Vars(r)["uuid"]
	err := h.service.DeleteChatMessageByUUID(r.Context(), uuidStr)
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(dto.MapDatabaseError(err), "Failed to delete chat message"))
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *ChatMessageHandler) GetChatMessagesBySessionUUID(w http.ResponseWriter, r *http.Request) {
	uuidStr := mux.Vars(r)["uuid"]
	pageNum, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil {
		pageNum = 1
	}
	pageSize, err := strconv.Atoi(r.URL.Query().Get("page_size"))
	if err != nil {
		pageSize = 200
	}

	messages, err := h.service.GetChatMessagesBySessionUUID(r.Context(), uuidStr, int32(pageNum), int32(pageSize))
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(dto.MapDatabaseError(err), "Failed to get chat messages"))
		return
	}

	simpleMsgs := lo.Map(messages, func(message sqlc_queries.ChatMessage, _ int) dto.SimpleChatMessage {
		var artifacts []dto.Artifact
		if message.Artifacts != nil {
			err := json.Unmarshal(message.Artifacts, &artifacts)
			if err != nil {
				artifacts = []dto.Artifact{}
			}
		}

		return dto.SimpleChatMessage{
			DateTime:  message.UpdatedAt.Format(time.RFC3339),
			Text:      message.Content,
			Inversion: message.Role != "user",
			Error:     false,
			Loading:   false,
			Artifacts: artifacts,
		}
	})
	json.NewEncoder(w).Encode(simpleMsgs)
}

func (h *ChatMessageHandler) GetChatHistoryBySessionUUID(w http.ResponseWriter, r *http.Request) {
	uuidStr := mux.Vars(r)["uuid"]
	pageNum, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil {
		pageNum = 1
	}
	pageSize, err := strconv.Atoi(r.URL.Query().Get("page_size"))
	if err != nil {
		pageSize = 200
	}
	simpleMsgs, err := h.sessionSvc.GetChatHistoryBySessionUUID(r.Context(), uuidStr, int32(pageNum), int32(pageSize))
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(dto.MapDatabaseError(err), "Failed to get chat history"))
		return
	}
	json.NewEncoder(w).Encode(simpleMsgs)
}

func (h *ChatMessageHandler) DeleteChatMessagesBySesionUUID(w http.ResponseWriter, r *http.Request) {
	uuidStr := mux.Vars(r)["uuid"]
	err := h.service.DeleteChatMessagesBySesionUUID(r.Context(), uuidStr)
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(dto.MapDatabaseError(err), "Failed to delete chat messages"))
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *ChatMessageHandler) GenerateMoreSuggestions(w http.ResponseWriter, r *http.Request) {
	messageUUID := mux.Vars(r)["uuid"]

	message, err := h.service.GetChatMessageByUUID(r.Context(), messageUUID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			dto.RespondWithAPIError(w, dto.ErrChatMessageNotFound.WithMessage("Message not found").WithDebugInfo(err.Error()))
		} else {
			dto.RespondWithAPIError(w, dto.WrapError(dto.MapDatabaseError(err), "Failed to get message"))
		}
		return
	}

	if message.Role != "assistant" {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("Suggestions can only be generated for assistant messages"))
		return
	}

	session, err := h.sessionSvc.GetChatSessionByUUID(r.Context(), message.ChatSessionUuid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			dto.RespondWithAPIError(w, dto.ErrChatSessionNotFound.WithMessage("Session not found").WithDebugInfo(err.Error()))
		} else {
			dto.RespondWithAPIError(w, dto.WrapError(dto.MapDatabaseError(err), "Failed to get session"))
		}
		return
	}

	if !session.ExploreMode {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("Suggestions are only available in explore mode"))
		return
	}

	contextMessages, err := h.service.GetLatestMessagesBySessionID(r.Context(), session.Uuid, 6)
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(dto.MapDatabaseError(err), "Failed to get conversation context"))
		return
	}

	var msgs []models.Message
	for _, msg := range contextMessages {
		msgs = append(msgs, models.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	chatService := svc.NewChatService(h.service.Q(), h.openAIKey, h.openAIProxy)

	newSuggestions := chatService.GenerateSuggestedQuestions(message.Content, msgs)
	if len(newSuggestions) == 0 {
		dto.RespondWithAPIError(w, dto.CreateAPIError(dto.ErrInternalUnexpected, "Failed to generate suggestions", "no suggestions returned"))
		return
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
		dto.RespondWithAPIError(w, dto.CreateAPIError(dto.ErrInternalUnexpected, "Failed to serialize suggestions", err.Error()))
		return
	}

	_, err = h.sessionSvc.UpdateChatMessageSuggestions(r.Context(),
		sqlc_queries.UpdateChatMessageSuggestionsParams{
			Uuid:               messageUUID,
			SuggestedQuestions: suggestionsJSON,
		})
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(dto.MapDatabaseError(err), "Failed to update message with suggestions"))
		return
	}

	response := map[string]interface{}{
		"newSuggestions": newSuggestions,
		"allSuggestions": uniqueSuggestions,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
