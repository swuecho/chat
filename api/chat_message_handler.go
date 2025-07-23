package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/swuecho/chat_backend/pkg/errors"
	"github.com/swuecho/chat_backend/repository"
	"github.com/swuecho/chat_backend/service"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

type ChatMessageHandler struct {
	serviceManager service.ServiceManager
}

func NewChatMessageHandler(sqlc_q *sqlc_queries.Queries) *ChatMessageHandler {
	// Initialize repository and service layers
	repositoryManager := repository.NewCoreRepositoryManager(sqlc_q)
	serviceManager := service.NewServiceManager(repositoryManager)
	
	return &ChatMessageHandler{
		serviceManager: serviceManager,
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
	router.HandleFunc("/uuid/chat_messages/chat_sessions/{uuid}", h.GetChatHistoryBySessionUUID).Methods(http.MethodGet)
	router.HandleFunc("/uuid/chat_messages/chat_sessions/{uuid}", h.DeleteChatMessagesBySesionUUID).Methods(http.MethodDelete)
}

func (h *ChatMessageHandler) CreateChatMessage(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SessionUUID string `json:"session_uuid"`
		Content     string `json:"content"`
		Role        string `json:"role"`
		Model       string `json:"model"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteErrorResponse(w, errors.InvalidInput("Invalid request format"))
		return
	}
	
	ctx := r.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		errors.WriteErrorResponse(w, errors.Unauthorized("Authentication required"))
		return
	}
	
	// Generate UUID for the message
	messageUUID := generateUUID()
	
	message, err := h.serviceManager.Chat().CreateChatMessage(ctx, req.SessionUUID, messageUUID, req.Role, req.Content, req.Model, userID)
	if err != nil {
		errors.WriteErrorResponse(w, err)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(message)
}

func (h *ChatMessageHandler) GetChatMessageByID(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		errors.WriteErrorResponse(w, errors.InvalidInput("Invalid message ID"))
		return
	}
	
	ctx := r.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		errors.WriteErrorResponse(w, errors.Unauthorized("Authentication required"))
		return
	}
	
	message, err := h.serviceManager.Chat().GetChatMessageByID(ctx, int32(id))
	if err != nil {
		errors.WriteErrorResponse(w, err)
		return
	}
	
	// Validate ownership through session
	if err := h.serviceManager.Chat().ValidateChatSession(ctx, message.ChatSessionUuid, userID); err != nil {
		errors.WriteErrorResponse(w, err)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(message)
}

func (h *ChatMessageHandler) UpdateChatMessage(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		errors.WriteErrorResponse(w, errors.InvalidInput("Invalid message ID"))
		return
	}
	
	var req struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteErrorResponse(w, errors.InvalidInput("Invalid request format"))
		return
	}
	
	ctx := r.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		errors.WriteErrorResponse(w, errors.Unauthorized("Authentication required"))
		return
	}
	
	// Get the message first to validate ownership through session
	message, err := h.serviceManager.Chat().GetChatMessageByID(ctx, int32(id))
	if err != nil {
		errors.WriteErrorResponse(w, err)
		return
	}
	
	// Update using message UUID
	updatedMessage, err := h.serviceManager.Chat().UpdateChatMessage(ctx, message.Uuid, req.Content, userID)
	if err != nil {
		errors.WriteErrorResponse(w, err)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedMessage)
}

func (h *ChatMessageHandler) DeleteChatMessage(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		errors.WriteErrorResponse(w, errors.InvalidInput("Invalid message ID"))
		return
	}
	
	ctx := r.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		errors.WriteErrorResponse(w, errors.Unauthorized("Authentication required"))
		return
	}
	
	err = h.serviceManager.Chat().DeleteChatMessage(ctx, int32(id), userID)
	if err != nil {
		errors.WriteErrorResponse(w, err)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Chat message deleted successfully"})
}

func (h *ChatMessageHandler) GetAllChatMessages(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, err := getUserID(ctx)
	if err != nil {
		errors.WriteErrorResponse(w, errors.Unauthorized("Authentication required"))
		return
	}
	
	// Only allow admin users to get all messages
	// For now, we'll restrict this to prevent data leaks
	errors.WriteErrorResponse(w, errors.ErrForbidden.WithDetail("Operation not permitted"))
}

// GetChatMessageByUUID get chat message by uuid
func (h *ChatMessageHandler) GetChatMessageByUUID(w http.ResponseWriter, r *http.Request) {
	uuidStr := mux.Vars(r)["uuid"]
	if uuidStr == "" {
		errors.WriteErrorResponse(w, errors.InvalidInput("UUID is required"))
		return
	}
	
	ctx := r.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		errors.WriteErrorResponse(w, errors.Unauthorized("Authentication required"))
		return
	}
	
	message, err := h.serviceManager.Chat().GetChatMessageByUUID(ctx, uuidStr)
	if err != nil {
		errors.WriteErrorResponse(w, err)
		return
	}
	
	// Validate ownership through session
	if err := h.serviceManager.Chat().ValidateChatSession(ctx, message.ChatSessionUuid, userID); err != nil {
		errors.WriteErrorResponse(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(message)
}

// UpdateChatMessageByUUID update chat message by uuid
func (h *ChatMessageHandler) UpdateChatMessageByUUID(w http.ResponseWriter, r *http.Request) {
	uuidStr := mux.Vars(r)["uuid"]
	if uuidStr == "" {
		errors.WriteErrorResponse(w, errors.InvalidInput("UUID is required"))
		return
	}
	
	var simple_msg SimpleChatMessage
	err := json.NewDecoder(r.Body).Decode(&simple_msg)
	if err != nil {
		errors.WriteErrorResponse(w, errors.InvalidInput("Invalid request format"))
		return
	}
	
	ctx := r.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		errors.WriteErrorResponse(w, errors.Unauthorized("Authentication required"))
		return
	}
	
	var messageParams sqlc_queries.UpdateChatMessageByUUIDParams
	messageParams.Uuid = uuidStr
	messageParams.Content = simple_msg.Text
	tokenCount, _ := getTokenCount(simple_msg.Text)
	messageParams.TokenCount = int32(tokenCount)
	messageParams.IsPin = simple_msg.IsPin
	
	// Convert artifacts if present
	if len(simple_msg.Artifacts) > 0 {
		artifactsJSON, err := json.Marshal(simple_msg.Artifacts)
		if err == nil {
			messageParams.Artifacts = artifactsJSON
		}
	}
	
	message, err := h.serviceManager.Chat().UpdateChatMessageByUUID(ctx, messageParams, userID)
	if err != nil {
		errors.WriteErrorResponse(w, err)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(message)
}

// DeleteChatMessageByUUID delete chat message by uuid
func (h *ChatMessageHandler) DeleteChatMessageByUUID(w http.ResponseWriter, r *http.Request) {
	uuidStr := mux.Vars(r)["uuid"]
	if uuidStr == "" {
		errors.WriteErrorResponse(w, errors.InvalidInput("UUID is required"))
		return
	}
	
	ctx := r.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		errors.WriteErrorResponse(w, errors.Unauthorized("Authentication required"))
		return
	}
	
	err = h.serviceManager.Chat().DeleteChatMessageByUUID(ctx, uuidStr, userID)
	if err != nil {
		errors.WriteErrorResponse(w, err)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Chat message deleted successfully"})
}

// GetChatMessagesBySessionUUID get chat messages by session uuid
func (h *ChatMessageHandler) GetChatMessagesBySessionUUID(w http.ResponseWriter, r *http.Request) {
	uuidStr := mux.Vars(r)["uuid"]
	if uuidStr == "" {
		errors.WriteErrorResponse(w, errors.InvalidInput("UUID is required"))
		return
	}
	
	pageNum, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil {
		pageNum = 1
	}
	pageSize, err := strconv.Atoi(r.URL.Query().Get("page_size"))
	if err != nil {
		pageSize = 200
	}

	ctx := r.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		errors.WriteErrorResponse(w, errors.Unauthorized("Authentication required"))
		return
	}
	
	// Validate session ownership
	if err := h.serviceManager.Chat().ValidateChatSession(ctx, uuidStr, userID); err != nil {
		errors.WriteErrorResponse(w, err)
		return
	}

	messages, err := h.serviceManager.Chat().GetChatMessagesBySessionUUID(ctx, uuidStr, int32(pageNum), int32(pageSize))
	if err != nil {
		errors.WriteErrorResponse(w, err)
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
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(simple_msgs)
}

// GetChatHistoryBySessionUUID get chat messages by session uuid (legacy method)
func (h *ChatMessageHandler) GetChatHistoryBySessionUUID(w http.ResponseWriter, r *http.Request) {
	// This method appears to use a specific database query for chat history
	// For now, redirect to the main GetChatMessagesBySessionUUID method
	h.GetChatMessagesBySessionUUID(w, r)
}

// DeleteChatMessagesBySesionUUID delete chat messages by session uuid
func (h *ChatMessageHandler) DeleteChatMessagesBySesionUUID(w http.ResponseWriter, r *http.Request) {
	uuidStr := mux.Vars(r)["uuid"]
	if uuidStr == "" {
		errors.WriteErrorResponse(w, errors.InvalidInput("UUID is required"))
		return
	}
	
	ctx := r.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		errors.WriteErrorResponse(w, errors.Unauthorized("Authentication required"))
		return
	}
	
	err = h.serviceManager.Chat().DeleteChatMessagesBySessionUUID(ctx, uuidStr, userID)
	if err != nil {
		errors.WriteErrorResponse(w, err)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "All chat messages deleted successfully"})
}

// generateUUID creates a proper UUID 
func generateUUID() string {
	return uuid.New().String()
}