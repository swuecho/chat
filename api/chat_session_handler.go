package main

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

type ChatSessionHandler struct {
	service     *ChatSessionService
	wsService   *ChatWorkspaceService
	activeSvc   *UserActiveChatSessionService
}

func NewChatSessionHandler(sqlc_q *sqlc_queries.Queries) *ChatSessionHandler {
	return &ChatSessionHandler{
		service:   NewChatSessionService(sqlc_q),
		wsService: NewChatWorkspaceService(sqlc_q),
		activeSvc: NewUserActiveChatSessionService(sqlc_q),
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
		}
		RespondWithAPIError(w, WrapError(MapDatabaseError(err), "Failed to get chat session"))
		return
	}

	json.NewEncoder(w).Encode(&ChatSessionResponse{
		Uuid:            session.Uuid,
		Topic:           session.Topic,
		MaxLength:       session.MaxLength,
		CreatedAt:       session.CreatedAt,
		UpdatedAt:       session.UpdatedAt,
		ArtifactEnabled: session.ArtifactEnabled,
	})
}

// createChatSessionByUUID creates a chat session by its UUID (idempotent)
func (h *ChatSessionHandler) createChatSessionByUUID(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Uuid                string `json:"uuid"`
		Topic               string `json:"topic"`
		Model               string `json:"model"`
		DefaultSystemPrompt string `json:"defaultSystemPrompt"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithAPIError(w, ErrValidationInvalidInput("Invalid request format").WithDebugInfo(err.Error()))
		return
	}

	ctx := r.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		RespondWithAPIError(w, ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}

	defaultWorkspace, err := h.wsService.EnsureDefaultWorkspaceExists(ctx, userID)
	if err != nil {
		RespondWithAPIError(w, WrapError(MapDatabaseError(err), "Failed to ensure default workspace exists"))
		return
	}

	session, err := h.service.CreateOrUpdateChatSessionByUUID(ctx, sqlc_queries.CreateOrUpdateChatSessionByUUIDParams{
		Uuid: req.Uuid, UserID: userID, Topic: req.Topic,
		MaxLength: DefaultMaxLength, Temperature: DefaultTemperature,
		Model: req.Model, MaxTokens: DefaultMaxTokens, TopP: DefaultTopP, N: DefaultN,
		Debug: false, SummarizeMode: false, ExploreMode: false, ArtifactEnabled: false,
		WorkspaceID: sql.NullInt32{Int32: defaultWorkspace.ID, Valid: true},
	})
	if err != nil {
		RespondWithAPIError(w, WrapError(MapDatabaseError(err), "Failed to create or update chat session"))
		return
	}

	if _, err := h.service.EnsureDefaultSystemPrompt(ctx, session.Uuid, userID, req.DefaultSystemPrompt); err != nil {
		RespondWithAPIError(w, WrapError(MapDatabaseError(err), "Failed to create default system prompt"))
		return
	}

	if _, err := h.service.UpsertUserActiveSession(ctx, sqlc_queries.UpsertUserActiveSessionParams{
		UserID: session.UserID, WorkspaceID: sql.NullInt32{},
		ChatSessionUuid: session.Uuid,
	}); err != nil {
		RespondWithAPIError(w, WrapError(MapDatabaseError(err), "Failed to update active user session"))
		return
	}

	json.NewEncoder(w).Encode(session)
}

// createOrUpdateChatSessionByUUID updates a chat session by its UUID
func (h *ChatSessionHandler) createOrUpdateChatSessionByUUID(w http.ResponseWriter, r *http.Request) {
	var sessionReq UpdateChatSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&sessionReq); err != nil {
		RespondWithAPIError(w, ErrValidationInvalidInput("Invalid request format").WithDebugInfo(err.Error()))
		return
	}
	if sessionReq.MaxLength == 0 {
		sessionReq.MaxLength = DefaultMaxLength
	}

	ctx := r.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		RespondWithAPIError(w, ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}

	params := sqlc_queries.CreateOrUpdateChatSessionByUUIDParams{
		Uuid: sessionReq.Uuid, UserID: userID, Topic: sessionReq.Topic,
		MaxLength: sessionReq.MaxLength, Temperature: sessionReq.Temperature,
		Model: sessionReq.Model, TopP: sessionReq.TopP, N: sessionReq.N,
		MaxTokens: sessionReq.MaxTokens, Debug: sessionReq.Debug,
		SummarizeMode: sessionReq.SummarizeMode, ArtifactEnabled: sessionReq.ArtifactEnabled,
		ExploreMode: sessionReq.ExploreMode,
	}

	if sessionReq.WorkspaceUUID != "" {
		workspace, err := h.wsService.GetWorkspaceByUUID(ctx, sessionReq.WorkspaceUUID)
		if err != nil {
			RespondWithAPIError(w, WrapError(MapDatabaseError(err), "Invalid workspace UUID"))
			return
		}
		params.WorkspaceID = sql.NullInt32{Int32: workspace.ID, Valid: true}
	} else {
		defaultWS, err := h.wsService.EnsureDefaultWorkspaceExists(ctx, userID)
		if err != nil {
			RespondWithAPIError(w, WrapError(MapDatabaseError(err), "Failed to ensure default workspace exists"))
			return
		}
		params.WorkspaceID = sql.NullInt32{Int32: defaultWS.ID, Valid: true}
	}

	session, err := h.service.CreateOrUpdateChatSessionByUUID(ctx, params)
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
	if err := h.service.DeleteChatSessionByUUID(r.Context(), uuid); err != nil {
		RespondWithAPIError(w, ErrInternalUnexpected.WithDetail("Failed to delete chat session").WithDebugInfo(err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
}

// getSimpleChatSessionsByUserID returns a list of simple chat sessions by user ID
func (h *ChatSessionHandler) getSimpleChatSessionsByUserID(w http.ResponseWriter, r *http.Request) {
	id, err := getUserID(r.Context())
	if err != nil {
		RespondWithAPIError(w, ErrValidationInvalidInput("Invalid user ID").WithDebugInfo(err.Error()))
		return
	}
	sessions, err := h.service.GetSimpleChatSessionsByUserID(r.Context(), id)
	if err != nil {
		RespondWithAPIError(w, ErrResourceNotFound("Chat sessions").WithDebugInfo(err.Error()))
		return
	}
	json.NewEncoder(w).Encode(sessions)
}

// updateChatSessionTopicByUUID updates a chat session topic by its UUID
func (h *ChatSessionHandler) updateChatSessionTopicByUUID(w http.ResponseWriter, r *http.Request) {
	uuid := mux.Vars(r)["uuid"]
	var params sqlc_queries.UpdateChatSessionTopicByUUIDParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		RespondWithAPIError(w, ErrValidationInvalidInput("Invalid request format").WithDebugInfo(err.Error()))
		return
	}
	params.Uuid = uuid

	userID, err := getUserID(r.Context())
	if err != nil {
		RespondWithAPIError(w, ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}
	params.UserID = userID

	session, err := h.service.UpdateChatSessionTopicByUUID(r.Context(), params)
	if err != nil {
		RespondWithAPIError(w, ErrInternalUnexpected.WithDetail("Failed to update chat session topic").WithDebugInfo(err.Error()))
		return
	}
	json.NewEncoder(w).Encode(session)
}

// updateSessionMaxLength updates the max length of a session
func (h *ChatSessionHandler) updateSessionMaxLength(w http.ResponseWriter, r *http.Request) {
	uuid := mux.Vars(r)["uuid"]
	var params sqlc_queries.UpdateSessionMaxLengthParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		RespondWithAPIError(w, ErrValidationInvalidInput("Invalid request format").WithDebugInfo(err.Error()))
		return
	}
	params.Uuid = uuid

	session, err := h.service.UpdateSessionMaxLength(r.Context(), params)
	if err != nil {
		RespondWithAPIError(w, ErrInternalUnexpected.WithDetail("Failed to update session max length").WithDebugInfo(err.Error()))
		return
	}
	json.NewEncoder(w).Encode(session)
}

// CreateChatSessionFromSnapshot creates a new session from a snapshot
func (h *ChatSessionHandler) createChatSessionFromSnapshot(w http.ResponseWriter, r *http.Request) {
	snapshotUUID := mux.Vars(r)["uuid"]

	userID, err := getUserID(r.Context())
	if err != nil {
		RespondWithAPIError(w, ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}

	snapshot, err := h.service.ChatSnapshotByUUID(r.Context(), snapshotUUID)
	if err != nil {
		RespondWithAPIError(w, ErrResourceNotFound("Chat snapshot").WithDebugInfo(err.Error()))
		return
	}

	var messages []SimpleChatMessage
	json.Unmarshal(snapshot.Conversation, &messages)
	promptMsg := messages[0]

	chatPrompt, err := h.service.GetChatPromptByUUID(r.Context(), promptMsg.Uuid)
	if err != nil {
		RespondWithAPIError(w, ErrResourceNotFound("Chat prompt").WithDebugInfo(err.Error()))
		return
	}

	originSession, err := h.service.GetChatSessionByUUIDWithInActive(r.Context(), chatPrompt.ChatSessionUuid)
	if err != nil {
		RespondWithAPIError(w, ErrResourceNotFound("Original chat session").WithDebugInfo(err.Error()))
		return
	}

	sessionUUID := uuid.New().String()
	session, err := h.service.CreateOrUpdateChatSessionByUUID(r.Context(), sqlc_queries.CreateOrUpdateChatSessionByUUIDParams{
		Uuid: sessionUUID, UserID: userID, Topic: snapshot.Title,
		MaxLength: originSession.MaxLength, Temperature: originSession.Temperature,
		Model: originSession.Model, MaxTokens: originSession.MaxTokens,
		TopP: originSession.TopP, Debug: originSession.Debug,
		SummarizeMode: originSession.SummarizeMode, ExploreMode: originSession.ExploreMode,
		WorkspaceID: originSession.WorkspaceID, N: 1,
	})
	if err != nil {
		RespondWithAPIError(w, ErrInternalUnexpected.WithDetail("Failed to create chat session from snapshot").WithDebugInfo(err.Error()))
		return
	}

	if _, err := h.service.CreateChatPrompt(r.Context(), sqlc_queries.CreateChatPromptParams{
		Uuid: NewUUID(), ChatSessionUuid: sessionUUID, Role: "system",
		Content: promptMsg.Text, UserID: userID, CreatedBy: userID, UpdatedBy: userID,
	}); err != nil {
		RespondWithAPIError(w, ErrInternalUnexpected.WithDetail("Failed to create prompt").WithDebugInfo(err.Error()))
		return
	}

	for _, msg := range messages[1:] {
		if _, err := h.service.CreateChatMessage(r.Context(), sqlc_queries.CreateChatMessageParams{
			ChatSessionUuid: sessionUUID, Uuid: NewUUID(),
			Role: msg.GetRole(), Content: msg.Text, UserID: userID,
			Raw: json.RawMessage([]byte("{}")),
		}); err != nil {
			RespondWithAPIError(w, ErrInternalUnexpected.WithDetail("Failed to create messages").WithDebugInfo(err.Error()))
			return
		}
	}

	if _, err := h.activeSvc.UpsertActiveSession(r.Context(), userID, nil, session.Uuid); err != nil {
		RespondWithAPIError(w, ErrInternalUnexpected.WithDetail("Failed to update active session").WithDebugInfo(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"SessionUuid": session.Uuid})
}
