package main

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/pkg/errors"
	"github.com/swuecho/chat_backend/repository"
	"github.com/swuecho/chat_backend/service"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

type ChatSessionHandler struct {
	serviceManager service.ServiceManager
}

func NewChatSessionHandler(sqlc_q *sqlc_queries.Queries) *ChatSessionHandler {
	// Initialize repository and service layers
	repositoryManager := repository.NewCoreRepositoryManager(sqlc_q)
	serviceManager := service.NewServiceManager(repositoryManager)
	
	return &ChatSessionHandler{
		serviceManager: serviceManager,
	}
}

func (h *ChatSessionHandler) Register(router *mux.Router) {
	router.HandleFunc("/chat_sessions/user", h.getSimpleChatSessionsByUserID).Methods(http.MethodGet)
	router.HandleFunc("/uuid/chat_sessions/max_length/{uuid}", h.updateSessionMaxLength).Methods("PUT")
	router.HandleFunc("/uuid/chat_sessions/topic/{uuid}", h.updateChatSessionTopicByUUID).Methods("PUT")
	router.HandleFunc("/uuid/chat_sessions/{uuid}", h.getChatSessionByUUID).Methods("GET")
	router.HandleFunc("/uuid/chat_sessions/{uuid}", h.createOrUpdateChatSessionByUUID).Methods("PUT")
	router.HandleFunc("/uuid/chat_sessions/{uuid}", h.deleteChatSessionByUUID).Methods("DELETE")
	router.HandleFunc("/uuid/chat_sessions", h.createChatSessionByUUID).Methods("POST")
	router.HandleFunc("/uuid/chat_session_from_snapshot/{uuid}", h.createChatSessionFromSnapshot).Methods(http.MethodPost)
}

// Using existing ChatSessionResponse from models.go

// CreateChatSessionRequest represents the request format for creating sessions
type CreateChatSessionRequest struct {
	Topic string `json:"topic"`
	Model string `json:"model"`
}

// UpdateTopicRequest represents the request format for updating session topic
type UpdateTopicRequest struct {
	Topic string `json:"topic"`
}

// getChatSessionByUUID returns a chat session by its UUID using service layer
func (h *ChatSessionHandler) getChatSessionByUUID(w http.ResponseWriter, r *http.Request) {
	uuid := mux.Vars(r)["uuid"]
	if uuid == "" {
		errors.WriteErrorResponse(w, errors.InvalidInput("UUID is required"))
		return
	}

	ctx := r.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		errors.WriteErrorResponse(w, errors.Unauthorized("Authentication required"))
		return
	}

	// Use service layer with validation
	session, err := h.serviceManager.Chat().GetChatSession(ctx, uuid)
	if err != nil {
		errors.WriteErrorResponse(w, err)
		return
	}

	// Validate ownership
	if session.UserID != userID {
		errors.WriteErrorResponse(w, errors.ErrForbidden.WithDetail("Session does not belong to user"))
		return
	}

	// Return structured response
	response := &ChatSessionResponse{
		Uuid:      session.Uuid,
		Topic:     session.Topic,
		MaxLength: session.MaxLength,
		CreatedAt: session.CreatedAt,
		UpdatedAt: session.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// getSimpleChatSessionsByUserID retrieves all sessions for the authenticated user
func (h *ChatSessionHandler) getSimpleChatSessionsByUserID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		errors.WriteErrorResponse(w, errors.Unauthorized("Authentication required"))
		return
	}

	// Use service layer
	sessions, err := h.serviceManager.Chat().GetUserChatSessions(ctx, userID)
	if err != nil {
		errors.WriteErrorResponse(w, err)
		return
	}

	// Convert to response format
	responses := make([]ChatSessionResponse, len(sessions))
	for i, session := range sessions {
		responses[i] = ChatSessionResponse{
			Uuid:      session.Uuid,
			Topic:     session.Topic,
			MaxLength: session.MaxLength,
			CreatedAt: session.CreatedAt,
			UpdatedAt: session.UpdatedAt,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responses)
}

// createChatSessionByUUID creates a chat session using service layer
func (h *ChatSessionHandler) createChatSessionByUUID(w http.ResponseWriter, r *http.Request) {
	var req CreateChatSessionRequest
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

	// Validate input
	if req.Topic == "" {
		errors.WriteErrorResponse(w, errors.ValidationFailed("topic", "Topic is required"))
		return
	}
	if req.Model == "" {
		errors.WriteErrorResponse(w, errors.ValidationFailed("model", "Model is required"))
		return
	}

	// Create session using service layer
	session, err := h.serviceManager.Chat().CreateChatSession(ctx, userID, req.Topic, req.Model)
	if err != nil {
		errors.WriteErrorResponse(w, err)
		return
	}

	response := &ChatSessionResponse{
		Uuid:      session.Uuid,
		Topic:     session.Topic,
		MaxLength: session.MaxLength,
		CreatedAt: session.CreatedAt,
		UpdatedAt: session.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// updateChatSessionTopicByUUID updates a session's topic using service layer
func (h *ChatSessionHandler) updateChatSessionTopicByUUID(w http.ResponseWriter, r *http.Request) {
	uuid := mux.Vars(r)["uuid"]
	if uuid == "" {
		errors.WriteErrorResponse(w, errors.InvalidInput("UUID is required"))
		return
	}

	var req UpdateTopicRequest
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

	// Validate input
	if req.Topic == "" {
		errors.WriteErrorResponse(w, errors.ValidationFailed("topic", "Topic is required"))
		return
	}

	// First validate session ownership
	if err := h.serviceManager.Chat().ValidateChatSession(ctx, uuid, userID); err != nil {
		errors.WriteErrorResponse(w, err)
		return
	}

	// Update using service layer
	updates := service.ChatSessionUpdate{
		Topic: &req.Topic,
	}
	
	session, err := h.serviceManager.Chat().UpdateChatSession(ctx, uuid, updates)
	if err != nil {
		errors.WriteErrorResponse(w, err)
		return
	}

	response := &ChatSessionResponse{
		Uuid:      session.Uuid,
		Topic:     session.Topic,
		MaxLength: session.MaxLength,
		CreatedAt: session.CreatedAt,
		UpdatedAt: session.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// deleteChatSessionByUUID deletes a chat session using service layer
func (h *ChatSessionHandler) deleteChatSessionByUUID(w http.ResponseWriter, r *http.Request) {
	uuid := mux.Vars(r)["uuid"]
	if uuid == "" {
		errors.WriteErrorResponse(w, errors.InvalidInput("UUID is required"))
		return
	}

	ctx := r.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		errors.WriteErrorResponse(w, errors.Unauthorized("Authentication required"))
		return
	}

	// Delete using service layer with validation
	err = h.serviceManager.Chat().DeleteChatSession(ctx, uuid, userID)
	if err != nil {
		errors.WriteErrorResponse(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Chat session deleted successfully",
		"uuid":    uuid,
	})
}

// Placeholder implementations for methods that need more complex migration
// These would need full implementation based on the original logic

func (h *ChatSessionHandler) createOrUpdateChatSessionByUUID(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement using service layer
	errors.WriteErrorResponse(w, errors.ErrInternalServer.WithDetail("Method not yet migrated to service layer"))
}

func (h *ChatSessionHandler) updateSessionMaxLength(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement using service layer
	errors.WriteErrorResponse(w, errors.ErrInternalServer.WithDetail("Method not yet migrated to service layer"))
}

func (h *ChatSessionHandler) createChatSessionFromSnapshot(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement using service layer
	errors.WriteErrorResponse(w, errors.ErrInternalServer.WithDetail("Method not yet migrated to service layer"))
}