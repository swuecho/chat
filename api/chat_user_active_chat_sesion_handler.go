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
	
	// Per-workspace active session endpoints
	router.HandleFunc("/workspaces/{workspaceUuid}/active-session", h.GetWorkspaceActiveSessionHandler).Methods(http.MethodGet)
	router.HandleFunc("/workspaces/{workspaceUuid}/active-session", h.SetWorkspaceActiveSessionHandler).Methods(http.MethodPut)
	router.HandleFunc("/workspaces/active-sessions", h.GetAllWorkspaceActiveSessionsHandler).Methods(http.MethodGet)
}

// GetUserActiveChatSessionHandler handles GET requests to get a session by user_id
func (h *UserActiveChatSessionHandler) GetUserActiveChatSessionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get and validate user ID
	userID, err := getUserID(ctx)
	if err != nil {
		RespondWithAPIError(w, ErrAuthInvalidCredentials.WithMessage("missing or invalid user ID"))
		return
	}

	log.Printf("Getting active chat session for user %d", userID)

	// Get session from service
	session, err := h.service.GetUserActiveChatSession(r.Context(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			RespondWithAPIError(w, ErrChatSessionNotFound.WithMessage(fmt.Sprintf("no active session for user %d", userID)))
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
		RespondWithAPIError(w, ErrAuthInvalidCredentials.WithMessage("missing or invalid user ID"))
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
		RespondWithAPIError(w, ErrChatSessionInvalid.WithMessage(
			fmt.Sprintf("invalid session UUID format: %s", sessionParams.ChatSessionUuid)))
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

// Per-workspace active session handlers

// GetWorkspaceActiveSessionHandler gets the active session for a specific workspace
func (h *UserActiveChatSessionHandler) GetWorkspaceActiveSessionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	workspaceUuid := mux.Vars(r)["workspaceUuid"]

	// Get and validate user ID
	userID, err := getUserID(ctx)
	if err != nil {
		RespondWithAPIError(w, ErrAuthInvalidCredentials.WithMessage("missing or invalid user ID"))
		return
	}

	// Get workspace to get its ID
	workspaceService := NewChatWorkspaceService(h.service.q)
	workspace, err := workspaceService.GetWorkspaceByUUID(ctx, workspaceUuid)
	if err != nil {
		RespondWithAPIError(w, ErrResourceNotFound("Workspace").WithMessage("workspace not found"))
		return
	}

	// Get workspace active session
	session, err := h.service.GetActiveSession(ctx, userID, &workspace.ID)
	if err != nil {
		RespondWithAPIError(w, ErrResourceNotFound("Active Session").WithMessage("no active session for workspace"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"chatSessionUuid": session.ChatSessionUuid,
		"workspaceUuid":   workspaceUuid,
		"updatedAt":       session.UpdatedAt,
	})
}

// SetWorkspaceActiveSessionHandler sets the active session for a specific workspace
func (h *UserActiveChatSessionHandler) SetWorkspaceActiveSessionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	workspaceUuid := mux.Vars(r)["workspaceUuid"]

	// Get and validate user ID
	userID, err := getUserID(ctx)
	if err != nil {
		RespondWithAPIError(w, ErrAuthInvalidCredentials.WithMessage("missing or invalid user ID"))
		return
	}

	// Parse request body
	var requestBody struct {
		ChatSessionUuid string `json:"chatSessionUuid"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		RespondWithAPIError(w, ErrValidationInvalidInput("failed to parse request body"))
		return
	}

	// Validate session UUID format
	if !uuidRegex.MatchString(requestBody.ChatSessionUuid) {
		RespondWithAPIError(w, ErrChatSessionInvalid.WithMessage("invalid session UUID format"))
		return
	}

	// Get workspace to get its ID
	workspaceService := NewChatWorkspaceService(h.service.q)
	workspace, err := workspaceService.GetWorkspaceByUUID(ctx, workspaceUuid)
	if err != nil {
		RespondWithAPIError(w, ErrResourceNotFound("Workspace").WithMessage("workspace not found"))
		return
	}

	// Set workspace active session
	session, err := h.service.UpsertActiveSession(ctx, userID, &workspace.ID, requestBody.ChatSessionUuid)
	if err != nil {
		RespondWithAPIError(w, WrapError(err, "failed to set workspace active session"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"chatSessionUuid": session.ChatSessionUuid,
		"workspaceUuid":   workspaceUuid,
		"updatedAt":       session.UpdatedAt,
	})
}

// GetAllWorkspaceActiveSessionsHandler gets all workspace active sessions for a user
func (h *UserActiveChatSessionHandler) GetAllWorkspaceActiveSessionsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get and validate user ID
	userID, err := getUserID(ctx)
	if err != nil {
		RespondWithAPIError(w, ErrAuthInvalidCredentials.WithMessage("missing or invalid user ID"))
		return
	}

	// Get all workspace active sessions
	sessions, err := h.service.GetAllActiveSessions(ctx, userID)
	if err != nil {
		RespondWithAPIError(w, WrapError(err, "failed to get workspace active sessions"))
		return
	}

	// Convert to response format with workspace UUIDs
	workspaceService := NewChatWorkspaceService(h.service.q)
	workspaces, err := workspaceService.GetWorkspacesByUserID(ctx, userID)
	if err != nil {
		RespondWithAPIError(w, WrapError(err, "failed to get workspaces"))
		return
	}

	// Create a map for workspace ID to UUID lookup
	workspaceMap := make(map[int32]string)
	for _, workspace := range workspaces {
		workspaceMap[workspace.ID] = workspace.Uuid
	}

	// Build response
	var response []map[string]interface{}
	for _, session := range sessions {
		if session.WorkspaceID.Valid {
			if workspaceUuid, exists := workspaceMap[session.WorkspaceID.Int32]; exists {
				response = append(response, map[string]interface{}{
					"workspaceUuid":   workspaceUuid,
					"chatSessionUuid": session.ChatSessionUuid,
					"updatedAt":       session.UpdatedAt,
				})
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
