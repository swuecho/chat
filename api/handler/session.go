package handler

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/svc"
)

type ChatSessionHandler struct {
	service    *svc.ChatSessionService
	wsService  *svc.ChatWorkspaceService
	activeSvc  *svc.UserActiveChatSessionService
	promptSvc  *svc.ChatPromptService
	messageSvc *svc.ChatMessageService
}

func NewChatSessionHandler(service *svc.ChatSessionService, wsService *svc.ChatWorkspaceService, activeSvc *svc.UserActiveChatSessionService, promptSvc *svc.ChatPromptService, messageSvc *svc.ChatMessageService) *ChatSessionHandler {
	return &ChatSessionHandler{
		service: service, wsService: wsService, activeSvc: activeSvc,
		promptSvc: promptSvc, messageSvc: messageSvc,
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

func (h *ChatSessionHandler) getChatSessionByUUID(w http.ResponseWriter, r *http.Request) {
	uuid := mux.Vars(r)["uuid"]
	session, err := h.service.GetChatSessionByUUID(r.Context(), uuid)
	if err != nil {
		dto.RespondWithAPIError(w, err)
		return
	}

	json.NewEncoder(w).Encode(&dto.ChatSessionResponse{
		Uuid:            session.Uuid,
		Topic:           session.Topic,
		MaxLength:       session.MaxLength,
		CreatedAt:       session.CreatedAt,
		UpdatedAt:       session.UpdatedAt,
		ArtifactEnabled: session.ArtifactEnabled,
	})
}

func (h *ChatSessionHandler) createChatSessionByUUID(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Uuid                string `json:"uuid"`
		Topic               string `json:"topic"`
		Model               string `json:"model"`
		DefaultSystemPrompt string `json:"defaultSystemPrompt"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("Invalid request format").WithDebugInfo(err.Error()))
		return
	}

	ctx := r.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}

	defaultWorkspace, err := h.wsService.EnsureDefaultWorkspaceExists(ctx, userID)
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(dto.MapDatabaseError(err), "Failed to ensure default workspace exists"))
		return
	}

	workspaceID := defaultWorkspace.ID
	session, err := h.service.CreateOrUpdateChatSessionByUUID(ctx, svc.CreateOrUpdateChatSessionInput{
		Uuid: req.Uuid, UserID: userID, Topic: req.Topic,
		MaxLength: dto.DefaultMaxLength, Temperature: dto.DefaultTemperature,
		Model: req.Model, MaxTokens: dto.DefaultMaxTokens, TopP: dto.DefaultTopP, N: dto.DefaultN,
		Debug: false, SummarizeMode: false, ExploreMode: false, ArtifactEnabled: false,
		WorkspaceID: &workspaceID,
	})
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(dto.MapDatabaseError(err), "Failed to create or update chat session"))
		return
	}

	if _, err := h.service.EnsureDefaultSystemPrompt(ctx, session.Uuid, userID, req.DefaultSystemPrompt); err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(dto.MapDatabaseError(err), "Failed to create default system prompt"))
		return
	}

	if _, err := h.activeSvc.UpsertActiveSession(ctx, session.UserID, nil, session.Uuid); err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(dto.MapDatabaseError(err), "Failed to update active user session"))
		return
	}

	json.NewEncoder(w).Encode(session)
}

func (h *ChatSessionHandler) createOrUpdateChatSessionByUUID(w http.ResponseWriter, r *http.Request) {
	var sessionReq dto.UpdateChatSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&sessionReq); err != nil {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("Invalid request format").WithDebugInfo(err.Error()))
		return
	}
	if sessionReq.MaxLength == 0 {
		sessionReq.MaxLength = dto.DefaultMaxLength
	}

	ctx := r.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}

	params := svc.CreateOrUpdateChatSessionInput{
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
			dto.RespondWithAPIError(w, dto.WrapError(dto.MapDatabaseError(err), "Invalid workspace UUID"))
			return
		}
		params.WorkspaceID = &workspace.ID
	} else {
		defaultWS, err := h.wsService.EnsureDefaultWorkspaceExists(ctx, userID)
		if err != nil {
			dto.RespondWithAPIError(w, dto.WrapError(dto.MapDatabaseError(err), "Failed to ensure default workspace exists"))
			return
		}
		params.WorkspaceID = &defaultWS.ID
	}

	session, err := h.service.CreateOrUpdateChatSessionByUUID(ctx, params)
	if err != nil {
		apiErr := dto.ErrInternalUnexpected
		apiErr.Detail = "Failed to create or update chat session"
		apiErr.DebugInfo = err.Error()
		dto.RespondWithAPIError(w, apiErr)
		return
	}
	json.NewEncoder(w).Encode(session)
}

func (h *ChatSessionHandler) deleteChatSessionByUUID(w http.ResponseWriter, r *http.Request) {
	uuid := mux.Vars(r)["uuid"]

	userID, err := getUserID(r.Context())
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}

	// Verify session ownership before deletion
	session, err := h.service.GetChatSessionByUUID(r.Context(), uuid)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrResourceNotFound("Chat session").WithDebugInfo(err.Error()))
		return
	}
	if session.UserID != userID {
		dto.RespondWithAPIError(w, dto.ErrAuthAccessDenied.WithMessage("You do not own this session"))
		return
	}

	if err := h.service.DeleteChatSessionByUUID(r.Context(), uuid); err != nil {
		dto.RespondWithAPIError(w, dto.ErrInternalUnexpected.WithDetail("Failed to delete chat session").WithDebugInfo(err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *ChatSessionHandler) getSimpleChatSessionsByUserID(w http.ResponseWriter, r *http.Request) {
	id, err := getUserID(r.Context())
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("Invalid user ID").WithDebugInfo(err.Error()))
		return
	}
	sessions, err := h.service.GetSimpleChatSessionsByUserID(r.Context(), id)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrResourceNotFound("Chat sessions").WithDebugInfo(err.Error()))
		return
	}
	response := make([]dto.SimpleChatSession, 0, len(sessions))
	for _, session := range sessions {
		response = append(response, dto.SimpleChatSession{
			Uuid: session.UUID, IsEdit: false, Title: session.Title,
			MaxLength: session.MaxLength, Temperature: session.Temperature,
			TopP: session.TopP, N: session.N, MaxTokens: session.MaxTokens,
			Debug: session.Debug, Model: session.Model,
			SummarizeMode: session.SummarizeMode, ArtifactEnabled: session.ArtifactEnabled,
			WorkspaceUuid: session.WorkspaceUUID,
		})
	}
	json.NewEncoder(w).Encode(response)
}

func (h *ChatSessionHandler) updateChatSessionTopicByUUID(w http.ResponseWriter, r *http.Request) {
	uuid := mux.Vars(r)["uuid"]
	var req struct {
		Topic string `json:"topic"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("Invalid request format").WithDebugInfo(err.Error()))
		return
	}
	userID, err := getUserID(r.Context())
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}
	session, err := h.service.UpdateChatSessionTopicByUUID(r.Context(), uuid, userID, req.Topic)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrInternalUnexpected.WithDetail("Failed to update chat session topic").WithDebugInfo(err.Error()))
		return
	}
	json.NewEncoder(w).Encode(session)
}

