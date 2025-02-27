package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"

	"github.com/gorilla/mux"
	sqlc "github.com/swuecho/chat_backend/sqlc_queries"
)

// UserActiveChatSessionHandler handles requests related to active chat sessions
type UserActiveChatSessionHandler struct {
	service *UserActiveChatSessionService
}

// NewUserActiveChatSessionHandler creates a new handler instance
func NewUserActiveChatSessionHandler(sqlc_q *sqlc.Queries) *UserActiveChatSessionHandler {
	activeSessionService := NewUserActiveChatSessionService(sqlc_q)
	return &UserActiveChatSessionHandler{
		service: activeSessionService,
	}
}

// Register sets up the handler routes
func (h *UserActiveChatSessionHandler) Register(router *mux.Router) {
	router.HandleFunc("/uuid/user_active_chat_session", h.GetUserActiveChatSessionHandler).Methods(http.MethodGet)
	router.HandleFunc("/uuid/user_active_chat_session", h.CreateOrUpdateUserActiveChatSessionHandler).Methods(http.MethodPut)
}

// GetUserActiveChatSessionHandler handles GET requests to get a session by user_id
func (h *UserActiveChatSessionHandler) GetUserActiveChatSessionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get and validate user ID
	userID, err := getUserID(ctx)
	if err != nil {
		RespondWithAPIError(w, ErrAuthInvalidCredentials.WithDetail("missing or invalid user ID"))
		return
	}

	log.Printf("Getting active chat session for user %d", userID)

	// Get session from service
	session, err := h.service.GetUserActiveChatSession(r.Context(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			RespondWithAPIError(w, ErrChatSessionNotFound.WithDetail(fmt.Sprintf("no active session for user %d", userID)))
		} else {
			RespondWithAPIError(w, WrapError(err, "failed to get active chat session"))
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(session); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

// UUID validation regex
var uuidRegex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// CreateOrUpdateUserActiveChatSessionHandler handles PUT requests to create/update a session
func (h *UserActiveChatSessionHandler) CreateOrUpdateUserActiveChatSessionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get and validate user ID
	userID, err := getUserID(ctx)
	if err != nil {
		RespondWithAPIError(w, ErrAuthInvalidCredentials.WithDetail("missing or invalid user ID"))
		return
	}

	// Parse request body
	var sessionParams sqlc.CreateOrUpdateUserActiveChatSessionParams
	if err := json.NewDecoder(r.Body).Decode(&sessionParams); err != nil {
		RespondWithAPIError(w, ErrValidationInvalidInput("failed to parse request body"))
		return
	}

	// Validate session UUID format
	if !uuidRegex.MatchString(sessionParams.ChatSessionUuid) {
		RespondWithAPIError(w, ErrValidationInvalidInput("invalid session UUID format"))
		return
	}

	// Use the user_id from token
	sessionParams.UserID = userID

	log.Printf("Creating/updating active chat session for user %d", userID)

	// Create/update session
	session, err := h.service.CreateOrUpdateUserActiveChatSession(r.Context(), sessionParams)
	if err != nil {
		RespondWithAPIError(w, WrapError(err, "failed to create or update active chat session"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(session); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}
