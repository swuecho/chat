package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

type ChatSessionHandler struct {
	service *ChatSessionService
}

func NewChatSessionHandler(sqlc_q *sqlc_queries.Queries) *ChatSessionHandler {
	// create a new ChatSessionService instance
	chatSessionService := NewChatSessionService(sqlc_q)
	return &ChatSessionHandler{
		service: chatSessionService,
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

// getChatSessionByUUID returns a chat session by its UUID
func (h *ChatSessionHandler) getChatSessionByUUID(w http.ResponseWriter, r *http.Request) {
	uuid := mux.Vars(r)["uuid"]
	session, err := h.service.GetChatSessionByUUID(r.Context(), uuid)
	if err != nil {
		if err == sql.ErrNoRows {
			apiErr := ErrResourceNotFound("Chat session")
			apiErr.Message = "Session not found with UUID: " + uuid
			RespondWithAPIError(w, apiErr)
			return
		} else {
			apiErr := WrapError(MapDatabaseError(err), "Failed to get chat session")
			RespondWithAPIError(w, apiErr)
			return
		}
	}

	session_resp := &ChatSessionResponse{
		Uuid:      session.Uuid,
		Topic:     session.Topic,
		MaxLength: session.MaxLength,
		CreatedAt: session.CreatedAt,
		UpdatedAt: session.UpdatedAt,
	}
	json.NewEncoder(w).Encode(session_resp)
}

// createChatSessionByUUID creates a chat session by its UUID (idempotent)
func (h *ChatSessionHandler) createChatSessionByUUID(w http.ResponseWriter, r *http.Request) {
	var sessionParams sqlc_queries.CreateChatSessionParams
	err := json.NewDecoder(r.Body).Decode(&sessionParams)
	if err != nil {
		apiErr := ErrValidationInvalidInput("Invalid request format")
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}
	ctx := r.Context()
	userIDInt, err := getUserID(ctx)
	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}

	// Get or create default workspace for the user
	workspaceService := NewChatWorkspaceService(h.service.q)
	defaultWorkspace, err := workspaceService.EnsureDefaultWorkspaceExists(ctx, userIDInt)
	if err != nil {
		apiErr := WrapError(MapDatabaseError(err), "Failed to ensure default workspace exists")
		RespondWithAPIError(w, apiErr)
		return
	}

	// Use CreateOrUpdateChatSessionByUUID for idempotent session creation
	createOrUpdateParams := sqlc_queries.CreateOrUpdateChatSessionByUUIDParams{
		Uuid:          sessionParams.Uuid,
		UserID:        userIDInt,
		Topic:         sessionParams.Topic,
		MaxLength:     10,
		Temperature:   0.7, // Default values
		Model:         sessionParams.Model,
		MaxTokens:     4096, // Default values
		TopP:          1.0,  // Default values
		N:             1,    // Default values
		Debug:         false,
		SummarizeMode: false,
		WorkspaceID:   sql.NullInt32{Int32: defaultWorkspace.ID, Valid: true},
	}

	session, err := h.service.CreateOrUpdateChatSessionByUUID(r.Context(), createOrUpdateParams)
	if err != nil {
		apiErr := WrapError(MapDatabaseError(err), "Failed to create or update chat session")
		RespondWithAPIError(w, apiErr)
		return
	}

	// set active chat session when creating a new chat session
	_, err = h.service.q.CreateOrUpdateUserActiveChatSession(r.Context(),
		sqlc_queries.CreateOrUpdateUserActiveChatSessionParams{
			UserID:          session.UserID,
			ChatSessionUuid: session.Uuid,
		})
	if err != nil {
		apiErr := WrapError(MapDatabaseError(err), "Failed to update or create active user session record")
		RespondWithAPIError(w, apiErr)
		return
	}
	json.NewEncoder(w).Encode(session)
}

type UpdateChatSessionRequest struct {
	Uuid          string  `json:"uuid"`
	Topic         string  `json:"topic"`
	MaxLength     int32   `json:"maxLength"`
	Temperature   float64 `json:"temperature"`
	Model         string  `json:"model"`
	TopP          float64 `json:"topP"`
	N             int32   `json:"n"`
	MaxTokens     int32   `json:"maxTokens"`
	Debug         bool    `json:"debug"`
	SummarizeMode bool    `json:"summarizeMode"`
	WorkspaceUUID string  `json:"workspaceUuid,omitempty"`
}

// UpdateChatSessionByUUID updates a chat session by its UUID
func (h *ChatSessionHandler) createOrUpdateChatSessionByUUID(w http.ResponseWriter, r *http.Request) {
	var sessionReq UpdateChatSessionRequest
	err := json.NewDecoder(r.Body).Decode(&sessionReq)
	if err != nil {
		apiErr := ErrValidationInvalidInput("Invalid request format")
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}
	if sessionReq.MaxLength == 0 {
		sessionReq.MaxLength = 10
	}

	ctx := r.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}
	var sessionParams sqlc_queries.CreateOrUpdateChatSessionByUUIDParams

	sessionParams.MaxLength = sessionReq.MaxLength
	sessionParams.Topic = sessionReq.Topic
	sessionParams.Uuid = sessionReq.Uuid
	sessionParams.UserID = userID
	sessionParams.Temperature = sessionReq.Temperature
	sessionParams.Model = sessionReq.Model
	sessionParams.TopP = sessionReq.TopP
	sessionParams.N = sessionReq.N
	sessionParams.MaxTokens = sessionReq.MaxTokens
	sessionParams.Debug = sessionReq.Debug
	sessionParams.SummarizeMode = sessionReq.SummarizeMode

	// Handle workspace
	if sessionReq.WorkspaceUUID != "" {
		workspaceService := NewChatWorkspaceService(h.service.q)
		workspace, err := workspaceService.GetWorkspaceByUUID(ctx, sessionReq.WorkspaceUUID)
		if err != nil {
			apiErr := WrapError(MapDatabaseError(err), "Invalid workspace UUID")
			RespondWithAPIError(w, apiErr)
			return
		}
		sessionParams.WorkspaceID = sql.NullInt32{Int32: workspace.ID, Valid: true}
	} else {
		// Ensure default workspace exists
		workspaceService := NewChatWorkspaceService(h.service.q)
		defaultWorkspace, err := workspaceService.EnsureDefaultWorkspaceExists(ctx, userID)
		if err != nil {
			apiErr := WrapError(MapDatabaseError(err), "Failed to ensure default workspace exists")
			RespondWithAPIError(w, apiErr)
			return
		}
		sessionParams.WorkspaceID = sql.NullInt32{Int32: defaultWorkspace.ID, Valid: true}
	}
	session, err := h.service.CreateOrUpdateChatSessionByUUID(r.Context(), sessionParams)
	if err != nil {
		apiErr := ErrInternalUnexpected
		apiErr.Detail = "Failed to create or update chat session"
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}
	json.NewEncoder(w).Encode(session)
}

