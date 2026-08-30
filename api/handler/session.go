package handler

import (
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
	router.HandleFunc("/chat_sessions/user", endpoint(h.getSimpleChatSessionsByUserID)).Methods(http.MethodGet)
	router.HandleFunc("/uuid/chat_sessions/max_length/{uuid}", endpoint(h.updateSessionMaxLength)).Methods(http.MethodPut)
	router.HandleFunc("/uuid/chat_sessions/topic/{uuid}", endpoint(h.updateChatSessionTopicByUUID)).Methods(http.MethodPut)
	router.HandleFunc("/uuid/chat_sessions/{uuid}", endpoint(h.getChatSessionByUUID)).Methods(http.MethodGet)
	router.HandleFunc("/uuid/chat_sessions/{uuid}", endpoint(h.createOrUpdateChatSessionByUUID)).Methods(http.MethodPut)
	router.HandleFunc("/uuid/chat_sessions/{uuid}", endpoint(h.deleteChatSessionByUUID)).Methods(http.MethodDelete)
	router.HandleFunc("/uuid/chat_sessions", endpoint(h.createChatSessionByUUID)).Methods(http.MethodPost)
	router.HandleFunc("/uuid/chat_session_from_snapshot/{uuid}", endpoint(h.createChatSessionFromSnapshot)).Methods(http.MethodPost)
}

func (h *ChatSessionHandler) getChatSessionByUUID(w http.ResponseWriter, r *http.Request) error {
	uuid := mux.Vars(r)["uuid"]
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}
	session, err := h.service.GetOwnedChatSession(r.Context(), svc.GetOwnedChatSessionQuery{UUID: uuid, UserID: userID})
	if err != nil {
		return err
	}
	return respondJSON(w, http.StatusOK, dto.ChatSessionResponse{
		Uuid:            session.UUID,
		Topic:           session.Topic,
		MaxLength:       session.MaxLength,
		CreatedAt:       session.CreatedAt,
		UpdatedAt:       session.UpdatedAt,
		ArtifactEnabled: session.ArtifactEnabled,
	})
}

func (h *ChatSessionHandler) createChatSessionByUUID(w http.ResponseWriter, r *http.Request) error {
	var req createChatSessionRequest
	if err := DecodeJSON(r, &req); err != nil {
		return dto.ErrValidationInvalidInput("Invalid request format").WithDebugInfo(err.Error())
	}

	ctx := r.Context()
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}

	session, err := h.service.SaveSession(ctx, svc.SaveSessionCommand{
		SessionUUID: req.UUID, UserID: userID, Topic: req.Topic,
		MaxLength: dto.DefaultMaxLength, Temperature: dto.DefaultTemperature,
		Model: req.Model, MaxTokens: dto.DefaultMaxTokens, TopP: dto.DefaultTopP, N: dto.DefaultN,
		Debug: false, SummarizeMode: false, ExploreMode: false, ArtifactEnabled: false,
		EnsureSystemPrompt: true, DefaultSystemPrompt: req.DefaultSystemPrompt, ActivateGlobally: true,
	})
	if err != nil {
		return err
	}
	return respondJSON(w, http.StatusCreated, chatSessionResponse(session))
}

func (h *ChatSessionHandler) createOrUpdateChatSessionByUUID(w http.ResponseWriter, r *http.Request) error {
	pathUUID := mux.Vars(r)["uuid"]
	var sessionReq dto.UpdateChatSessionRequest
	if err := DecodeJSON(r, &sessionReq); err != nil {
		return dto.ErrValidationInvalidInput("Invalid request format").WithDebugInfo(err.Error())
	}
	if sessionReq.Uuid != pathUUID {
		return dto.ErrValidationInvalidInput("request uuid must match path uuid")
	}
	if sessionReq.MaxLength == 0 {
		sessionReq.MaxLength = dto.DefaultMaxLength
	}

	ctx := r.Context()
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
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
		return err
	}
	return respondJSON(w, http.StatusOK, chatSessionResponse(session))
}

func (h *ChatSessionHandler) deleteChatSessionByUUID(w http.ResponseWriter, r *http.Request) error {
	uuid := mux.Vars(r)["uuid"]
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}

	if err := h.service.DeleteChatSessionByUUID(r.Context(), svc.DeleteChatSessionCommand{UUID: uuid, UserID: userID}); err != nil {
		return err
	}
	return respondStatus(w, http.StatusOK)
}

func (h *ChatSessionHandler) getSimpleChatSessionsByUserID(w http.ResponseWriter, r *http.Request) error {
	id, err := authenticatedUserID(r)
	if err != nil {
		return err
	}
	sessions, err := h.service.GetSimpleChatSessionsByUserID(r.Context(), id)
	if err != nil {
		return dto.ErrResourceNotFound("Chat sessions").WithDebugInfo(err.Error())
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
	return respondJSON(w, http.StatusOK, response)
}

func (h *ChatSessionHandler) updateChatSessionTopicByUUID(w http.ResponseWriter, r *http.Request) error {
	uuid := mux.Vars(r)["uuid"]
	var req updateSessionTopicRequest
	if err := DecodeJSON(r, &req); err != nil {
		return dto.ErrValidationInvalidInput("Invalid request format").WithDebugInfo(err.Error())
	}
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}
	session, err := h.service.UpdateChatSessionTopicByUUID(r.Context(), uuid, userID, req.Topic)
	if err != nil {
		return err
	}
	return respondJSON(w, http.StatusOK, chatSessionResponse(session))
}

func (h *ChatSessionHandler) updateSessionMaxLength(w http.ResponseWriter, r *http.Request) error {
	uuid := mux.Vars(r)["uuid"]
	var req updateSessionMaxLengthRequest
	if err := DecodeJSON(r, &req); err != nil {
		return dto.ErrValidationInvalidInput("Invalid request format").WithDebugInfo(err.Error())
	}
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}
	updatedSession, err := h.service.UpdateSessionMaxLength(r.Context(), svc.UpdateSessionMaxLengthCommand{UUID: uuid, UserID: userID, MaxLength: req.MaxLength})
	if err != nil {
		return err
	}
	return respondJSON(w, http.StatusOK, chatSessionResponse(updatedSession))
}

func (h *ChatSessionHandler) createChatSessionFromSnapshot(w http.ResponseWriter, r *http.Request) error {
	snapshotUUID := mux.Vars(r)["uuid"]
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}

	result, err := h.service.CreateSessionFromSnapshot(r.Context(), svc.CreateSessionFromSnapshotCommand{
		SnapshotUUID: snapshotUUID,
		UserID:       userID,
	})
	if err != nil {
		return err
	}
	return respondJSON(w, http.StatusCreated, sessionCreatedHTTPResponse{SessionUUID: result.SessionUUID})
}
