package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/svc"
)

type ChatSessionHandler struct {
	service *svc.ChatSessionService
}

func NewChatSessionHandler(service *svc.ChatSessionService) *ChatSessionHandler {
	return &ChatSessionHandler{service: service}
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
	userID, err := getUserID(r.Context())
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}
	session, err := h.service.GetOwnedChatSession(r.Context(), svc.GetOwnedChatSessionQuery{UUID: uuid, UserID: userID})
	if err != nil {
		dto.RespondWithAPIError(w, dto.ToAPIError(err))
		return
	}

	json.NewEncoder(w).Encode(&dto.ChatSessionResponse{
		Uuid:            session.UUID,
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

	session, err := h.service.SaveSession(ctx, svc.SaveSessionCommand{
		SessionUUID: req.Uuid, UserID: userID, Topic: req.Topic,
		MaxLength: dto.DefaultMaxLength, Temperature: dto.DefaultTemperature,
		Model: req.Model, MaxTokens: dto.DefaultMaxTokens, TopP: dto.DefaultTopP, N: dto.DefaultN,
		Debug: false, SummarizeMode: false, ExploreMode: false, ArtifactEnabled: false,
		EnsureSystemPrompt: true, DefaultSystemPrompt: req.DefaultSystemPrompt, ActivateGlobally: true,
	})
	if err != nil {
		dto.RespondWithAPIError(w, dto.ToAPIError(err))
		return
	}

	json.NewEncoder(w).Encode(chatSessionResponse(session))
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

	command := svc.SaveSessionCommand{
		SessionUUID: sessionReq.Uuid, UserID: userID, Topic: sessionReq.Topic,
		MaxLength: sessionReq.MaxLength, Temperature: sessionReq.Temperature,
		Model: sessionReq.Model, TopP: sessionReq.TopP, N: sessionReq.N,
		MaxTokens: sessionReq.MaxTokens, Debug: sessionReq.Debug,
		SummarizeMode: sessionReq.SummarizeMode, ArtifactEnabled: sessionReq.ArtifactEnabled,
		ExploreMode: sessionReq.ExploreMode, WorkspaceUUID: sessionReq.WorkspaceUUID,
	}

	session, err := h.service.SaveSession(ctx, command)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ToAPIError(err))
		return
	}
	json.NewEncoder(w).Encode(chatSessionResponse(session))
}

func (h *ChatSessionHandler) deleteChatSessionByUUID(w http.ResponseWriter, r *http.Request) {
	uuid := mux.Vars(r)["uuid"]

	userID, err := getUserID(r.Context())
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}

	if err := h.service.DeleteChatSessionByUUID(r.Context(), svc.DeleteChatSessionCommand{UUID: uuid, UserID: userID}); err != nil {
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
	userID, err := getUserID(r.Context())
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}
	updatedSession, err := h.service.UpdateSessionMaxLength(r.Context(), svc.UpdateSessionMaxLengthCommand{UUID: uuid, UserID: userID, MaxLength: req.MaxLength})
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

	result, err := h.service.CreateSessionFromSnapshot(r.Context(), svc.CreateSessionFromSnapshotCommand{
		SnapshotUUID: snapshotUUID,
		UserID:       userID,
	})
	if err != nil {
		dto.RespondWithAPIError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"SessionUuid": result.SessionUUID})
}