// deleteChatSessionByUUID deletes a chat session by its UUID
func (h *ChatSessionHandler) deleteChatSessionByUUID(w http.ResponseWriter, r *http.Request) {
	uuid := mux.Vars(r)["uuid"]
	err := h.service.DeleteChatSessionByUUID(r.Context(), uuid)
	if err != nil {
		apiErr := ErrInternalUnexpected
		apiErr.Detail = "Failed to delete chat session"
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// getSimpleChatSessionsByUserID returns a list of simple chat sessions by user ID
func (h *ChatSessionHandler) getSimpleChatSessionsByUserID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := ctx.Value(userContextKey).(string)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		apiErr := ErrValidationInvalidInput("Invalid user ID")
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}

	sessions, err := h.service.GetSimpleChatSessionsByUserID(ctx, int32(id))
	if err != nil {
		apiErr := ErrResourceNotFound("Chat sessions")
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}
	json.NewEncoder(w).Encode(sessions)
}

// updateChatSessionTopicByUUID updates a chat session topic by its UUID
func (h *ChatSessionHandler) updateChatSessionTopicByUUID(w http.ResponseWriter, r *http.Request) {
	uuid := mux.Vars(r)["uuid"]
	var sessionParams sqlc_queries.UpdateChatSessionTopicByUUIDParams
	err := json.NewDecoder(r.Body).Decode(&sessionParams)
	if err != nil {
		apiErr := ErrValidationInvalidInput("Invalid request format")
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}
	sessionParams.Uuid = uuid

	ctx := r.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}

	sessionParams.UserID = userID

	session, err := h.service.UpdateChatSessionTopicByUUID(r.Context(), sessionParams)
	if err != nil {
		apiErr := ErrInternalUnexpected
		apiErr.Detail = "Failed to update chat session topic"
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}
	json.NewEncoder(w).Encode(session)
}

