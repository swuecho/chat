package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/samber/lo"
	"github.com/swuecho/chat_backend/models"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

type ChatMessageHandler struct {
	service *ChatMessageService
}

func NewChatMessageHandler(sqlc_q *sqlc_queries.Queries) *ChatMessageHandler {
	chatMessageService := NewChatMessageService(sqlc_q)
	return &ChatMessageHandler{
		service: chatMessageService,
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

//type userIdContextKey string

//const userIDKey = userIdContextKey("userID")

func (h *ChatMessageHandler) CreateChatMessage(w http.ResponseWriter, r *http.Request) {
	var messageParams sqlc_queries.CreateChatMessageParams
	err := json.NewDecoder(r.Body).Decode(&messageParams)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	message, err := h.service.CreateChatMessage(r.Context(), messageParams)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(message)
}

func (h *ChatMessageHandler) GetChatMessageByID(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid chat message ID", http.StatusBadRequest)
		return
	}
	message, err := h.service.GetChatMessageByID(r.Context(), int32(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(message)
}

func (h *ChatMessageHandler) UpdateChatMessage(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid chat message ID", http.StatusBadRequest)
		return
	}
	var messageParams sqlc_queries.UpdateChatMessageParams
	err = json.NewDecoder(r.Body).Decode(&messageParams)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	messageParams.ID = int32(id)
	message, err := h.service.UpdateChatMessage(r.Context(), messageParams)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(message)
}

func (h *ChatMessageHandler) DeleteChatMessage(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid chat message ID", http.StatusBadRequest)
		return
	}
	err = h.service.DeleteChatMessage(r.Context(), int32(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *ChatMessageHandler) GetAllChatMessages(w http.ResponseWriter, r *http.Request) {
	messages, err := h.service.GetAllChatMessages(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(messages)
}

// GetChatMessageByUUID get chat message by uuid
func (h *ChatMessageHandler) GetChatMessageByUUID(w http.ResponseWriter, r *http.Request) {
	uuidStr := mux.Vars(r)["uuid"]
	message, err := h.service.GetChatMessageByUUID(r.Context(), uuidStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(message)
}

// UpdateChatMessageByUUID update chat message by uuid
func (h *ChatMessageHandler) UpdateChatMessageByUUID(w http.ResponseWriter, r *http.Request) {
	var simple_msg SimpleChatMessage
	err := json.NewDecoder(r.Body).Decode(&simple_msg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var messageParams sqlc_queries.UpdateChatMessageByUUIDParams
	messageParams.Uuid = simple_msg.Uuid
	messageParams.Content = simple_msg.Text
	tokenCount, _ := getTokenCount(simple_msg.Text)
	messageParams.TokenCount = int32(tokenCount)
	messageParams.IsPin = simple_msg.IsPin
	message, err := h.service.UpdateChatMessageByUUID(r.Context(), messageParams)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(message)
}

// DeleteChatMessageByUUID delete chat message by uuid
func (h *ChatMessageHandler) DeleteChatMessageByUUID(w http.ResponseWriter, r *http.Request) {
	uuidStr := mux.Vars(r)["uuid"]
	err := h.service.DeleteChatMessageByUUID(r.Context(), uuidStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// GetChatMessagesBySessionUUID get chat messages by session uuid
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
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	simple_msgs := lo.Map(messages, func(message sqlc_queries.ChatMessage, _ int) SimpleChatMessage {
		// Extract artifacts from database
		var artifacts []Artifact
		if message.Artifacts != nil {
			err := json.Unmarshal(message.Artifacts, &artifacts)
			if err != nil {
				// Log error but don't fail the request
				artifacts = []Artifact{}
			}
		}

		return SimpleChatMessage{
			DateTime:  message.UpdatedAt.Format(time.RFC3339),
			Text:      message.Content,
			Inversion: message.Role != "user",
			Error:     false,
			Loading:   false,
			Artifacts: artifacts,
		}
	})
	json.NewEncoder(w).Encode(simple_msgs)
}

// GetChatMessagesBySessionUUID get chat messages by session uuid
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
	simple_msgs, err := h.service.q.GetChatHistoryBySessionUUID(r.Context(), uuidStr, int32(pageNum), int32(pageSize))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(simple_msgs)
}

// DeleteChatMessagesBySesionUUID delete chat messages by session uuid
func (h *ChatMessageHandler) DeleteChatMessagesBySesionUUID(w http.ResponseWriter, r *http.Request) {
	uuidStr := mux.Vars(r)["uuid"]
	err := h.service.DeleteChatMessagesBySesionUUID(r.Context(), uuidStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// GenerateMoreSuggestions generates additional suggested questions for a message
func (h *ChatMessageHandler) GenerateMoreSuggestions(w http.ResponseWriter, r *http.Request) {
	messageUUID := mux.Vars(r)["uuid"]

	// Get the existing message
	message, err := h.service.q.GetChatMessageByUUID(r.Context(), messageUUID)
	if err != nil {
		http.Error(w, "Message not found", http.StatusNotFound)
		return
	}

	// Only allow suggestions for assistant messages
	if message.Role != "assistant" {
		http.Error(w, "Suggestions can only be generated for assistant messages", http.StatusBadRequest)
		return
	}

	// Get the session to check if explore mode is enabled
	session, err := h.service.q.GetChatSessionByUUID(r.Context(), message.ChatSessionUuid)
	if err != nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Check if explore mode is enabled
	if !session.ExploreMode {
		http.Error(w, "Suggestions are only available in explore mode", http.StatusBadRequest)
		return
	}

	// Get conversation context - last 6 messages
	contextMessages, err := h.service.q.GetLatestMessagesBySessionUUID(r.Context(),
		sqlc_queries.GetLatestMessagesBySessionUUIDParams{
			ChatSessionUuid: session.Uuid,
			Limit:           6,
		})
	if err != nil {
		http.Error(w, "Failed to get conversation context", http.StatusInternalServerError)
		return
	}

	// Convert to models.Message format for suggestion generation
	var msgs []models.Message
	for _, msg := range contextMessages {
		msgs = append(msgs, models.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// Create a new ChatService to access suggestion generation methods
	chatService := NewChatService(h.service.q)

	// Generate new suggested questions
	newSuggestions := chatService.generateSuggestedQuestions(message.Content, msgs)
	if len(newSuggestions) == 0 {
		http.Error(w, "Failed to generate suggestions", http.StatusInternalServerError)
		return
	}

	// Parse existing suggestions
	var existingSuggestions []string
	if len(message.SuggestedQuestions) > 0 {
		if err := json.Unmarshal(message.SuggestedQuestions, &existingSuggestions); err != nil {
			// If unmarshal fails, treat as empty array
			existingSuggestions = []string{}
		}
	}

	// Combine existing and new suggestions (avoiding duplicates)
	allSuggestions := append(existingSuggestions, newSuggestions...)

	// Remove duplicates
	seenSuggestions := make(map[string]bool)
	var uniqueSuggestions []string
	for _, suggestion := range allSuggestions {
		if !seenSuggestions[suggestion] {
			seenSuggestions[suggestion] = true
			uniqueSuggestions = append(uniqueSuggestions, suggestion)
		}
	}

	// Update the message with new suggestions
	suggestionsJSON, err := json.Marshal(uniqueSuggestions)
	if err != nil {
		http.Error(w, "Failed to serialize suggestions", http.StatusInternalServerError)
		return
	}

	_, err = h.service.q.UpdateChatMessageSuggestions(r.Context(),
		sqlc_queries.UpdateChatMessageSuggestionsParams{
			Uuid:               messageUUID,
			SuggestedQuestions: suggestionsJSON,
		})
	if err != nil {
		http.Error(w, "Failed to update message with suggestions", http.StatusInternalServerError)
		return
	}

	// Return the new suggestions to the client
	response := map[string]interface{}{
		"newSuggestions": newSuggestions,
		"allSuggestions": uniqueSuggestions,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