func (h *ChatSessionHandler) updateSessionMaxLength(w http.ResponseWriter, r *http.Request) {
	uuid := mux.Vars(r)["uuid"]
	var req struct {
		MaxLength int32 `json:"maxLength"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("Invalid request format").WithDebugInfo(err.Error()))
		return
	}
	// Verify session ownership
	userID, err := getUserID(r.Context())
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}
	session, err := h.service.GetChatSessionByUUID(r.Context(), uuid)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrResourceNotFound("Chat session").WithDebugInfo(err.Error()))
		return
	}
	if session.UserID != userID {
		dto.RespondWithAPIError(w, dto.ErrAuthAccessDenied.WithMessage("You do not own this session"))
		return
	}

	updatedSession, err := h.service.UpdateSessionMaxLength(r.Context(), uuid, req.MaxLength)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrInternalUnexpected.WithDetail("Failed to update session max length").WithDebugInfo(err.Error()))
		return
	}
	json.NewEncoder(w).Encode(updatedSession)
}

func (h *ChatSessionHandler) createChatSessionFromSnapshot(w http.ResponseWriter, r *http.Request) {
	snapshotUUID := mux.Vars(r)["uuid"]

	userID, err := getUserID(r.Context())
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}

	snapshot, err := h.service.ChatSnapshotByUUID(r.Context(), snapshotUUID)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrResourceNotFound("Chat snapshot").WithDebugInfo(err.Error()))
		return
	}

	var messages []dto.SimpleChatMessage
	if err := json.Unmarshal(snapshot.Conversation, &messages); err != nil || len(messages) == 0 {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("Snapshot has no messages"))
		return
	}
	promptMsg := messages[0]

	chatPrompt, err := h.service.GetChatPromptByUUID(r.Context(), promptMsg.Uuid)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrResourceNotFound("Chat prompt").WithDebugInfo(err.Error()))
		return
	}

	originSession, err := h.service.GetChatSessionByUUIDWithInActive(r.Context(), chatPrompt.ChatSessionUuid)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrResourceNotFound("Original chat session").WithDebugInfo(err.Error()))
		return
	}

	sessionUUID := uuid.New().String()
	var workspaceID *int32
	if originSession.WorkspaceID.Valid {
		id := originSession.WorkspaceID.Int32
		workspaceID = &id
	}
	session, err := h.service.CreateOrUpdateChatSessionByUUID(r.Context(), svc.CreateOrUpdateChatSessionInput{
		Uuid: sessionUUID, UserID: userID, Topic: snapshot.Title,
		MaxLength: originSession.MaxLength, Temperature: originSession.Temperature,
		Model: originSession.Model, MaxTokens: originSession.MaxTokens,
		TopP: originSession.TopP, Debug: originSession.Debug,
		SummarizeMode: originSession.SummarizeMode, ExploreMode: originSession.ExploreMode,
		WorkspaceID: workspaceID, N: 1,
	})
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrInternalUnexpected.WithDetail("Failed to create chat session from snapshot").WithDebugInfo(err.Error()))
		return
	}

	if _, err := h.promptSvc.CreateChatPrompt(r.Context(), svc.CreateChatPromptInput{
		Uuid: NewUUID(), ChatSessionUuid: sessionUUID, Role: "system",
		Content: promptMsg.Text, UserID: userID, CreatedBy: userID, UpdatedBy: userID,
	}); err != nil {
		dto.RespondWithAPIError(w, dto.ErrInternalUnexpected.WithDetail("Failed to create prompt").WithDebugInfo(err.Error()))
		return
	}

	for _, msg := range messages[1:] {
		if _, err := h.messageSvc.CreateChatMessage(r.Context(), svc.CreateChatMessageInput{
			ChatSessionUuid: sessionUUID, Uuid: NewUUID(),
			Role: msg.GetRole(), Content: msg.Text, UserID: userID,
			Raw: json.RawMessage([]byte("{}")),
		}); err != nil {
			dto.RespondWithAPIError(w, dto.ErrInternalUnexpected.WithDetail("Failed to create messages").WithDebugInfo(err.Error()))
			return
		}
	}

	if _, err := h.activeSvc.UpsertActiveSession(r.Context(), userID, nil, session.Uuid); err != nil {
		dto.RespondWithAPIError(w, dto.ErrInternalUnexpected.WithDetail("Failed to update active session").WithDebugInfo(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"SessionUuid": session.Uuid})
}