// updateSessionMaxLength
func (h *ChatSessionHandler) updateSessionMaxLength(w http.ResponseWriter, r *http.Request) {
	uuid := mux.Vars(r)["uuid"]
	var sessionParams sqlc_queries.UpdateSessionMaxLengthParams
	err := json.NewDecoder(r.Body).Decode(&sessionParams)
	if err != nil {
		apiErr := ErrValidationInvalidInput("Invalid request format")
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}
	sessionParams.Uuid = uuid

	session, err := h.service.UpdateSessionMaxLength(r.Context(), sessionParams)
	if err != nil {
		apiErr := ErrInternalUnexpected
		apiErr.Detail = "Failed to update session max length"
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}
	json.NewEncoder(w).Encode(session)
}

// CreateChatSessionFromSnapshot ($uuid)
// create a new session with title of snapshot,
// create a prompt with the first message of snapshot
// create messages based on the rest of messages.
// return the new session uuid

func (h *ChatSessionHandler) createChatSessionFromSnapshot(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	snapshot_uuid := vars["uuid"]

	userID, err := getUserID(r.Context())
	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}

	snapshot, err := h.service.q.ChatSnapshotByUUID(r.Context(), snapshot_uuid)
	if err != nil {
		apiErr := ErrResourceNotFound("Chat snapshot")
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}

	sessionTitle := snapshot.Title
	conversions := snapshot.Conversation
	var conversionsSimpleMessages []SimpleChatMessage
	json.Unmarshal(conversions, &conversionsSimpleMessages)
	promptMsg := conversionsSimpleMessages[0]
	chatPrompt, err := h.service.q.GetChatPromptByUUID(r.Context(), promptMsg.Uuid)
	if err != nil {
		apiErr := ErrResourceNotFound("Chat prompt")
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}
	originSession, err := h.service.q.GetChatSessionByUUIDWithInActive(r.Context(), chatPrompt.ChatSessionUuid)
	if err != nil {
		apiErr := ErrResourceNotFound("Original chat session")
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}

	sessionUUID := uuid.New().String()

	session, err := h.service.q.CreateOrUpdateChatSessionByUUID(r.Context(), sqlc_queries.CreateOrUpdateChatSessionByUUIDParams{
		Uuid:        sessionUUID,
		UserID:      userID,
		Topic:       sessionTitle,
		MaxLength:   originSession.MaxLength,
		Temperature: originSession.Temperature,
		Model:       originSession.Model,
		MaxTokens:   originSession.MaxTokens,
		TopP:        originSession.TopP,
		Debug:       originSession.Debug,
		N:           1,
	})
	if err != nil {
		apiErr := ErrInternalUnexpected
		apiErr.Detail = "Failed to create chat session from snapshot"
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}

	_, err = h.service.q.CreateChatPrompt(r.Context(), sqlc_queries.CreateChatPromptParams{
		Uuid:            NewUUID(),
		ChatSessionUuid: sessionUUID,
		Role:            "system",
		Content:         promptMsg.Text,
		UserID:          userID,
		CreatedBy:       userID,
		UpdatedBy:       userID,
	})

	if err != nil {
		apiErr := ErrInternalUnexpected
		apiErr.Detail = "Failed to create prompt for chat session from snapshot"
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}

	for _, message := range conversionsSimpleMessages[1:] {
		// if inversion is true, the role is user, otherwise assistant
		// Determine the role based on the inversion flag

		messageParam := sqlc_queries.CreateChatMessageParams{
			ChatSessionUuid: sessionUUID,
			Uuid:            NewUUID(),
			Role:            message.GetRole(),
			Content:         message.Text,
			UserID:          userID,
			Raw:             json.RawMessage([]byte("{}")),
		}
		_, err = h.service.q.CreateChatMessage(r.Context(), messageParam)
		if err != nil {
			apiErr := ErrInternalUnexpected
			apiErr.Detail = "Failed to create messages for chat session from snapshot"
			apiErr.DebugInfo = err.Error()
			RespondWithAPIError(w, apiErr)
			return
		}

	}

	// set active session using simplified service
	activeSessionService := NewUserActiveChatSessionService(h.service.q)
	_, err = activeSessionService.UpsertActiveSession(r.Context(), userID, nil, session.Uuid)
	if err != nil {
		apiErr := ErrInternalUnexpected
		apiErr.Detail = "Failed to update active session"
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"SessionUuid": session.Uuid})
}
